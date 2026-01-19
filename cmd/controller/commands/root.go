package commands

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/asjdf/p2p-playground-lite/pkg/config"
	"github.com/asjdf/p2p-playground-lite/pkg/logging"
	"github.com/asjdf/p2p-playground-lite/pkg/p2p"
	"github.com/asjdf/p2p-playground-lite/pkg/types"
	"github.com/spf13/cobra"
)

const (
	deployProtocolID = "/p2p-playground/deploy/1.0.0"
)

var (
	cfgFile string
)

var rootCmd = &cobra.Command{
	Use:   "controller",
	Short: "P2P Playground controller",
	Long:  `Controller for P2P Playground - deploy and manage applications across P2P nodes.`,
}

var (
	deployNodeID    string
	deployAutoStart bool
)

var deployCmd = &cobra.Command{
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

		// Create minimal config for controller
		logCfg := &config.LoggingConfig{
			Level:  "info",
			Format: "console",
		}
		logger, err := logging.New(logCfg)
		if err != nil {
			return fmt.Errorf("failed to create logger: %w", err)
		}

		// Create P2P host
		ctx := context.Background()
		host, err := p2p.NewHost(ctx, []string{"/ip4/0.0.0.0/tcp/0"}, logger)
		if err != nil {
			return fmt.Errorf("failed to create P2P host: %w", err)
		}
		defer host.Close()

		fmt.Printf("Controller ID: %s\n", host.ID())

		// Enable mDNS discovery
		if err := host.EnableMDNS(ctx); err != nil {
			return fmt.Errorf("failed to enable mDNS: %w", err)
		}

		// Wait for peer discovery
		fmt.Println("Discovering nodes...")
		time.Sleep(3 * time.Second)

		// Get target node
		var targetPeerID string
		if deployNodeID != "" {
			targetPeerID = deployNodeID
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
		appID, err := deployPackage(ctx, host, targetPeerID, packagePath, fileInfo.Size(), deployAutoStart, logger)
		if err != nil {
			return fmt.Errorf("deployment failed: %w", err)
		}

		fmt.Printf("\n✓ Deployment successful!\n")
		fmt.Printf("  Application ID: %s\n", appID)
		if deployAutoStart {
			fmt.Printf("  Status: Started\n")
		} else {
			fmt.Printf("  Status: Deployed (not started)\n")
		}

		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List deployed applications",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Listing applications...")
		fmt.Println("Not implemented yet")
		return nil
	},
}

var nodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "List available nodes",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Discovering P2P nodes...")

		// Create minimal config for controller
		logCfg := &config.LoggingConfig{
			Level:  "info",
			Format: "console",
		}
		logger, err := logging.New(logCfg)
		if err != nil {
			return fmt.Errorf("failed to create logger: %w", err)
		}

		// Create P2P host with random port
		ctx := context.Background()
		host, err := p2p.NewHost(ctx, []string{"/ip4/0.0.0.0/tcp/0"}, logger)
		if err != nil {
			return fmt.Errorf("failed to create P2P host: %w", err)
		}
		defer host.Close()

		fmt.Printf("Controller ID: %s\n", host.ID())
		fmt.Printf("Controller addresses:\n")
		for _, addr := range host.Addrs() {
			fmt.Printf("  - %s\n", addr)
		}
		fmt.Println()

		// Enable mDNS discovery
		if err := host.EnableMDNS(ctx); err != nil {
			return fmt.Errorf("failed to enable mDNS: %w", err)
		}

		fmt.Println("Scanning for nodes via mDNS (waiting 5 seconds)...")
		time.Sleep(5 * time.Second)

		// Get discovered peers
		peers := host.Peers()
		if len(peers) == 0 {
			fmt.Println("No peers discovered.")
			return nil
		}

		fmt.Printf("\nDiscovered %d peer(s):\n", len(peers))
		for i, peer := range peers {
			fmt.Printf("%d. Peer ID: %s\n", i+1, peer.ID)
			fmt.Printf("   Addresses:\n")
			for _, addr := range peer.Addrs {
				fmt.Printf("     - %s\n", addr)
			}
			fmt.Println()
		}

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file")

	deployCmd.Flags().StringVar(&deployNodeID, "node", "", "target node peer ID")
	deployCmd.Flags().BoolVar(&deployAutoStart, "start", true, "automatically start the application after deployment")

	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(nodesCmd)
}

func Execute() error {
	return rootCmd.Execute()
}

// DeployRequest represents a deployment request
type DeployRequest struct {
	FileName  string `json:"file_name"`
	FileSize  int64  `json:"file_size"`
	AutoStart bool   `json:"auto_start"`
}

// DeployResponse represents a deployment response
type DeployResponse struct {
	Success bool   `json:"success"`
	AppID   string `json:"app_id,omitempty"`
	Error   string `json:"error,omitempty"`
}

// deployPackage deploys a package to a target node
func deployPackage(ctx context.Context, host *p2p.Host, peerID string, packagePath string, fileSize int64, autoStart bool, logger types.Logger) (string, error) {
	// Open package file
	file, err := os.Open(packagePath)
	if err != nil {
		return "", fmt.Errorf("failed to open package: %w", err)
	}
	defer file.Close()

	// Create stream to target peer
	stream, err := host.NewStream(ctx, peerID, deployProtocolID)
	if err != nil {
		return "", fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()

	// Prepare request
	req := DeployRequest{
		FileName:  filepath.Base(packagePath),
		FileSize:  fileSize,
		AutoStart: autoStart,
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request header size
	reqSize := uint32(len(reqBytes))
	if err := binary.Write(stream, binary.BigEndian, reqSize); err != nil {
		return "", fmt.Errorf("failed to send header size: %w", err)
	}

	// Send request header
	if _, err := stream.Write(reqBytes); err != nil {
		return "", fmt.Errorf("failed to send header: %w", err)
	}

	logger.Info("sending package", "file", req.FileName, "size", fileSize)

	// Send file content
	buf := make([]byte, 64*1024) // 64KB chunks
	var sent int64
	lastProgress := 0

	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("failed to read file: %w", err)
		}

		if n == 0 {
			break
		}

		if _, err := stream.Write(buf[:n]); err != nil {
			return "", fmt.Errorf("failed to send chunk: %w", err)
		}

		sent += int64(n)
		progress := int(float64(sent) / float64(fileSize) * 100)
		if progress > lastProgress && progress%10 == 0 {
			fmt.Printf("  Progress: %d%%\n", progress)
			lastProgress = progress
		}
	}

	fmt.Printf("  Progress: 100%%\n")
	logger.Info("package sent", "size", sent)

	// Read response header size
	var respSize uint32
	if err := binary.Read(stream, binary.BigEndian, &respSize); err != nil {
		return "", fmt.Errorf("failed to read response size: %w", err)
	}

	// Read response
	respBytes := make([]byte, respSize)
	if _, err := io.ReadFull(stream, respBytes); err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var resp DeployResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if !resp.Success {
		return "", fmt.Errorf("deployment failed on node: %s", resp.Error)
	}

	return resp.AppID, nil
}
