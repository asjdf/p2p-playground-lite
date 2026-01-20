package discovery

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/asjdf/p2p-playground-lite/pkg/types"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
)

const (
	// DiscoveryTopic is the pubsub topic for node discovery
	DiscoveryTopic = "p2p-playground/discovery"

	// AnnounceInterval is how often nodes announce themselves
	AnnounceInterval = 10 * time.Second

	// NodeTimeout is how long before a node is considered offline
	NodeTimeout = 30 * time.Second
)

// NodeAnnouncement is broadcast by nodes to announce their presence
type NodeAnnouncement struct {
	PeerID    string            `json:"peer_id"`
	Name      string            `json:"name"`
	Labels    map[string]string `json:"labels,omitempty"`
	Addrs     []string          `json:"addrs"`
	Version   string            `json:"version,omitempty"`
	Timestamp int64             `json:"timestamp"`
}

// DiscoveredNode represents a discovered p2p-playground node
type DiscoveredNode struct {
	PeerID   peer.ID
	Name     string
	Labels   map[string]string
	Addrs    []string
	Version  string
	LastSeen time.Time
}

// Service handles node discovery via pubsub
type Service struct {
	host   host.Host
	pubsub *pubsub.PubSub
	topic  *pubsub.Topic
	sub    *pubsub.Subscription
	logger types.Logger

	// DHT-based peer discovery
	routingDiscovery *drouting.RoutingDiscovery

	// Node info for announcements
	nodeName   string
	nodeLabels map[string]string
	version    string

	// Discovered nodes
	nodes   map[peer.ID]*DiscoveredNode
	nodesMu sync.RWMutex

	// Callbacks
	onNodeDiscovered func(*DiscoveredNode)
	onNodeLost       func(peer.ID)

	ctx    context.Context
	cancel context.CancelFunc
}

// Config contains discovery service configuration
type Config struct {
	NodeName   string
	NodeLabels map[string]string
	Version    string
	Routing    routing.ContentRouting // Optional: DHT routing for peer discovery
}

// NewService creates a new discovery service
func NewService(h host.Host, logger types.Logger, cfg *Config) (*Service, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create pubsub with gossipsub
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		cancel()
		return nil, err
	}

	// Join the discovery topic
	topic, err := ps.Join(DiscoveryTopic)
	if err != nil {
		cancel()
		return nil, err
	}

	// Subscribe to the topic
	sub, err := topic.Subscribe()
	if err != nil {
		cancel()
		return nil, err
	}

	s := &Service{
		host:       h,
		pubsub:     ps,
		topic:      topic,
		sub:        sub,
		logger:     logger,
		nodeName:   cfg.NodeName,
		nodeLabels: cfg.NodeLabels,
		version:    cfg.Version,
		nodes:      make(map[peer.ID]*DiscoveredNode),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Set up DHT-based routing discovery if routing is provided
	if cfg.Routing != nil {
		s.routingDiscovery = drouting.NewRoutingDiscovery(cfg.Routing)
		logger.Info("DHT-based peer discovery enabled for topic", "topic", DiscoveryTopic)
	}

	return s, nil
}

// SetOnNodeDiscovered sets the callback for when a new node is discovered
func (s *Service) SetOnNodeDiscovered(cb func(*DiscoveredNode)) {
	s.onNodeDiscovered = cb
}

// SetOnNodeLost sets the callback for when a node is lost
func (s *Service) SetOnNodeLost(cb func(peer.ID)) {
	s.onNodeLost = cb
}

// Start begins the discovery service
func (s *Service) Start() {
	// Start listening for announcements
	go s.listenLoop()

	// Start announcing ourselves
	go s.announceLoop()

	// Start cleanup loop for stale nodes
	go s.cleanupLoop()

	// Start DHT-based peer discovery if routing is available
	if s.routingDiscovery != nil {
		go s.dhtPeerDiscoveryLoop()
	}

	s.logger.Info("discovery service started", "topic", DiscoveryTopic)
}

// Stop stops the discovery service
func (s *Service) Stop() {
	s.cancel()
	s.sub.Cancel()
	if err := s.topic.Close(); err != nil {
		s.logger.Warn("failed to close topic", "error", err)
	}
	s.logger.Info("discovery service stopped")
}

// GetNodes returns all discovered nodes
func (s *Service) GetNodes() []*DiscoveredNode {
	s.nodesMu.RLock()
	defer s.nodesMu.RUnlock()

	nodes := make([]*DiscoveredNode, 0, len(s.nodes))
	for _, node := range s.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// GetNode returns a specific node by peer ID
func (s *Service) GetNode(peerID peer.ID) *DiscoveredNode {
	s.nodesMu.RLock()
	defer s.nodesMu.RUnlock()
	return s.nodes[peerID]
}

// Announce broadcasts our presence to the network
func (s *Service) Announce() error {
	addrs := s.host.Addrs()
	addrStrs := make([]string, len(addrs))
	for i, addr := range addrs {
		addrStrs[i] = addr.String()
	}

	announcement := NodeAnnouncement{
		PeerID:    s.host.ID().String(),
		Name:      s.nodeName,
		Labels:    s.nodeLabels,
		Addrs:     addrStrs,
		Version:   s.version,
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(announcement)
	if err != nil {
		return err
	}

	return s.topic.Publish(s.ctx, data)
}

// listenLoop listens for node announcements
func (s *Service) listenLoop() {
	for {
		msg, err := s.sub.Next(s.ctx)
		if err != nil {
			if s.ctx.Err() != nil {
				return // Context cancelled
			}
			s.logger.Warn("error receiving message", "error", err)
			continue
		}

		// Ignore our own messages
		if msg.ReceivedFrom == s.host.ID() {
			continue
		}

		var announcement NodeAnnouncement
		if err := json.Unmarshal(msg.Data, &announcement); err != nil {
			s.logger.Warn("invalid announcement", "error", err, "from", msg.ReceivedFrom)
			continue
		}

		peerID, err := peer.Decode(announcement.PeerID)
		if err != nil {
			s.logger.Warn("invalid peer ID in announcement", "error", err)
			continue
		}

		s.handleAnnouncement(peerID, &announcement)
	}
}

// handleAnnouncement processes a node announcement
func (s *Service) handleAnnouncement(peerID peer.ID, announcement *NodeAnnouncement) {
	s.nodesMu.Lock()
	defer s.nodesMu.Unlock()

	existing := s.nodes[peerID]
	isNew := existing == nil

	node := &DiscoveredNode{
		PeerID:   peerID,
		Name:     announcement.Name,
		Labels:   announcement.Labels,
		Addrs:    announcement.Addrs,
		Version:  announcement.Version,
		LastSeen: time.Now(),
	}
	s.nodes[peerID] = node

	if isNew {
		s.logger.Info("discovered new node",
			"peer_id", peerID,
			"name", announcement.Name,
			"addrs", announcement.Addrs,
		)
		if s.onNodeDiscovered != nil {
			go s.onNodeDiscovered(node)
		}
	}
}

// announceLoop periodically announces our presence
func (s *Service) announceLoop() {
	// Announce immediately
	if err := s.Announce(); err != nil {
		s.logger.Warn("failed to announce", "error", err)
	}

	ticker := time.NewTicker(AnnounceInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			if err := s.Announce(); err != nil {
				s.logger.Warn("failed to announce", "error", err)
			}
		}
	}
}

// cleanupLoop removes stale nodes
func (s *Service) cleanupLoop() {
	ticker := time.NewTicker(NodeTimeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.cleanupStaleNodes()
		}
	}
}

// cleanupStaleNodes removes nodes that haven't announced recently
func (s *Service) cleanupStaleNodes() {
	s.nodesMu.Lock()
	defer s.nodesMu.Unlock()

	now := time.Now()
	for peerID, node := range s.nodes {
		if now.Sub(node.LastSeen) > NodeTimeout {
			delete(s.nodes, peerID)
			s.logger.Info("node lost", "peer_id", peerID, "name", node.Name)
			if s.onNodeLost != nil {
				go s.onNodeLost(peerID)
			}
		}
	}
}

// dhtPeerDiscoveryLoop uses DHT to discover peers subscribed to the same topic
func (s *Service) dhtPeerDiscoveryLoop() {
	// Advertise ourselves as a provider for this topic
	dutil.Advertise(s.ctx, s.routingDiscovery, DiscoveryTopic)
	s.logger.Info("advertising topic via DHT", "topic", DiscoveryTopic)

	// Periodically find peers
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.findDHTPeers()
		}
	}
}

// findDHTPeers finds and connects to peers via DHT
func (s *Service) findDHTPeers() {
	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()

	peerChan, err := s.routingDiscovery.FindPeers(ctx, DiscoveryTopic)
	if err != nil {
		s.logger.Warn("failed to find peers via DHT", "error", err)
		return
	}

	foundCount := 0
	for p := range peerChan {
		if p.ID == s.host.ID() {
			continue // Skip ourselves
		}
		if len(p.Addrs) == 0 {
			continue // Skip peers without addresses
		}

		// Check if already connected
		if s.host.Network().Connectedness(p.ID) == 1 { // Connected
			continue
		}

		// Try to connect
		if err := s.host.Connect(ctx, p); err != nil {
			s.logger.Debug("failed to connect to DHT peer", "peer", p.ID, "error", err)
			continue
		}

		foundCount++
		s.logger.Info("connected to peer via DHT discovery", "peer", p.ID, "addrs", p.Addrs)
	}

	if foundCount > 0 {
		s.logger.Info("DHT peer discovery completed", "new_connections", foundCount)
	}
}
