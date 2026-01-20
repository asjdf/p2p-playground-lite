package run

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/asjdf/p2p-playground-lite/cmd/controller/commands/common"
	"github.com/asjdf/p2p-playground-lite/pkg/consts"
	"github.com/asjdf/p2p-playground-lite/pkg/p2p"
	pkgmanager "github.com/asjdf/p2p-playground-lite/pkg/package"
	"github.com/asjdf/p2p-playground-lite/pkg/security"
	"github.com/asjdf/p2p-playground-lite/pkg/types"
	"github.com/spf13/cobra"
)

var (
	nodeID     string
	cleanup    bool
	noSign     bool
	privateKey string
)

// Cmd represents the run command
var Cmd = &cobra.Command{
	Use:   "run [app-directory]",
	Short: "Build, deploy and run an application (like go run)",
	Long: `Build an application package, deploy it to nodes, and monitor logs in real-time.

This command is similar to 'go run' but for P2P Playground applications:
1. Discovers available nodes in the network
2. Builds the application package from the directory
3. Optionally signs the package
4. Deploys to all discovered nodes (or specified node with --node)
5. Streams logs in real-time with format: [node-id] original log

By default, the application is deployed to ALL discovered nodes in the network.
Use --node to deploy to a specific node only.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appDir := args[0]
		ctx := context.Background()

		// Verify app directory exists and has manifest
		manifestPath := filepath.Join(appDir, "manifest.yaml")
		if _, err := os.Stat(manifestPath); err != nil {
			return fmt.Errorf("manifest.yaml not found in %s: %w", appDir, err)
		}

		// Simply display the directory for now
		fmt.Printf("Building and running application from: %s\n", appDir)

		// Create P2P host
		host, err := common.CreateP2PHost(ctx)
		if err != nil {
			return err
		}
		defer func() { _ = host.Close() }()

		fmt.Printf("Controller ID: %s\n", host.ID())

		// Discover nodes
		fmt.Println("\nDiscovering nodes...")
		time.Sleep(3 * time.Second)

		var targetPeerIDs []string
		if nodeID != "" {
			targetPeerIDs = []string{nodeID}
			fmt.Printf("Using specified node: %s\n", nodeID)
		} else {
			peers := host.Peers()
			if len(peers) == 0 {
				return fmt.Errorf("no nodes discovered")
			}

			// List all discovered nodes
			fmt.Printf("\nDiscovered %d node(s):\n", len(peers))
			for i, peer := range peers {
				fmt.Printf("%d. Peer ID: %s\n", i+1, peer.ID)
				fmt.Printf("   Addresses:\n")
				for _, addr := range peer.Addrs {
					fmt.Printf("     - %s\n", addr)
				}
				targetPeerIDs = append(targetPeerIDs, peer.ID)
			}
			fmt.Println()
			fmt.Printf("Deploying to all %d node(s)\n", len(targetPeerIDs))
		}

		// Build package
		fmt.Println("\nBuilding application package...")
		pkgMgr := pkgmanager.New()
		pkgPath, err := pkgMgr.Pack(ctx, appDir)
		if err != nil {
			return fmt.Errorf("failed to build package: %w", err)
		}
		fmt.Printf("Package created: %s\n", pkgPath)

		// Cleanup package file after deployment if requested
		if cleanup {
			defer func() {
				_ = os.Remove(pkgPath)
				_ = os.Remove(pkgPath + ".sig")
			}()
		}

		// Sign package if requested
		var signature []byte
		if !noSign && privateKey != "" {
			fmt.Println("\nSigning package...")
			signer, err := security.LoadSigner(privateKey)
			if err != nil {
				return fmt.Errorf("failed to load private key: %w", err)
			}

			signature, err = signer.SignFile(pkgPath)
			if err != nil {
				return fmt.Errorf("failed to sign package: %w", err)
			}

			// Save signature
			sigPath := pkgPath + ".sig"
			if err := os.WriteFile(sigPath, signature, 0644); err != nil {
				common.GlobalLogger.Warn("failed to save signature file", "error", err)
			} else {
				common.GlobalLogger.Info("package signed", "sig_path", sigPath)
			}
		} else if !noSign {
			common.GlobalLogger.Warn("no private key specified, deploying without signature")
		}

		// Get package info
		fileInfo, err := os.Stat(pkgPath)
		if err != nil {
			return fmt.Errorf("failed to get package info: %w", err)
		}

		// Deploy package to all target nodes
		fmt.Printf("\nDeploying package to %d node(s)...\n", len(targetPeerIDs))

		type deploymentResult struct {
			peerID string
			appID  string
			err    error
		}

		results := make(chan deploymentResult, len(targetPeerIDs))

		for _, peerID := range targetPeerIDs {
			go func(pid string) {
				appID, err := common.DeployPackage(ctx, host, pid, pkgPath, fileInfo.Size(), true, common.GlobalLogger)
				results <- deploymentResult{peerID: pid, appID: appID, err: err}
			}(peerID)
		}

		// Collect deployment results
		deployments := make(map[string]string) // peerID -> appID
		var deployErrors []error

		for i := 0; i < len(targetPeerIDs); i++ {
			result := <-results
			if result.err != nil {
				deployErrors = append(deployErrors, fmt.Errorf("node %s: %w", result.peerID, result.err))
			} else {
				deployments[result.peerID] = result.appID
				fmt.Printf("  ✓ Deployed to node: %s (app: %s)\n", result.peerID, result.appID)
			}
		}

		if len(deployErrors) > 0 {
			fmt.Println("\nDeployment errors:")
			for _, err := range deployErrors {
				fmt.Printf("  ✗ %v\n", err)
			}
		}

		if len(deployments) == 0 {
			return fmt.Errorf("failed to deploy to any nodes")
		}

		fmt.Printf("\n✓ Application deployed and started on %d node(s)!\n\n", len(deployments))

		// Stream logs from all deployed nodes
		fmt.Println("Streaming logs from all nodes (Ctrl+C to stop):")
		fmt.Println("─────────────────────────────────────────────────────────────")

		// Create a context that we can cancel
		logsCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		// Set up signal handling for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		// Start log streaming for each node in separate goroutines
		for peerID, appID := range deployments {
			go func(pid, aid string) {
				if err := streamLogs(logsCtx, host, pid, aid, common.GlobalLogger); err != nil {
					// Only log errors if context wasn't cancelled
					if logsCtx.Err() == nil {
						common.GlobalLogger.Warn("log streaming stopped", "peer", pid, "error", err)
					}
				}
			}(peerID, appID)
		}

		// Wait for interrupt signal
		<-sigChan
		fmt.Println("\n\nReceived interrupt signal, stopping...")

		return nil
	},
}

// streamLogs streams logs from the application with [node-id] prefix
func streamLogs(ctx context.Context, host *p2p.Host, peerID string, appID string, logger types.Logger) error {
	// Create stream to target peer
	stream, err := host.NewStream(ctx, peerID, consts.LogsProtocolID)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}
	defer func() { _ = stream.Close() }()

	// Prepare logs request (follow mode)
	req := common.LogsRequest{
		AppID:  appID,
		Follow: true,
		Tail:   0, // Get all logs
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request header size
	reqSize := uint32(len(reqBytes))
	if err := binary.Write(stream, binary.BigEndian, reqSize); err != nil {
		return fmt.Errorf("failed to send header size: %w", err)
	}

	// Send request header
	if _, err := stream.Write(reqBytes); err != nil {
		return fmt.Errorf("failed to send header: %w", err)
	}

	logger.Info("requesting logs", "app_id", appID, "follow", true)

	// Read response header size
	var respSize uint32
	if err := binary.Read(stream, binary.BigEndian, &respSize); err != nil {
		return fmt.Errorf("failed to read response size: %w", err)
	}

	// Read response
	respBytes := make([]byte, respSize)
	if _, err := io.ReadFull(stream, respBytes); err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var resp common.LogsResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("logs request failed on node: %s", resp.Error)
	}

	// Shorten peer ID for display (first 8 characters)
	shortPeerID := peerID
	if len(peerID) > 8 {
		shortPeerID = peerID[:8]
	}

	// Output initial logs with prefix
	if resp.Logs != "" {
		lines := strings.Split(strings.TrimSpace(resp.Logs), "\n")
		for _, line := range lines {
			if line != "" {
				fmt.Printf("[%s] %s\n", shortPeerID, line)
			}
		}
	}

	// For follow mode, keep reading from stream
	// Note: Current implementation returns all logs at once
	// In a real streaming implementation, we would keep reading chunks
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			fmt.Printf("[%s] %s\n", shortPeerID, line)
		}
	}

	if err := scanner.Err(); err != nil {
		// Don't treat EOF as an error
		if err != io.EOF {
			return fmt.Errorf("error reading log stream: %w", err)
		}
	}

	return nil
}

func init() {
	Cmd.Flags().StringVar(&nodeID, "node", "", "target node peer ID")
	Cmd.Flags().BoolVar(&cleanup, "cleanup", true, "remove package file after deployment")
	Cmd.Flags().BoolVar(&noSign, "no-sign", false, "skip package signing")
	Cmd.Flags().StringVar(&privateKey, "private-key", "", "path to private key file for signing")
}
