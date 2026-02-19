#!/bin/bash

# OpenSearchProtocol (OSP) Integration Test Script
# This script spins up a local Master and Worker, performs a search, and cleans up.

set -e

# 1. Paths and Setup
cd "$(dirname "$0")/.."
MASTER_BIN="./bin/master"
WORKER_BIN="./bin/worker"
LOG_DIR="./logs"
mkdir -p $LOG_DIR

# 2. Cleanup Function
cleanup() {
    echo "Stopping OSP cluster..."
    kill $MASTER_PID $WORKER_PID 2>/dev/null || true
    echo "Cleanup complete."
}
trap cleanup EXIT

# 3. Start Master Node
echo "Starting OSP Master..."
$MASTER_BIN -http :19000 -grpc :19001 > $LOG_DIR/master.log 2>&1 &
MASTER_PID=$!
sleep 2 # Wait for Master to bind ports

# 4. Start Worker Node
echo "Starting OSP Worker..."
$WORKER_BIN -master localhost:19001 -id "test-worker-01" > $LOG_DIR/worker.log 2>&1 &
WORKER_PID=$!
sleep 3 # Wait for Worker to register and connect

# 5. Perform Search Request
echo "Performing Search (Brave): 'golang'..."
RESPONSE=$(curl -s "http://localhost:19000/web/search?q=golang&engine=3&count=5")

# 6. Verify Results
if [[ $(echo "$RESPONSE" | grep -c "type") -ge 1 ]]; then
    echo "SUCCESS: Received valid search JSON from OSP Master."
    echo "--------------------------------------------------"
    echo "$RESPONSE" | python3 -m json.tool || echo "$RESPONSE"
    echo "--------------------------------------------------"
else
    echo "FAILURE: Invalid response from OSP Master."
    echo "Response: $RESPONSE"
    echo "Master Logs:"
    cat $LOG_DIR/master.log
    echo "Worker Logs:"
    cat $LOG_DIR/worker.log
    exit 1
fi
