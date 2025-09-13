# Environment Variables

The Go Matter Server supports configuration via environment variables, CLI flags, and configuration files. This document describes all available environment variables.

## Configuration Precedence

Configuration is loaded in the following order (highest to lowest precedence):

1. **CLI flags** (highest precedence)
2. **Environment variables**
3. **Configuration file** (YAML)
4. **Default values** (lowest precedence)

## Environment Variable Format

All environment variables use the prefix `MATTER_` followed by the configuration path in uppercase with dots and hyphens replaced by underscores.

**Format:** `MATTER_<SECTION>_<KEY>`

Example: `--port` flag becomes `MATTER_SERVER_PORT` environment variable

## Server Configuration

| Environment Variable | CLI Flag | Description | Default |
|---------------------|----------|-------------|---------|
| `MATTER_SERVER_PORT` | `--port`, `-p` | WebSocket server port | `5580` |
| `MATTER_SERVER_LISTEN_ADDRESSES` | `--listen`, `-l` | Comma-separated list of listen addresses | All interfaces |

## Storage Configuration

| Environment Variable | CLI Flag | Description | Default |
|---------------------|----------|-------------|---------|
| `MATTER_STORAGE_PATH` | `--storage-path` | Storage path for persistent data | `$PWD/.matter_server` |

## Matter Configuration

| Environment Variable | CLI Flag | Description | Default |
|---------------------|----------|-------------|---------|
| `MATTER_MATTER_VENDOR_ID` | `--vendor-id` | Vendor ID for the Fabric | `65521` (0xFFF1) |
| `MATTER_MATTER_FABRIC_ID` | `--fabric-id` | Fabric ID for the Fabric | `1` |
| `MATTER_MATTER_PAA_ROOT_CERT_DIR` | `--paa-root-cert-dir` | Directory for PAA root certificates | _(empty)_ |
| `MATTER_MATTER_ENABLE_TEST_NET_DCL` | `--enable-test-net-dcl` | Enable test-net DCL certificates | `false` |
| `MATTER_MATTER_DISABLE_SERVER_INTERACTIONS` | `--disable-server-interactions` | Disable server cluster interactions | `false` |

## Network Configuration

| Environment Variable | CLI Flag | Description | Default |
|---------------------|----------|-------------|---------|
| `MATTER_NETWORK_PRIMARY_INTERFACE` | `--primary-interface` | Primary network interface for link-local addresses | _(empty)_ |

## Bluetooth Configuration

Bluetooth is internal-only and auto-enables when an adapter ID is provided.

| Environment Variable | CLI Flag | Description | Default |
|---------------------|----------|-------------|---------|
| `MATTER_BLUETOOTH_ADAPTER_ID` | `--bluetooth-adapter` | Bluetooth adapter ID. When `>= 0`, Bluetooth is enabled; when `-1`, disabled. | `-1` |

Notes:
- `MATTER_BLUETOOTH_ENABLED` is ignored. Availability is determined solely by `BLUETOOTH_ADAPTER_ID` and runtime BlueZ/DBus readiness. The `server_info.bluetooth_enabled` field reflects actual availability, not just configuration.

## OTA Configuration

| Environment Variable | CLI Flag | Description | Default |
|---------------------|----------|-------------|---------|
| `MATTER_OTA_PROVIDER_DIR` | `--ota-provider-dir` | Directory for OTA Provider software updates | _(empty)_ |

## mDNS Configuration

| Environment Variable | CLI Flag | Description | Default |
|---------------------|----------|-------------|---------|
| `MATTER_MDNS_ENABLED` | `--mdns-enabled` | Enable mDNS hostname advertisement | `true` |
| `MATTER_MDNS_HOSTNAME` | `--mdns-hostname` | Hostname to advertise via mDNS | System hostname + `.local` |

## Logging Configuration

| Environment Variable | CLI Flag | Description | Default | Options |
|---------------------|----------|-------------|---------|---------|
| `MATTER_LOG_LEVEL` | `--log-level` | Log level | `info` | `trace`, `debug`, `info`, `warn`, `error`, `fatal` |
| `MATTER_LOG_FORMAT` | `--log-format` | Log format | `console` | `console`, `json` |

## Global Configuration

| Environment Variable | CLI Flag | Description | Default |
|---------------------|----------|-------------|---------|
| _(none)_ | `--config` | Configuration file path | `$HOME/.matter_server/config.yaml` |
| _(none)_ | `--env-file` | Environment file to load | _(none - autoloads .env if present)_ |

## .env File Support

The server automatically loads environment variables from a `.env` file in the current directory if it exists. You can also specify a custom .env file path using the `--env-file` flag.

### Example .env file:

```bash
# Server Configuration
MATTER_SERVER_PORT=8080
MATTER_SERVER_LISTEN_ADDRESSES=127.0.0.1,::1

# Matter Configuration
MATTER_MATTER_VENDOR_ID=1234
MATTER_MATTER_FABRIC_ID=5

# Logging
MATTER_LOG_LEVEL=debug
MATTER_LOG_FORMAT=json
```

### Using a custom .env file:

```bash
# Load from custom file
./go-matter-server --env-file /path/to/custom.env

# Auto-load from .env in current directory
./go-matter-server
```

## Boolean Values

Boolean environment variables accept the following values:
- **True:** `true`, `1`, `yes`, `on`, `True`, `TRUE`
- **False:** `false`, `0`, `no`, `off`, `False`, `FALSE`

## Array Values

Array values (like `MATTER_SERVER_LISTEN_ADDRESSES`) should be comma-separated:

```bash
MATTER_SERVER_LISTEN_ADDRESSES=127.0.0.1,::1,192.168.1.100
```

## Examples

### Basic Development Setup

```bash
export MATTER_LOG_LEVEL=debug
export MATTER_SERVER_PORT=15580
export MATTER_STORAGE_PATH=./dev_storage
./go-matter-server
```

### Production Setup

```bash
export MATTER_LOG_FORMAT=json
export MATTER_LOG_LEVEL=info
export MATTER_SERVER_PORT=5580
export MATTER_SERVER_LISTEN_ADDRESSES=0.0.0.0
export MATTER_STORAGE_PATH=/var/lib/matter-server
./go-matter-server
```

### Using with systemd

Create a systemd environment file `/etc/matter-server/environment`:

```bash
MATTER_LOG_FORMAT=json
MATTER_LOG_LEVEL=info
MATTER_STORAGE_PATH=/var/lib/matter-server
MATTER_SERVER_PORT=5580
```

Then reference it in your systemd service:

```ini
[Service]
EnvironmentFile=/etc/matter-server/environment
ExecStart=/usr/local/bin/go-matter-server
```

## Validation

All environment variables are validated according to the same rules as CLI flags:
- Port must be between 1 and 65535
- Vendor ID must be between 0 and 65535 (0xFFFF)
- Fabric ID must be greater than 0
- Log level must be valid
- Log format must be `console` or `json`

Invalid values will cause the server to exit with an error message.
