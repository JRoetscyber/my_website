#!/usr/bin/env bash
set -e

# ==============================================================================
# JO4 DEV - ZERO-DOWNTIME RED/BLUE (BLUE/GREEN) DEPLOYMENT SCRIPT
# ==============================================================================

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$PROJECT_DIR"

UPSTREAM_FILE="nginx/upstream.inc"

if [ ! -f "$UPSTREAM_FILE" ]; then
    mkdir -p nginx
    echo "server web-blue:5000;" > "$UPSTREAM_FILE"
fi

if [ ! -f ".env" ]; then
    echo "⚠️ .env file not found! Copying from .env.example..."
    cp .env.example .env
    echo "⚠️ Please review .env with your production credentials!"
fi

# Detect currently active environment
CURRENT_UPSTREAM=$(grep -oE 'web-[a-z]+' "$UPSTREAM_FILE" || echo "web-blue")

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
echo "   Current Active Environment : $ACTIVE ($CURRENT_UPSTREAM)"
echo "   Target Deployment Target   : $TARGET ($TARGET_SERVICE)"
echo "=========================================================="

# 1. Ensure Nginx and network are running
echo "📦 Ensuring Nginx reverse proxy is running..."
docker compose up -d nginx

# 2. Build and launch target container
echo "🔨 Building and starting $TARGET_SERVICE..."
docker compose build "$TARGET_SERVICE"
docker compose up -d "$TARGET_SERVICE"

# 3. Perform Health Check
echo "🔍 Waiting for $TARGET_SERVICE to become healthy on port $TARGET_PORT..."
MAX_ATTEMPTS=20
ATTEMPT=0
HEALTHY=0

while [ $ATTEMPT -lt $MAX_ATTEMPTS ]; do
    ATTEMPT=$((ATTEMPT + 1))
    if curl -s -f "http://127.0.0.1:$TARGET_PORT/health" > /dev/null 2>&1; then
        HEALTHY=1
        break
    fi
    echo "   Attempt $ATTEMPT/$MAX_ATTEMPTS: Waiting for server response..."
    sleep 2
done

if [ $HEALTHY -eq 0 ]; then
    echo "❌ HEALTH CHECK FAILED on $TARGET_SERVICE!"
    echo "   Aborting deployment. Active environment remains $ACTIVE."
    docker compose logs --tail=50 "$TARGET_SERVICE"
    docker compose stop "$TARGET_SERVICE"
    exit 1
fi

echo "✅ Health check PASSED for $TARGET_SERVICE!"

# 4. Atomically switch Nginx upstream
echo "🔄 Switching Nginx traffic to $TARGET_SERVICE..."
echo "server $TARGET_SERVICE:5000;" > "$UPSTREAM_FILE"

# 5. Reload Nginx without dropping connections
docker compose exec -T nginx nginx -s reload

echo "⏳ Traffic switched. Draining connections from $OLD_SERVICE (5s)..."
sleep 5

# 6. Stop idle environment to conserve resources
echo "🛑 Stopping old environment ($OLD_SERVICE)..."
docker compose stop "$OLD_SERVICE"

echo "=========================================================="
echo "🎉 DEPLOYMENT COMPLETE! ZERO DOWNTIME ACHIEVED."
echo "   Now serving live traffic on: $TARGET ($TARGET_SERVICE)"
echo "=========================================================="
