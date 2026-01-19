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
	"github.com/asjdf/p2p-playground-lite/pkg/consts"
	"github.com/asjdf/p2p-playground-lite/pkg/logging"
	"github.com/asjdf/p2p-playground-lite/pkg/p2p"
	"github.com/asjdf/p2p-playground-lite/pkg/types"
	"github.com/spf13/cobra"
)

var (
	cfgFile       string
	globalConfig  *config.ControllerConfig
	globalLogger  types.Logger
)

var rootCmd = &cobra.Command{
	Use:   "controller",
	Short: "P2P Playground controller",
	Long:  `Controller for P2P Playground - deploy and manage applications across P2P nodes.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		var err error
		globalConfig, err = loadConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Initialize logger
		globalLogger, err = logging.New(&globalConfig.Logging)
		if err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}

		return nil
	},
}

var (
	deployNodeID    string
	deployAutoStart bool
	listNodeID      string
	logsNodeID      string
	logsFollow      bool
	logsTail        int
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

		// Use global logger
		logger := globalLogger

		// Create P2P host using configuration
		ctx := context.Background()
		host, err := p2p.NewHost(ctx, globalConfig.Node.ListenAddrs, logger)
		if err != nil {
			return fmt.Errorf("failed to create P2P host: %w", err)
		}
		defer host.Close()

		fmt.Printf("Controller ID: %s\n", host.ID())

		// Enable mDNS discovery if configured
		if globalConfig.Node.EnableMDNS {
			if err := host.EnableMDNS(ctx); err != nil {
				return fmt.Errorf("failed to enable mDNS: %w", err)
			}
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
	Long: `List all deployed applications on a target node.

If --node is not specified, applications from the first discovered node will be listed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Listing applications...")

		// Use global logger
		logger := globalLogger

		// Create P2P host using configuration
		ctx := context.Background()
		host, err := p2p.NewHost(ctx, globalConfig.Node.ListenAddrs, logger)
		if err != nil {
			return fmt.Errorf("failed to create P2P host: %w", err)
		}
		defer host.Close()

		// Enable mDNS discovery if configured
		if globalConfig.Node.EnableMDNS {
			if err := host.EnableMDNS(ctx); err != nil {
				return fmt.Errorf("failed to enable mDNS: %w", err)
			}
		}

		// Wait for peer discovery
		fmt.Println("Discovering nodes...")
		time.Sleep(3 * time.Second)

		// Get target node
		var targetPeerID string
		if listNodeID != "" {
			targetPeerID = listNodeID
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
		apps, err := listApplications(ctx, host, targetPeerID, logger)
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

var logsCmd = &cobra.Command{
	Use:   "logs [app-id]",
	Short: "View application logs",
	Long: `View logs from a deployed application.

If --node is not specified, logs will be fetched from the first discovered node.
Use --tail to limit the number of lines shown.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error{
		appID := args[0]
		fmt.Printf("Fetching logs for application: %s\n", appID)

		// Use global logger
		logger := globalLogger

		// Create P2P host using configuration
		ctx := context.Background()
		host, err := p2p.NewHost(ctx, globalConfig.Node.ListenAddrs, logger)
		if err != nil {
			return fmt.Errorf("failed to create P2P host: %w", err)
		}
		defer host.Close()

		// Enable mDNS discovery if configured
		if globalConfig.Node.EnableMDNS {
			if err := host.EnableMDNS(ctx); err != nil {
				return fmt.Errorf("failed to enable mDNS: %w", err)
			}
		}

		// Wait for peer discovery
		fmt.Println("Discovering nodes...")
		time.Sleep(3 * time.Second)

		// Get target node
		var targetPeerID string
		if logsNodeID != "" {
			targetPeerID = logsNodeID
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
		fmt.Println("\nFetching logs...\n")
		logs, err := fetchLogs(ctx, host, targetPeerID, appID, logsFollow, logsTail, logger)
		if err != nil {
			return fmt.Errorf("failed to fetch logs: %w", err)
		}

		// Display logs
		fmt.Println(logs)

		return nil
	},
}

var nodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "List available nodes",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Discovering P2P nodes...")

		// Use global logger
		logger := globalLogger

		// Create P2P host using configuration
		ctx := context.Background()
		host, err := p2p.NewHost(ctx, globalConfig.Node.ListenAddrs, logger)
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

		// Enable mDNS discovery if configured
		if globalConfig.Node.EnableMDNS {
			if err := host.EnableMDNS(ctx); err != nil {
				return fmt.Errorf("failed to enable mDNS: %w", err)
			}
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
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default: ~/.p2p-playground/controller.yaml)")

	deployCmd.Flags().StringVar(&deployNodeID, "node", "", "target node peer ID")
	deployCmd.Flags().BoolVar(&deployAutoStart, "start", true, "automatically start the application after deployment")

	listCmd.Flags().StringVar(&listNodeID, "node", "", "target node peer ID")

	logsCmd.Flags().StringVar(&logsNodeID, "node", "", "target node peer ID")
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "follow log output")
	logsCmd.Flags().IntVar(&logsTail, "tail", 50, "number of lines to show from the end")

	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(nodesCmd)
	rootCmd.AddCommand(keygenCmd)
	rootCmd.AddCommand(signCmd)
}

// loadConfig loads the controller configuration
func loadConfig(configPath string) (*config.ControllerConfig, error) {
	// If no config file specified, try default location
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			defaultPath := filepath.Join(homeDir, ".p2p-playground", "controller.yaml")
			if _, err := os.Stat(defaultPath); err == nil {
				configPath = defaultPath
			}
		}
	}

	// Load config from file if it exists
	if configPath != "" {
		cfg, err := config.LoadControllerConfig(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load config from %s: %w", configPath, err)
		}
		return cfg, nil
	}

	// Use defaults if no config file
	cfg, err := config.LoadControllerConfig("")
	if err != nil {
		return nil, fmt.Errorf("failed to load default config: %w", err)
	}
	return cfg, nil
}

func Execute() error {
	return rootCmd.Execute()
}

// DeployRequest represents a deployment request
type DeployRequest struct {
	FileName  string `json:"file_name"`
	FileSize  int64  `json:"file_size"`
	AutoStart bool   `json:"auto_start"`
	Signature []byte `json:"signature,omitempty"` // Ed25519 signature of the package file
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
	stream, err := host.NewStream(ctx, peerID, consts.DeployProtocolID)
	if err != nil {
		return "", fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()

	// Load signature if exists
	var signature []byte
	sigPath := packagePath + ".sig"
	if sigData, err := os.ReadFile(sigPath); err == nil {
		signature = sigData
		logger.Info("package signature found", "sig_path", sigPath)
	} else {
		logger.Warn("no package signature found, deploying without signature verification")
	}

	// Prepare request
	req := DeployRequest{
		FileName:  filepath.Base(packagePath),
		FileSize:  fileSize,
		AutoStart: autoStart,
		Signature: signature,
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

// ListAppsResponse represents the response for list apps request
type ListAppsResponse struct {
	Success bool                 `json:"success"`
	Apps    []*types.Application `json:"apps,omitempty"`
	Error   string               `json:"error,omitempty"`
}

// listApplications lists applications on a target node
func listApplications(ctx context.Context, host *p2p.Host, peerID string, logger types.Logger) ([]*types.Application, error) {
	// Create stream to target peer
	stream, err := host.NewStream(ctx, peerID, consts.ListProtocolID)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()

	logger.Info("requesting application list", "peer", peerID)

	// Read response header size
	var respSize uint32
	if err := binary.Read(stream, binary.BigEndian, &respSize); err != nil {
		return nil, fmt.Errorf("failed to read response size: %w", err)
	}

	// Read response
	respBytes := make([]byte, respSize)
	if _, err := io.ReadFull(stream, respBytes); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var resp ListAppsResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("list failed on node: %s", resp.Error)
	}

	logger.Info("received application list", "count", len(resp.Apps))
	return resp.Apps, nil
}

// LogsRequest represents a logs request
type LogsRequest struct {
	AppID  string `json:"app_id"`
	Follow bool   `json:"follow"`
	Tail   int    `json:"tail"`
}

// LogsResponse represents a logs response
type LogsResponse struct {
	Success bool   `json:"success"`
	Logs    string `json:"logs,omitempty"`
	Error   string `json:"error,omitempty"`
}

// fetchLogs fetches logs from an application on a target node
func fetchLogs(ctx context.Context, host *p2p.Host, peerID string, appID string, follow bool, tail int, logger types.Logger) (string, error) {
	// Create stream to target peer
	stream, err := host.NewStream(ctx, peerID, consts.LogsProtocolID)
	if err != nil {
		return "", fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()

	// Prepare request
	req := LogsRequest{
		AppID:  appID,
		Follow: follow,
		Tail:   tail,
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

	logger.Info("requesting logs", "app_id", appID, "tail", tail)

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

	var resp LogsResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if !resp.Success {
		return "", fmt.Errorf("logs request failed on node: %s", resp.Error)
	}

	logger.Info("received logs", "size", len(resp.Logs))
	return resp.Logs, nil
}
