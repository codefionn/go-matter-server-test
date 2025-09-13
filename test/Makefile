# Container Orchestration Testing Makefile
# Automatically detects and uses podman-compose or docker-compose

# Include compose detection
SHELL := /bin/bash
include_compose = $(shell source ./detect-compose.sh >/dev/null 2>&1 && echo "COMPOSE_CMD=$$(get_compose_command)" && echo "COMPOSE_FILE=$$(get_compose_file)")
$(eval $(include_compose))

PROJECT_NAME = go-matter-server

# Fallback if detection fails
ifeq ($(COMPOSE_CMD),)
COMPOSE_CMD = docker-compose
endif
ifeq ($(COMPOSE_FILE),)
COMPOSE_FILE = docker-compose.yml
endif

.PHONY: help build up down logs clean test client server health

help:	## Show this help
	@echo "Go Matter Server Container Test Environment"
	@echo "=========================================="
	@echo "Using: $(COMPOSE_CMD)"
	@echo "Compose file: $(COMPOSE_FILE)"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build:	## Build container images
	$(COMPOSE_CMD) -f $(COMPOSE_FILE) build

server:	## Start matter-server
	$(COMPOSE_CMD) -f $(COMPOSE_FILE) up matter-server

server-bg:	## Start matter-server in background
	$(COMPOSE_CMD) -f $(COMPOSE_FILE) up -d matter-server

client:	## Run example client
	$(COMPOSE_CMD) -f $(COMPOSE_FILE) --profile client up example-client

client-bg:	## Run example client in background
	$(COMPOSE_CMD) -f $(COMPOSE_FILE) --profile client up -d example-client

test:	## Run integration tests
	$(COMPOSE_CMD) -f $(COMPOSE_FILE) --profile test up --abort-on-container-exit test-runner

test-all:	## Run all services including tests
	$(COMPOSE_CMD) -f $(COMPOSE_FILE) --profile client --profile test up --abort-on-container-exit

health:	## Check server health
	@echo "Checking matter-server health..."
	@$(COMPOSE_CMD) -f $(COMPOSE_FILE) exec matter-server wget -q --spider http://localhost:5580/health && echo "✅ Server is healthy" || echo "❌ Server is not healthy"

logs:	## Show logs from all services
	$(COMPOSE_CMD) -f $(COMPOSE_FILE) logs -f

logs-server:	## Show server logs
	$(COMPOSE_CMD) -f $(COMPOSE_FILE) logs -f matter-server

logs-client:	## Show client logs
	$(COMPOSE_CMD) -f $(COMPOSE_FILE) logs -f example-client

logs-test:	## Show test logs
	$(COMPOSE_CMD) -f $(COMPOSE_FILE) logs -f test-runner

shell-server:	## Open shell in server container
	$(COMPOSE_CMD) -f $(COMPOSE_FILE) exec matter-server sh

shell-client:	## Run client container with shell
	$(COMPOSE_CMD) -f $(COMPOSE_FILE) run --rm example-client sh

up:	## Start all services
	$(COMPOSE_CMD) -f $(COMPOSE_FILE) up

up-bg:	## Start all services in background
	$(COMPOSE_CMD) -f $(COMPOSE_FILE) up -d

down:	## Stop all services
	$(COMPOSE_CMD) -f $(COMPOSE_FILE) down

restart:	## Restart all services
	$(COMPOSE_CMD) -f $(COMPOSE_FILE) restart

clean:	## Clean up containers, volumes, and networks
	$(COMPOSE_CMD) -f $(COMPOSE_FILE) down -v --remove-orphans

clean-all:	## Clean up everything including images
	$(COMPOSE_CMD) -f $(COMPOSE_FILE) down --rmi all -v --remove-orphans

rebuild:	## Rebuild and restart
	$(MAKE) clean
	$(MAKE) build
	$(MAKE) up-bg

# Development targets
dev-start:	## Start development environment
	$(MAKE) server-bg
	@echo "Waiting for server to be ready..."
	@sleep 5
	$(MAKE) health

dev-test:	## Run development tests
	$(MAKE) dev-start
	$(MAKE) client
	$(MAKE) down

# CI/CD targets
ci-test:	## Run CI tests
	$(MAKE) build
	$(MAKE) test
	$(MAKE) clean

# Monitoring targets
ps:	## Show running containers
	docker-compose -f $(COMPOSE_FILE) ps

top:	## Show container resource usage
	docker-compose -f $(COMPOSE_FILE) top

# Quick commands
quick-test: build server-bg client down	## Quick test cycle