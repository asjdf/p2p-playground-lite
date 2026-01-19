package runtime

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/asjdf/p2p-playground-lite/pkg/health"
	"github.com/asjdf/p2p-playground-lite/pkg/types"
)

// appInfo holds application runtime information
type appInfo struct {
	app           *types.Application
	healthChecker *health.Checker
	cancelHealth  context.CancelFunc
	autoRestart   bool
}

// Runtime manages application processes
type Runtime struct {
	apps   map[string]*appInfo
	mu     sync.RWMutex
	logger types.Logger
}

// New creates a new runtime
func New(logger types.Logger) *Runtime {
	return &Runtime{
		apps:   make(map[string]*appInfo),
		logger: logger,
	}
}

// Start starts an application
func (r *Runtime) Start(ctx context.Context, app *types.Application) error {
	return r.start(ctx, app, false)
}

// StartWithAutoRestart starts an application with auto-restart enabled
func (r *Runtime) StartWithAutoRestart(ctx context.Context, app *types.Application) error {
	return r.start(ctx, app, true)
}

// start is the internal start implementation
func (r *Runtime) start(ctx context.Context, app *types.Application, autoRestart bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if already running
	if existing, exists := r.apps[app.ID]; exists {
		if existing.app.Status == types.AppStatusRunning {
			return types.ErrAppAlreadyRunning
		}
	}

	// Update status
	app.Status = types.AppStatusStarting

	// Build command
	cmdPath := filepath.Join(app.WorkDir, app.Manifest.Entrypoint)
	cmd := exec.CommandContext(ctx, cmdPath, app.Manifest.Args...)
	cmd.Dir = app.WorkDir

	// Set environment variables
	cmd.Env = os.Environ()
	for k, v := range app.Manifest.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Create log directory
	logDir := filepath.Join(app.WorkDir, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return types.WrapError(err, "failed to create log directory")
	}

	// Set up log files
	stdoutLog := filepath.Join(logDir, "stdout.log")
	stderrLog := filepath.Join(logDir, "stderr.log")

	stdoutFile, err := os.OpenFile(stdoutLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return types.WrapError(err, "failed to create stdout log")
	}

	stderrFile, err := os.OpenFile(stderrLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		stdoutFile.Close()
		return types.WrapError(err, "failed to create stderr log")
	}

	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile

	// Start process
	if err := cmd.Start(); err != nil {
		stdoutFile.Close()
		stderrFile.Close()
		return types.WrapError(err, "failed to start process")
	}

	// Update application info
	app.PID = cmd.Process.Pid
	app.Status = types.AppStatusRunning
	app.StartedAt = time.Now()

	// Create appInfo
	info := &appInfo{
		app:         app,
		autoRestart: autoRestart,
	}

	// Set up health monitoring if configured
	if app.Manifest.HealthCheck != nil {
		healthCfg := convertHealthCheckConfig(app.Manifest.HealthCheck)
		checker := health.New(healthCfg, app.PID, r.logger)

		healthCtx, healthCancel := context.WithCancel(context.Background())
		info.healthChecker = checker
		info.cancelHealth = healthCancel

		// Start health monitoring in background
		go checker.StartMonitoring(healthCtx, func(result *health.Result) {
			r.logger.Warn("application unhealthy, triggering restart",
				"app_id", app.ID,
				"message", result.Message,
				"failures", result.FailureCount,
			)

			// Auto-restart if enabled
			if autoRestart {
				go func() {
					if err := r.Restart(context.Background(), app.ID); err != nil {
						r.logger.Error("failed to auto-restart application",
							"app_id", app.ID,
							"error", err,
						)
					}
				}()
			}
		})

		r.logger.Info("health monitoring started",
			"app_id", app.ID,
			"type", healthCfg.Type,
			"interval", healthCfg.Interval,
		)
	}

	// Store application info
	r.apps[app.ID] = info

	// Monitor process in background
	go func() {
		defer stdoutFile.Close()
		defer stderrFile.Close()

		err := cmd.Wait()

		r.mu.Lock()
		defer r.mu.Unlock()

		if info, exists := r.apps[app.ID]; exists {
			// Cancel health monitoring
			if info.cancelHealth != nil {
				info.cancelHealth()
			}

			if err != nil {
				info.app.Status = types.AppStatusFailed
				r.logger.Error("application exited with error",
					"app_id", info.app.ID,
					"error", err,
				)
			} else {
				info.app.Status = types.AppStatusStopped
				r.logger.Info("application stopped",
					"app_id", info.app.ID,
				)
			}
			info.app.PID = 0
		}
	}()

	r.logger.Info("application started",
		"app_id", app.ID,
		"pid", app.PID,
	)

	return nil
}

// convertHealthCheckConfig converts manifest health check config to health package config
func convertHealthCheckConfig(hc *types.HealthCheckConfig) *health.Config {
	cfg := &health.Config{
		Type:     health.CheckType(hc.Type),
		Interval: hc.Interval,
		Timeout:  hc.Timeout,
		Retries:  hc.Retries,
	}

	// Set defaults
	if cfg.Interval == 0 {
		cfg.Interval = 30 * time.Second
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}
	if cfg.Retries == 0 {
		cfg.Retries = 3
	}

	// Parse endpoint for HTTP/TCP
	if hc.Endpoint != "" {
		// Simple parsing: port number for TCP, full URL for HTTP
		if cfg.Type == health.CheckTypeHTTP {
			// Extract port from endpoint (assuming format like ":8080/health")
			// For now, use default port 8080
			cfg.HTTPPort = 8080
			cfg.HTTPPath = hc.Endpoint
		} else if cfg.Type == health.CheckTypeTCP {
			// Extract port from endpoint (assuming format like ":8080")
			cfg.TCPPort = 8080
		}
	}

	return cfg
}

// Stop stops a running application
func (r *Runtime) Stop(ctx context.Context, appID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	info, exists := r.apps[appID]
	if !exists {
		return types.ErrNotFound
	}

	if info.app.Status != types.AppStatusRunning {
		return types.ErrAppNotRunning
	}

	// Cancel health monitoring
	if info.cancelHealth != nil {
		info.cancelHealth()
		info.cancelHealth = nil
	}

	// Find process
	process, err := os.FindProcess(info.app.PID)
	if err != nil {
		return types.WrapError(err, "failed to find process")
	}

	// Send SIGTERM
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return types.WrapError(err, "failed to stop process")
	}

	// Wait for graceful shutdown (with timeout)
	done := make(chan struct{})
	go func() {
		process.Wait()
		close(done)
	}()

	select {
	case <-done:
		r.logger.Info("application stopped gracefully", "app_id", appID)
	case <-time.After(10 * time.Second):
		// Force kill
		r.logger.Warn("application did not stop gracefully, forcing kill", "app_id", appID)
		process.Kill()
	}

	info.app.Status = types.AppStatusStopped
	info.app.PID = 0

	return nil
}

// Restart restarts an application
func (r *Runtime) Restart(ctx context.Context, appID string) error {
	// Get autoRestart setting before stopping
	r.mu.RLock()
	info, exists := r.apps[appID]
	r.mu.RUnlock()

	if !exists {
		return types.ErrNotFound
	}

	autoRestart := info.autoRestart

	// Stop first
	if err := r.Stop(ctx, appID); err != nil && err != types.ErrAppNotRunning {
		return err
	}

	// Wait a bit
	time.Sleep(time.Second)

	// Get app info again
	r.mu.RLock()
	info, exists = r.apps[appID]
	r.mu.RUnlock()

	if !exists {
		return types.ErrNotFound
	}

	// Start again with same autoRestart setting
	return r.start(ctx, info.app, autoRestart)
}

// Status returns the status of an application
func (r *Runtime) Status(ctx context.Context, appID string) (*types.AppStatus, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info, exists := r.apps[appID]
	if !exists {
		return nil, types.ErrNotFound
	}

	status := &types.AppStatus{
		App:     info.app,
		Healthy: info.app.Status == types.AppStatusRunning,
		Message: string(info.app.Status),
	}

	// Include health check information if available
	if info.healthChecker != nil {
		lastResult := info.healthChecker.LastResult()
		if lastResult != nil {
			status.Healthy = lastResult.Healthy
			status.Message = lastResult.Message
			status.LastHealthCheck = lastResult.Timestamp
		}
	}

	return status, nil
}

// Logs returns a stream of application logs
func (r *Runtime) Logs(ctx context.Context, appID string, follow bool) (io.ReadCloser, error) {
	r.mu.RLock()
	info, exists := r.apps[appID]
	r.mu.RUnlock()

	if !exists {
		return nil, types.ErrNotFound
	}

	logPath := filepath.Join(info.app.WorkDir, "logs", "stdout.log")

	if !follow {
		// Just return the file
		return os.Open(logPath)
	}

	// Follow logs (tail -f style)
	file, err := os.Open(logPath)
	if err != nil {
		return nil, types.WrapError(err, "failed to open log file")
	}

	// Seek to end
	file.Seek(0, io.SeekEnd)

	pr, pw := io.Pipe()

	go func() {
		defer file.Close()
		defer pw.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			pw.Write(scanner.Bytes())
			pw.Write([]byte("\n"))
		}

		// Keep checking for new lines
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				for scanner.Scan() {
					pw.Write(scanner.Bytes())
					pw.Write([]byte("\n"))
				}
			}
		}
	}()

	return pr, nil
}

// List returns all managed applications
func (r *Runtime) List(ctx context.Context) ([]*types.Application, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	apps := make([]*types.Application, 0, len(r.apps))
	for _, info := range r.apps {
		apps = append(apps, info.app)
	}

	return apps, nil
}
