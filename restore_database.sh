#!/usr/bin/env bash
set -e

echo "=========================================================="
echo "🔍 JO4 Dev - Automatic Database Recovery & Restore"
echo "=========================================================="

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$PROJECT_DIR"

# 1. Identify running web service container
TARGET_CONTAINER=$(docker ps --filter "name=jo4-web" --format "{{.Names}}" | head -n 1)

if [ -z "$TARGET_CONTAINER" ]; then
    echo "⚠️ No running jo4-web container found. Starting with docker compose up -d..."
    docker compose up -d web-green
    TARGET_CONTAINER="jo4-web-green"
fi

echo "🎯 Target live container: $TARGET_CONTAINER"

# 2. Look for candidate databases across the server
CANDIDATES=()

if [ -f "jo4dev_backup.db" ]; then
    CANDIDATES+=("$(pwd)/jo4dev_backup.db")
fi

if [ -f "/root/my_website/jo4dev_backup.db" ]; then
    CANDIDATES+=("/root/my_website/jo4dev_backup.db")
fi

if [ -f "/root/jo4dev_backup.db" ]; then
    CANDIDATES+=("/root/jo4dev_backup.db")
fi

# Try extracting from legacy jo4-site container if it exists
if docker ps -a --format '{{.Names}}' | grep -q '^jo4-site$'; then
    echo "📦 Detected legacy container 'jo4-site'! Extracting /app/instance/jo4dev.db..."
    docker cp jo4-site:/app/instance/jo4dev.db ./jo4_site_extracted.db 2>/dev/null || true
    if [ -f "./jo4_site_extracted.db" ]; then
        CANDIDATES+=("$(pwd)/jo4_site_extracted.db")
    fi
fi

# Scan Docker volumes for older sqlite databases
for vdir in /var/lib/docker/volumes/*db_data*/_data; do
    if [ -f "$vdir/jo4dev.db" ]; then
        CANDIDATES+=("$vdir/jo4dev.db")
    fi
done

# Scan any .db in current directory
for f in *.db; do
    if [ -f "$f" ]; then
        CANDIDATES+=("$(pwd)/$f")
    fi
done

# Deduplicate candidates
UNIQUE_CANDIDATES=($(echo "${CANDIDATES[@]}" | tr ' ' '\n' | sort -u | tr '\n' ' '))

echo "🔍 Found ${#UNIQUE_CANDIDATES[@]} potential database file(s) on this host:"

BEST_DB=""
MAX_ITEMS=-1

for db_file in "${UNIQUE_CANDIDATES[@]}"; do
    if [ -f "$db_file" ]; then
        COUNTS=$(python3 -c "
import sqlite3
try:
    c = sqlite3.connect('$db_file')
    b = c.execute(\"SELECT count(*) FROM blog_posts\").fetchone()[0]
    p = c.execute(\"SELECT count(*) FROM projects\").fetchone()[0]
    l = c.execute(\"SELECT count(*) FROM leads\").fetchone()[0]
    print(f'{b} blogs, {p} projects, {l} leads')
    print(b + p + l)
except Exception as e:
    print(f'unreadable ({e})')
    print(-1)
" 2>/dev/null || echo -e "sqlite unreadable\n-1")

        DESC=$(echo "$COUNTS" | head -n 1)
        TOTAL=$(echo "$COUNTS" | tail -n 1)

        echo "   📄 $db_file -> $DESC"

        if [[ "$TOTAL" =~ ^[0-9]+$ ]] && [ "$TOTAL" -gt "$MAX_ITEMS" ]; then
            MAX_ITEMS=$TOTAL
            BEST_DB="$db_file"
        fi
    fi
done

if [ -z "$BEST_DB" ] || [ "$MAX_ITEMS" -le 0 ]; then
    echo "⚠️ No populated database candidate was found with blog posts or projects."
    echo "   Checking current database in $TARGET_CONTAINER..."
    docker exec "$TARGET_CONTAINER" stat /app/data/jo4dev.db 2>/dev/null || true
    exit 1
fi

echo ""
echo "=========================================================="
echo "✅ Selected Best Database Source: $BEST_DB"
echo "   Contains: $MAX_ITEMS total items (blogs + projects + leads)"
echo "📥 Restoring directly into $TARGET_CONTAINER:/app/data/jo4dev.db..."
echo "=========================================================="

# Create local safety backup of current container db before overwrite
docker cp "$TARGET_CONTAINER":/app/data/jo4dev.db ./jo4dev_pre_restore.db.bak 2>/dev/null || true

# Copy best database into active container
docker cp "$BEST_DB" "$TARGET_CONTAINER":/app/data/jo4dev.db

# Ensure correct non-root permissions for Go Fiber application
docker exec -u 0 "$TARGET_CONTAINER" chown -R appuser:appgroup /app/data
docker exec -u 0 "$TARGET_CONTAINER" chmod -R 775 /app/data
docker exec -u 0 "$TARGET_CONTAINER" chmod 664 /app/data/jo4dev.db

# Keep copy as jo4dev_backup.db in project root so future deploy.sh runs keep it
cp "$BEST_DB" ./jo4dev_backup.db

echo "🔄 Restarting $TARGET_CONTAINER to reload database connections..."
docker restart "$TARGET_CONTAINER"

sleep 3
echo "=========================================================="
echo "🎉 RESTORATION COMPLETED SUCCESSFULLY!"
echo "   Active service container: $TARGET_CONTAINER"
echo "   Persistent backup saved:  ./jo4dev_backup.db"
echo "   Check: https://jo4.co.za/blog and https://jo4.co.za/projects"
echo "=========================================================="
