package p2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/asjdf/p2p-playground-lite/pkg/security"
	"github.com/asjdf/p2p-playground-lite/pkg/types"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/pnet"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-libp2p/p2p/host/autorelay"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/multiformats/go-multiaddr"
)

// DefaultBootstrapPeers are the default IPFS bootstrap nodes
var DefaultBootstrapPeers = []string{
	"/dnsaddr/bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN",
	"/dnsaddr/bootstrap.libp2p.io/p2p/QmQCU2EcMqAqQPR2i9bChDtGNJchTbq5TbXJJ16u19uLTa",
	"/dnsaddr/bootstrap.libp2p.io/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb",
	"/dnsaddr/bootstrap.libp2p.io/p2p/QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt",
	"/ip4/104.131.131.82/tcp/4001/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
	"/ip4/104.131.131.82/udp/4001/quic/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
}

// Host wraps libp2p host
type Host struct {
	host   host.Host
	dht    *dht.IpfsDHT
	logger types.Logger
}

// HostConfig contains configuration for creating a P2P host
type HostConfig struct {
	// ListenAddrs are the multiaddrs to listen on
	ListenAddrs []string

	// PSK is the pre-shared key for private network (hex-encoded)
	PSK string

	// EnableAuth enables PSK authentication
	EnableAuth bool

	// TrustedPeers are peer IDs allowed to connect (if non-empty)
	TrustedPeers []string

	// BootstrapPeers are initial peers to connect to
	BootstrapPeers []string

	// DisableDHT disables Distributed Hash Table for peer discovery
	DisableDHT bool

	// DHTMode is the DHT mode: "client" or "server" (default: "server")
	DHTMode string

	// DisableNATService disables NAT traversal service
	DisableNATService bool

	// DisableAutoRelay disables automatic relay for NAT traversal
	DisableAutoRelay bool

	// DisableHolePunching disables hole punching for direct connections
	DisableHolePunching bool

	// DisableRelayService disables this node from acting as a relay server
	DisableRelayService bool

	// StaticRelays are static relay addresses for NAT traversal
	// If provided, these will be used instead of DHT-based relay discovery
	StaticRelays []string
}

// NewHost creates a new P2P host
func NewHost(ctx context.Context, config *HostConfig, logger types.Logger) (*Host, error) {
	// Parse listen addresses
	var maddrs []multiaddr.Multiaddr
	for _, addr := range config.ListenAddrs {
		maddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return nil, fmt.Errorf("invalid multiaddr %s: %w", addr, err)
		}
		maddrs = append(maddrs, maddr)
	}

	// Build libp2p options
	opts := []libp2p.Option{
		libp2p.ListenAddrs(maddrs...),
		// Enable TLS 1.3 and Noise security transports
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		libp2p.Security(noise.ID, noise.New),
	}

	// Add NAT traversal options (enabled by default)
	if !config.DisableNATService {
		opts = append(opts, libp2p.EnableNATService())
		logger.Info("NAT service enabled")
	}
	if !config.DisableHolePunching {
		opts = append(opts, libp2p.EnableHolePunching())
		logger.Info("hole punching enabled")
	}

	// Enable relay service (enabled by default, allows this node to relay connections for others)
	if !config.DisableRelayService {
		opts = append(opts, libp2p.EnableRelayService())
		logger.Info("relay service enabled - this node can relay connections for other peers")
	}

	// Add DHT routing (enabled by default)
	var kadDHT *dht.IpfsDHT
	if !config.DisableDHT {
		opts = append(opts, libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			// Determine DHT mode (default: server)
			var dhtMode dht.ModeOpt
			if config.DHTMode == "client" {
				dhtMode = dht.ModeClient
			} else {
				dhtMode = dht.ModeServer
			}

			var err error
			kadDHT, err = dht.New(ctx, h, dht.Mode(dhtMode))
			if err != nil {
				return nil, err
			}
			return kadDHT, nil
		}))
		dhtModeStr := config.DHTMode
		if dhtModeStr == "" {
			dhtModeStr = "server"
		}
		logger.Info("DHT enabled", "mode", dhtModeStr)

		// Enable AutoRelay (only when DHT is enabled, unless static relays are configured)
		if !config.DisableAutoRelay {
			// Use static relays if provided, otherwise use DHT peer source
			if len(config.StaticRelays) > 0 {
				staticRelays := parseStaticRelays(config.StaticRelays, logger)
				if len(staticRelays) > 0 {
					opts = append(opts, libp2p.EnableAutoRelayWithStaticRelays(staticRelays,
						autorelay.WithBackoff(30*time.Second),
						autorelay.WithMinInterval(time.Minute),
					))
					logger.Info("auto relay enabled with static relays", "count", len(staticRelays))
				}
			} else {
				// Use DHT as peer source for relay discovery
				opts = append(opts, libp2p.EnableAutoRelayWithPeerSource(
					func(ctx context.Context, numPeers int) <-chan peer.AddrInfo {
						ch := make(chan peer.AddrInfo)
						go func() {
							defer close(ch)
							// Wait a bit for DHT to initialize
							time.Sleep(2 * time.Second)
							if kadDHT == nil {
								return
							}
							// Find peers from DHT
							for _, p := range kadDHT.RoutingTable().ListPeers() {
								select {
								case ch <- peer.AddrInfo{ID: p}:
								case <-ctx.Done():
									return
								}
								if numPeers--; numPeers <= 0 {
									return
								}
							}
						}()
						return ch
					},
					autorelay.WithBackoff(30*time.Second),
					autorelay.WithMinInterval(time.Minute),
				))
				logger.Info("auto relay enabled with DHT peer source")
			}
		}
	} else if !config.DisableAutoRelay {
		// DHT is disabled, but we can still use static relays if provided
		if len(config.StaticRelays) > 0 {
			staticRelays := parseStaticRelays(config.StaticRelays, logger)
			if len(staticRelays) > 0 {
				opts = append(opts, libp2p.EnableAutoRelayWithStaticRelays(staticRelays,
					autorelay.WithBackoff(30*time.Second),
					autorelay.WithMinInterval(time.Minute),
				))
				logger.Info("auto relay enabled with static relays (DHT disabled)", "count", len(staticRelays))
			}
		} else {
			logger.Warn("auto relay disabled: requires DHT or static relays for peer discovery")
		}
	}

	// Add PSK if authentication is enabled
	if config.EnableAuth && config.PSK != "" {
		psk, err := security.DecodePSK(config.PSK)
		if err != nil {
			return nil, types.WrapError(err, "failed to decode PSK")
		}

		// PSK is just a []byte in libp2p
		opts = append(opts, libp2p.PrivateNetwork(pnet.PSK(psk)))
		logger.Info("PSK authentication enabled")
	}

	// Add connection gating if trusted peers are specified
	if len(config.TrustedPeers) > 0 {
		gater := newConnectionGater(config.TrustedPeers, logger)
		opts = append(opts, libp2p.ConnectionGater(gater))
		logger.Info("connection gating enabled", "trusted_peers", len(config.TrustedPeers))
	}

	// Create libp2p host
	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, types.WrapError(err, "failed to create libp2p host")
	}

	logger.Info("libp2p host created",
		"id", h.ID().String(),
		"addrs", h.Addrs(),
		"psk_enabled", config.EnableAuth && config.PSK != "",
		"trusted_peers", len(config.TrustedPeers),
		"dht_enabled", !config.DisableDHT,
	)

	// Bootstrap DHT if enabled
	if !config.DisableDHT && kadDHT != nil {
		if err := kadDHT.Bootstrap(ctx); err != nil {
			logger.Warn("failed to bootstrap DHT", "error", err)
		} else {
			logger.Info("DHT bootstrap started")
		}
	}

	// Connect to bootstrap peers
	// If DHT is enabled and no bootstrap peers are configured, use default IPFS bootstrap nodes
	bootstrapPeers := config.BootstrapPeers
	if !config.DisableDHT && len(bootstrapPeers) == 0 {
		bootstrapPeers = DefaultBootstrapPeers
		logger.Info("no bootstrap peers configured, using default IPFS bootstrap nodes")
	}

	if len(bootstrapPeers) > 0 {
		logger.Info("connecting to bootstrap peers", "count", len(bootstrapPeers))
		go connectToBootstrapPeers(ctx, h, bootstrapPeers, logger)
	}

	return &Host{
		host:   h,
		dht:    kadDHT,
		logger: logger,
	}, nil
}

// ID returns the host's peer ID
func (h *Host) ID() string {
	return h.host.ID().String()
}

// LibP2PHost returns the underlying libp2p host for advanced usage
func (h *Host) LibP2PHost() host.Host {
	return h.host
}

// DHT returns the DHT routing table (may be nil if DHT is disabled)
func (h *Host) DHT() *dht.IpfsDHT {
	return h.dht
}

// Addrs returns the host's listening addresses
func (h *Host) Addrs() []string {
	addrs := h.host.Addrs()
	result := make([]string, len(addrs))
	for i, addr := range addrs {
		result[i] = addr.String()
	}
	return result
}

// Connect establishes a connection to a peer
func (h *Host) Connect(ctx context.Context, addr string) error {
	maddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return types.WrapError(err, "invalid multiaddr")
	}

	peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return types.WrapError(err, "failed to parse peer info")
	}

	if err := h.host.Connect(ctx, *peerInfo); err != nil {
		return types.WrapError(err, "failed to connect to peer")
	}

	h.logger.Info("connected to peer", "peer", peerInfo.ID)

	return nil
}

// NewStream creates a new stream to a peer
func (h *Host) NewStream(ctx context.Context, peerID string, protocolID string) (types.Stream, error) {
	pid, err := peer.Decode(peerID)
	if err != nil {
		return nil, types.WrapError(err, "invalid peer ID")
	}

	stream, err := h.host.NewStream(ctx, pid, protocol.ID(protocolID))
	if err != nil {
		return nil, types.WrapError(err, "failed to create stream")
	}

	return &streamWrapper{stream: stream}, nil
}

// SetStreamHandler registers a handler for incoming streams
func (h *Host) SetStreamHandler(protocolID string, handler types.StreamHandler) {
	h.host.SetStreamHandler(protocol.ID(protocolID), func(s network.Stream) {
		handler(&streamWrapper{stream: s})
	})
}

// Close shuts down the host
func (h *Host) Close() error {
	return h.host.Close()
}

// EnableMDNS enables mDNS discovery
func (h *Host) EnableMDNS(ctx context.Context) error {
	service := mdns.NewMdnsService(h.host, "p2p-playground", &discoveryNotifee{
		h:      h.host,
		logger: h.logger,
	})

	if err := service.Start(); err != nil {
		return types.WrapError(err, "failed to start mDNS")
	}

	h.logger.Info("mDNS discovery enabled")
	return nil
}

// PeerInfo contains information about a peer
type PeerInfo struct {
	ID    string
	Addrs []string
}

// Peers returns a list of connected peers
func (h *Host) Peers() []PeerInfo {
	peers := h.host.Network().Peers()
	result := make([]PeerInfo, 0, len(peers))

	for _, p := range peers {
		conns := h.host.Network().ConnsToPeer(p)
		addrs := make([]string, 0)
		for _, conn := range conns {
			addrs = append(addrs, conn.RemoteMultiaddr().String())
		}

		result = append(result, PeerInfo{
			ID:    p.String(),
			Addrs: addrs,
		})
	}

	return result
}

// NetworkStats contains network diagnostic information
type NetworkStats struct {
	ConnectedPeers  int
	DHTRoutingTable int
	DHTMode         string
}

// GetNetworkStats returns current network statistics
func (h *Host) GetNetworkStats() NetworkStats {
	stats := NetworkStats{
		ConnectedPeers: len(h.host.Network().Peers()),
	}

	if h.dht != nil {
		stats.DHTRoutingTable = h.dht.RoutingTable().Size()
		// Convert DHT mode to string
		mode := h.dht.Mode()
		switch mode {
		case dht.ModeServer:
			stats.DHTMode = "server"
		case dht.ModeClient:
			stats.DHTMode = "client"
		case dht.ModeAuto:
			stats.DHTMode = "auto"
		case dht.ModeAutoServer:
			stats.DHTMode = "autoserver"
		default:
			stats.DHTMode = "unknown"
		}
	}

	return stats
}

// StartDiagnosticLogging starts periodic logging of network status
func (h *Host) StartDiagnosticLogging(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				stats := h.GetNetworkStats()
				h.logger.Info("network status",
					"connected_peers", stats.ConnectedPeers,
					"dht_routing_table_size", stats.DHTRoutingTable,
					"dht_mode", stats.DHTMode,
				)

				// Log peer details if there are connections
				peers := h.Peers()
				if len(peers) > 0 {
					for _, p := range peers {
						h.logger.Debug("connected peer", "id", p.ID, "addrs", p.Addrs)
					}
				}
			}
		}
	}()
}

// discoveryNotifee handles peer discovery
type discoveryNotifee struct {
	h      host.Host
	logger types.Logger
}

func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.logger.Info("discovered peer via mDNS", "peer", pi.ID)

	if err := n.h.Connect(context.Background(), pi); err != nil {
		n.logger.Warn("failed to connect to discovered peer",
			"peer", pi.ID,
			"error", err,
		)
	}
}

// streamWrapper wraps libp2p stream to implement types.Stream
type streamWrapper struct {
	stream network.Stream
}

func (s *streamWrapper) Read(p []byte) (n int, err error) {
	return s.stream.Read(p)
}

func (s *streamWrapper) Write(p []byte) (n int, err error) {
	return s.stream.Write(p)
}

func (s *streamWrapper) Close() error {
	return s.stream.Close()
}

func (s *streamWrapper) Reset() error {
	return s.stream.Reset()
}

// connectionGater implements connection gating based on trusted peers
type connectionGater struct {
	trustedPeers map[peer.ID]bool
	logger       types.Logger
}

// newConnectionGater creates a new connection gater
func newConnectionGater(trustedPeerIDs []string, logger types.Logger) *connectionGater {
	trustedMap := make(map[peer.ID]bool)
	for _, pidStr := range trustedPeerIDs {
		pid, err := peer.Decode(pidStr)
		if err != nil {
			logger.Warn("invalid trusted peer ID", "peer", pidStr, "error", err)
			continue
		}
		trustedMap[pid] = true
	}

	return &connectionGater{
		trustedPeers: trustedMap,
		logger:       logger,
	}
}

// InterceptPeerDial is called before dialing a peer
func (g *connectionGater) InterceptPeerDial(p peer.ID) bool {
	// If no trusted peers configured, allow all
	if len(g.trustedPeers) == 0 {
		return true
	}

	// Check if peer is trusted
	if g.trustedPeers[p] {
		return true
	}

	g.logger.Warn("blocked outbound connection to untrusted peer", "peer", p)
	return false
}

// InterceptAddrDial is called before dialing an address
func (g *connectionGater) InterceptAddrDial(_ peer.ID, _ multiaddr.Multiaddr) bool {
	return true // Let InterceptPeerDial handle the decision
}

// InterceptAccept is called when accepting an inbound connection
func (g *connectionGater) InterceptAccept(addrs network.ConnMultiaddrs) bool {
	return true // Let InterceptSecured handle the decision
}

// InterceptSecured is called after the connection has been secured
func (g *connectionGater) InterceptSecured(_ network.Direction, p peer.ID, _ network.ConnMultiaddrs) bool {
	// If no trusted peers configured, allow all
	if len(g.trustedPeers) == 0 {
		return true
	}

	// Check if peer is trusted
	if g.trustedPeers[p] {
		return true
	}

	g.logger.Warn("blocked connection from untrusted peer", "peer", p)
	return false
}

// InterceptUpgraded is called after the connection has been upgraded
func (g *connectionGater) InterceptUpgraded(_ network.Conn) (bool, control.DisconnectReason) {
	return true, 0
}

// connectToBootstrapPeers connects to bootstrap peers in the background
func connectToBootstrapPeers(ctx context.Context, h host.Host, bootstrapPeers []string, logger types.Logger) {
	var wg sync.WaitGroup

	for _, addrStr := range bootstrapPeers {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()

			// Parse multiaddr
			maddr, err := multiaddr.NewMultiaddr(addr)
			if err != nil {
				logger.Warn("invalid bootstrap peer address", "addr", addr, "error", err)
				return
			}

			// Extract peer info
			peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
			if err != nil {
				logger.Warn("failed to parse bootstrap peer info", "addr", addr, "error", err)
				return
			}

			// Connect with timeout
			connectCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			if err := h.Connect(connectCtx, *peerInfo); err != nil {
				logger.Warn("failed to connect to bootstrap peer", "peer", peerInfo.ID, "error", err)
				return
			}

			logger.Info("connected to bootstrap peer", "peer", peerInfo.ID)
		}(addrStr)
	}

	wg.Wait()
	logger.Info("bootstrap peer connections completed")
}

// parseStaticRelays parses static relay addresses into peer.AddrInfo structs
func parseStaticRelays(relayAddrs []string, logger types.Logger) []peer.AddrInfo {
	var relays []peer.AddrInfo

	for _, addrStr := range relayAddrs {
		// Parse multiaddr
		maddr, err := multiaddr.NewMultiaddr(addrStr)
		if err != nil {
			logger.Warn("invalid static relay address", "addr", addrStr, "error", err)
			continue
		}

		// Extract peer info
		peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			logger.Warn("failed to parse static relay peer info", "addr", addrStr, "error", err)
			continue
		}

		relays = append(relays, *peerInfo)
		logger.Debug("added static relay", "peer", peerInfo.ID, "addrs", peerInfo.Addrs)
	}

	return relays
}
