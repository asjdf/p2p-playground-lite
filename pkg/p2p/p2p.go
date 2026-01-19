package p2p

import (
	"context"
	"fmt"

	"github.com/asjdf/p2p-playground-lite/pkg/types"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/multiformats/go-multiaddr"
)

// Host wraps libp2p host
type Host struct {
	host   host.Host
	logger types.Logger
}

// NewHost creates a new P2P host
func NewHost(ctx context.Context, listenAddrs []string, logger types.Logger) (*Host, error) {
	// Parse listen addresses
	var maddrs []multiaddr.Multiaddr
	for _, addr := range listenAddrs {
		maddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return nil, fmt.Errorf("invalid multiaddr %s: %w", addr, err)
		}
		maddrs = append(maddrs, maddr)
	}

	// Create libp2p host with basic options (no QUIC for now)
	h, err := libp2p.New(
		libp2p.ListenAddrs(maddrs...),
	)
	if err != nil {
		return nil, types.WrapError(err, "failed to create libp2p host")
	}

	logger.Info("libp2p host created",
		"id", h.ID().String(),
		"addrs", h.Addrs(),
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

