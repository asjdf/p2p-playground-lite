package p2p

import (
	"context"
	"fmt"

	"github.com/asjdf/p2p-playground-lite/pkg/security"
	"github.com/asjdf/p2p-playground-lite/pkg/types"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-libp2p/core/pnet"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/multiformats/go-multiaddr"
)

// Host wraps libp2p host
type Host struct {
	host   host.Host
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
	)

	return &Host{
		host:   h,
		logger: logger,
	}, nil
}

// ID returns the host's peer ID
func (h *Host) ID() string {
	return h.host.ID().String()
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

