# Nix Flake Usage Guide

This project includes Nix flake configuration for reproducible development environments and builds.

## Prerequisites

- [Nix package manager](https://nixos.org/download.html) with flakes enabled
- Git (for tracking flake files)

### Enable Nix Flakes

Add to your `~/.config/nix/nix.conf` or `/etc/nix/nix.conf`:

```
experimental-features = nix-command flakes
```

## Quick Start

### Enter Development Shell

```bash
# Enter the default development shell
nix develop

# Or use a specific shell variant
nix develop .#minimal     # Minimal Go environment
```

### Build the Project

```bash
# Build the project
nix build

# Run the built binary
./result/bin/go-matter-server --help
```

### Run the Application

```bash
# Run directly without building
nix run
```

## Development Shells

### Default Shell (`nix develop`)

Includes:
- Go 1.24.1 (matching go.mod)
- Go language server (gopls)
- Go tools (gofmt, goimports, etc.)
- golangci-lint for linting
- Development utilities (git, make, jq, netcat, lsof)

### Minimal Shell (`nix develop .#minimal`)

Includes only:
- Go 1.24.1
- Git

## Using with direnv

If you have [direnv](https://direnv.net/) installed:

1. The `.envrc` file is already configured
2. Run `direnv allow` in the project directory
3. The development environment will be automatically loaded when you enter the directory

## Building and Testing

Once in the development shell:

```bash
# Run tests
go test ./...

# Build the binary
go build -o go-matter-server ./cmd/matter-server

# Run the server
go run ./cmd/matter-server --help

# Lint the code
golangci-lint run

# Use Makefile for common tasks
make help
```

## Project Structure

```
.
├── flake.nix          # Nix flake configuration
├── .envrc             # direnv configuration
├── go.mod             # Go module definition
├── go.sum             # Go module checksums
├── cmd/matter-server/  # Application entry point
│   └── main.go        # CLI and server initialization
├── internal/          # Internal packages
│   ├── config/        # Configuration management
│   ├── logger/        # Logging functionality
│   ├── models/        # Data models
│   ├── server/        # HTTP/WebSocket server
│   ├── storage/       # JSON storage backend
│   ├── websocket/     # WebSocket handler
│   └── mdns/          # mDNS implementation
└── NIX_USAGE.md       # This file
```

## Troubleshooting

### GitHub API Rate Limits

If you encounter GitHub API rate limits when running `nix flake check`, try:

1. Wait a few minutes for the rate limit to reset
2. Use a GitHub token for authentication
3. Use cached flakes with `nix develop --offline` (after first successful run)

### Vendor Hash Issues

If you encounter vendor hash mismatches when building:

1. The flake uses `lib.fakeHash` initially
2. Run `nix build` to get the correct hash
3. Replace `lib.fakeHash` with the correct hash in `flake.nix`

### Storage Directory

The application uses `.matter_server/` in the current directory for storage by default. This directory is automatically created by the development shell.

## Environment Variables

Available environment variables:
- `MATTER_LOG_LEVEL`: Set log level (debug, info, warn, error)
- `MATTER_SERVER_PORT`: Set server port (default: 5580)
- `PROJECT_NAME`: Set to "go-matter-server" by `.envrc`
- `MATTER_SERVER_DEV`: Set to 1 in development

## Integration with IDEs

### VS Code

For VS Code integration with the Nix environment:

1. Install the direnv VS Code extension
2. Ensure direnv is working (`direnv allow`)
3. The Go extension should automatically use the Go version from Nix

### Other Editors

Most editors that support direnv or can read environment variables will work automatically with the Nix development shell.