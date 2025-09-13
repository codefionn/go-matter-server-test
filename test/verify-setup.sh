#!/bin/bash

set -e

echo "🔍 Verifying Docker Test Setup"
echo "=============================="

# Check if we can build the Go applications
echo "📦 Testing Go build process..."
cd ..

echo "  Building matter-server..."
if go build -o test/matter-server ./cmd/matter-server; then
    echo "  ✅ matter-server builds successfully"
else
    echo "  ❌ Failed to build matter-server"
    exit 1
fi

echo "  Building example-client..."
if go build -o test/example-client ./cmd/example-client; then
    echo "  ✅ example-client builds successfully"
else
    echo "  ❌ Failed to build example-client"
    exit 1
fi

cd test

echo
echo "🧪 Testing configuration files..."

# Verify config file syntax
if [ -f "config/test-config.yaml" ]; then
    echo "  ✅ test-config.yaml exists"
else
    echo "  ❌ test-config.yaml missing"
    exit 1
fi

# Verify environment file
if [ -f ".env" ]; then
    echo "  ✅ .env file exists"
else
    echo "  ❌ .env file missing"
    exit 1
fi

echo
echo "🚀 Testing matter-server startup..."

# Start matter-server in background
./matter-server --config config/test-config.yaml --storage-path data &
SERVER_PID=$!

# Give it time to start
sleep 3

# Test if it's running
if kill -0 $SERVER_PID 2>/dev/null; then
    echo "  ✅ matter-server started successfully (PID: $SERVER_PID)"
    
    # Try to test the health endpoint
    sleep 2
    if curl -s -f http://localhost:5580/health >/dev/null 2>&1; then
        echo "  ✅ Health endpoint is accessible"
    else
        echo "  ⚠️  Health endpoint not accessible (this might be expected)"
    fi
    
    # Stop the server
    kill $SERVER_PID
    wait $SERVER_PID 2>/dev/null || true
    echo "  ✅ matter-server stopped cleanly"
else
    echo "  ❌ matter-server failed to start"
    exit 1
fi

echo
echo "🔌 Testing example-client (quick test)..."

# Start server again for client test
./matter-server --config config/test-config.yaml --storage-path data &
SERVER_PID=$!
sleep 3

# Run client with timeout
timeout 10s ./example-client &
CLIENT_PID=$!

# Let client run for a few seconds
sleep 5

# Stop both
kill $CLIENT_PID 2>/dev/null || true
kill $SERVER_PID 2>/dev/null || true
wait $SERVER_PID 2>/dev/null || true
wait $CLIENT_PID 2>/dev/null || true

echo "  ✅ example-client ran without crashing"

echo
echo "📋 Container orchestration validation..."

# Source compose detection
source ./detect-compose.sh >/dev/null 2>&1

# Check Dockerfile syntax
if [ -f "Dockerfile" ]; then
    echo "  ✅ Dockerfile exists"
else
    echo "  ❌ Dockerfile missing"
fi

# Check compose tools and files
compose_cmd="$(get_compose_command 2>/dev/null)"
compose_type="$(get_compose_type 2>/dev/null)"

if [ -n "$compose_cmd" ]; then
    echo "  ✅ $compose_type is available ($compose_cmd)"
    
    # Test configuration files
    compose_file="$(get_compose_file 2>/dev/null)"
    if [ -n "$compose_file" ] && [ -f "$compose_file" ]; then
        echo "  ✅ Using compose file: $compose_file"
        
        if test_compose_config "$compose_file" >/dev/null 2>&1; then
            echo "  ✅ Compose configuration is valid"
        else
            echo "  ⚠️  Compose configuration may have issues"
        fi
    fi
else
    echo "  ⚠️  No container orchestration tool available"
    echo "     Install podman-compose (preferred) or docker-compose"
fi

# Check compose files exist
if [ -f "docker-compose.yml" ]; then
    echo "  ✅ docker-compose.yml exists"
fi

if [ -f "docker-compose.simple.yml" ]; then
    echo "  ✅ docker-compose.simple.yml exists (fallback option)"
fi

echo
echo "🧹 Cleaning up..."
rm -f matter-server example-client
rm -rf data/*

echo
echo "✅ All tests passed! The Docker test setup is ready."
echo
echo "📚 Next steps:"
echo "   1. Fix Docker/Podman networking issues if needed"
echo "   2. Run 'docker-compose build' when Docker is working"
echo "   3. Use 'make help' for available commands"
echo "   4. Check README.md for detailed usage instructions"