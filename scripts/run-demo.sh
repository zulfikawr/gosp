#!/bin/bash

# GOSP Live Demo & Experience Script 🦞
# This script runs a local cluster using the unified binary.

set -e

# 1. Paths and Setup
cd "$(dirname "$0")/.."
GOSP_BIN="./bin/gosp"
LOG_DIR="./logs"
mkdir -p $LOG_DIR

# 2. Cleanup Function
cleanup() {
    echo ""
    echo "🦞 Stopping GOSP cluster..."
    pkill -f "$GOSP_BIN master" || true
    pkill -f "$GOSP_BIN worker" || true
    echo "✨ Cleanup complete."
}
trap cleanup EXIT

echo "🚀 Starting Go OpenSearchProtocol (GOSP) Cluster..."

# 3. Start Master Node (Custom Ports to avoid conflict)
# HTTP API: 19000 | gRPC: 19004
$GOSP_BIN master --http :19000 --grpc :19004 > $LOG_DIR/master.log 2>&1 &
MASTER_PID=$!
echo "📡 Master API: http://localhost:19000/web/search"

# 4. Start Worker Node
# Connects to Master gRPC
$GOSP_BIN worker --master localhost:19004 --id "worker-local-01" > $LOG_DIR/worker.log 2>&1 &
WORKER_PID=$!
echo "🦾 Worker 'worker-local-01' connected."

echo "⏳ Waiting for cluster stabilization..."
sleep 4

echo ""
echo "🔥 GOSP IS ONLINE. You can now test the API via CLI."
echo "----------------------------------------------------------------"
echo "Example Command:"
echo "./bin/gosp search --query \"golang concurrency\""
echo "----------------------------------------------------------------"
echo ""
echo "Press [CTRL+C] to stop the cluster and exit."

# Keep the script running to hold the background processes
while true; do
    sleep 1
done
