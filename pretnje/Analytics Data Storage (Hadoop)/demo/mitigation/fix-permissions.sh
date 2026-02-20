#!/bin/bash
#
set +e

# Check if containers are running
echo "[1/4] Checking if Hadoop cluster is running..."
if ! docker ps | grep -q "edgenode-1"; then
    echo "ERROR: Hadoop cluster is not running!"
    echo "Please start the cluster first with: docker compose up -d"
    exit 1
fi
echo "Cluster is running."
echo ""

# Show current vulnerable permissions
echo "[2/4] Checking current keytab permissions..."
# Use MSYS_NO_PATHCONV to prevent Git Bash path translation on Windows
PERMS=$(MSYS_NO_PATHCONV=1 docker exec edgenode-1 stat -c "%a" /etc/security/keytabs/hadoop/tester.keytab 2>/dev/null || echo "000")
MSYS_NO_PATHCONV=1 docker exec edgenode-1 ls -la /etc/security/keytabs/hadoop/tester.keytab 2>/dev/null || true

if [ "$PERMS" == "644" ] || [ "$PERMS" == "664" ] || [ "$PERMS" == "666" ]; then
    echo "VULNERABILITY DETECTED: Keytab has world-readable permissions ($PERMS)"
elif [ "$PERMS" == "600" ] || [ "$PERMS" == "400" ]; then
    echo "Keytab already has secure permissions ($PERMS)"
    echo "No fix needed - system is already secure."
    exit 0
fi
echo ""

# Apply fix
echo "[3/4] Applying permission fix..."

# List of containers that have keytabs (only those actually running)
CONTAINERS=(
    "edgenode-1"
    "namenode-1"
    "datanode-1"
    "datanode-2"
    "resourcemanager-1"
    "nodemanager-1"
    "nodemanager-2"
)

FIXED_COUNT=0
for CONTAINER in "${CONTAINERS[@]}"; do
    if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER}$"; then
        echo "Fixing ${CONTAINER}..."
        
        # Fix keytab permissions (600 = owner read/write only)
        MSYS_NO_PATHCONV=1 docker exec "$CONTAINER" bash -c "
            if [ -d /etc/security/keytabs/hadoop ]; then
                chmod 600 /etc/security/keytabs/hadoop/*.keytab 2>/dev/null || true
            fi
        " 2>/dev/null || true
        
        ((FIXED_COUNT++))
    fi
done

echo "Fixed permissions on $FIXED_COUNT containers."
echo ""

# Verify fix
echo "[4/4] Verifying fix..."
MSYS_NO_PATHCONV=1 docker exec edgenode-1 ls -la /etc/security/keytabs/hadoop/tester.keytab 2>/dev/null

NEW_PERMS=$(MSYS_NO_PATHCONV=1 docker exec edgenode-1 stat -c "%a" /etc/security/keytabs/hadoop/tester.keytab 2>/dev/null || echo "000")
echo ""
if [ "$NEW_PERMS" == "600" ]; then
    echo "Keytab permissions are now secure: 600 (-rw-------)"
    echo "The keytab leak vulnerability has been mitigated."
else
    echo "Warning: Permissions are $NEW_PERMS instead of 600"
    echo "Verify manually if needed."
fi
echo ""
