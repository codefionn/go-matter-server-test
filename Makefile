# Go Matter Server Makefile

.PHONY: help build test clean run dev-shell nix-build nix-run format lint install

# Default target
help: ## Show this help message
	@echo "Go Matter Server - Available make targets:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Go targets
build: ## Build the Go binary
	go build -o go-matter-server ./cmd/matter-server

test: ## Run all tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	rm -f go-matter-server matter-server coverage.out coverage.html
	rm -rf result result-*

run: ## Run the application
	go run ./cmd/matter-server --help

format: ## Format Go code
	go fmt ./...
	goimports -w .

lint: ## Run linter
	golangci-lint run

# Nix targets
nix-build: ## Build using Nix
	nix build

nix-run: ## Run using Nix
	nix run

nix-dev: ## Enter Nix development shell
	nix develop

nix-dev-minimal: ## Enter minimal Nix development shell
	nix develop .#minimal

nix-format: ## Format Nix files
	nix fmt

# Development targets
dev-shell: ## Enter development environment (same as nix-dev)
	nix develop

install: ## Install dependencies and setup development environment
	@echo "Setting up Go Matter Server development environment..."
	@if command -v nix >/dev/null 2>&1; then \
		echo "✓ Nix found - using Nix development environment"; \
		echo "Run 'make nix-dev' to enter the development shell"; \
	else \
		echo "✗ Nix not found - using local Go installation"; \
		echo "Please ensure Go 1.24+ is installed"; \
	fi
	@if [ -f .envrc ] && command -v direnv >/dev/null 2>&1; then \
		echo "✓ direnv found - run 'direnv allow' to auto-enter dev environment"; \
	fi
	go mod download
	@echo "✓ Dependencies downloaded"

# Git and project management
git-setup: ## Initialize git repository with initial commit
	git add .
	git commit -m "Initial commit: Go Matter Server with Nix flake support"

# Documentation
docs: ## Open documentation
	@echo "Documentation files:"
	@echo "  README.md         - Main project documentation"
	@echo "  NIX_USAGE.md      - Nix flake usage guide"
	@echo "  internal/         - Go package documentation"

# Advanced Nix operations
nix-update: ## Update Nix flake inputs
	nix flake update

nix-check: ## Check Nix flake configuration
	nix flake check

# Server operations
server-dev: ## Run server in development mode
	MATTER_LOG_LEVEL=debug go run ./cmd/matter-server

server-test: ## Run server with test configuration
	go run ./cmd/matter-server --log-level debug --port 15580

# Testing variations
test-unit: ## Run only unit tests
	go test -short ./...

test-integration: ## Run integration tests
	go test -run TestE2E ./...

test-verbose: ## Run tests with verbose output
	go test -v -race ./...

# Build variations
build-release: ## Build optimized release binary
	CGO_ENABLED=0 go build -ldflags="-s -w" -o go-matter-server ./cmd/matter-server

build-debug: ## Build binary with debug symbols
	go build -gcflags="-N -l" -o go-matter-server-debug ./cmd/matter-server

# Platform-specific builds
build-linux: ## Build for Linux
	GOOS=linux GOARCH=amd64 go build -o go-matter-server-linux ./cmd/matter-server

build-macos: ## Build for macOS
	GOOS=darwin GOARCH=amd64 go build -o go-matter-server-macos ./cmd/matter-server

build-windows: ## Build for Windows
	GOOS=windows GOARCH=amd64 go build -o go-matter-server.exe ./cmd/matter-server