#!/usr/bin/env bash

set -e

cd "$(dirname "$0")/.."

echo "🐳 Go Matter Server Docker Demo"
echo "=============================="

# Check if Docker and Docker Compose are available
if ! command -v docker >/dev/null 2>&1; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

if ! command -v docker-compose >/dev/null 2>&1; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

echo
echo "🔨 Building Docker images..."
docker-compose build

echo
echo "🚀 Starting matter-server in background..."
docker-compose up -d matter-server

echo
echo "⏳ Waiting for server to be ready (this may take up to 60 seconds)..."
timeout=60
while [ $timeout -gt 0 ]; do
    if docker-compose exec -T matter-server wget --quiet --tries=1 --spider http://localhost:5580/health 2>/dev/null; then
        break
    fi
    echo -n "."
    sleep 2
    timeout=$((timeout-2))
done

if [ $timeout -le 0 ]; then
    echo
    echo "❌ Server failed to become healthy within 60 seconds"
    echo "📋 Server logs:"
    docker-compose logs matter-server
    exit 1
fi

echo
echo "✅ Matter-server is healthy and ready!"

echo
echo "🔍 Server status:"
docker-compose ps

echo
echo "🧪 Running example client..."
echo "================================"
docker-compose --profile client up example-client

echo
echo "🧪 Running integration tests..."
echo "==============================="
docker-compose --profile test up test-runner

echo
echo "📊 Final status:"
docker-compose ps

echo
echo "📋 Cleaning up..."
docker-compose down

echo
echo "✅ Demo completed successfully!"
echo
echo "📚 Next steps:"
echo "  - Use 'make help' in the test directory for more commands"
echo "  - Check test/README.md for detailed documentation"
echo "  - Customize test/config/test-config.yaml for your needs"