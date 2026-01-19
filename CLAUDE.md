# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

P2P Playground Lite is a distributed application testing and deployment tool that enables automated distribution of Go applications across P2P networks.

**Module:** `github.com/asjdf/p2p-playground-lite`
**Go Version:** 1.24.2

### Core Components

1. **Controller**: Initiates application distribution and manages nodes (CLI + optional Web/API)
2. **Node Daemon**: Runs on test nodes to receive, deploy, and manage applications

### Key Features

- P2P-based application distribution using libp2p
- Automated process lifecycle management (start/stop/restart)
- Multi-version application support with rollback capability
- Health checking and auto-restart
- Resource limiting (CPU/memory)
- Log collection and forwarding
- Security: node authentication, file integrity checking, code signing, encrypted transmission

## Development Commands

### Building
```bash
go build ./...
```

### Running Tests
```bash
go test ./...
```

### Running a Single Test
```bash
go test -run TestName ./path/to/package
```

### Formatting Code
```bash
go fmt ./...
```

### Linting
```bash
go vet ./...
```

## Architecture

### P2P Network
- **Library**: go-libp2p
- **Node Discovery**: mDNS (local network broadcast) + manual configuration
- **Transport**: TCP + QUIC with TLS 1.3
- **Use Case**: Hybrid scenarios (LAN + WAN)

### Application Package Format
Applications are distributed as `tar.gz` packages containing:
- `manifest.yaml`: Application metadata (name, version, entrypoint, resources, health check)
- `bin/`: Executable files
- `config/`: Configuration files
- `resources/`: Additional resources
- `scripts/`: Lifecycle hooks (pre-start, post-stop)
- `app-signature`: Code signature for verification

### Security Model
1. **Node Authentication**: PSK or certificate-based, trusted node whitelist
2. **File Integrity**: SHA-256 checksums validated on transfer
3. **Code Signing**: GPG/Ed25519 signatures verified before execution
4. **Transport Encryption**: libp2p TLS 1.3

### Version Management
- Semantic versioning (semver)
- Multi-version coexistence on same node
- Automatic updates with configurable strategies (immediate/graceful/manual)
- Version rollback on health check failure
- Version tags (latest, stable, dev)

## Project Structure

```
p2p-playground-lite/
├── cmd/
│   ├── controller/      # Controller CLI and server
│   └── daemon/          # Node daemon
├── pkg/
│   ├── p2p/             # P2P network layer (libp2p wrapper)
│   ├── package/         # Application packaging and manifest
│   ├── runtime/         # Process management and lifecycle
│   ├── security/        # Authentication, signing, verification
│   ├── api/             # gRPC and HTTP API definitions
│   └── version/         # Version management
├── api/
│   └── proto/           # Protobuf definitions
├── web/                 # Web dashboard (full version)
├── docs/
│   └── DESIGN.md        # Detailed product design document
└── configs/             # Example configurations
```

## Design Decisions

### Two Editions
1. **Lite Edition**: CLI only, minimal dependencies, suitable for CI/CD
2. **Full Edition**: CLI + gRPC API + HTTP API + Web Dashboard

### Node Grouping
Nodes can be labeled (e.g., `env=test`, `region=cn-north`) for selective deployment and batch operations.

### Log Management
- Local log storage with rotation
- Real-time log streaming via gRPC
- Optional aggregated viewing across nodes
- Log retention policies by time/size

## Development Roadmap

See `docs/DESIGN.md` for detailed phases. High-level stages:
1. **Phase 1 (MVP)**: P2P networking, mDNS discovery, basic packaging, file transfer, process management, CLI
2. **Phase 2**: Security (authentication, signing), health checks, resource limits, logging
3. **Phase 3**: Version management, multi-version support, auto-updates, node labels
4. **Phase 4**: gRPC/HTTP APIs, Web Dashboard
5. **Phase 5**: Production optimization, monitoring, documentation
