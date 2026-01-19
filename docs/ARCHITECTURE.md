# P2P Playground Lite - Architecture & Implementation Plan

## Overview

This document outlines the architecture and implementation strategy for P2P Playground Lite, a distributed application testing and deployment tool that enables automated distribution of Go applications across P2P networks.

**Technology Stack**: Go 1.24.2, libp2p v0.46.0, cobra CLI framework
**Implementation Scope**: Phase 1 (MVP) + Phase 2 (Security) + Phase 3 (Advanced Features)
**Last Updated**: 2026-01-19

## Implementation Status

### ✓ Completed

**Phase 1A: Foundation** (100%)
- ✓ Updated go.mod with all dependencies (libp2p v0.46.0, cobra, viper, zap)
- ✓ Created Makefile with build, test, Docker targets
- ✓ Implemented pkg/types/ (interfaces.go, models.go, errors.go)
- ✓ All unit tests passing

**Phase 1B: Core Services** (100%)
- ✓ Implemented pkg/config/ - YAML configuration with viper (with mapstructure tags)
- ✓ Implemented pkg/logging/ - Structured logging with zap
- ✓ Implemented pkg/package/ - Manifest parsing, tar.gz handling
- ✓ Implemented pkg/p2p/ - libp2p host wrapper with mDNS discovery
- ✓ Implemented pkg/transfer/ - File transfer protocol
- ✓ Implemented pkg/storage/ - Filesystem-based storage

**Phase 1C: Process Management** (100%)
- ✓ Implemented pkg/runtime/ - Process start/stop/restart with log capture
- ✓ Implemented pkg/daemon/ - Daemon orchestration
- ✓ All core functionality working

**Phase 1D: CLI Implementation** (100%)
- ✓ Implemented cmd/daemon/ - Daemon CLI (start command fully functional)
- ✓ Implemented cmd/controller/ - Controller CLI framework
- ✓ Implemented controller nodes command (P2P node discovery)
- ✓ Implemented controller deploy command (with progress tracking)
- ✓ Implemented controller list command (application listing)
- ✓ Implemented controller logs command (with --tail and --follow support)
- ✓ Created example configurations
- ✓ Both binaries build successfully (daemon: 35MB, controller: 5.3MB)

**Testing Infrastructure** (100%)
- ✓ Created Docker multi-node environment (3 daemons + 1 controller)
- ✓ Created example hello-world application
- ✓ Verified P2P network formation (all 3 nodes discovering each other)
- ✓ Controller successfully discovers all daemon nodes via mDNS

### 🚧 In Progress

**Phase 1D: Verification and Documentation**
- ⏳ Complete end-to-end testing documentation
- ⏳ Prepare for Phase 2 (Security)

**Phase 2: Security** (70%)
- ✓ Ed25519 key generation and management
- ✓ Package signing with controller sign command
- ✓ Signature verification in daemon
- ✓ Configurable security policies
- ✓ Multi-key trust support
- ⏸️ PSK (Pre-Shared Key) authentication for P2P network
- ⏸️ TLS 1.3 transport encryption (libp2p built-in, needs configuration)

### ⏸️ Not Started

**Phase 3A: Health & Resources** (0%)
- ⏸️ HTTP/TCP/process health checks
- ⏸️ cgroups resource limiting
- ⏸️ Auto-restart on failure

**Phase 3B: Version Management** (0%)
- ⏸️ Semver parsing and comparison
- ⏸️ Node labeling and filtering
- ⏸️ Multi-version storage
- ⏸️ Auto-rollback on failure

### Current State Summary

**Working Features:**
- P2P networking with libp2p v0.46.0
- mDNS automatic node discovery
- Daemon process management with log capture
- Application deployment via P2P (with progress tracking)
- Application listing across nodes
- Log streaming from deployed applications
- **Ed25519 package signing and verification**
- **Configurable security policies**
- **Multi-key trust model**
- Configuration system
- Logging infrastructure
- Complete CLI tooling (deploy, list, logs, nodes, keygen, sign)

**Network Topology (Docker):**
```
daemon1 (172.20.0.3:9010) ←→ daemon2 (172.20.0.5:9012)
         ↕                              ↕
       daemon3 (172.20.0.2:9013)

All nodes discovered via mDNS
Controller can discover all 3 daemons
```

**Next Steps:**
1. ✓ Phase 1 (MVP) Complete - All core functionality working
2. Document end-to-end testing procedures
3. Begin Phase 2 (Security) - Ed25519 signing, PSK authentication
4. Implement Phase 3A (Health & Resources) - Health checks, resource limits

## Project Structure

```
p2p-playground-lite/
├── cmd/
│   ├── controller/              # Controller CLI
│   │   ├── main.go
│   │   └── commands/           # Cobra commands
│   │       ├── root.go
│   │       ├── deploy.go
│   │       ├── app.go
│   │       ├── node.go
│   │       ├── logs.go
│   │       └── version.go
│   └── daemon/                  # Node daemon
│       ├── main.go
│       └── commands/
│           ├── root.go
│           ├── start.go
│           └── status.go
│
├── pkg/                         # Public packages
│   ├── types/                   # Core interfaces and models
│   ├── config/                  # Configuration management
│   ├── logging/                 # Logging infrastructure
│   ├── package/                 # Package management
│   ├── security/                # Security layer
│   ├── version/                 # Version management
│   ├── labels/                  # Node labeling
│   ├── p2p/                     # P2P networking
│   ├── transfer/                # File transfer
│   ├── runtime/                 # Process management
│   ├── health/                  # Health checking
│   ├── resources/               # Resource limiting
│   ├── storage/                 # Storage management
│   ├── daemon/                  # Daemon orchestration
│   ├── controller/              # Controller orchestration
│   └── updater/                 # Auto-update
│
├── internal/                    # Private utilities
│   ├── testutil/                # Testing utilities
│   └── util/                    # Common utilities
│
├── configs/                     # Example configurations
│   ├── controller.example.yaml
│   └── daemon.example.yaml
│
└── docs/
    ├── DESIGN.md               # Product design
    ├── ARCHITECTURE.md         # This file
    └── examples/               # Usage examples
```

## Package Dependency Graph

Implementation should follow this dependency order (lower levels first):

**Level 1 (Foundation - No dependencies)**:
- `pkg/types/` - Core interfaces and domain models
- `internal/util/` - Common utilities

**Level 2 (Configuration & Logging)**:
- `pkg/config/` - Depends on: types
- `pkg/logging/` - Depends on: types, config

**Level 3 (Core Functionality)**:
- `pkg/package/` - Depends on: types, config, logging
- `pkg/security/` - Depends on: types, config, logging
- `pkg/storage/` - Depends on: types, config, logging
- `pkg/version/` - Depends on: types, config, logging
- `pkg/labels/` - Depends on: types

**Level 4 (Services)**:
- `pkg/p2p/` - Depends on: types, config, logging, security
- `pkg/transfer/` - Depends on: types, p2p, package, security
- `pkg/runtime/` - Depends on: types, config, logging, package
- `pkg/health/` - Depends on: types, config, logging, runtime
- `pkg/resources/` - Depends on: types, config, logging

**Level 5 (Orchestration)**:
- `pkg/daemon/` - Depends on: all level 3-4 packages
- `pkg/controller/` - Depends on: all level 3-4 packages
- `pkg/updater/` - Depends on: version, transfer, runtime

**Level 6 (Entry Points)**:
- `cmd/daemon/` - Depends on: daemon orchestrator
- `cmd/controller/` - Depends on: controller orchestrator

## Core Interfaces

All major interfaces are defined in `pkg/types/interfaces.go` to enable parallel development:

### P2P Layer
```go
type Host interface {
    ID() string
    Addrs() []string
    Connect(ctx context.Context, addr string) error
    NewStream(ctx context.Context, peerID string, protocol string) (Stream, error)
    SetStreamHandler(protocol string, handler StreamHandler)
    Close() error
}

type Stream interface {
    Read(p []byte) (n int, err error)
    Write(p []byte) (n int, err error)
    Close() error
}
```

### Runtime Layer
```go
type Runtime interface {
    Start(ctx context.Context, app *Application) error
    Stop(ctx context.Context, appID string) error
    Restart(ctx context.Context, appID string) error
    Status(ctx context.Context, appID string) (*AppStatus, error)
    Logs(ctx context.Context, appID string, follow bool) (io.ReadCloser, error)
}

type HealthChecker interface {
    Check(ctx context.Context, app *Application) error
}

type ResourceLimiter interface {
    Apply(ctx context.Context, pid int, limits *ResourceLimits) error
    Release(ctx context.Context, pid int) error
}
```

### Package Layer
```go
type PackageManager interface {
    Pack(ctx context.Context, appDir string) (string, error)
    Unpack(ctx context.Context, pkgPath string, destDir string) (*Manifest, error)
    Verify(ctx context.Context, pkgPath string, signature []byte) error
}

type Manifest struct {
    Name        string
    Version     string
    Entrypoint  string
    Args        []string
    Env         map[string]string
    Resources   *ResourceLimits
    HealthCheck *HealthCheckConfig
}
```

### Security Layer
```go
type Signer interface {
    Sign(data []byte) ([]byte, error)
    Verify(data []byte, signature []byte, publicKey []byte) error
}

type Authenticator interface {
    Authenticate(ctx context.Context, peerID string) error
}
```

### Storage Layer
```go
type Storage interface {
    Save(ctx context.Context, key string, data []byte) error
    Load(ctx context.Context, key string) ([]byte, error)
    Delete(ctx context.Context, key string) error
    List(ctx context.Context, prefix string) ([]string, error)
}
```

## Domain Models

Key domain models defined in `pkg/types/models.go`:

```go
type Application struct {
    ID           string
    Name         string
    Version      string
    PackagePath  string
    Manifest     *Manifest
    Status       AppStatus
    PID          int
    StartedAt    time.Time
    Labels       map[string]string
}

type AppStatus string
const (
    AppStatusStopped  AppStatus = "stopped"
    AppStatusStarting AppStatus = "starting"
    AppStatusRunning  AppStatus = "running"
    AppStatusFailed   AppStatus = "failed"
)

type NodeInfo struct {
    ID       string
    Addrs    []string
    Labels   map[string]string
    Apps     []*Application
}

type ResourceLimits struct {
    CPUPercent float64
    MemoryMB   int64
}

type HealthCheckConfig struct {
    Type        string // "http", "tcp", "process"
    Endpoint    string
    Interval    time.Duration
    Timeout     time.Duration
    Retries     int
}
```

## Implementation Phases

### Phase 1A: Foundation (Steps 1-3)
**Goal**: Set up project structure and core interfaces

**Tasks**:
1. Update `go.mod` with all required dependencies
2. Create `Makefile` for common tasks (build, test, lint)
3. Implement `pkg/types/` (interfaces.go, models.go, errors.go)
4. Implement `internal/util/` (common utilities)

**Acceptance Criteria**:
- All packages can be imported without circular dependencies
- Unit tests pass for types package

### Phase 1B: Core Services (Steps 4-6)
**Goal**: Implement package management, P2P networking, and file transfer

**Tasks**:
1. Implement `pkg/config/` - YAML configuration with viper
2. Implement `pkg/logging/` - Structured logging with zap
3. Implement `pkg/package/` - Manifest parsing, tar.gz handling
4. Implement `pkg/p2p/` - libp2p host wrapper, mDNS discovery
5. Implement `pkg/transfer/` - File transfer protocol
6. Implement `pkg/storage/` - Filesystem-based storage

**Acceptance Criteria**:
- Can create and parse application packages
- Can establish P2P connections between two nodes
- Can transfer files over P2P network
- Unit tests pass for all packages

### Phase 1C: Process Management (Step 7)
**Goal**: Implement application lifecycle management

**Tasks**:
1. Implement `pkg/runtime/` - Process start/stop/restart
2. Implement log capture from stdout/stderr
3. Implement `pkg/daemon/` - Orchestrate runtime + storage
4. Implement `pkg/controller/` - Orchestrate P2P + transfer

**Acceptance Criteria**:
- Can start/stop/restart applications
- Can capture and retrieve application logs
- Unit tests pass

### Phase 1D: CLI Implementation (Step 8)
**Goal**: Build working controller and daemon binaries

**Tasks**:
1. Implement `cmd/daemon/` - Daemon CLI with cobra
2. Implement `cmd/controller/` - Controller CLI with cobra
3. Create example configurations in `configs/`
4. Wire all packages together

**Acceptance Criteria**:
- Both binaries build successfully
- Can start daemon and deploy app from controller
- End-to-end manual test passes

### Phase 2: Security (Step 9)
**Goal**: Add authentication and code signing

**Tasks**:
1. Implement `pkg/security/` - Ed25519 signing/verification
2. Add PSK authentication to P2P layer
3. Integrate signature verification into package deployment
4. Add TLS 1.3 to libp2p transports

**Acceptance Criteria**:
- Unsigned packages are rejected
- Unauthenticated nodes cannot connect
- All security tests pass

### Phase 3A: Health & Resources (Step 10)
**Goal**: Add health checking and resource limiting

**Tasks**:
1. Implement `pkg/health/` - HTTP/TCP/process health checks
2. Implement `pkg/resources/` - cgroups on Linux
3. Integrate auto-restart on health check failure
4. Add resource limits to runtime

**Acceptance Criteria**:
- Unhealthy apps are automatically restarted
- Apps respect CPU/memory limits (on Linux)
- Health check tests pass

### Phase 3B: Version Management (Step 11)
**Goal**: Support multiple versions and auto-updates

**Tasks**:
1. Implement `pkg/version/` - Semver parsing and comparison
2. Implement `pkg/labels/` - Node labeling and filtering
3. Implement `pkg/updater/` - Auto-update strategies
4. Add multi-version storage to daemon
5. Add version rollback on health check failure

**Acceptance Criteria**:
- Can deploy multiple versions of same app
- Can switch between versions
- Auto-rollback works on deployment failure
- Version management tests pass

## Key Dependencies

```go
require (
    // P2P
    github.com/libp2p/go-libp2p v0.33.0
    github.com/multiformats/go-multiaddr v0.12.2

    // CLI
    github.com/spf13/cobra v1.8.0
    github.com/spf13/viper v1.18.2

    // Config & Logging
    gopkg.in/yaml.v3 v3.0.1
    go.uber.org/zap v1.26.0

    // Security
    golang.org/x/crypto v0.18.0

    // Versioning
    github.com/Masterminds/semver/v3 v3.2.1

    // Testing
    github.com/stretchr/testify v1.8.4
)
```

## Architectural Decisions

### 1. Storage: Filesystem-based
**Rationale**: Simple, inspectable, no external dependencies
**Trade-off**: Not suitable for distributed controller (out of scope for lite edition)

**Structure**:
```
~/.p2p-playground/
├── packages/           # Stored application packages
│   └── {name}/
│       └── {version}/
│           ├── package.tar.gz
│           └── manifest.yaml
├── apps/              # Running applications
│   └── {app-id}/
│       ├── bin/
│       ├── config/
│       └── logs/
└── keys/              # Security keys
    ├── node.key
    └── node.pub
```

### 2. Configuration: YAML with Viper
**Rationale**: Standard Go approach, supports env vars and flags
**Files**: `controller.yaml`, `daemon.yaml`

### 3. Logging: Structured with Zap
**Rationale**: High performance, structured logging standard
**Levels**: DEBUG, INFO, WARN, ERROR

### 4. P2P Protocol: Custom Length-Prefixed
**Rationale**: Simple, efficient for large file transfers
**Format**: `[4-byte length][payload]`

### 5. Health Checks: Pull-based
**Rationale**: Daemon checks its own apps, simpler than push
**Types**: HTTP (GET endpoint), TCP (port open), Process (pid alive)

### 6. Resource Limits: cgroups on Linux
**Rationale**: Native OS support, no-op on other platforms
**Fallback**: Best-effort limits or no-op

### 7. Concurrency: Context-based
**Rationale**: Standard Go pattern for cancellation and timeouts
**Pattern**: All long-running operations accept `context.Context`

### 8. Error Handling: Wrapped Errors
**Rationale**: Preserve error chain with context
**Pattern**: `fmt.Errorf("operation failed: %w", err)`

## Testing Strategy

### Unit Tests
- Location: `pkg/*/\*_test.go`
- Coverage: >80% for core packages
- Mocking: Use interfaces, no external dependencies

### Integration Tests
- Location: `test/integration/`
- Coverage: End-to-end workflows
- Setup: In-memory libp2p transports

### Test Utilities
- Location: `internal/testutil/`
- Utilities: Mock implementations, test fixtures

### Testing Approach per Package
- `pkg/package/`: Test manifest parsing, tar.gz creation
- `pkg/p2p/`: Use libp2p in-memory transports
- `pkg/runtime/`: Test process lifecycle with dummy apps
- `pkg/security/`: Test signing/verification with test keys
- `pkg/version/`: Test semver parsing and comparison

## Verification Plan

After implementation, verify with this end-to-end test:

1. **Setup**:
   ```bash
   # Build binaries
   make build

   # Create test app
   echo 'package main; import "fmt"; func main() { fmt.Println("Hello P2P") }' > test.go
   go build -o testapp test.go
   ```

2. **Start Daemon**:
   ```bash
   ./bin/daemon start --config configs/daemon.yaml
   ```

3. **Package Application**:
   ```bash
   ./bin/controller package --app testapp --name hello --version 1.0.0
   ```

4. **Deploy Application**:
   ```bash
   ./bin/controller deploy --package hello-1.0.0.tar.gz --node <node-id>
   ```

5. **Verify**:
   ```bash
   # Check status
   ./bin/controller app status hello

   # View logs
   ./bin/controller logs hello

   # Stop app
   ./bin/controller app stop hello

   # Restart app
   ./bin/controller app restart hello
   ```

6. **Test Multi-Version**:
   ```bash
   # Deploy v2.0.0
   ./bin/controller deploy --package hello-2.0.0.tar.gz --node <node-id>

   # Switch versions
   ./bin/controller app switch hello --version 1.0.0
   ```

7. **Test Security**:
   ```bash
   # Generate keys
   ./bin/controller keygen

   # Sign package
   ./bin/controller sign --package hello-1.0.0.tar.gz

   # Deploy signed package (should verify)
   ./bin/controller deploy --package hello-1.0.0.tar.gz --node <node-id>
   ```

## Performance Considerations

### File Transfer
- Chunk size: 64KB (balance between memory and throughput)
- Progress tracking: Every 1MB transferred
- Timeout: 30s per chunk

### Health Checks
- Default interval: 30s
- Default timeout: 5s
- Max retries: 3

### Resource Limits
- CPU: Percentage of single core (e.g., 50% = 0.5 cores)
- Memory: Hard limit in MB

### Storage
- Log rotation: 10MB per file, keep 5 files
- Package cleanup: Keep last 3 versions per app

## Security Design Decisions

### Phase 2.1 Implementation (Current)

**Decision**: Package signature verification is **optional by default**

**Implementation**:
- `security.require_signed_packages` defaults to `false`
- When signature is present, it is always verified
- When signature is absent:
  - If `require_signed_packages: true` → deployment rejected
  - If `require_signed_packages: false` → warning logged, deployment allowed

**Rationale**:
1. **Backward Compatibility** - Existing deployments continue to work
2. **Gradual Adoption** - Users can test signature functionality first
3. **Flexibility** - Suitable for different environments (dev vs prod)

**Future Path**:
- Phase 2.2: Consider changing default to `true` based on user feedback
- Phase 3: Add fine-grained policies (per-app, per-label, etc.)

### Signature Algorithm

**Choice**: Ed25519

**Rationale**:
- High performance (fast signing and verification)
- Small key sizes (32B public, 64B private)
- Strong security (128-bit security level)
- Native Go standard library support
- Modern cryptographic standard

### Multi-Key Trust Model

**Implementation**: daemon accepts packages signed by any trusted public key

**Use Cases**:
1. Key rotation - trust both old and new keys during transition
2. Multi-team deployments - each team has their own signing key
3. Multi-environment - different environments can use different keys

**Security**: Any compromise of a single trusted key requires key revocation (future feature)

## Security Considerations

### Authentication
- PSK (Pre-Shared Key) for simple deployments (not yet implemented)
- Node whitelist (peer IDs in config)

### Code Signing
- Ed25519 signatures (fast, secure)
- Package signing via `controller sign` command
- Signature verification with configurable enforcement
- Public key distribution (manual - copy `.pub` files to daemon nodes)

### Transport Security
- libp2p TLS 1.3 by default
- No custom crypto implementations

### File Integrity
- SHA-256 checksums for all transfers
- Verify before unpacking

## Future Considerations (Out of Scope for Lite)

- gRPC/HTTP APIs (Full edition)
- Web Dashboard (Full edition)
- Distributed controller (multiple instances)
- Log aggregation service
- Metrics collection (Prometheus)
- Container support (Docker/Podman)
- Kubernetes operator

## References

- Product Design: [DESIGN.md](DESIGN.md)
- libp2p Documentation: https://docs.libp2p.io/
- Cobra CLI: https://github.com/spf13/cobra
- Go Best Practices: https://go.dev/doc/effective_go

---

**Status**: Implementation in progress
**Last Updated**: 2026-01-19
