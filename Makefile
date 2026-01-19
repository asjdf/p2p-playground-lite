# P2P Playground Lite - Makefile
.PHONY: help build test lint clean install deps run-daemon run-controller

# Variables
BINARY_DIR=bin
CONTROLLER_BINARY=$(BINARY_DIR)/controller
DAEMON_BINARY=$(BINARY_DIR)/daemon
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags="-s -w"

# Default target
help:
	@echo "P2P Playground Lite - Available targets:"
	@echo "  make build          - Build both controller and daemon binaries"
	@echo "  make controller     - Build controller binary only"
	@echo "  make daemon         - Build daemon binary only"
	@echo "  make test           - Run all tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make lint           - Run linters (vet + fmt check)"
	@echo "  make fmt            - Format all Go code"
	@echo "  make clean          - Remove built binaries and caches"
	@echo "  make install        - Install binaries to GOPATH/bin"
	@echo "  make deps           - Download and verify dependencies"
	@echo "  make run-daemon     - Run daemon with example config"
	@echo "  make run-controller - Run controller with example config"

# Build both binaries
build: controller daemon
	@echo "✓ Build complete"

# Build controller
controller:
	@echo "Building controller..."
	@mkdir -p $(BINARY_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(CONTROLLER_BINARY) ./cmd/controller

# Build daemon
daemon:
	@echo "Building daemon..."
	@mkdir -p $(BINARY_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(DAEMON_BINARY) ./cmd/daemon

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -race -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Lint code
lint:
	@echo "Running linters..."
	$(GO) vet ./...
	@echo "Checking formatting..."
	@test -z "$$(gofmt -l .)" || (echo "Code needs formatting. Run 'make fmt'" && exit 1)
	@echo "✓ Lint passed"

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo "✓ Code formatted"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BINARY_DIR)
	rm -f coverage.out coverage.html
	$(GO) clean -cache -testcache
	@echo "✓ Clean complete"

# Install binaries
install:
	@echo "Installing binaries..."
	$(GO) install ./cmd/controller
	$(GO) install ./cmd/daemon
	@echo "✓ Installed to $(GOPATH)/bin"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod verify
	@echo "✓ Dependencies ready"

# Run daemon (for development)
run-daemon:
	@if [ ! -f configs/daemon.yaml ]; then \
		echo "Error: configs/daemon.yaml not found"; \
		exit 1; \
	fi
	$(GO) run ./cmd/daemon start --config configs/daemon.yaml

# Run controller (for development)
run-controller:
	$(GO) run ./cmd/controller

# Development: watch and rebuild (requires entr)
watch:
	@which entr > /dev/null || (echo "Install entr: brew install entr" && exit 1)
	find . -name '*.go' | entr -r make build

# Check if tools are installed
check-tools:
	@which go > /dev/null || (echo "Go is not installed" && exit 1)
	@echo "✓ Go version: $$(go version)"

# Docker commands
docker-build:
	@echo "Building Docker images..."
	docker-compose build

docker-up:
	@echo "Starting Docker containers..."
	docker-compose up -d
	@echo "✓ Containers started"
	@echo "Run 'make docker-logs' to view logs"

docker-down:
	@echo "Stopping Docker containers..."
	docker-compose down

docker-down-clean:
	@echo "Stopping containers and removing volumes..."
	docker-compose down -v

docker-logs:
	docker-compose logs -f

docker-logs-daemon1:
	docker-compose logs -f daemon1

docker-logs-daemon2:
	docker-compose logs -f daemon2

docker-logs-daemon3:
	docker-compose logs -f daemon3

docker-restart:
	docker-compose restart

docker-ps:
	docker-compose ps

docker-exec-daemon1:
	docker exec -it p2p-daemon1 sh

docker-exec-daemon2:
	docker exec -it p2p-daemon2 sh

docker-exec-daemon3:
	docker exec -it p2p-daemon3 sh

docker-exec-controller:
	docker exec -it p2p-controller sh

# Package example app
package-example:
	@echo "Packaging hello-world example..."
	cd examples/hello-world && ./build.sh && tar -czf hello-world-1.0.0.tar.gz manifest.yaml bin/
	@echo "✓ Package created: examples/hello-world/hello-world-1.0.0.tar.gz"

# Full test workflow
docker-test: docker-build package-example docker-up
	@echo ""
	@echo "✓ Test environment ready!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Check logs: make docker-logs"
	@echo "  2. View daemon1: make docker-logs-daemon1"
	@echo "  3. Enter controller: make docker-exec-controller"
	@echo "  4. Clean up: make docker-down-clean"
