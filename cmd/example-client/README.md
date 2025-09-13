# Matter Server Example Client

This program demonstrates how to discover and communicate with the go-matter-server using mDNS and WebSocket.

## Features

- **mDNS Discovery**: Automatically discovers the matter-server on the local network using multicast DNS
- **WebSocket Communication**: Connects to the matter-server's WebSocket API
- **Example Commands**: Sends various example commands to demonstrate the API
- **Event Handling**: Listens for server events and responses

## Usage

1. Start the matter-server:
   ```bash
   ./matter-server
   ```

2. In another terminal, run the example client:
   ```bash
   ./example-client
   ```

## How it Works

### mDNS Discovery
The client queries for `matter-server.local` on the mDNS multicast address (224.0.0.251:5353). If no response is received within 5 seconds, it falls back to connecting to localhost.

### WebSocket Connection
Once the server is discovered, the client connects to the WebSocket endpoint at `ws://<server-ip>:5580/ws`.

### Example Commands
The client sends several example commands:
- `server_info` - Get server information
- `diagnostics` - Get server diagnostics
- `get_nodes` - Get all Matter nodes
- `start_listening` - Start listening for events
- `discover` - Discover new devices

### Message Handling
The client handles three types of messages:
- **Result Messages**: Command responses with success/error status
- **Event Messages**: Server-sent events (device updates, etc.)
- **Raw Messages**: Other message types

## Building

```bash
go build -o example-client ./cmd/example-client
```

## Network Requirements

- The client needs to be on the same network as the matter-server for mDNS discovery
- UDP port 5353 must be accessible for mDNS multicast
- TCP port 5580 must be accessible for WebSocket connection

## Example Output

```
🚀 Matter Server Example Client
===============================
🔍 Discovering matter-server via mDNS...
📡 Sent mDNS query for matter-server.local...
✅ Found matter-server at 192.168.1.100 (from 192.168.1.100:5353)
🔌 Connecting to matter-server at ws://192.168.1.100:5580/ws
✅ Connected to matter-server WebSocket
📨 Raw message: {"fabric_id":1,"compressed_fabric_id":...}

📋 Sending example commands...

1. Get Server Info
📤 Sending command: server_info [a1b2c3d4-...]
✅ Command success [a1b2c3d4-...]: map[fabric_id:1 ...]

2. Get Server Diagnostics
📤 Sending command: diagnostics [e5f6g7h8-...]
✅ Command success [e5f6g7h8-...]: map[info:map[...] nodes:[] events:[]]

...
```