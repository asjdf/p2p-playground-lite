# P2P Playground Lite - 产品设计文档

## 1. 项目概述

P2P Playground Lite 是一个基于 P2P 网络的分布式应用测试和部署工具，旨在简化 Golang 应用的开发测试流程。通过 P2P 网络将应用程序自动分发到多个测试节点，实现快速部署和分布式测试。

### 1.1 核心价值

- **快速分发**：通过 P2P 网络将应用快速分发到多个测试节点
- **自动化测试**：daemon 进程自动接收、部署和运行应用
- **分布式管理**：统一管理多个测试节点上的应用生命周期
- **版本控制**：支持多版本共存、版本回滚和自动更新

### 1.2 使用场景

- **局域网开发测试**：公司内部快速分发应用到多台测试机
- **跨公网测试**：在分布式环境中测试应用的网络特性
- **持续集成**：集成到 CI/CD 流程中进行自动化测试

## 2. 系统架构

### 2.1 组件设计

系统分为两个主要组件：

#### 2.1.1 控制端 (Controller)
负责发起应用分发和管理操作的客户端，提供以下功能：
- 打包和分发应用
- 管理节点和应用版本
- 收集和查看日志
- 监控节点状态

#### 2.1.2 节点端 (Node/Daemon)
运行在各个测试节点上的守护进程，提供以下功能：
- 接收应用包
- 管理应用生命周期（启动、停止、重启）
- 健康检查和自动重启
- 日志收集和上报
- 资源限制和监控

### 2.2 版本划分

#### 轻量版 (Lite)
- 仅提供 CLI 工具
- 核心功能：应用分发、进程管理、日志收集
- 适合自动化脚本和 CI/CD 集成

#### 全量版 (Full)
- CLI 工具
- gRPC API 服务
- HTTP REST API（用于兼容性）
- Web Dashboard（基于 gRPC API）
- 适合需要图形界面和 API 集成的场景

## 3. 技术选型

### 3.1 P2P 网络层

**选择：libp2p**

使用 go-libp2p 作为 P2P 网络基础设施，原因：
- 成熟稳定，被 IPFS 等大型项目验证
- 内置节点发现、NAT 穿透、多传输协议支持
- 自带 TLS 加密，满足安全传输需求
- 良好的 Go 语言支持

**关键配置**：
- 节点发现：mDNS (本地网络广播)
- 传输协议：TCP + QUIC
- 安全传输：libp2p TLS 1.3
- 内容路由：支持混合场景（局域网+公网）

### 3.2 进程管理

**资源限制**：
- 使用 cgroups (Linux) 或等效机制限制 CPU/内存
- 配置可调的资源上限

**健康检查**：
- 进程存活检查（定期检查进程状态）
- 可选的 HTTP/TCP 健康检查端点
- 配置重启策略（立即重启、延迟重启、重启次数限制）

**日志管理**：
- 捕获 stdout/stderr
- 支持日志轮转和大小限制
- 实时日志流式传输
- 持久化本地日志

### 3.3 安全机制

#### 3.3.1 节点身份认证
- 基于 libp2p PeerID 的节点识别
- 使用预共享密钥 (PSK) 或证书进行节点认证
- 维护可信节点白名单

#### 3.3.2 应用包签名验证
- 对整个应用包（app.tar.gz）计算 SHA-256 哈希并使用私钥签名
- 签名算法：Ed25519（推荐）或 RSA-PSS 2048/4096
- 节点接收后使用可信公钥验证签名
- 维护可信发布者公钥白名单
- **一次签名验证同时保证**：
  - ✅ 验证应用包来源可信（发布者身份）
  - ✅ 验证应用包完整性（任何文件被篡改都会导致签名验证失败）
  - ✅ 防止中间人攻击
- 签名验证失败则拒绝部署

#### 3.3.3 加密传输
- libp2p 内置 TLS 1.3 加密
- 可选的应用包额外加密层

## 4. 核心功能设计

### 4.1 应用包格式

应用包采用双层打包结构，确保完整性和防篡改：

#### 4.1.1 包结构

```
# 最终分发包（外层）
myapp-1.0.0.p2p
├── app.tar.gz           # 应用内容包（内层）
└── app.tar.gz.sig       # 整个 tar.gz 的数字签名

# 应用内容包（app.tar.gz 解压后）
app.tar.gz
├── manifest.yaml        # 应用元数据和文件清单
├── bin/
│   └── myapp            # 主可执行文件
├── config/
│   └── config.yaml      # 配置文件
├── resources/
│   └── ...              # 资源文件
└── scripts/
    ├── pre-start.sh     # 启动前脚本
    └── post-stop.sh     # 停止后脚本
```

**manifest.yaml 结构**：
```yaml
name: myapp
version: 1.0.0
description: My application

entrypoint: bin/myapp
args:
  - --config
  - config/config.yaml

env:
  APP_ENV: test

resources:
  cpu_limit: "1.0"
  memory_limit: 512Mi

health_check:
  type: http
  endpoint: http://localhost:8080/health
  interval: 10s
  timeout: 5s

created_at: "2026-01-19T10:00:00Z"
```

#### 4.1.2 打包和签名流程

**Controller 端打包流程**：
1. 准备应用文件（可执行文件、配置、资源、脚本等）
2. 编写 manifest.yaml，包含应用元数据和运行配置
3. 将应用文件 + manifest.yaml 打包成 app.tar.gz（内层包）
4. 计算 app.tar.gz 的 SHA-256 哈希
5. 使用私钥对 app.tar.gz 的哈希进行签名，生成 app.tar.gz.sig
6. 将 app.tar.gz 和 app.tar.gz.sig 一起打包成 myapp-1.0.0.p2p（外层包）

**签名算法**：
- 推荐：Ed25519（性能好，密钥小，安全性高）
- 备选：RSA-PSS 2048/4096（兼容性好）
- 签名格式：detached signature

#### 4.1.3 验证流程

**Node 端验证流程**（简洁高效）：

1. **解压外层包**
   - 解压 .p2p 文件，得到 app.tar.gz 和 app.tar.gz.sig

2. **验证签名**
   - 计算 app.tar.gz 的 SHA-256 哈希
   - 使用可信公钥验证 app.tar.gz.sig 签名
   - **签名验证失败 → 拒绝部署**

3. **解压内层包**
   - 签名验证通过后，解压 app.tar.gz 到应用目录
   - 读取 manifest.yaml
   - 设置可执行文件权限

4. **验证通过**
   - 应用准备就绪，可以启动

**安全保证**：
- ✅ **外层签名验证**：确保应用包来源可信和整体完整性
- ✅ **防篡改**：任何文件被修改都会导致 tar.gz 哈希变化，签名验证失败
- ✅ **防中间人攻击**：签名验证确保包未被替换
- ✅ **防重放攻击**：可结合版本号和时间戳检查
- ✅ **简洁高效**：只需一次签名验证，性能好

### 4.2 版本管理

#### 4.2.1 版本标识
- 使用语义化版本号 (Semantic Versioning)
- 每个应用包有唯一的版本标识
- 支持版本标签 (latest, stable, dev)

#### 4.2.2 多版本共存
- 每个版本独立存储
- 可以同时运行多个版本的应用（不同端口/配置）
- 版本间隔离，互不干扰

#### 4.2.3 自动更新
- 控制端推送新版本到节点
- 节点可配置自动更新策略：
  - 立即更新：收到后立即停止旧版本，启动新版本
  - 优雅更新：等待当前版本停止后更新
  - 手动更新：需要手动触发
- 支持灰度发布：先更新部分节点

#### 4.2.4 版本回滚
- 保留历史版本
- 一键回滚到指定版本
- 自动回滚：新版本健康检查失败时回滚

### 4.3 节点发现和管理

#### 4.3.1 节点发现
- mDNS 本地网络自动发现
- 支持手动添加远程节点（跨网段）
- 节点上线/下线事件通知

#### 4.3.2 节点状态
每个节点维护以下状态：
- 节点 ID 和网络地址
- 运行中的应用列表和版本
- 资源使用情况（CPU、内存、磁盘）
- 最后心跳时间

#### 4.3.3 节点分组
- 支持节点打标签（env=test, region=cn-north）
- 基于标签选择性分发应用
- 分组管理和批量操作

### 4.4 日志收集

#### 4.4.1 日志类型
- 应用标准输出/错误输出
- 节点 daemon 自身日志
- 系统事件日志（应用启动、停止、更新等）

#### 4.4.2 日志传输
- 实时日志流：通过 gRPC stream 推送
- 历史日志查询：支持时间范围和关键词搜索
- 日志聚合：控制端可聚合查看多节点日志

#### 4.4.3 日志存储
- 节点本地存储（滚动日志）
- 可选的远程日志存储（如 S3、OSS）
- 日志保留策略（按时间或大小）

## 5. API 设计

### 5.1 CLI 命令（轻量版 + 全量版）

```bash
# 节点管理
p2p-playground node list                      # 列出所有节点
p2p-playground node info <node-id>            # 查看节点详情
p2p-playground node add <address>             # 手动添加节点

# 应用部署
p2p-playground deploy <app-package>           # 部署应用到所有节点
p2p-playground deploy <app-package> --nodes=node1,node2  # 部署到指定节点
p2p-playground deploy <app-package> --labels=env=test    # 部署到标签匹配的节点

# 应用管理
p2p-playground app list [--node=<node-id>]    # 列出应用
p2p-playground app start <app-name>           # 启动应用
p2p-playground app stop <app-name>            # 停止应用
p2p-playground app restart <app-name>         # 重启应用
p2p-playground app remove <app-name>          # 删除应用

# 版本管理
p2p-playground app versions <app-name>        # 查看应用版本
p2p-playground app rollback <app-name> <version>  # 回滚到指定版本
p2p-playground app upgrade <app-name>         # 升级到最新版本

# 日志查看
p2p-playground logs <app-name> [--node=<node-id>]  # 查看日志
p2p-playground logs <app-name> --follow       # 实时跟踪日志
p2p-playground logs <app-name> --since=1h     # 查看最近1小时日志

# Daemon 管理
p2p-playground daemon start                   # 启动 daemon
p2p-playground daemon stop                    # 停止 daemon
p2p-playground daemon status                  # 查看 daemon 状态
```

### 5.2 gRPC API（全量版）

```protobuf
service P2PPlayground {
  // 节点管理
  rpc ListNodes(ListNodesRequest) returns (ListNodesResponse);
  rpc GetNodeInfo(GetNodeInfoRequest) returns (NodeInfo);
  rpc StreamNodeEvents(StreamNodeEventsRequest) returns (stream NodeEvent);

  // 应用部署
  rpc DeployApp(stream DeployAppRequest) returns (DeployAppResponse);
  rpc ListApps(ListAppsRequest) returns (ListAppsResponse);
  rpc StartApp(StartAppRequest) returns (StartAppResponse);
  rpc StopApp(StopAppRequest) returns (StopAppResponse);
  rpc RemoveApp(RemoveAppRequest) returns (RemoveAppResponse);

  // 版本管理
  rpc ListVersions(ListVersionsRequest) returns (ListVersionsResponse);
  rpc RollbackApp(RollbackAppRequest) returns (RollbackAppResponse);

  // 日志管理
  rpc GetLogs(GetLogsRequest) returns (GetLogsResponse);
  rpc StreamLogs(StreamLogsRequest) returns (stream LogEntry);
}
```

### 5.3 HTTP REST API（全量版）

```
GET    /api/v1/nodes                     # 列出节点
GET    /api/v1/nodes/:id                 # 节点详情
POST   /api/v1/apps/deploy               # 部署应用
GET    /api/v1/apps                      # 列出应用
POST   /api/v1/apps/:name/start          # 启动应用
POST   /api/v1/apps/:name/stop           # 停止应用
POST   /api/v1/apps/:name/restart        # 重启应用
DELETE /api/v1/apps/:name                # 删除应用
GET    /api/v1/apps/:name/versions       # 应用版本列表
POST   /api/v1/apps/:name/rollback       # 版本回滚
GET    /api/v1/apps/:name/logs           # 获取日志
GET    /api/v1/apps/:name/logs/stream    # 实时日志流 (WebSocket)
```

### 5.4 Web Dashboard（全量版）

Web 界面功能模块：

1. **节点管理面板**
   - 节点列表和状态
   - 节点资源监控（CPU、内存、磁盘）
   - 节点标签管理

2. **应用管理面板**
   - 应用列表（按节点分组）
   - 应用状态和版本信息
   - 启动/停止/重启操作
   - 应用部署向导

3. **日志查看器**
   - 多节点日志聚合查看
   - 实时日志流
   - 日志搜索和过滤

4. **版本管理**
   - 版本历史
   - 版本对比
   - 一键回滚

## 6. 部署架构

### 6.1 典型部署场景

#### 场景 1：局域网测试环境
```
开发机 (Controller)
    │
    └─── P2P Network (mDNS)
           ├─── 测试机1 (Node Daemon)
           ├─── 测试机2 (Node Daemon)
           └─── 测试机3 (Node Daemon)
```

#### 场景 2：跨公网分布式测试
```
CI/CD Server (Controller)
    │
    └─── P2P Network (libp2p + NAT穿透)
           ├─── 云主机1 (AWS, Node Daemon)
           ├─── 云主机2 (阿里云, Node Daemon)
           └─── 本地机房 (Node Daemon)
```

#### 场景 3：混合环境
```
开发机 (Controller)
    │
    └─── P2P Network
           ├─── 局域网节点组 (mDNS)
           │     ├─── 测试机1
           │     └─── 测试机2
           └─── 远程节点组 (手动配置)
                 ├─── 云主机1
                 └─── 云主机2
```

### 6.2 配置文件

#### 6.2.1 密钥说明

**Controller（控制端）**：
- 持有**私钥**（signing key），用于签名打包的应用
- 不需要验证应用包，因此不需要公钥列表

**Node（节点端）**：
- 持有**可信公钥列表**（trusted public keys），用于验证接收到的应用包
- 不需要私钥，因为节点不签名应用包

**密钥生成和分发流程**：
1. 开发者生成密钥对（私钥 + 公钥）
2. 私钥保存在 Controller，用于签名
3. 公钥分发给所有 Node，加入可信列表
4. 支持多个可信公钥（多个开发者或团队）

#### 6.2.2 Controller 配置 (controller.yaml)
```yaml
controller:
  name: "dev-controller"

p2p:
  listen_addresses:
    - "/ip4/0.0.0.0/tcp/4001"
    - "/ip4/0.0.0.0/udp/4001/quic"
  enable_mdns: true
  psk: "your-pre-shared-key"  # P2P 网络的预共享密钥

security:
  # 私钥路径，用于签名应用包
  signing_key_path: "/etc/p2p-playground/signing.key"
  # 私钥类型: ed25519 或 rsa
  signing_key_type: "ed25519"

api:
  grpc:
    enabled: true
    listen: "0.0.0.0:9090"
  http:
    enabled: true
    listen: "0.0.0.0:8080"
  web:
    enabled: true
    listen: "0.0.0.0:3000"
```

#### 6.2.3 Node Daemon 配置 (daemon.yaml)
```yaml
node:
  name: "test-node-1"
  labels:
    env: "test"
    region: "cn-north"

p2p:
  listen_addresses:
    - "/ip4/0.0.0.0/tcp/4001"
    - "/ip4/0.0.0.0/udp/4001/quic"
  enable_mdns: true
  bootstrap_peers: []  # 可选的引导节点
  psk: "your-pre-shared-key"  # P2P 网络的预共享密钥（与 Controller 相同）

security:
  # 可信公钥列表，用于验证应用包签名
  trusted_public_keys:
    - name: "team-lead"
      key_type: "ed25519"
      public_key: "b5a2c9e1f3d7a8b6c4e2f0a1b3c5d7e9f1a3b5c7d9e1f3a5b7c9d1e3f5a7b9"

    - name: "ci-cd-server"
      key_type: "ed25519"
      public_key: "c6b3d0e2f4d8a9b7c5e3f1a2b4c6d8e0f2a4b6c8d0e2f4a6b8c0d2e4f6a8b0"

runtime:
  app_dir: "/var/lib/p2p-playground/apps"
  log_dir: "/var/log/p2p-playground"
  max_apps: 10

  # 默认资源限制
  default_limits:
    cpu: "1.0"
    memory: "1Gi"

logging:
  level: "info"
  max_size: "100Mi"
  max_age_days: 7
```

## 7. 开发路线图

### Phase 1: 核心功能（MVP）
- [ ] P2P 网络基础（libp2p 集成）
- [ ] 节点发现（mDNS）
- [ ] 应用包定义和打包
- [ ] 文件传输和校验
- [ ] 基础进程管理（启动/停止）
- [ ] CLI 工具（轻量版）

### Phase 2: 安全和稳定性
- [ ] 节点身份认证
- [ ] 代码签名和验证
- [ ] 健康检查和自动重启
- [ ] 资源限制
- [ ] 日志收集

### Phase 3: 高级功能
- [ ] 版本管理
- [ ] 多版本共存
- [ ] 自动更新
- [ ] 节点标签和分组

### Phase 4: API 和 UI（全量版）
- [ ] gRPC API
- [ ] HTTP REST API
- [ ] Web Dashboard
- [ ] 实时日志流

### Phase 5: 生产优化
- [ ] 性能优化
- [ ] 大规模节点支持
- [ ] 监控和告警
- [ ] 文档和示例

## 8. 未来考虑

### 8.1 可能的扩展功能
- **容器化支持**：支持分发和运行 Docker 容器
- **插件系统**：允许扩展自定义功能
- **调度策略**：根据节点资源自动调度应用
- **流量控制**：限制 P2P 网络带宽使用
- **加密存储**：应用包落盘加密

### 8.2 集成能力
- **CI/CD 集成**：提供 GitHub Actions、GitLab CI 插件
- **监控集成**：输出 Prometheus metrics
- **日志集成**：支持 Loki、ElasticSearch 等日志系统

## 9. 安全注意事项

1. **网络安全**
   - 使用 PSK 或证书限制节点加入
   - 定期轮换密钥
   - 监控异常节点行为

2. **代码安全**
   - 强制签名验证
   - 定期审计可信签名者列表
   - 实施最小权限原则

3. **运行时安全**
   - 应用进程沙箱化
   - 资源限制防止 DoS
   - 定期安全扫描

4. **数据安全**
   - 敏感配置加密存储
   - 日志脱敏
   - 安全删除旧版本应用

## 10. 性能目标

- **分发速度**：100MB 应用包分发到 10 个节点 < 30秒
- **启动延迟**：应用接收到启动后 < 5秒
- **日志延迟**：实时日志延迟 < 1秒
- **资源开销**：Daemon 进程空闲内存 < 50MB
- **扩展性**：单个控制端支持管理 >= 100 个节点
