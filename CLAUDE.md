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
golangci-lint run ./...
```

## Git Commit Convention

This project follows the [Conventional Commits](https://www.conventionalcommits.org/) specification for commit messages.

### Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- **feat**: 新功能 (A new feature)
- **fix**: 修复 Bug (A bug fix)
- **docs**: 文档变更 (Documentation only changes)
- **style**: 代码格式调整 (Changes that do not affect the meaning of the code)
- **refactor**: 重构 (A code change that neither fixes a bug nor adds a feature)
- **perf**: 性能优化 (A code change that improves performance)
- **test**: 测试相关 (Adding missing tests or correcting existing tests)
- **build**: 构建系统或外部依赖变更 (Changes that affect the build system or external dependencies)
- **ci**: CI 配置变更 (Changes to our CI configuration files and scripts)
- **chore**: 其他不影响源码的变更 (Other changes that don't modify src or test files)
- **revert**: 回退之前的提交 (Reverts a previous commit)

### Scope (Optional)

表示影响的范围，例如：
- `controller`: controller 相关
- `daemon`: daemon 相关
- `p2p`: P2P 网络层
- `security`: 安全相关
- `runtime`: 运行时
- `package`: 包管理
- etc.

### Subject

- 使用祈使句，现在时态："add" 而非 "added" 或 "adds"
- 不要大写首字母
- 结尾不加句号
- 简明扼要（建议不超过 50 个字符）

### Body (Optional)

- 详细描述改动的动机和与之前行为的对比
- 可以包含多个段落
- 每行建议不超过 72 个字符

### Footer (Optional)

- Breaking Changes: 以 `BREAKING CHANGE:` 开头
- 关闭的 Issue: `Closes #123, #456`

### Examples

```bash
# 新功能
feat: 添加 controller run 命令用于快速构建部署和日志监听

# Bug 修复
fix: 修复所有 golangci-lint 检查问题

# 文档更新
docs: 更新 DESIGN.md，反映 PSK、TLS 和 Phase 1/2 完成情况

# CI 配置
ci: 修复 golangci-lint CI 依赖问题

# 杂项
chore: 更新 .gitignore 忽略 controller 和 daemon 二进制文件

# 带 scope 的提交
feat(controller): 添加日志实时监听功能
fix(daemon): 修复包签名验证逻辑
```

### Important Notes

- **不要使用 Co-Authored-By**: 不要在提交信息中添加 `Co-Authored-By: Claude` 等协作者信息
- **语言选择**: 优先使用中文提交信息
- **自动化 Release**: 本项目使用 semantic-release 根据提交信息自动生成版本号和 CHANGELOG
  - `feat:` 触发 minor 版本升级 (0.x.0)
  - `fix:` 触发 patch 版本升级 (0.0.x)
  - `BREAKING CHANGE:` 触发 major 版本升级 (x.0.0)

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
