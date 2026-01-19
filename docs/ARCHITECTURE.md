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

**Phase 2: Security & Stability** (85%)
- ✓ Ed25519 key generation and management
- ✓ Package signing with controller sign command
- ✓ Signature verification in daemon
- ✓ **Controller configuration file support** (kubectl-style --config flag)
- ✓ **Default mandatory signature verification** (allow_unsigned_packages: false)
- ✓ **Health check module** (pkg/health: Process/HTTP/TCP checks)
- ✓ Multi-key trust support
- ✓ **Health check integration with runtime** (monitoring from manifest config)
- ✓ **Auto-restart mechanism** (StartWithAutoRestart() method)
- ✓ **PSK (Pre-Shared Key) authentication** (private P2P network)
- ✓ **TLS 1.3 transport encryption** (libp2p native support)
- ✓ **Connection gating** (trusted peers whitelist)
- ⏸️ Resource limiting (cgroups on Linux)

### ⏸️ Not Started

**Phase 3A: Health & Resources** (5%)
- ✓ Health check base module (pkg/health/)
- ⏸️ Health check integration with runtime
- ⏸️ cgroups resource limiting
- ⏸️ Auto-restart with backoff strategies

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
- **Mandatory signature verification by default** (configurable)
- **Controller configuration file support** (--config flag, default: ~/.p2p-playground/controller.yaml)
- **Health check module** (Process/HTTP/TCP check types with retry logic)
- **Multi-key trust model** (multiple trusted public keys support)
- Configuration system (YAML with viper)
- Logging infrastructure (structured logging with zap)
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

### Phase 2.1 Implementation (Current) - ⚠️ UPDATED

**Decision**: Package signature verification is **mandatory by default** (as of 2026-01-19)

**Implementation**:
- `security.require_signed_packages` defaults to `true` in all configs
- When signature is present, it is always verified
- When signature is absent:
  - If `require_signed_packages: true` (default) → deployment **rejected**
  - If `require_signed_packages: false` → warning logged, deployment allowed
- Development/testing can set `require_signed_packages: false` in config

**Rationale**:
1. **Security by Default** - Safe default behavior prevents unsigned code execution
2. **Clear Security Posture** - Users must explicitly opt-out of security
3. **Best Practice** - Aligns with industry standards (container signing, code signing)

**Configuration**:
```yaml
# Production (default)
security:
  require_signed_packages: true  # Reject unsigned packages

# Development/Testing (explicit opt-out)
security:
  require_signed_packages: false  # Allow unsigned packages with warning
```

**Migration**: Existing users upgrading will need to either:
1. Generate signing keys and sign packages (recommended)
2. Explicitly set `require_signed_packages: false` in config

**Future Path**:
- Phase 3: Add fine-grained policies (per-app, per-label, etc.)
- Phase 3: Key revocation and rotation automation

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

### Controller Configuration System

**Implementation**: kubectl-style configuration file support (as of 2026-01-19)

**Features**:
- Global `--config/-c` flag for all commands
- Default config location: `~/.p2p-playground/controller.yaml`
- Automatic fallback to hardcoded defaults if no config found
- Shared configuration across all controller commands

**Configuration Structure**:
```yaml
node:
  listen_addrs: [...]  # P2P listening addresses
  enable_mdns: true    # mDNS discovery

logging:
  level: info          # debug/info/warn/error
  format: console      # console/json

storage:
  data_dir: ...        # Base data directory
  keys_dir: ...        # Signing keys location

security:
  require_signed_packages: true

deployment:
  timeout: 5m          # Deployment timeout
  retry_attempts: 3    # Retry count
  retry_delay: 10s     # Delay between retries
```

**Usage**:
```bash
# Use default config (~/.p2p-playground/controller.yaml)
controller deploy app.tar.gz

# Specify config file
controller --config /path/to/config.yaml deploy app.tar.gz
controller -c prod-config.yaml nodes
```

**Design Rationale**:
- Familiar UX for kubectl/docker users
- Environment-specific configurations (dev/staging/prod)
- Reduces command-line flag clutter
- Enables configuration management via version control

### Health Check Module

**Implementation**: Standalone health check module `pkg/health/` (as of 2026-01-19)

**Check Types**:
1. **Process Check**: Verifies process is alive using `syscall.Signal(0)`
2. **HTTP Check**: HTTP endpoint health check (GET request, status code validation)
3. **TCP Check**: TCP port connectivity check

**Configuration** (from manifest.yaml):
```yaml
health_check:
  type: process           # process|http|tcp
  interval: 30s           # Check interval
  timeout: 5s             # Timeout per check
  retries: 3              # Failures before unhealthy
  http_port: 8080         # For HTTP checks
  http_path: /health      # For HTTP checks
  tcp_port: 9090          # For TCP checks
```

**Features**:
- Consecutive failure tracking
- Configurable retry threshold
- Timeout support for each check
- Callback mechanism for unhealthy state
- Continuous monitoring via `StartMonitoring()`

**Status**: ✅ Complete - Integrated with runtime (as of 2026-01-19)

**Integration Details**:
- Health checker created automatically from manifest config
- Monitoring starts in background when app starts
- Cancels automatically when app stops
- Reports health status in `Status()` API
- Triggers auto-restart on failure (if enabled)

### Auto-Restart Mechanism

**Implementation**: Integrated into `pkg/runtime/runtime.go` (as of 2026-01-19)

**Usage**:
```go
// Start with auto-restart enabled
runtime.StartWithAutoRestart(ctx, app)

// Or use normal Start (no auto-restart)
runtime.Start(ctx, app)
```

**Behavior**:
- When health check fails (exceeds retry threshold), runtime triggers automatic restart
- Auto-restart preserves the setting across manual restarts
- Unhealthy app is stopped gracefully before restart
- Restart uses same configuration as original start
- Logged as "application unhealthy, triggering restart"

**Configuration** (from manifest.yaml):
```yaml
health_check:
  type: process
  interval: 30s
  retries: 3

# Auto-restart enabled via deployment flag
# (passed to runtime.StartWithAutoRestart)
```

**Implementation Architecture**:
- `appInfo` struct stores health checker and cancellation function
- Health monitoring runs in separate goroutine
- Unhealthy callback triggers `Restart()` in background goroutine
- Process monitor goroutine cancels health monitoring on exit

**Testing**:
```bash
# Create a manifest with health check
cat > manifest.yaml <<EOF
name: test-app
version: 1.0.0
entrypoint: bin/test
health_check:
  type: process
  interval: 30s
  timeout: 5s
  retries: 3
EOF

# Deploy with auto-restart (future: via --auto-restart flag)
# Currently integrated in daemon deployment logic
```

## Security Considerations

### Authentication

**PSK (Pre-Shared Key) Authentication** ✅ Implemented (2026-01-19)

P2P Playground支持PSK认证以创建私有P2P网络：

```yaml
security:
  enable_auth: true
  psk: "hex-encoded-32-bytes-key"
  trusted_peers:
    - "peer-id-1"
    - "peer-id-2"
```

**生成PSK**:
```bash
# 生成新的PSK（32字节，256位）
controller psk

# 指定输出路径
controller psk --output /path/to/psk
```

**工作原理**:
- PSK是32字节（256位）随机密钥
- 使用hex编码存储（64个十六进制字符）
- 所有节点（controller和daemon）必须使用相同的PSK
- 没有PSK或PSK不匹配的节点无法加入网络
- 使用libp2p的PrivateNetwork功能实现

**连接门控（Connection Gating）**:
- 基于peer ID的访问控制
- 支持可信节点白名单（`trusted_peers`）
- 拦截入站和出站连接
- 详细日志记录被拒绝的连接
- 如果trusted_peers为空，则允许所有连接

**使用场景**:
- 开发环境：无PSK，快速迭代
- 测试环境：PSK + 无trusted_peers，团队共享
- 生产环境：PSK + trusted_peers白名单，最高安全级别

### Code Signing
- Ed25519 signatures (fast, secure)
- Package signing via `controller sign` command
- Signature verification with configurable enforcement
- Public key distribution (manual - copy `.pub` files to daemon nodes)
- Multi-key trust support (multiple public keys in trusted directory)
- Default: reject unsigned packages (allow_unsigned_packages: false)

### Transport Security

**TLS 1.3 + Noise** ✅ Implemented (2026-01-19)

libp2p提供多层安全传输：

```go
// 配置TLS 1.3和Noise作为安全传输
libp2p.Security(libp2ptls.ID, libp2ptls.New),  // TLS 1.3 (primary)
libp2p.Security(noise.ID, noise.New),          // Noise (fallback)
```

**特性**:
- TLS 1.3作为主要安全传输协议
- Noise protocol作为备选（更轻量）
- libp2p原生支持，自动协商
- 无需手动配置证书（使用peer identity）
- 支持前向保密（Forward Secrecy）
- 所有P2P连接默认加密

**与PSK的关系**:
- PSK用于网络级访问控制（谁能加入）
- TLS/Noise用于传输级加密（数据保密性）
- 两者可以同时启用，提供双重保护

### File Integrity
- SHA-256 checksums for all transfers
- Verify before unpacking
- Signature verification for package authenticity

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
