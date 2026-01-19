# Docker Testing Guide

This guide shows how to test P2P Playground Lite using Docker.

## Prerequisites

- Docker 20.10+
- Docker Compose 1.29+

## Quick Start

### 1. Build and Start the Network

```bash
# Build images and start all nodes
docker-compose up -d

# Check logs
docker-compose logs -f daemon1
```

This will start:
- 3 daemon nodes (daemon1, daemon2, daemon3)
- 1 controller node (for manual testing)

### 2. Verify Nodes are Running

```bash
# Check daemon logs
docker-compose logs daemon1 | grep "daemon started"
docker-compose logs daemon2 | grep "daemon started"
docker-compose logs daemon3 | grep "daemon started"

# Check if daemons discovered each other via mDNS
docker-compose logs | grep "discovered peer"
```

### 3. Package the Example Application

```bash
# Build and package hello-world app
cd examples/hello-world
./build.sh
tar -czf hello-world-1.0.0.tar.gz manifest.yaml bin/
cd ../..
```

### 4. Deploy Application (Manual Testing)

Enter the controller container:

```bash
docker exec -it p2p-controller sh
```

Inside the container:

```bash
# Check available nodes (TODO: implement node discovery command)
# For now, manually note the peer IDs from daemon logs

# Copy package to container (from host)
# docker cp examples/hello-world/hello-world-1.0.0.tar.gz p2p-controller:/data/

# Deploy to a node (TODO: implement deploy command)
# controller deploy /data/hello-world-1.0.0.tar.gz --node <peer-id>
```

### 5. Verify Deployment

```bash
# Check daemon logs for deployment
docker-compose logs daemon1 | grep "hello-world"

# Check running applications
docker exec p2p-daemon1 ps aux | grep hello-world
```

## Architecture

```
┌─────────────────────────────────────────────────┐
│              Docker Network (bridge)            │
│                  172.20.0.0/16                  │
│                                                 │
│  ┌──────────────┐    ┌──────────────┐         │
│  │ Controller   │    │  Daemon 1    │         │
│  │ (manual)     │    │  :9000       │         │
│  │ :9001        │    │              │         │
│  └──────────────┘    └──────────────┘         │
│                                                 │
│  ┌──────────────┐    ┌──────────────┐         │
│  │  Daemon 2    │    │  Daemon 3    │         │
│  │  :9002       │    │  :9003       │         │
│  └──────────────┘    └──────────────┘         │
│                                                 │
│         P2P Discovery via mDNS                  │
└─────────────────────────────────────────────────┘
```

## Useful Commands

### View Logs

```bash
# All nodes
docker-compose logs -f

# Specific node
docker-compose logs -f daemon1

# Last 50 lines
docker-compose logs --tail=50 daemon1
```

### Execute Commands

```bash
# Enter daemon container
docker exec -it p2p-daemon1 sh

# Check daemon status
docker exec p2p-daemon1 ps aux

# View daemon data
docker exec p2p-daemon1 ls -la /data
```

### Restart Nodes

```bash
# Restart specific node
docker-compose restart daemon1

# Restart all
docker-compose restart
```

### Clean Up

```bash
# Stop all containers
docker-compose down

# Stop and remove volumes (clean slate)
docker-compose down -v

# Rebuild images
docker-compose build --no-cache
```

## Testing Scenarios

### 1. Basic Connectivity Test

```bash
# Start daemons and watch for peer discovery
docker-compose up -d
docker-compose logs -f | grep "discovered peer"
```

Expected: Each daemon should discover the other 2 daemons via mDNS.

### 2. Controller Node Discovery Test ✓ VERIFIED

```bash
# Use controller to discover all daemon nodes
docker exec p2p-controller controller nodes
```

Expected output:
```
Discovering P2P nodes...
Controller ID: <peer-id>
Controller addresses:
  - /ip4/127.0.0.1/tcp/<port>
  - /ip4/172.20.0.4/tcp/<port>

Scanning for nodes via mDNS (waiting 5 seconds)...

Discovered 3 peer(s):
1. Peer ID: 12D3KooWAw8VzZsnnK43xqaXgRBmXE4YhgkpToJ9WeX7bGFuKjXg
   Addresses:
     - /ip4/172.20.0.3/tcp/9000

2. Peer ID: 12D3KooWSuoRnuqBdo4e4PJrSYMCdGpFTSbmaTQUjyECRd37orLK
   Addresses:
     - /ip4/172.20.0.5/tcp/9000

3. Peer ID: 12D3KooWNgysE5Jhkqyjj6jAbsHYRLj4KSqTGssJdEBnmatgtddZ
   Addresses:
     - /ip4/172.20.0.2/tcp/9000
```

**Status**: ✓ Verified working - Controller successfully discovers all 3 daemon nodes via mDNS

### 3. Application Deployment Test ✓ VERIFIED

```bash
# 1. Package application
cd examples/hello-world && ./build.sh && tar -czf hello-world-1.0.0.tar.gz manifest.yaml bin/ && cd ../..

# 2. Copy to controller
docker cp examples/hello-world/hello-world-1.0.0.tar.gz p2p-controller:/data/

# 3. Deploy to a node
docker exec p2p-controller controller deploy /data/hello-world-1.0.0.tar.gz
```

Expected output:
```
Deploying package: /data/hello-world-1.0.0.tar.gz
Controller ID: <peer-id>
Discovering nodes...
Using discovered node: <target-peer-id>

Deploying package...
  Progress: 50%
  Progress: 100%

✓ Deployment successful!
  Application ID: hello-world-1.0.0
  Status: Started
```

**Status**: ✓ Verified working - Application successfully deployed, unpacked, and started on target node

### 4. Application Listing Test ✓ VERIFIED

```bash
# List applications on a specific node
docker exec p2p-controller controller list --node <peer-id>

# Or list on first discovered node (default)
docker exec p2p-controller controller list
```

Expected output:
```
Listing applications...
Discovering nodes...
Using specified node: <peer-id>

Fetching applications...

Found 1 application(s):

1. Application: hello-world
   ID: hello-world-1.0.0
   Version: 1.0.0
   Status: running
   PID: 19
   Started: 2026-01-19 08:40:00
   Labels: map[app:hello-world env:demo]
```

**Status**: ✓ Verified working - Successfully lists all deployed applications with full details

### 5. Application Logs Test ✓ VERIFIED

```bash
# View logs from a deployed application
docker exec p2p-controller controller logs hello-world-1.0.0 --node <peer-id>

# Specify number of lines to show
docker exec p2p-controller controller logs hello-world-1.0.0 --tail 20

# Follow logs in real-time (TODO: streaming not yet implemented)
docker exec p2p-controller controller logs hello-world-1.0.0 --follow
```

Expected output:
```
Fetching logs for application: hello-world-1.0.0
Discovering nodes...
Using specified node: <peer-id>

Fetching logs...

Hello from P2P Playground!
Version: 1.0.0
Hostname: daemon3
Started at: 2026-01-19T08:40:00Z
[08:40:10] Application is running...
[08:40:20] Application is running...
[08:40:30] Application is running...
...
```

**Status**: ✓ Verified working - Successfully fetches and displays application logs

### 6. Package Signature Verification Test ✓ VERIFIED

#### 6.1. Generate Signing Keys

```bash
# Generate Ed25519 key pair for package signing
./bin/controller keygen -o /tmp/test-controller-keys
```

Expected output:
```
✓ Key pair generated successfully!
  Private key: /tmp/test-controller-keys/controller.key
  Public key:  /tmp/test-controller-keys/controller.pub

⚠️  Keep the private key secure and never share it.
📤 Distribute the public key to nodes for signature verification.

Public key (hex): <hex-encoded-public-key>
```

#### 6.2. Sign Application Package

```bash
# Sign the packaged application
./bin/controller sign examples/hello-world/hello-world-1.0.0.tar.gz \
  --key /tmp/test-controller-keys/controller.key
```

Expected output:
```
✓ Package signed successfully!
  Signature: examples/hello-world/hello-world-1.0.0.tar.gz.sig
```

#### 6.3. Deploy Signed Package to Daemon

```bash
# Set up trusted keys directory for testing
mkdir -p configs/keys/trusted
cp /tmp/test-controller-keys/controller.pub configs/keys/trusted/

# Copy trusted key to daemon container
docker exec p2p-daemon2 mkdir -p /data/keys/trusted
docker cp configs/keys/trusted/controller.pub p2p-daemon2:/data/keys/trusted/

# Copy signed package and signature to controller
docker cp examples/hello-world/hello-world-1.0.0.tar.gz p2p-controller:/data/
docker cp examples/hello-world/hello-world-1.0.0.tar.gz.sig p2p-controller:/data/

# Deploy with signature verification
docker exec p2p-controller controller deploy /data/hello-world-1.0.0.tar.gz
```

Expected output:
```
Deploying package: /data/hello-world-1.0.0.tar.gz
...
✓ Deployment successful!
  Application ID: hello-world-1.0.0
  Status: Started
```

Verify signature verification in daemon logs:
```bash
docker logs p2p-daemon2 | grep -E "(signature|Signature|verify)"
```

Expected log entries:
```
info  verifying package signature
info  signature verified  {"public_key": "controller.pub"}
info  package signature verified successfully
```

**Status**: ✓ Verified working - Signature verification works with Ed25519 keys

#### 6.4. Deploy Unsigned Package (Optional Verification)

```bash
# Remove signature file
docker exec p2p-controller rm /data/hello-world-1.0.0.tar.gz.sig

# Deploy without signature
docker exec p2p-controller controller deploy /data/hello-world-1.0.0.tar.gz
```

Expected output (deployment succeeds with warning):
```
Deploying package: /data/hello-world-1.0.0.tar.gz
...
warn  no package signature found, deploying without signature verification
...
✓ Deployment successful!
```

Verify warning in daemon logs:
```bash
docker logs p2p-daemon3 | grep signature
```

Expected log:
```
warn  package deployed without signature verification
```

**Status**: ✓ Verified working - Unsigned packages deploy with warning when require_signed_packages=false (⚠️ default is true, will reject unsigned packages)

#### 6.5. Multi-Key Trust Scenario

The daemon supports multiple trusted public keys in the `/data/keys/trusted/` directory. This allows:
- Key rotation without disrupting deployments
- Multiple teams/developers signing packages
- Gradual key migration

```bash
# Add multiple trusted keys
docker cp /path/to/team-a.pub p2p-daemon1:/data/keys/trusted/
docker cp /path/to/team-b.pub p2p-daemon1:/data/keys/trusted/

# Daemon will accept packages signed by either key
```

**Security Notes**:
- Private keys (`.key` files) should never be committed to version control
- Public keys (`.pub` files) can be safely distributed
- **⚠️ By default, signature verification is MANDATORY** (`require_signed_packages: true`)
- Unsigned packages will be **rejected** unless you explicitly set `require_signed_packages: false`
- `daemon-docker.yaml` sets `require_signed_packages: false` for easier local testing
- For production, keep the default `require_signed_packages: true`
- See `docs/SECURITY.md` for detailed security guidelines

### 7. Multi-Node Deployment Test

Deploy the same application to all 3 daemons and verify they all run correctly.

### 8. Network Partition Test

```bash
# Disconnect daemon3 from network
docker network disconnect p2p-playground_p2p-network p2p-daemon3

# Verify other nodes continue working
docker-compose logs daemon1 daemon2

# Reconnect
docker network connect p2p-playground_p2p-network p2p-daemon3
```

### 9. Node Failure Test

```bash
# Stop daemon2
docker-compose stop daemon2

# Verify applications on daemon1 and daemon3 continue running

# Restart daemon2
docker-compose start daemon2

# Verify it rejoins the network
```

## Troubleshooting

### Daemons Not Discovering Each Other

```bash
# Check if containers are on the same network
docker network inspect p2p-playground_p2p-network

# Check mDNS logs
docker-compose logs | grep -i mdns

# Verify listen addresses
docker-compose logs daemon1 | grep "libp2p host created"
```

### Port Conflicts

If you see "address already in use" errors:

```bash
# Check what's using the ports
lsof -i :9000
lsof -i :9001

# Change ports in docker-compose.yaml
```

### Container Won't Start

```bash
# Check logs
docker-compose logs daemon1

# Check container status
docker-compose ps

# Rebuild
docker-compose build daemon1
docker-compose up -d daemon1
```

### Trusted Keys Not Present in Daemon Container

If signature verification fails with "trusted public keys directory not found":

```bash
# Check if trusted keys directory exists in container
docker exec p2p-daemon1 ls -la /data/keys/trusted/

# If directory missing, manually create and copy keys
docker exec p2p-daemon1 mkdir -p /data/keys/trusted
docker cp configs/keys/trusted/*.pub p2p-daemon1:/data/keys/trusted/

# Restart daemon to pick up new keys
docker-compose restart daemon1
```

**Note**: The Dockerfile's `COPY configs/keys/trusted/*.pub` command only works if `.pub` files exist at build time. For testing, you may need to manually copy keys to running containers as shown above.

## Development Workflow

1. Make code changes
2. Rebuild images: `docker-compose build`
3. Restart services: `docker-compose up -d`
4. Test changes: `docker-compose logs -f`
5. Iterate

## Next Steps

### Phase 1D - CLI Implementation ✓ COMPLETED
- [x] Implement controller deploy command (✓ `controller deploy` working)
- [x] Implement node discovery/list command (✓ `controller nodes` working)
- [x] Implement application list command (✓ `controller list` working)
- [x] Implement logs viewing command (✓ `controller logs` working)
- [x] End-to-end deployment tested and verified

### Phase 2 - Security ✓ PARTIALLY COMPLETED
- [ ] Add PSK authentication
- [x] Add package signature verification (✓ Ed25519 signing and verification working)
- [ ] Add TLS 1.3 transport encryption

### Phase 3 - Advanced Features (Planned)
- [ ] Add health checks to docker-compose
- [ ] Implement auto-restart on health check failure
- [ ] Add resource limiting (CPU/memory)
- [ ] Add Prometheus metrics
- [ ] Create Grafana dashboard

### Completed ✓
- [x] Docker multi-node environment setup
- [x] mDNS peer discovery working
- [x] Controller node discovery verified
- [x] All 3 daemons successfully connecting
- [x] Application deployment working end-to-end
- [x] Application listing and status viewing
- [x] Log fetching and viewing from deployed apps
- [x] Complete Phase 1 (MVP) implementation
