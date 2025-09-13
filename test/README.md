# Container Testing Environment

This directory contains a complete container orchestration setup for testing the go-matter-server and example client in an isolated environment. It automatically detects and prefers **podman-compose** over **docker-compose** for better compatibility and performance.

## Container Orchestration Detection

The system automatically detects and uses the best available container orchestration tool:

1. **podman-compose** (preferred) - Better security, rootless containers, systemd integration
2. **docker-compose** (fallback) - Traditional Docker Compose
3. **docker compose** (plugin) - Docker CLI with compose plugin

Use the detection script to see what's available:
```bash
cd test
./detect-compose.sh
```

## Quick Start

### 1. Start the Matter Server
```bash
# From the test directory (auto-detects best compose tool)
make server

# Or use compose directly (uses detected tool)
make build
make server-bg
```

### 2. Run the Example Client
```bash
# Run client once
make client

# Or run interactively
make shell-client
```

### 3. Run Integration Tests
```bash
# Run full test suite
make test

# Run all tests and services
make test-all
```

## Services

### matter-server
- **Purpose**: Main Matter server application
- **Port**: 5580 (WebSocket API)
- **Health Check**: HTTP endpoint at `/health`
- **Configuration**: Uses `config/test-config.yaml`
- **Data**: Persisted in `matter-data` volume

### example-client
- **Purpose**: Example client demonstrating API usage
- **Profile**: `client` (run with `--profile client`)
- **Dependencies**: Waits for matter-server to be healthy

### test-runner
- **Purpose**: Automated integration test runner
- **Profile**: `test` (run with `--profile test`)
- **Scripts**: Executes tests from `scripts/run-tests.sh`

## Directory Structure

```
test/
├── docker-compose.yml      # Main compose configuration
├── Dockerfile              # Multi-stage build for Go applications
├── README.md              # This file
├── .env                   # Environment variables
├── config/
│   └── test-config.yaml   # Test configuration for matter-server
├── logs/                  # Log files (mounted volume)
└── scripts/
    └── run-tests.sh       # Integration test script
```

## Configuration

### Environment Variables
Set in `test/.env`:
- `MATTER_LOG_LEVEL=debug` - Logging level
- `MATTER_SERVER_PORT=5580` - WebSocket port
- `MATTER_MDNS_HOSTNAME=matter-server` - mDNS hostname

### Server Configuration
Located in `test/config/test-config.yaml`:
- Enables debug logging
- Configures mDNS with hostname "matter-server"
- Sets up test-friendly defaults

## Usage Examples

### Development Testing
```bash
# Start server in background
docker-compose up -d matter-server

# View logs
docker-compose logs -f matter-server

# Run client tests
docker-compose --profile client up example-client

# Cleanup
docker-compose down -v
```

### Automated Testing (CI/CD)
```bash
# Run complete test suite
docker-compose --profile test up --abort-on-container-exit

# Check exit codes
echo $?
```

### Debug Mode
```bash
# Start with interactive shell
docker-compose run --rm matter-server sh

# Or attach to running container
docker-compose exec matter-server sh
```

## Networking

- **Network**: Custom bridge network `matter-network` (172.20.0.0/16)
- **mDNS**: Works within the Docker network
- **Port Mapping**: Host port 5580 maps to container port 5580

## Health Checks

The matter-server includes a health check that:
- Tests HTTP endpoint availability
- Runs every 30 seconds
- Times out after 10 seconds
- Allows 3 retries
- Waits 40 seconds before starting

## Data Persistence

- **Volume**: `matter-data` persists server data
- **Logs**: `./logs` directory for log files
- **Config**: `./config` mounted read-only

## Troubleshooting

### Container Won't Start
```bash
# Check logs
docker-compose logs matter-server

# Check if ports are in use
netstat -tulpn | grep 5580

# Rebuild images
docker-compose build --no-cache
```

### Client Can't Connect
```bash
# Verify server health
docker-compose exec matter-server wget -q --spider http://localhost:5580/health

# Check network connectivity
docker-compose exec example-client ping matter-server

# Test WebSocket manually
docker-compose exec example-client nc -z matter-server 5580
```

### mDNS Issues
```bash
# Check mDNS logs
docker-compose logs matter-server | grep -i mdns

# Verify hostname resolution
docker-compose exec example-client nslookup matter-server
```

## Advanced Usage

### Custom Configuration
```bash
# Use custom config file
docker-compose run -v /path/to/config.yaml:/app/config.yaml:ro matter-server
```

### Performance Testing
```bash
# Run multiple clients
for i in {1..5}; do
  docker-compose run -d --name client-$i example-client
done
```

### Integration with Host Network
```bash
# Use host networking (for mDNS discovery from host)
docker run --rm --network host go-matter-server
```

## Cleanup

```bash
# Stop and remove containers
docker-compose down

# Remove volumes and networks
docker-compose down -v --remove-orphans

# Remove built images
docker-compose down --rmi all -v --remove-orphans
```