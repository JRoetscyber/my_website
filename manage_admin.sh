#!/usr/bin/env bash
set -e

# ==============================================================================
# JO4 DEV - Admin Account Management Utility
# ==============================================================================

echo "=========================================================="
echo "🔐 JO4 Dev Admin Management Utility"
echo "=========================================================="

# Find live container
CONTAINER=""
for c in jo4-web-blue jo4-web-green; do
    if docker ps --format '{{.Names}}' | grep -q "^${c}$"; then
        CONTAINER="$c"
        break
    fi
done

if [ -z "$CONTAINER" ]; then
    echo "❌ No running JO4 web container found (checked jo4-web-blue, jo4-web-green)."
    exit 1
fi

echo "🎯 Connected to live container: $CONTAINER"
echo ""

# Display current admin users in SQLite database
echo "📋 Existing Users in Database:"
docker exec "$CONTAINER" sqlite3 /app/data/jo4dev.db \
    "SELECT '   ID: ' || id || ' | Username: ' || username || ' | Hash type: ' || substr(password_hash, 1, 15) || '...' FROM login;" 2>/dev/null || echo "   (No users found or table empty)"
echo ""

# Check .env configuration
ADMIN_USER=""
ADMIN_PASS=""
if [ -f ".env" ]; then
    ADMIN_USER=$(grep -E '^ADMIN_USERNAME=' .env | cut -d '=' -f2- | tr -d '\r"' || true)
    ADMIN_PASS=$(grep -E '^ADMIN_PASSWORD=' .env | cut -d '=' -f2- | tr -d '\r"' || true)
fi

echo "⚙️ Credentials configured in .env:"
echo "   ADMIN_USERNAME : ${ADMIN_USER:-admin}"
echo "   ADMIN_PASSWORD : ${ADMIN_PASS:-(hidden or not set)}"
echo ""

if [ -n "$1" ] && [ -n "$2" ]; then
    TARGET_USER="$1"
    NEW_PASS="$2"
else
    TARGET_USER="${ADMIN_USER:-admin}"
    NEW_PASS="${ADMIN_PASS:-ChangeMeNow_JO4Secure2026!}"
fi

echo "🔧 Setting password for user '$TARGET_USER'..."

# Use python inside the host or container to generate standard bcrypt hash and update SQLite
python3 -c "
import sqlite3, subprocess, sys

user = '$TARGET_USER'
pwd = '$NEW_PASS'

# Generate bcrypt hash using python or container
try:
    import bcrypt
    hashed = bcrypt.hashpw(pwd.encode('utf-8'), bcrypt.gensalt(10)).decode('utf-8')
except ImportError:
    # Fallback to python standard library or container
    import hashlib, os
    # Create Werkzeug compatible PBKDF2 hash
    salt = os.urandom(8).hex()
    dk = hashlib.pbkdf2_hmac('sha256', pwd.encode('utf-8'), salt.encode('utf-8'), 260000)
    hashed = f'pbkdf2:sha256:260000\${salt}\${dk.hex()}'

cmd = [
    'docker', 'exec', '$CONTAINER', 'sqlite3', '/app/data/jo4dev.db',
    f'''
    CREATE TABLE IF NOT EXISTS login (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT UNIQUE, password_hash TEXT);
    INSERT INTO login (username, password_hash) VALUES ('{user}', '{hashed}')
    ON CONFLICT(username) DO UPDATE SET password_hash='{hashed}';
    '''
]
subprocess.run(cmd, check=True)
print(f'✅ Password successfully updated for user \'{user}\'')
"

echo ""
echo "=========================================================="
echo "✅ ADMIN CREDENTIALS READY:"
echo "   URL      : https://jo4.co.za/login"
echo "   Username : $TARGET_USER"
echo "   Password : $NEW_PASS"
echo "=========================================================="
