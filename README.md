# Go Matter Server

> This project is actively in initial development. Use at your own risk!

A Go implementation of the Python Matter Server, providing a WebSocket-based Matter controller server with mDNS service discovery and JSON storage backend.

## Default Ports

- 5580: WebSocket
- 5353: mDNS (so no avahi-daemon for you!)

## Features

- **WebSocket API**: Full-featured WebSocket server for real-time communication
- **Matter Protocol Support**: Complete Matter device controller implementation
- **mDNS Service Discovery**: Built-in multicast DNS server for device discovery
- **JSON Storage Backend**: Persistent storage using JSON files
- **RESTful HTTP API**: HTTP endpoints for basic operations
- **Event System**: Real-time event broadcasting to connected clients
- **Configuration Management**: Flexible configuration via files, environment variables, and command line flags

## Bluetooth (BlueZ)

- Backend: BlueZ D-Bus via `github.com/godbus/dbus/v5` (no `go-bluetooth`).
- Docs: BlueZ D-Bus API documentation lives under the BlueZ repo `doc` directory:
  https://git.kernel.org/pub/scm/bluetooth/bluez.git/tree/doc
- Interfaces used: `org.bluez.Adapter1`, `org.bluez.Device1`, with `org.freedesktop.DBus.ObjectManager` and `org.freedesktop.DBus.Properties`.
- Requirements: System D-Bus and `bluetoothd` running with at least one adapter (e.g., `hci0`).
- API surface: Internal-only (mirrors python-matter-server)
- Enabling: Set `--bluetooth-adapter <id>` (or `bluetooth.adapter_id >= 0` in config). If not set (`-1`), Bluetooth is disabled.
- Availability flag: `server_info.bluetooth_enabled` reflects actual availability (adapter + BlueZ/DBus present), not just configuration.

## Architecture

This implementation mirrors the Python Matter Server architecture:

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   WebSocket     │    │   HTTP Server    │    │   mDNS Server   │
│   Handler       │◄──►│   (REST API)     │◄──►│   Discovery     │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Matter        │    │   JSON Storage   │    │   Event System  │
│   Controller    │◄──►│   Backend        │◄──►│   (Pub/Sub)     │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## Installation

### Prerequisites

- Go 1.22 or later
- Git

### Building from Source

```bash
git clone https://github.com/codefionn/go-matter-server.git
cd go-matter-server
go mod tidy
go build -o matter-server ./cmd/matter-server
```

## Usage

### Basic Usage

Start the server with default settings:

```bash
./matter-server
```

### Command Line Options

```bash
# Start server on custom port
./matter-server --port 8080

# Enable debug logging
./matter-server --log-level debug

# Use custom storage path  
./matter-server --storage-path /var/lib/matter-server

# Use configuration file
./matter-server --config /etc/matter-server/config.yaml
```

### Configuration

The server can be configured via:

1. **Configuration file** (YAML format)
2. **Environment variables** (prefixed with `MATTER_`)
3. **Command line flags**

Configuration precedence: CLI flags > Environment variables > Config file > Defaults

#### Example Configuration File

```yaml
server:
  port: 5580
  listen_addresses: ["127.0.0.1", "::1"]
  
storage:
  path: "/var/lib/matter-server"
  
matter:
  vendor_id: 0xFFF1
  fabric_id: 1
  
log:
  level: "info"
  format: "json"
```

#### Environment Variables

```bash
export MATTER_SERVER_PORT=5580
export MATTER_LOG_LEVEL=debug  
export MATTER_STORAGE_PATH=/var/lib/matter-server
```

## API Reference

### WebSocket API

Connect to `ws://localhost:5580/ws` for real-time communication.

#### Message Format

**Command Message:**
```json
{
  "message_id": "uuid-string",
  "command": "command_name",
  "args": {
    "param1": "value1",
    "param2": "value2"
  }
}
```

**Success Response:**
```json
{
  "message_id": "uuid-string", 
  "result": {
    "data": "response data"
  }
}
```

**Error Response:**
```json
{
  "message_id": "uuid-string",
  "error_code": 500,
  "details": "Error description"
}
```

**Event Message:**
```json
{
  "event": "node_added",
  "data": {
    "node_id": 123,
    "available": true
  }
}
```

#### Available Commands

- `server_info` - Get server information
- `get_nodes` - Get all Matter nodes
- `get_node` - Get specific node by ID
- `ping_node` - Ping a Matter node
- `diagnostics` - Get server diagnostics
- `start_listening` - Start receiving events

### HTTP API

#### Endpoints

- `GET /api/info` - Server information
- `GET /api/nodes` - List all nodes
- `GET /api/diagnostics` - Server diagnostics  
- `GET /health` - Health check

#### Example Response

```bash
curl http://localhost:5580/api/info
```

```json
{
  "fabric_id": 1,
  "compressed_fabric_id": 1,
  "schema_version": 11,
  "min_supported_schema_version": 1,
  "sdk_version": "go-matter-server-1.0.0",
  "wifi_credentials_set": false,
  "thread_credentials_set": false,
  "bluetooth_enabled": false
}
```

## Development

### Project Structure

```
├── cmd/matter-server/          # Main application entry point
├── internal/
│   ├── config/                 # Configuration management
│   ├── mdns/                   # mDNS service discovery
│   ├── models/                 # Data models and types
│   ├── server/                 # Main server implementation
│   ├── storage/                # JSON storage backend
│   └── websocket/              # WebSocket handler
├── config.example.yaml         # Example configuration
├── go.mod                      # Go module definition
└── README.md                   # This file
```

### Running Tests

```bash
go test ./...
```

### Building for Different Platforms

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o matter-server-linux ./cmd/matter-server

# Windows
GOOS=windows GOARCH=amd64 go build -o matter-server.exe ./cmd/matter-server

# macOS
GOOS=darwin GOARCH=amd64 go build -o matter-server-macos ./cmd/matter-server
```

## Storage

The server uses JSON files for persistent storage:

- `nodes.json` - Matter node data
- `vendors.json` - Vendor information cache
- `settings.json` - Server settings

Storage files are located in:
- Linux/macOS: `$HOME/.matter_server/`
- Windows: `%USERPROFILE%\.matter_server\`

## Logging

The server supports structured logging with configurable levels:

- `debug` - Detailed debugging information
- `info` - General information (default)
- `warn` - Warning messages  
- `error` - Error messages only

Log formats:
- `console` - Human-readable format (default)
- `json` - Structured JSON format

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Based on the [Python Matter Server](https://github.com/home-assistant-libs/python-matter-server) project
- Uses the [Gorilla WebSocket](https://github.com/gorilla/websocket) library
- Built with [Cobra CLI](https://github.com/spf13/cobra) and [Viper](https://github.com/spf13/viper) configuration
