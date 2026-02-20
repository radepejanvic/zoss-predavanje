#!/bin/bash
#
# Upload Test Data to HDFS
# =========================
#
# This script uploads GDPR-sensitive travel data to the Hadoop cluster for testing.
#
# Usage:
#   bash upload-data.sh
#

set +e

# Check if cluster is running
echo "[1/3] Checking if Hadoop cluster is running..."
if ! docker ps | grep -q "edgenode-1"; then
    echo "ERROR: Hadoop cluster is not running!"
    echo "Please start the cluster first with: docker compose up -d"
    exit 1
fi
echo "Cluster is running."
echo ""

# Create HDFS directory
echo "[2/3] Creating HDFS directory /data/travels..."
MSYS_NO_PATHCONV=1 docker exec edgenode-1 bash -c "kinit -kt /etc/security/keytabs/hadoop/tester.keytab tester && hdfs dfs -mkdir -p /data/travels" 2>/dev/null
if [ $? -eq 0 ]; then
    echo "Directory created successfully."
else
    echo "Directory already exists or error occurred (continuing...)."
fi
echo ""

# Copy file to container
echo "[3/3] Uploading travels.json to HDFS..."
MSYS_NO_PATHCONV=1 docker cp ./data/travels.json edgenode-1:/tmp/travels.json 2>/dev/null

# Upload to HDFS
MSYS_NO_PATHCONV=1 docker exec edgenode-1 bash -c "kinit -kt /etc/security/keytabs/hadoop/tester.keytab tester && hdfs dfs -put -f /tmp/travels.json /data/travels/travels.json" 2>/dev/null

if [ $? -eq 0 ]; then
    echo "Upload successful."
else
    echo "ERROR: Upload failed!"
    exit 1
fi
echo ""

# Verify upload
echo "Verifying upload..."
FILE_COUNT=$(MSYS_NO_PATHCONV=1 docker exec edgenode-1 bash -c "kinit -kt /etc/security/keytabs/hadoop/tester.keytab tester && hdfs dfs -ls /data/travels/ 2>/dev/null | grep travels.json | wc -l")

if [ "$FILE_COUNT" -ge 1 ]; then
    echo "==================================================================="
    echo "SUCCESS: Test data uploaded to HDFS"
    echo "==================================================================="
    echo "Location: /data/travels/travels.json"
else
    echo "WARNING: Could not verify file in HDFS"
fi
