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

	"github.com/asjdf/p2p-playground-lite/pkg/types"
)

// Runtime manages application processes
type Runtime struct {
	apps   map[string]*types.Application
	mu     sync.RWMutex
	logger types.Logger
}

// New creates a new runtime
func New(logger types.Logger) *Runtime {
	return &Runtime{
		apps:   make(map[string]*types.Application),
		logger: logger,
	}
}

// Start starts an application
func (r *Runtime) Start(ctx context.Context, app *types.Application) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if already running
	if existing, exists := r.apps[app.ID]; exists {
		if existing.Status == types.AppStatusRunning {
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

	// Store application
	r.apps[app.ID] = app

	// Monitor process in background
	go func() {
		defer stdoutFile.Close()
		defer stderrFile.Close()

		err := cmd.Wait()

		r.mu.Lock()
		defer r.mu.Unlock()

		if app, exists := r.apps[app.ID]; exists {
			if err != nil {
				app.Status = types.AppStatusFailed
				r.logger.Error("application exited with error",
					"app_id", app.ID,
					"error", err,
				)
			} else {
				app.Status = types.AppStatusStopped
				r.logger.Info("application stopped",
					"app_id", app.ID,
				)
			}
			app.PID = 0
		}
	}()

	r.logger.Info("application started",
		"app_id", app.ID,
		"pid", app.PID,
	)

	return nil
}

// Stop stops a running application
func (r *Runtime) Stop(ctx context.Context, appID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	app, exists := r.apps[appID]
	if !exists {
		return types.ErrNotFound
	}

	if app.Status != types.AppStatusRunning {
		return types.ErrAppNotRunning
	}

	// Find process
	process, err := os.FindProcess(app.PID)
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

	app.Status = types.AppStatusStopped
	app.PID = 0

	return nil
}

// Restart restarts an application
func (r *Runtime) Restart(ctx context.Context, appID string) error {
	// Stop first
	if err := r.Stop(ctx, appID); err != nil && err != types.ErrAppNotRunning {
		return err
	}

	// Wait a bit
	time.Sleep(time.Second)

	// Get app info
	r.mu.RLock()
	app, exists := r.apps[appID]
	r.mu.RUnlock()

	if !exists {
		return types.ErrNotFound
	}

	// Start again
	return r.Start(ctx, app)
}

// Status returns the status of an application
func (r *Runtime) Status(ctx context.Context, appID string) (*types.AppStatus, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	app, exists := r.apps[appID]
	if !exists {
		return nil, types.ErrNotFound
	}

	return &types.AppStatus{
		App:     app,
		Healthy: app.Status == types.AppStatusRunning,
		Message: string(app.Status),
	}, nil
}

// Logs returns a stream of application logs
func (r *Runtime) Logs(ctx context.Context, appID string, follow bool) (io.ReadCloser, error) {
	r.mu.RLock()
	app, exists := r.apps[appID]
	r.mu.RUnlock()

	if !exists {
		return nil, types.ErrNotFound
	}

	logPath := filepath.Join(app.WorkDir, "logs", "stdout.log")

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
	for _, app := range r.apps {
		apps = append(apps, app)
	}

	return apps, nil
}
