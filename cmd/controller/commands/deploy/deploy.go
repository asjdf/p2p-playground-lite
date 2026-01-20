package deploy

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/asjdf/p2p-playground-lite/cmd/controller/commands/common"
	"github.com/spf13/cobra"
)

var (
	nodeID    string
	autoStart bool
)

// Cmd represents the deploy command
var Cmd = &cobra.Command{
	Use:   "deploy [package]",
	Short: "Deploy an application package",
	Long: `Deploy an application package to a target node.

If --node is not specified, the package will be deployed to the first discovered node.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		packagePath := args[0]
		fmt.Printf("Deploying package: %s\n", packagePath)

		// Check if file exists
		fileInfo, err := os.Stat(packagePath)
		if err != nil {
			return fmt.Errorf("failed to access package file: %w", err)
		}

		// Create P2P host using configuration
		ctx := context.Background()
		host, err := common.CreateP2PHost(ctx)
		if err != nil {
			return err
		}
		defer func() { _ = host.Close() }()

		fmt.Printf("Controller ID: %s\n", host.ID())

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

		// Deploy package
		fmt.Println("\nDeploying package...")
		appID, err := common.DeployPackage(ctx, host, targetPeerID, packagePath, fileInfo.Size(), autoStart, common.GlobalLogger)
		if err != nil {
			return fmt.Errorf("deployment failed: %w", err)
		}

		fmt.Printf("\n✓ Deployment successful!\n")
		fmt.Printf("  Application ID: %s\n", appID)
		if autoStart {
			fmt.Printf("  Status: Started\n")
		} else {
			fmt.Printf("  Status: Deployed (not started)\n")
		}

		return nil
	},
}

func init() {
	Cmd.Flags().StringVar(&nodeID, "node", "", "target node peer ID")
	Cmd.Flags().BoolVar(&autoStart, "start", true, "automatically start the application after deployment")
}
