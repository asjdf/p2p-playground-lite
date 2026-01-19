package health

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/asjdf/p2p-playground-lite/pkg/types"
)

// CheckType represents the type of health check
type CheckType string

const (
	// CheckTypeProcess checks if the process is running
	CheckTypeProcess CheckType = "process"
	// CheckTypeHTTP performs an HTTP health check
	CheckTypeHTTP CheckType = "http"
	// CheckTypeTCP performs a TCP health check
	CheckTypeTCP CheckType = "tcp"
)

// Config contains health check configuration
type Config struct {
	// Type of health check
	Type CheckType `yaml:"type" mapstructure:"type"`

	// Interval between health checks
	Interval time.Duration `yaml:"interval" mapstructure:"interval"`

	// Timeout for each health check
	Timeout time.Duration `yaml:"timeout" mapstructure:"timeout"`

	// Retries before considering the app unhealthy
	Retries int `yaml:"retries" mapstructure:"retries"`

	// HTTPPath is the HTTP endpoint path (for HTTP checks)
	HTTPPath string `yaml:"http_path,omitempty" mapstructure:"http_path"`

	// HTTPPort is the HTTP port to check (for HTTP checks)
	HTTPPort int `yaml:"http_port,omitempty" mapstructure:"http_port"`

	// TCPPort is the TCP port to check (for TCP checks)
	TCPPort int `yaml:"tcp_port,omitempty" mapstructure:"tcp_port"`
}

// Result represents the result of a health check
type Result struct {
	// Healthy indicates if the application is healthy
	Healthy bool

	// Message provides details about the health check result
	Message string

	// Timestamp of the health check
	Timestamp time.Time

	// FailureCount is the number of consecutive failures
	FailureCount int
}

// Checker performs health checks
type Checker struct {
	config *Config
	logger types.Logger
	pid    int

	// State
	lastResult       *Result
	consecutiveFails int
}

// New creates a new health checker
func New(config *Config, pid int, logger types.Logger) *Checker {
	return &Checker{
		config: config,
		logger: logger,
		pid:    pid,
		lastResult: &Result{
			Healthy:   true,
			Message:   "Not checked yet",
			Timestamp: time.Now(),
		},
	}
}

// Check performs a health check
func (c *Checker) Check(ctx context.Context) (*Result, error) {
	var err error
	var healthy bool
	var message string

	checkCtx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	switch c.config.Type {
	case CheckTypeProcess:
		healthy, message, err = c.checkProcess()
	case CheckTypeHTTP:
		healthy, message, err = c.checkHTTP(checkCtx)
	case CheckTypeTCP:
		healthy, message, err = c.checkTCP(checkCtx)
	default:
		return nil, fmt.Errorf("unsupported health check type: %s", c.config.Type)
	}

	if err != nil {
		return nil, err
	}

	// Update consecutive failure count
	if !healthy {
		c.consecutiveFails++
	} else {
		c.consecutiveFails = 0
	}

	result := &Result{
		Healthy:      healthy && c.consecutiveFails < c.config.Retries,
		Message:      message,
		Timestamp:    time.Now(),
		FailureCount: c.consecutiveFails,
	}

	c.lastResult = result
	return result, nil
}

// checkProcess checks if the process is running
func (c *Checker) checkProcess() (bool, string, error) {
	// Try to send signal 0 to check if process exists
	process, err := os.FindProcess(c.pid)
	if err != nil {
		return false, fmt.Sprintf("process not found: %v", err), nil
	}

	// Send signal 0 (doesn't actually send a signal, just checks if process exists)
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		return false, fmt.Sprintf("process not responding: %v", err), nil
	}

	return true, "process is running", nil
}

// checkHTTP performs an HTTP health check
func (c *Checker) checkHTTP(ctx context.Context) (bool, string, error) {
	if c.config.HTTPPort == 0 {
		return false, "HTTP port not configured", fmt.Errorf("HTTP port not configured")
	}

	path := c.config.HTTPPath
	if path == "" {
		path = "/health"
	}

	url := fmt.Sprintf("http://localhost:%d%s", c.config.HTTPPort, path)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, fmt.Sprintf("failed to create HTTP request: %v", err), nil
	}

	client := &http.Client{
		Timeout: c.config.Timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("HTTP check failed: %v", err), nil
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, fmt.Sprintf("HTTP check returned status %d", resp.StatusCode), nil
	}

	return true, fmt.Sprintf("HTTP check passed (status %d)", resp.StatusCode), nil
}

// checkTCP performs a TCP health check
func (c *Checker) checkTCP(ctx context.Context) (bool, string, error) {
	if c.config.TCPPort == 0 {
		return false, "TCP port not configured", fmt.Errorf("TCP port not configured")
	}

	addr := fmt.Sprintf("localhost:%d", c.config.TCPPort)

	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return false, fmt.Sprintf("TCP connection failed: %v", err), nil
	}
	_ = conn.Close()

	return true, "TCP connection successful", nil
}

// LastResult returns the last health check result
func (c *Checker) LastResult() *Result {
	return c.lastResult
}

// IsHealthy returns true if the application is healthy
func (c *Checker) IsHealthy() bool {
	return c.lastResult != nil && c.lastResult.Healthy
}

// StartMonitoring starts continuous health monitoring
func (c *Checker) StartMonitoring(ctx context.Context, onUnhealthy func(*Result)) {
	ticker := time.NewTicker(c.config.Interval)
	defer ticker.Stop()

	c.logger.Info("starting health monitoring",
		"type", c.config.Type,
		"interval", c.config.Interval,
		"retries", c.config.Retries)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("stopping health monitoring")
			return
		case <-ticker.C:
			result, err := c.Check(ctx)
			if err != nil {
				c.logger.Error("health check error", "error", err)
				continue
			}

			if !result.Healthy {
				c.logger.Warn("application unhealthy",
					"message", result.Message,
					"failures", result.FailureCount,
					"threshold", c.config.Retries)

				if onUnhealthy != nil {
					onUnhealthy(result)
				}
			} else {
				c.logger.Debug("health check passed", "message", result.Message)
			}
		}
	}
}
