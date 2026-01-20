package logs

import (
	"context"
	"fmt"
	"time"

	"github.com/asjdf/p2p-playground-lite/cmd/controller/commands/common"
	"github.com/spf13/cobra"
)

var (
	nodeID string
	follow bool
	tail   int
)

// Cmd represents the logs command
var Cmd = &cobra.Command{
	Use:   "logs [app-id]",
	Short: "View application logs",
	Long: `View logs from a deployed application.

If --node is not specified, logs will be fetched from the first discovered node.
Use --tail to limit the number of lines shown.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appID := args[0]
		fmt.Printf("Fetching logs for application: %s\n", appID)

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

		// Fetch logs
		fmt.Println("\nFetching logs...")
		logsContent, err := common.FetchLogs(ctx, host, targetPeerID, appID, follow, tail, common.GlobalLogger)
		if err != nil {
			return fmt.Errorf("failed to fetch logs: %w", err)
		}

		// Display logs
		fmt.Println(logsContent)

		return nil
	},
}

func init() {
	Cmd.Flags().StringVar(&nodeID, "node", "", "target node peer ID")
	Cmd.Flags().BoolVarP(&follow, "follow", "f", false, "follow log output")
	Cmd.Flags().IntVar(&tail, "tail", 50, "number of lines to show from the end")
}
