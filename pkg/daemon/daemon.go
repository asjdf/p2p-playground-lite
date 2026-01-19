package daemon

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"github.com/asjdf/p2p-playground-lite/pkg/config"
	"github.com/asjdf/p2p-playground-lite/pkg/logging"
	"github.com/asjdf/p2p-playground-lite/pkg/p2p"
	pkgmanager "github.com/asjdf/p2p-playground-lite/pkg/package"
	"github.com/asjdf/p2p-playground-lite/pkg/runtime"
	"github.com/asjdf/p2p-playground-lite/pkg/security"
	"github.com/asjdf/p2p-playground-lite/pkg/storage"
	"github.com/asjdf/p2p-playground-lite/pkg/transfer"
	"github.com/asjdf/p2p-playground-lite/pkg/types"
)

const (
	deployProtocolID = "/p2p-playground/deploy/1.0.0"
	listProtocolID   = "/p2p-playground/list/1.0.0"
	logsProtocolID   = "/p2p-playground/logs/1.0.0"
)

// Daemon coordinates all daemon components
type Daemon struct {
	config     *config.DaemonConfig
	logger     types.Logger
	host       *p2p.Host
	storage    *storage.FileStorage
	pkgMgr     *pkgmanager.Manager
	runtime    *runtime.Runtime
	transfer   *transfer.Manager
	signer     *security.Signer
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// New creates a new daemon
func New(cfg *config.DaemonConfig) (*Daemon, error) {
	// Initialize logger
	logger, err := logging.New(&cfg.Logging)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	d := &Daemon{
		config:     cfg,
		logger:     logger,
		ctx:        ctx,
		cancelFunc: cancel,
	}

	return d, nil
}

// Start starts the daemon
func (d *Daemon) Start() error {
	d.logger.Info("starting P2P Playground daemon")

	// Initialize storage
	storage, err := storage.NewFileStorage(d.config.Storage.DataDir)
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}
	d.storage = storage
	d.logger.Info("storage initialized", "path", d.config.Storage.DataDir)

	// Load or generate keys
	signer, err := security.LoadOrGenerateKeys(d.config.Storage.KeysDir)
	if err != nil {
		return fmt.Errorf("failed to load keys: %w", err)
	}
	d.signer = signer
	d.logger.Info("keys loaded")

	// Initialize P2P host
	host, err := p2p.NewHost(d.ctx, d.config.Node.ListenAddrs, d.logger)
	if err != nil {
		return fmt.Errorf("failed to create P2P host: %w", err)
	}
	d.host = host

	// Enable mDNS if configured
	if d.config.Node.EnableMDNS {
		if err := host.EnableMDNS(d.ctx); err != nil {
			d.logger.Warn("failed to enable mDNS", "error", err)
		}
	}

	// Initialize package manager
	d.pkgMgr = pkgmanager.New()

	// Initialize runtime
	d.runtime = runtime.New(d.logger)

	// Initialize transfer manager
	d.transfer = transfer.New(d.host, d.logger)

	// Register protocol handlers
	d.host.SetStreamHandler(deployProtocolID, d.handleDeployRequest)
	d.host.SetStreamHandler(listProtocolID, d.handleListRequest)
	d.host.SetStreamHandler(logsProtocolID, d.handleLogsRequest)

	d.logger.Info("daemon started",
		"peer_id", host.ID(),
		"addrs", host.Addrs(),
	)

	return nil
}

// Stop stops the daemon
func (d *Daemon) Stop() error {
	d.logger.Info("stopping daemon")

	if d.cancelFunc != nil {
		d.cancelFunc()
	}

	if d.host != nil {
		d.host.Close()
	}

	d.logger.Info("daemon stopped")
	return nil
}

// DeployPackage deploys a package
func (d *Daemon) DeployPackage(ctx context.Context, pkgPath string) (*types.Application, error) {
	d.logger.Info("deploying package", "path", pkgPath)

	// Get manifest
	manifest, err := d.pkgMgr.GetManifest(ctx, pkgPath)
	if err != nil {
		return nil, types.WrapError(err, "failed to get manifest")
	}

	// Create application directory
	appID := fmt.Sprintf("%s-%s", manifest.Name, manifest.Version)
	appDir := filepath.Join(d.config.Storage.AppsDir, appID)

	// Unpack package
	_, err = d.pkgMgr.Unpack(ctx, pkgPath, appDir)
	if err != nil {
		return nil, types.WrapError(err, "failed to unpack package")
	}

	// Create application
	app := &types.Application{
		ID:          appID,
		Name:        manifest.Name,
		Version:     manifest.Version,
		PackagePath: pkgPath,
		Manifest:    manifest,
		Status:      types.AppStatusStopped,
		WorkDir:     appDir,
		Labels:      manifest.Labels,
	}

	d.logger.Info("package deployed", "app_id", appID)

	return app, nil
}

// StartApp starts an application
func (d *Daemon) StartApp(ctx context.Context, appID string) error {
	// For now, assume app is already deployed
	// In real implementation, look up from storage
	return types.ErrNotImplemented
}

// StopApp stops an application
func (d *Daemon) StopApp(ctx context.Context, appID string) error {
	return d.runtime.Stop(ctx, appID)
}

// ListApps lists all applications
func (d *Daemon) ListApps(ctx context.Context) ([]*types.Application, error) {
	return d.runtime.List(ctx)
}

// GetNodeInfo returns node information
func (d *Daemon) GetNodeInfo() *types.NodeInfo {
	apps, _ := d.runtime.List(d.ctx)

	return &types.NodeInfo{
		ID:     d.host.ID(),
		Addrs:  d.host.Addrs(),
		Labels: d.config.Node.Labels,
		Apps:   apps,
	}
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

// handleDeployRequest handles incoming deploy requests
func (d *Daemon) handleDeployRequest(stream types.Stream) {
	defer stream.Close()

	d.logger.Info("received deploy request")

	// Read request header (JSON)
	var headerSize uint32
	if err := binary.Read(stream, binary.BigEndian, &headerSize); err != nil {
		d.logger.Error("failed to read header size", "error", err)
		d.sendDeployResponse(stream, false, "", err.Error())
		return
	}

	headerBytes := make([]byte, headerSize)
	if _, err := io.ReadFull(stream, headerBytes); err != nil {
		d.logger.Error("failed to read header", "error", err)
		d.sendDeployResponse(stream, false, "", err.Error())
		return
	}

	var req DeployRequest
	if err := json.Unmarshal(headerBytes, &req); err != nil {
		d.logger.Error("failed to parse request", "error", err)
		d.sendDeployResponse(stream, false, "", err.Error())
		return
	}

	d.logger.Info("deploy request details",
		"file_name", req.FileName,
		"file_size", req.FileSize,
		"auto_start", req.AutoStart,
	)

	// Save package to packages directory
	pkgPath := filepath.Join(d.config.Storage.PackagesDir, req.FileName)
	if err := d.receiveFile(stream, pkgPath, req.FileSize); err != nil {
		d.logger.Error("failed to receive file", "error", err)
		d.sendDeployResponse(stream, false, "", err.Error())
		return
	}

	// Deploy package
	app, err := d.DeployPackage(d.ctx, pkgPath)
	if err != nil {
		d.logger.Error("failed to deploy package", "error", err)
		d.sendDeployResponse(stream, false, "", err.Error())
		return
	}

	// Auto-start if requested
	if req.AutoStart {
		if err := d.runtime.Start(d.ctx, app); err != nil {
			d.logger.Warn("failed to auto-start application", "error", err)
			// Don't fail the deployment, just log the warning
		} else {
			d.logger.Info("application started", "app_id", app.ID)
		}
	}

	d.sendDeployResponse(stream, true, app.ID, "")
}

// receiveFile receives file content from stream
func (d *Daemon) receiveFile(stream types.Stream, destPath string, expectedSize int64) error {
	file, err := d.storage.CreateFile(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	buf := make([]byte, 64*1024) // 64KB chunks
	var received int64

	for received < expectedSize {
		n, err := stream.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read chunk: %w", err)
		}

		if n == 0 {
			break
		}

		if _, err := file.Write(buf[:n]); err != nil {
			return fmt.Errorf("failed to write chunk: %w", err)
		}

		received += int64(n)
	}

	if received != expectedSize {
		return fmt.Errorf("incomplete transfer: received %d of %d bytes", received, expectedSize)
	}

	d.logger.Info("file received", "path", destPath, "size", received)
	return nil
}

// sendDeployResponse sends deployment response
func (d *Daemon) sendDeployResponse(stream types.Stream, success bool, appID string, errMsg string) {
	resp := DeployResponse{
		Success: success,
		AppID:   appID,
		Error:   errMsg,
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		d.logger.Error("failed to marshal response", "error", err)
		return
	}

	// Send response size
	respSize := uint32(len(respBytes))
	if err := binary.Write(stream, binary.BigEndian, respSize); err != nil {
		d.logger.Error("failed to send response size", "error", err)
		return
	}

	// Send response
	if _, err := stream.Write(respBytes); err != nil {
		d.logger.Error("failed to send response", "error", err)
		return
	}

	d.logger.Info("deploy response sent", "success", success, "app_id", appID)
}

// ListAppsResponse represents the response for list apps request
type ListAppsResponse struct {
	Success bool                 `json:"success"`
	Apps    []*types.Application `json:"apps,omitempty"`
	Error   string               `json:"error,omitempty"`
}

// handleListRequest handles incoming list apps requests
func (d *Daemon) handleListRequest(stream types.Stream) {
	defer stream.Close()

	d.logger.Info("received list apps request")

	// Get all applications
	apps, err := d.runtime.List(d.ctx)
	if err != nil {
		d.logger.Error("failed to list apps", "error", err)
		d.sendListResponse(stream, false, nil, err.Error())
		return
	}

	d.sendListResponse(stream, true, apps, "")
}

// sendListResponse sends list apps response
func (d *Daemon) sendListResponse(stream types.Stream, success bool, apps []*types.Application, errMsg string) {
	resp := ListAppsResponse{
		Success: success,
		Apps:    apps,
		Error:   errMsg,
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		d.logger.Error("failed to marshal response", "error", err)
		return
	}

	// Send response size
	respSize := uint32(len(respBytes))
	if err := binary.Write(stream, binary.BigEndian, respSize); err != nil {
		d.logger.Error("failed to send response size", "error", err)
		return
	}

	// Send response
	if _, err := stream.Write(respBytes); err != nil {
		d.logger.Error("failed to send response", "error", err)
		return
	}

	d.logger.Info("list response sent", "app_count", len(apps))
}

// LogsRequest represents a logs request
type LogsRequest struct {
	AppID  string `json:"app_id"`
	Follow bool   `json:"follow"`
	Tail   int    `json:"tail"` // Number of lines from end, 0 for all
}

// LogsResponse represents a logs response
type LogsResponse struct {
	Success bool   `json:"success"`
	Logs    string `json:"logs,omitempty"`
	Error   string `json:"error,omitempty"`
}

// handleLogsRequest handles incoming logs requests
func (d *Daemon) handleLogsRequest(stream types.Stream) {
	defer stream.Close()

	d.logger.Info("received logs request")

	// Read request header
	var headerSize uint32
	if err := binary.Read(stream, binary.BigEndian, &headerSize); err != nil {
		d.logger.Error("failed to read header size", "error", err)
		d.sendLogsResponse(stream, false, "", err.Error())
		return
	}

	headerBytes := make([]byte, headerSize)
	if _, err := io.ReadFull(stream, headerBytes); err != nil {
		d.logger.Error("failed to read header", "error", err)
		d.sendLogsResponse(stream, false, "", err.Error())
		return
	}

	var req LogsRequest
	if err := json.Unmarshal(headerBytes, &req); err != nil {
		d.logger.Error("failed to parse request", "error", err)
		d.sendLogsResponse(stream, false, "", err.Error())
		return
	}

	d.logger.Info("logs request details", "app_id", req.AppID, "follow", req.Follow, "tail", req.Tail)

	// Get logs
	logsReader, err := d.runtime.Logs(d.ctx, req.AppID, req.Follow)
	if err != nil {
		d.logger.Error("failed to get logs", "error", err)
		d.sendLogsResponse(stream, false, "", err.Error())
		return
	}
	defer logsReader.Close()

	// Read all logs
	logsBytes, err := io.ReadAll(logsReader)
	if err != nil {
		d.logger.Error("failed to read logs", "error", err)
		d.sendLogsResponse(stream, false, "", err.Error())
		return
	}

	logs := string(logsBytes)

	// Apply tail if requested
	if req.Tail > 0 {
		lines := make([]string, 0)
		for _, line := range splitLines(logs) {
			if line != "" {
				lines = append(lines, line)
			}
		}
		if len(lines) > req.Tail {
			lines = lines[len(lines)-req.Tail:]
		}
		logs = joinLines(lines)
	}

	d.sendLogsResponse(stream, true, logs, "")
}

// sendLogsResponse sends logs response
func (d *Daemon) sendLogsResponse(stream types.Stream, success bool, logs string, errMsg string) {
	resp := LogsResponse{
		Success: success,
		Logs:    logs,
		Error:   errMsg,
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		d.logger.Error("failed to marshal response", "error", err)
		return
	}

	// Send response size
	respSize := uint32(len(respBytes))
	if err := binary.Write(stream, binary.BigEndian, respSize); err != nil {
		d.logger.Error("failed to send response size", "error", err)
		return
	}

	// Send response
	if _, err := stream.Write(respBytes); err != nil {
		d.logger.Error("failed to send response", "error", err)
		return
	}

	d.logger.Info("logs response sent", "log_size", len(logs))
}

// Helper functions
func splitLines(s string) []string {
	lines := make([]string, 0)
	start := 0
	for i, c := range s {
		if c == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		result += line
		if i < len(lines)-1 {
			result += "\n"
		}
	}
	return result
}
