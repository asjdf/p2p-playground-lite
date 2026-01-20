package common

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/asjdf/p2p-playground-lite/pkg/config"
	"github.com/asjdf/p2p-playground-lite/pkg/consts"
	"github.com/asjdf/p2p-playground-lite/pkg/logging"
	"github.com/asjdf/p2p-playground-lite/pkg/p2p"
	"github.com/asjdf/p2p-playground-lite/pkg/types"
)

var (
	CfgFile      string
	GlobalConfig *config.ControllerConfig
	GlobalLogger types.Logger
)

// InitConfig initializes configuration and logger
func InitConfig(cfgFile string) error {
	CfgFile = cfgFile

	// Load configuration
	var err error
	GlobalConfig, err = LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize logger
	GlobalLogger, err = logging.New(&GlobalConfig.Logging)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	return nil
}

// LoadConfig loads the controller configuration
func LoadConfig(configPath string) (*config.ControllerConfig, error) {
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

// CreateP2PHost creates a P2P host using global configuration
func CreateP2PHost(ctx context.Context) (*p2p.Host, error) {
	hostConfig := &p2p.HostConfig{
		ListenAddrs:         GlobalConfig.Node.ListenAddrs,
		PSK:                 GlobalConfig.Security.PSK,
		EnableAuth:          GlobalConfig.Security.EnableAuth,
		TrustedPeers:        []string{}, // Controller doesn't restrict trusted peers
		BootstrapPeers:      GlobalConfig.Node.BootstrapPeers,
		DisableDHT:          GlobalConfig.Node.DisableDHT,
		DHTMode:             GlobalConfig.Node.DHTMode,
		DisableNATService:   GlobalConfig.Node.DisableNATService,
		DisableAutoRelay:    GlobalConfig.Node.DisableAutoRelay,
		DisableHolePunching: GlobalConfig.Node.DisableHolePunching,
		DisableRelayService: GlobalConfig.Node.DisableRelayService,
		StaticRelays:        GlobalConfig.Node.StaticRelays,
	}

	host, err := p2p.NewHost(ctx, hostConfig, GlobalLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to create P2P host: %w", err)
	}

	// Enable mDNS discovery if configured
	if GlobalConfig.Node.EnableMDNS {
		if err := host.EnableMDNS(ctx); err != nil {
			GlobalLogger.Warn("failed to enable mDNS", "error", err)
		}
	}

	return host, nil
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

// ListAppsResponse represents the response for list apps request
type ListAppsResponse struct {
	Success bool                 `json:"success"`
	Apps    []*types.Application `json:"apps,omitempty"`
	Error   string               `json:"error,omitempty"`
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

// DeployPackage deploys a package to a target node
func DeployPackage(ctx context.Context, host *p2p.Host, peerID string, packagePath string, fileSize int64, autoStart bool, logger types.Logger) (string, error) {
	// Open package file
	file, err := os.Open(packagePath)
	if err != nil {
		return "", fmt.Errorf("failed to open package: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Create stream to target peer
	stream, err := host.NewStream(ctx, peerID, consts.DeployProtocolID)
	if err != nil {
		return "", fmt.Errorf("failed to create stream: %w", err)
	}
	defer func() { _ = stream.Close() }()

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

// ListApplications lists applications on a target node
func ListApplications(ctx context.Context, host *p2p.Host, peerID string, logger types.Logger) ([]*types.Application, error) {
	// Create stream to target peer
	stream, err := host.NewStream(ctx, peerID, consts.ListProtocolID)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}
	defer func() { _ = stream.Close() }()

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

// FetchLogs fetches logs from an application on a target node
func FetchLogs(ctx context.Context, host *p2p.Host, peerID string, appID string, follow bool, tail int, logger types.Logger) (string, error) {
	// Create stream to target peer
	stream, err := host.NewStream(ctx, peerID, consts.LogsProtocolID)
	if err != nil {
		return "", fmt.Errorf("failed to create stream: %w", err)
	}
	defer func() { _ = stream.Close() }()

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
