#!/bin/bash

# OSP Live Demo & Experience Script 🦞
# This script runs a local cluster and lets you query the Master via HTTP.

set -e

# 1. Paths and Setup
cd "$(dirname "$0")/.."
MASTER_BIN="./bin/master"
WORKER_BIN="./bin/worker"
LOG_DIR="./logs"
mkdir -p $LOG_DIR

# 2. Cleanup Function
cleanup() {
    echo ""
    echo "🦞 Stopping OSP cluster..."
    kill $MASTER_PID $WORKER_PID 2>/dev/null || true
    echo "✨ Cleanup complete."
}
trap cleanup EXIT

echo "🚀 Starting Open Search Protocol (OSP) Cluster..."

# 3. Start Master Node (Custom Ports to avoid conflict)
# HTTP API: 19000 | gRPC: 19004
$MASTER_BIN -http :19000 -grpc :19004 > $LOG_DIR/master.log 2>&1 &
MASTER_PID=$!
echo "📡 Master API: http://localhost:19000/web/search"

# 4. Start Worker Node
# Connects to Master gRPC
$WORKER_BIN -master localhost:19004 -id "worker-local-01" > $LOG_DIR/worker.log 2>&1 &
WORKER_PID=$!
echo "🦾 Worker 'worker-local-01' connected."

echo "⏳ Waiting for cluster stabilization..."
sleep 4

echo ""
echo "🔥 OSP IS ONLINE. You can now test the API."
echo "----------------------------------------------------------------"
echo "Example Command (Copy & Paste in another terminal):"
echo "curl -s \"http://localhost:19000/web/search?q=golang&count=5\" | python3 -m json.tool"
echo "----------------------------------------------------------------"
echo ""
echo "Press [CTRL+C] to stop the cluster and exit."

# Keep the script running to hold the background processes
while true; do
    sleep 1
done
