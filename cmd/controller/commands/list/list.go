package list

import (
	"context"
	"fmt"
	"time"

	"github.com/asjdf/p2p-playground-lite/cmd/controller/commands/common"
	"github.com/spf13/cobra"
)

var (
	nodeID string
)

// Cmd represents the list command
var Cmd = &cobra.Command{
	Use:   "list",
	Short: "List deployed applications",
	Long: `List all deployed applications on a target node.

If --node is not specified, applications from the first discovered node will be listed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Listing applications...")

		// Create P2P host using configuration
		ctx := context.Background()
		host, err := common.CreateP2PHost(ctx)
		if err != nil {
			return err
		}
		defer func() { _ = host.Close() }()

		// Wait for peer discovery
		fmt.Println("Discovering nodes...")
		time.Sleep(3 * time.Second)

		// Get target node
		var targetPeerID string
		if nodeID != "" {
			targetPeerID = nodeID
			fmt.Printf("Using specified node: %s\n", targetPeerID)
		} else {
			// Use first discovered peer
			peers := host.Peers()
			if len(peers) == 0 {
				return fmt.Errorf("no nodes discovered")
			}
			targetPeerID = peers[0].ID
			fmt.Printf("Using discovered node: %s\n", targetPeerID)
		}

		// List applications
		fmt.Println("\nFetching applications...")
		apps, err := common.ListApplications(ctx, host, targetPeerID, common.GlobalLogger)
		if err != nil {
			return fmt.Errorf("failed to list applications: %w", err)
		}

		// Display results
		fmt.Printf("\nFound %d application(s):\n\n", len(apps))
		if len(apps) == 0 {
			fmt.Println("  (no applications deployed)")
			return nil
		}

		for i, app := range apps {
			fmt.Printf("%d. Application: %s\n", i+1, app.Name)
			fmt.Printf("   ID: %s\n", app.ID)
			fmt.Printf("   Version: %s\n", app.Version)
			fmt.Printf("   Status: %s\n", app.Status)
			if app.PID > 0 {
				fmt.Printf("   PID: %d\n", app.PID)
			}
			if !app.StartedAt.IsZero() {
				fmt.Printf("   Started: %s\n", app.StartedAt.Format("2006-01-02 15:04:05"))
			}
			if len(app.Labels) > 0 {
				fmt.Printf("   Labels: %v\n", app.Labels)
			}
			fmt.Println()
		}

		return nil
	},
}

func init() {
	Cmd.Flags().StringVar(&nodeID, "node", "", "target node peer ID")
}
