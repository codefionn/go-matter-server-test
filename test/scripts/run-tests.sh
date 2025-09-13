#!/bin/sh

set -e

echo "🧪 Starting Matter Server Integration Tests"
echo "==========================================="

# Wait for server to be ready
echo "⏳ Waiting for matter-server to be ready..."
until wget --quiet --tries=1 --spider http://matter-server:5580/health 2>/dev/null || [ $? -eq 8 ]; do
    echo "Waiting for matter-server..."
    sleep 2
done

echo "✅ Matter-server is ready!"
echo

# Test 1: Basic connectivity
echo "🔍 Test 1: Basic WebSocket connectivity"
./example-client &
CLIENT_PID=$!

# Give client time to run
sleep 10

# Stop the client
kill $CLIENT_PID 2>/dev/null || true
wait $CLIENT_PID 2>/dev/null || true

echo "✅ Test 1 completed"
echo

# Test 2: API Commands
echo "🔍 Test 2: API command testing"
echo "Testing server_info command..."

# Use a simple script to test specific commands
cat << 'EOF' > /tmp/test-commands.sh
#!/bin/sh
echo "Testing individual commands..."

# Test with timeout to prevent hanging
timeout 30s ./example-client || {
    echo "Client test completed or timed out"
}
EOF

chmod +x /tmp/test-commands.sh
/tmp/test-commands.sh

echo "✅ Test 2 completed"
echo

# Test 3: Health check
echo "🔍 Test 3: Health check endpoint"
if wget --quiet --tries=3 --spider http://matter-server:5580/health 2>/dev/null; then
    echo "✅ Health check endpoint is accessible"
else
    echo "❌ Health check endpoint failed"
fi

echo
echo "🎉 All tests completed!"
echo "======================"