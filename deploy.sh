#!/usr/bin/env bash
set -eo pipefail

# ==============================================================================
# JO4 DEV - ZERO-DOWNTIME RED/BLUE (BLUE/GREEN) DEPLOYMENT SCRIPT
# ==============================================================================

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$PROJECT_DIR"

UPSTREAM_FILE="nginx/upstream.inc"
TARGET_SERVICE=""
OLD_SERVICE=""
ACTIVE=""
TARGET=""

# Comprehensive Error Handler
on_error() {
    local exit_code="$1"
    local line_num="$2"
    echo ""
    echo "=========================================================="
    echo "❌ DEPLOYMENT ERROR: Command failed with exit code $exit_code at line $line_num"
    echo "=========================================================="
    if [ -n "$TARGET_SERVICE" ]; then
        echo "🔍 Checking container logs for $TARGET_SERVICE..."
        docker compose logs --tail=50 "$TARGET_SERVICE" || true
        echo "🛑 Stopping failed target container ($TARGET_SERVICE)..."
        docker compose stop "$TARGET_SERVICE" 2>/dev/null || true
    fi
    if [ -n "$ACTIVE" ]; then
        echo "🛡️ Current active environment ($ACTIVE) remains untouched and running."
    fi
    exit "$exit_code"
}

trap 'on_error $? $LINENO' ERR

# 1. Sanity check: Ensure Docker is accessible
if ! command -v docker >/dev/null 2>&1; then
    echo "❌ Error: 'docker' CLI is not found or not in PATH."
    exit 1
fi

if ! docker info >/dev/null 2>&1; then
    echo "❌ Error: Docker daemon is not running or current user lacks docker permissions."
    exit 1
fi

# 2. Check Environment Configuration
if [ ! -f ".env" ]; then
    if [ -f ".env.example" ]; then
        echo "⚠️ .env file not found! Initializing from .env.example..."
        cp .env.example .env
        echo "⚠️ Created .env. Please review it with your production secrets."
    else
        echo "❌ Error: Neither .env nor .env.example exists."
        exit 1
    fi
fi

# Ensure external Cloudflare tunnel network exists
if ! docker network inspect tunnel-net >/dev/null 2>&1; then
    echo "🌐 Initializing Docker network 'tunnel-net' for Cloudflare Tunnel..."
    docker network create tunnel-net 2>/dev/null || true
fi

# 3. Ensure Nginx configuration directories and upstream file exist
mkdir -p nginx/conf.d
if [ ! -f "$UPSTREAM_FILE" ]; then
    echo "server web-blue:5000;" > "$UPSTREAM_FILE"
fi

# 4. Determine Active and Target Environments
CURRENT_UPSTREAM=$(grep -oE 'web-[a-z]+' "$UPSTREAM_FILE" 2>/dev/null || echo "web-blue")

if [ "$CURRENT_UPSTREAM" == "web-blue" ]; then
    ACTIVE="BLUE"
    TARGET="GREEN"
    TARGET_SERVICE="web-green"
    TARGET_PORT=5002
    OLD_SERVICE="web-blue"
else
    ACTIVE="GREEN"
    TARGET="BLUE"
    TARGET_SERVICE="web-blue"
    TARGET_PORT=5001
    OLD_SERVICE="web-green"
fi

echo "=========================================================="
echo "🚀 JO4 Dev Zero-Downtime Deployment"
echo "   Current Live Service       : $ACTIVE ($CURRENT_UPSTREAM)"
echo "   Deploying New Release To   : $TARGET ($TARGET_SERVICE on port $TARGET_PORT)"
echo "=========================================================="

# 5. Build Target Container
echo "🔨 Step 1/5: Building $TARGET_SERVICE..."
docker compose build "$TARGET_SERVICE"

# 6. Start Target Container & Supporting Services
echo "🚀 Step 2/5: Launching redis & $TARGET_SERVICE..."
docker compose up -d redis "$TARGET_SERVICE"

# Auto-restore existing database if container has 0 blog posts and backup is available
BLOG_COUNT=$(docker compose exec -T "$TARGET_SERVICE" sqlite3 /app/data/jo4dev.db "SELECT count(*) FROM blog_posts;" 2>/dev/null || echo "0")
if [ "$BLOG_COUNT" -eq 0 ]; then
    BACKUP_SRC=""
    if [ -f "jo4dev_backup.db" ]; then
        BACKUP_SRC="jo4dev_backup.db"
    elif [ -f "/root/my_website/jo4dev_backup.db" ]; then
        BACKUP_SRC="/root/my_website/jo4dev_backup.db"
    elif docker ps -a --format '{{.Names}}' | grep -q '^jo4-site$'; then
        echo "📦 Extracting database from legacy jo4-site container..."
        docker cp jo4-site:/app/instance/jo4dev.db ./jo4dev_backup.db 2>/dev/null || true
        [ -f "./jo4dev_backup.db" ] && BACKUP_SRC="./jo4dev_backup.db"
    fi

    if [ -n "$BACKUP_SRC" ]; then
        echo "📥 Restoring historical blogs, projects, and leads from $BACKUP_SRC into $TARGET_SERVICE..."
        docker compose cp "$BACKUP_SRC" "$TARGET_SERVICE":/app/data/jo4dev.db
        docker compose exec -u 0 -T "$TARGET_SERVICE" chown -R appuser:appgroup /app/data
        docker compose exec -u 0 -T "$TARGET_SERVICE" chmod 664 /app/data/jo4dev.db
        docker compose restart "$TARGET_SERVICE"
        echo "✅ Database restored successfully."
    fi
fi

# 7. Health Check Target Container
echo "🔍 Step 3/5: Running health checks on http://127.0.0.1:$TARGET_PORT/health..."
MAX_ATTEMPTS=25
ATTEMPT=0
HEALTHY=0

while [ $ATTEMPT -lt $MAX_ATTEMPTS ]; do
    ATTEMPT=$((ATTEMPT + 1))
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "http://127.0.0.1:$TARGET_PORT/health" 2>/dev/null || echo "000")
    if [ "$HTTP_CODE" -eq 200 ]; then
        HEALTHY=1
        break
    fi
    echo "   Attempt $ATTEMPT/$MAX_ATTEMPTS (HTTP $HTTP_CODE) - Waiting for service..."
    sleep 2
done

if [ $HEALTHY -eq 0 ]; then
    echo "❌ Health check failed after $MAX_ATTEMPTS attempts."
    echo "   Dumping last 50 log lines from $TARGET_SERVICE:"
    docker compose logs --tail=50 "$TARGET_SERVICE"
    echo "🛑 Halting deployment and stopping $TARGET_SERVICE..."
    docker compose stop "$TARGET_SERVICE" 2>/dev/null || true
    echo "🛡️ Live traffic remains on $ACTIVE ($OLD_SERVICE)."
    exit 1
fi

echo "✅ Health check PASSED! $TARGET_SERVICE is healthy and responding with HTTP 200."

# 8. Ensure Nginx is running
echo "🌐 Step 4/5: Ensuring Nginx reverse proxy is active..."
docker compose up -d nginx

# 9. Switch Traffic Atomically
echo "🔄 Step 5/5: Switching Nginx upstream traffic to $TARGET_SERVICE..."
echo "server $TARGET_SERVICE:5000;" > "$UPSTREAM_FILE"

# Test Nginx syntax before reload
if ! docker compose exec -T nginx nginx -t >/dev/null 2>&1; then
    echo "❌ Nginx configuration test failed! Reverting upstream..."
    echo "server $OLD_SERVICE:5000;" > "$UPSTREAM_FILE"
    docker compose stop "$TARGET_SERVICE" 2>/dev/null || true
    exit 1
fi

# Graceful reload: zero dropped requests
docker compose exec -T nginx nginx -s reload

echo "⏳ Waiting 5 seconds to drain in-flight connections from $OLD_SERVICE..."
sleep 5

# Stop previous container to free server memory
echo "🛑 Stopping previous container ($OLD_SERVICE)..."
docker compose stop "$OLD_SERVICE" 2>/dev/null || true

echo "=========================================================="
echo "🎉 DEPLOYMENT SUCCESSFUL!"
echo "   Live Environment is now: $TARGET ($TARGET_SERVICE)"
echo "   Zero downtime achieved."
echo "=========================================================="
