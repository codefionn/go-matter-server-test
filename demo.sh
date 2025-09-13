#!/bin/bash

set -e

echo "🚀 Matter Server Demo"
echo "===================="
echo

# Build the server and client
echo "🔨 Building matter-server and example-client..."
go build -o matter-server ./cmd/matter-server
go build -o example-client ./cmd/example-client
echo "✅ Build complete!"
echo

# Start the server in background
echo "🚀 Starting matter-server in background..."
./matter-server --log-level debug &
SERVER_PID=$!

# Give server time to start
echo "⏳ Waiting for server to start..."
sleep 3

# Function to cleanup server on exit
cleanup() {
    echo ""
    echo "🛑 Stopping matter-server (PID: $SERVER_PID)..."
    kill $SERVER_PID 2>/dev/null || true
    wait $SERVER_PID 2>/dev/null || true
    echo "✅ Demo complete!"
}
trap cleanup EXIT

echo "✅ Matter-server is running!"
echo

# Run the client
echo "🔌 Running example client..."
echo "=========================================="
./example-client

echo
echo "Demo finished. Press Ctrl+C to exit or wait for automatic cleanup..."
sleep 5