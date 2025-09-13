# Docker Test Environment Troubleshooting

## Common Issues and Solutions

### Docker/Podman Networking Issues

**Problem**: `netavark (exit code 1): create bridge: Netlink error: Operation not supported`

**Solutions**:
1. **Use simple compose file**:
   ```bash
   docker-compose -f docker-compose.simple.yml up
   ```

2. **Use host networking**:
   ```bash
   docker run --rm --network host matter-server-test
   ```

3. **Switch from Podman to Docker**:
   ```bash
   # On systems using Podman
   sudo systemctl disable podman.socket
   # Install Docker following official docs
   ```

### Container Build Failures

**Problem**: Build context too large or slow builds

**Solutions**:
1. **Use .dockerignore**:
   - Already provided in `test/.dockerignore`
   - Excludes test files, docs, and build artifacts

2. **Multi-stage build optimization**:
   - Uses minimal Alpine base image
   - Only copies necessary files to final image

3. **Build with specific target**:
   ```bash
   docker build --target builder -f test/Dockerfile ..
   ```

### Health Check Failures

**Problem**: Container fails health checks

**Solutions**:
1. **Check if server is listening**:
   ```bash
   docker exec matter-server netstat -tlnp | grep 5580
   ```

2. **Test health endpoint manually**:
   ```bash
   docker exec matter-server wget -q --spider http://localhost:5580/health
   ```

3. **Check server logs**:
   ```bash
   docker-compose logs matter-server
   ```

### mDNS Discovery Issues

**Problem**: Client can't discover server via mDNS

**Causes and Solutions**:
1. **Docker networking isolation**:
   - mDNS multicast doesn't cross Docker bridge networks
   - Use `--network host` for testing mDNS

2. **Container hostname mismatch**:
   - Server advertises container hostname, not "matter-server.local"
   - Client falls back to localhost connection

3. **Firewall blocking multicast**:
   - Ensure UDP port 5353 is open
   - Check iptables/firewalld rules

### WebSocket Connection Issues

**Problem**: Client can't connect to server WebSocket

**Solutions**:
1. **Check port mapping**:
   ```bash
   docker-compose ps
   # Should show 0.0.0.0:5580->5580/tcp
   ```

2. **Test connectivity**:
   ```bash
   # From host
   telnet localhost 5580
   
   # From container
   docker exec example-client nc -z matter-server 5580
   ```

3. **Check server binding**:
   - Ensure server binds to 0.0.0.0:5580, not 127.0.0.1:5580
   - Check configuration: `listen_addresses: ["0.0.0.0"]`

### Storage/Permission Issues

**Problem**: Server can't write to storage directory

**Solutions**:
1. **Fix volume permissions**:
   ```bash
   sudo chown -R $(id -u):$(id -g) test/data
   chmod 755 test/data
   ```

2. **Use bind mounts instead of volumes**:
   ```yaml
   volumes:
     - ./data:/app/data  # Instead of named volume
   ```

### Memory/Resource Issues

**Problem**: Containers consuming too much memory

**Solutions**:
1. **Add resource limits**:
   ```yaml
   deploy:
     resources:
       limits:
         memory: 512M
         cpus: '0.5'
   ```

2. **Use lighter base image**:
   - Already using `alpine:latest`
   - Consider `scratch` for even smaller size

### Go Build Issues

**Problem**: Go modules or build failures

**Solutions**:
1. **Clear module cache**:
   ```bash
   go clean -modcache
   go mod download
   ```

2. **Update dependencies**:
   ```bash
   go mod tidy
   go mod vendor  # If using vendor directory
   ```

3. **Cross-compilation issues**:
   ```bash
   CGO_ENABLED=0 GOOS=linux go build
   ```

## Debugging Tools

### Container Inspection
```bash
# Enter running container
docker exec -it matter-server sh

# Inspect container details
docker inspect matter-server

# Check resource usage
docker stats matter-server
```

### Network Debugging
```bash
# List Docker networks
docker network ls

# Inspect network details
docker network inspect test_matter-network

# Test connectivity between containers
docker exec example-client ping matter-server
```

### Log Analysis
```bash
# Follow all logs
docker-compose logs -f

# Filter specific service logs
docker-compose logs -f matter-server | grep ERROR

# Export logs to file
docker-compose logs --no-color > debug.log
```

### Environment Verification
```bash
# Check environment variables
docker exec matter-server env | grep MATTER

# Verify configuration
docker exec matter-server cat /app/config.yaml

# Check running processes
docker exec matter-server ps aux
```

## Performance Optimization

### Build Optimization
1. **Use BuildKit**:
   ```bash
   DOCKER_BUILDKIT=1 docker build ...
   ```

2. **Multi-stage builds**:
   - Separate build and runtime stages
   - Minimize final image size

3. **Layer caching**:
   - Copy `go.mod`/`go.sum` first
   - Download dependencies in separate layer

### Runtime Optimization
1. **Resource limits**:
   - Set appropriate memory/CPU limits
   - Use swap accounting if needed

2. **Volume optimization**:
   - Use tmpfs for temporary data
   - Optimize I/O with proper volume types

## Alternative Setups

### Without Docker Compose
```bash
# Build image
docker build -f test/Dockerfile -t matter-server .

# Run server
docker run -d -p 5580:5580 --name matter-server matter-server

# Run client
docker run --rm --link matter-server matter-server ./example-client
```

### With Docker Swarm
```bash
# Initialize swarm
docker swarm init

# Deploy stack
docker stack deploy -c docker-compose.yml matter-stack
```

### With Kubernetes
```yaml
# Basic deployment (example)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: matter-server
spec:
  replicas: 1
  selector:
    matchLabels:
      app: matter-server
  template:
    spec:
      containers:
      - name: matter-server
        image: matter-server:latest
        ports:
        - containerPort: 5580
```

## Getting Help

If you encounter issues not covered here:

1. **Check logs first**:
   ```bash
   docker-compose logs --tail=50 matter-server
   ```

2. **Run verification script**:
   ```bash
   ./verify-setup.sh
   ```

3. **Test without Docker**:
   ```bash
   # Build and run locally
   make quick-test
   ```

4. **Create minimal reproduction case**:
   ```bash
   # Use simple compose file
   docker-compose -f docker-compose.simple.yml up
   ```