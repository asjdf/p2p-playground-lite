package nodes

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/asjdf/p2p-playground-lite/cmd/controller/commands/common"
	"github.com/asjdf/p2p-playground-lite/pkg/discovery"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/spf13/cobra"
)

// Cmd represents the nodes command
var Cmd = &cobra.Command{
	Use:   "nodes",
	Short: "Discover and list P2P Playground nodes",
	Long: `Continuously discover P2P Playground nodes using gossip protocol.

This command will keep running until interrupted (Ctrl+C).
It discovers nodes that are running the p2p-playground daemon.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Discovering P2P Playground nodes...")

		// Create P2P host using configuration
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		host, err := common.CreateP2PHost(ctx)
		if err != nil {
			return err
		}
		defer func() { _ = host.Close() }()

		fmt.Printf("Controller ID: %s\n", host.ID())
		fmt.Printf("Controller addresses:\n")
		for _, addr := range host.Addrs() {
			fmt.Printf("  - %s\n", addr)
		}
		fmt.Println()

		// Initialize discovery service
		discoverySvc, err := discovery.NewService(host.LibP2PHost(), common.GlobalLogger, &discovery.Config{
			NodeName:   "controller",
			NodeLabels: nil,
			Version:    "0.1.0",
			Routing:    host.DHT(),
		})
		if err != nil {
			return fmt.Errorf("failed to create discovery service: %w", err)
		}

		// Track discovered nodes
		discoveredNodes := make(map[string]*discovery.DiscoveredNode)

		discoverySvc.SetOnNodeDiscovered(func(node *discovery.DiscoveredNode) {
			discoveredNodes[node.PeerID.String()] = node
			fmt.Printf("\n✓ New node discovered:\n")
			fmt.Printf("  Peer ID: %s\n", node.PeerID)
			fmt.Printf("  Name: %s\n", node.Name)
			if len(node.Labels) > 0 {
				fmt.Printf("  Labels: %v\n", node.Labels)
			}
			fmt.Printf("  Addresses: %v\n", node.Addrs)
			fmt.Printf("  (Total nodes: %d)\n", len(discoveredNodes))
		})

		discoverySvc.SetOnNodeLost(func(peerID peer.ID) {
			if node, ok := discoveredNodes[peerID.String()]; ok {
				fmt.Printf("\n✗ Node lost: %s (%s)\n", node.Name, peerID)
				delete(discoveredNodes, peerID.String())
				fmt.Printf("  (Total nodes: %d)\n", len(discoveredNodes))
			}
		})

		discoverySvc.Start()
		defer discoverySvc.Stop()

		fmt.Println("\nListening for P2P Playground nodes... (Press Ctrl+C to stop)")
		fmt.Println("Nodes will announce themselves every 10 seconds.")

		// Wait for interrupt signal
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		fmt.Println("\n\nStopping discovery...")

		// Print final summary
		nodes := discoverySvc.GetNodes()
		if len(nodes) == 0 {
			fmt.Println("\nNo P2P Playground nodes discovered.")
		} else {
			fmt.Printf("\nDiscovered %d P2P Playground node(s):\n", len(nodes))
			for i, node := range nodes {
				fmt.Printf("%d. %s (%s)\n", i+1, node.Name, node.PeerID)
				fmt.Printf("   Labels: %v\n", node.Labels)
				fmt.Printf("   Addresses: %v\n", node.Addrs)
				fmt.Printf("   Last seen: %s\n", node.LastSeen.Format("15:04:05"))
			}
		}

		return nil
	},
}
