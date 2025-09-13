# CI/CD Pipeline Integration

## Overview

The Docker test environment has been fully integrated into the GitHub Actions CI/CD pipeline. The new `docker-test` job provides comprehensive testing of the containerized matter-server and client applications.

## New CI Jobs Added

### `docker-test` Job

**Purpose**: Test the complete Docker environment including containerized server, client, and integration tests.

**Dependencies**: Runs after `format-check`, `lint`, and `test` jobs pass.

**Key Features**:
- ✅ Validates Docker Compose configurations
- ✅ Builds and tests Docker images
- ✅ Runs health checks on containerized services
- ✅ Tests WebSocket API endpoints
- ✅ Validates mDNS functionality
- ✅ Runs integration test suite
- ✅ Collects comprehensive test artifacts
- ✅ Handles cleanup automatically

### Pipeline Steps

#### 1. **Setup and Validation** 
```yaml
- name: Verify Docker test setup
- name: Validate Docker Compose configurations
```
- Runs our custom `verify-setup.sh` script
- Validates both main and simple Docker Compose files
- Checks for all required configuration files

#### 2. **Image Building**
```yaml
- name: Build Docker test images
```
- Builds Docker images using Docker Compose
- Fallback to simple compose file if main build fails
- Uses parallel building for efficiency

#### 3. **Health Check Testing**
```yaml
- name: Run Docker Compose health checks
```
- Starts matter-server container in background
- Waits up to 2 minutes for health check to pass
- Validates server is running and responsive

#### 4. **Client and API Testing**
```yaml
- name: Run example client test
- name: Run integration test suite
- name: Test WebSocket API endpoints
```
- Tests example client connection
- Runs full integration test suite
- Validates HTTP health endpoint
- Tests WebSocket connectivity

#### 5. **mDNS and Network Testing**
```yaml
- name: Test mDNS functionality
```
- Validates mDNS server logs and functionality
- Checks network service discovery

#### 6. **Artifact Collection and Cleanup**
```yaml
- name: Collect test artifacts
- name: Upload test artifacts
- name: Cleanup Docker resources
```
- Collects all container logs and system information
- Uploads artifacts for debugging (7-day retention)
- Performs comprehensive cleanup

## Integration Points

### Updated Job Dependencies

The `docker-test` job has been integrated into the pipeline dependencies:

```yaml
# These jobs now depend on docker-test:
docker-build:
  needs: [format-check, lint, test, docker-test]

docker-dev-build:
  needs: [format-check, lint, test, docker-test]

release:
  needs: [format-check, lint, test, integration-test, docker-test, docker-build]
```

This ensures Docker tests must pass before:
- Building production Docker images
- Creating releases
- Building development images for PRs

## Artifact Collection

### Test Artifacts Collected
- `docker-compose.log` - All container logs
- `matter-server.log` - Server-specific logs
- `example-client.log` - Client-specific logs  
- `test-runner.log` - Integration test logs
- `container-status.txt` - Container status information
- `docker-system.txt` - Docker system information
- `docker-images.txt` - Built image information

### Artifact Access
Artifacts are available in GitHub Actions:
1. Go to the "Actions" tab
2. Select the workflow run
3. Download "docker-test-artifacts" 
4. Artifacts are retained for 7 days

## Local Testing

### Test CI Pipeline Locally
```bash
# Run CI simulation
./test/ci-test.sh

# Run specific tests
cd test
make test
make client
make health
```

### Validate Changes Before Pushing
```bash
# Validate configurations
cd test
docker-compose config --quiet
docker-compose -f docker-compose.simple.yml config --quiet

# Test setup
./verify-setup.sh

# Run comprehensive test
./ci-test.sh
```

## Error Handling

### Common Scenarios Handled
1. **Docker build failures** - Fallback to simple compose
2. **Health check timeouts** - Detailed logging and graceful failure
3. **Network issues** - Alternative test methods
4. **Resource cleanup** - Always runs regardless of test results

### Debugging Failed Runs
1. **Check job logs** in GitHub Actions
2. **Download artifacts** for detailed investigation
3. **Run locally** using `./test/ci-test.sh`
4. **Use simple compose** if networking issues occur

## Performance Considerations

### Optimizations
- **Parallel building** of Docker images
- **Multi-stage Dockerfile** for smaller images
- **BuildKit** enabled for faster builds
- **Efficient layer caching** strategy

### Resource Usage
- **Memory**: ~512MB per container
- **Build time**: ~5-10 minutes total
- **Storage**: Minimal due to cleanup
- **Network**: Uses bridge networking

## Security

### Safety Measures
- **No secrets** required for testing
- **Isolated networking** prevents external access
- **Cleanup guaranteed** via `if: always()` blocks
- **Read-only** configuration mounting

### Best Practices
- All test data is ephemeral
- No persistent volumes in CI
- Container processes run as non-root
- Network isolation between test runs

## Future Enhancements

### Planned Additions
- **Multi-platform testing** (AMD64, ARM64)
- **Performance benchmarking** tests
- **Security scanning** of images
- **Load testing** with multiple clients

### Configuration
All CI behavior can be customized via:
- `test/.env` - Environment variables
- `test/config/test-config.yaml` - Server configuration
- `test/docker-compose.yml` - Service definitions
- `test/scripts/run-tests.sh` - Test logic

## Troubleshooting

### Common Issues

1. **Build failures**:
   ```bash
   # Check Docker daemon
   docker version
   
   # Validate compose files
   docker-compose config
   ```

2. **Health check failures**:
   ```bash
   # Test local health endpoint
   curl http://localhost:5580/health
   
   # Check server logs
   docker-compose logs matter-server
   ```

3. **Network connectivity**:
   ```bash
   # Test container networking
   docker network ls
   docker network inspect test_matter-network
   ```

### Getting Help

- **Check GitHub Actions logs** first
- **Download and examine artifacts**
- **Run `./test/ci-test.sh` locally**
- **Use `make help` for available commands**
- **Review `test/TROUBLESHOOTING.md`** for detailed solutions

## Summary

The Docker test integration provides:
- ✅ **Complete containerized testing**
- ✅ **Automated health checks**
- ✅ **Comprehensive artifact collection**
- ✅ **Robust error handling**
- ✅ **Efficient resource usage**
- ✅ **Easy local reproduction**

This ensures that both the matter-server and example client work correctly in containerized environments, matching production deployment scenarios.