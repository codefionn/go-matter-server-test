# Podman Integration Guide

## Overview

The container test environment automatically detects and prefers **podman-compose** over **docker-compose** for better security, performance, and compatibility. This document explains the integration and benefits.

## Why Podman?

### Security Benefits
- **Rootless containers** - No daemon running as root
- **Better isolation** - Uses user namespaces and cgroups v2
- **SELinux integration** - Native support for enhanced security
- **No privileged daemon** - Reduced attack surface

### Performance Benefits
- **Lower resource usage** - No background daemon
- **Better networking** - Uses CNI plugins
- **Systemd integration** - Native container management
- **Faster startup** - No daemon initialization overhead

### Compatibility Benefits
- **Drop-in replacement** - Same CLI interface as Docker
- **OCI compliance** - Full compatibility with container standards
- **Multi-architecture** - Better ARM64 and other architecture support

## Automatic Detection

The system uses `detect-compose.sh` to automatically choose the best available tool:

```bash
# Detection priority order:
1. podman-compose (preferred)
2. docker-compose (fallback)
3. docker compose (plugin fallback)
```

### Detection Logic

```bash
# Test podman-compose first
if command -v podman-compose >/dev/null 2>&1; then
    if timeout 5s podman-compose version >/dev/null 2>&1; then
        USE_PODMAN_COMPOSE=true
    fi
fi

# Fallback to docker-compose
if [ "$USE_PODMAN_COMPOSE" != "true" ]; then
    # Use docker-compose or docker compose
fi
```

## Installation

### Ubuntu/Debian
```bash
# Install Podman
sudo apt-get update
sudo apt-get install -y podman python3-pip

# Install podman-compose
pip3 install podman-compose

# Optional: Configure for rootless
sudo usermod --add-subuids 100000-165535 --add-subgids 100000-165535 $USER
```

### RHEL/CentOS/Fedora
```bash
# Install Podman (usually pre-installed)
sudo dnf install -y podman python3-pip

# Install podman-compose
pip3 install podman-compose
```

### Arch Linux
```bash
# Install Podman
sudo pacman -S podman python-pip

# Install podman-compose
pip install podman-compose
```

## Configuration Differences

### Compose File Selection

The system automatically selects the appropriate compose file:

- **podman-compose**: Prefers `docker-compose.simple.yml` (avoids networking issues)
- **docker-compose**: Uses `docker-compose.yml` (full networking support)

### Network Configuration

**Podman Compose:**
```yaml
# Uses simpler networking to avoid privilege issues
networks:
  matter-network:
    driver: bridge  # No custom IPAM
```

**Docker Compose:**
```yaml
# Can use advanced networking features
networks:
  matter-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
```

### Volume Handling

**Podman advantages:**
- Rootless volume mounts
- Better permission handling
- SELinux context preservation

## Usage Examples

### Makefile Integration

```bash
# All Makefile targets automatically use detected tool
make help        # Shows which tool is being used
make build       # Uses podman-compose if available
make server      # Starts with detected orchestration
make test        # Runs tests with best available tool
```

### Manual Usage

```bash
# Detect what's available
cd test
./detect-compose.sh

# Use detected command
source ./detect-compose.sh
compose_cmd="$(get_compose_command)"
$compose_cmd up matter-server
```

### CI/CD Integration

The GitHub Actions pipeline automatically:

1. **Installs podman-compose** as preferred tool
2. **Detects available tools** using the detection script
3. **Uses appropriate compose file** based on tool capabilities
4. **Handles networking differences** automatically

```yaml
- name: Install podman-compose (preferred)
  run: |
    sudo apt-get install -y podman python3-pip
    pip3 install podman-compose

- name: Detect and validate container orchestration
  run: |
    source ./detect-compose.sh
    echo "Using: $(get_compose_command)"
```

## Troubleshooting

### Common Issues

#### 1. podman-compose not found
```bash
# Install podman-compose
pip3 install podman-compose

# Or use package manager
sudo apt-get install podman-compose
```

#### 2. Permission issues with volumes
```bash
# Configure subuid/subgid for rootless
sudo usermod --add-subuids 100000-165535 --add-subgids 100000-165535 $USER
podman system migrate  # Migrate to new UID ranges
```

#### 3. Network connectivity issues
```bash
# Use simple compose file
podman-compose -f docker-compose.simple.yml up
```

#### 4. SELinux context issues
```bash
# Fix SELinux contexts
sudo setsebool -P container_manage_cgroup true
```

### Debugging Tools

```bash
# Check podman status
podman system info

# Test container functionality
podman run --rm alpine echo "Hello World"

# Check compose functionality
podman-compose --version

# Debug networking
podman network ls
podman network inspect podman
```

### Performance Tuning

```bash
# Enable cgroups v2 (if not already)
sudo grubby --update-kernel=ALL --args="systemd.unified_cgroup_hierarchy=1"

# Configure storage driver
echo 'driver = "overlay"' >> ~/.config/containers/storage.conf

# Optimize for development
echo 'events_logger = "none"' >> ~/.config/containers/containers.conf
```

## Migration from Docker

### Automatic Migration
The system handles migration automatically - no changes needed to compose files or commands.

### Manual Migration Steps
If you want to fully migrate:

```bash
# 1. Install podman and podman-compose
sudo apt-get install -y podman python3-pip
pip3 install podman-compose

# 2. Test detection
cd test
./detect-compose.sh

# 3. Use Makefile (automatically uses podman-compose)
make build
make server

# 4. Verify everything works
make test
```

### Verification

```bash
# Check what's being used
make help
# Should show: "Using: podman-compose"

# Test full functionality
make quick-test

# Verify in CI
./ci-test.sh
```

## Advanced Configuration

### Custom Detection Logic

Modify `detect-compose.sh` to customize tool selection:

```bash
# Force specific tool
export FORCE_COMPOSE_CMD="podman-compose"

# Skip podman-compose
export SKIP_PODMAN_COMPOSE=true
```

### Compose File Selection

Override automatic file selection:

```bash
# Force specific compose file
export FORCE_COMPOSE_FILE="docker-compose.yml"

# Use in Makefile
make COMPOSE_FILE=docker-compose.simple.yml server
```

### Environment Variables

```bash
# Podman-specific settings
export PODMAN_COMPOSE_WARNING_LOGS=false
export PODMAN_COMPOSE_PARALLEL=true

# Docker compatibility
export COMPOSE_DOCKER_CLI_BUILD=1
export DOCKER_BUILDKIT=1
```

## Benefits Summary

| Feature | Docker Compose | Podman Compose |
|---------|---------------|----------------|
| Root privileges | Required for daemon | Rootless by default |
| Resource usage | Higher (daemon) | Lower (no daemon) |
| Security | Standard | Enhanced (SELinux, user ns) |
| Networking | Full features | Simplified but secure |
| Performance | Good | Better (less overhead) |
| CI/CD | Standard | Preferred in secure environments |

## Support Matrix

| Environment | Docker Compose | Podman Compose | Recommended |
|-------------|---------------|----------------|-------------|
| Development | ✅ | ✅ | podman-compose |
| CI/CD | ✅ | ✅ | podman-compose |
| Production | ✅ | ✅ | podman-compose |
| Rootless | ❌ | ✅ | podman-compose |
| Air-gapped | ✅ | ✅ | podman-compose |

The container test environment provides seamless integration with both tools while preferring the more secure and efficient podman-compose approach.