package types

import (
	"time"
)

// Application represents a deployed application
type Application struct {
	// ID is the unique identifier for this application instance
	ID string `json:"id"`

	// Name is the application name
	Name string `json:"name"`

	// Version is the semantic version
	Version string `json:"version"`

	// PackagePath is the path to the package file
	PackagePath string `json:"package_path"`

	// Manifest contains application metadata
	Manifest *Manifest `json:"manifest"`

	// Status is the current status of the application
	Status AppStatusType `json:"status"`

	// PID is the process ID (0 if not running)
	PID int `json:"pid,omitempty"`

	// StartedAt is when the application was started
	StartedAt time.Time `json:"started_at,omitempty"`

	// Labels are key-value pairs for organization
	Labels map[string]string `json:"labels,omitempty"`

	// WorkDir is the working directory for the application
	WorkDir string `json:"work_dir"`
}

// AppStatusType represents the status of an application
type AppStatusType string

const (
	// AppStatusStopped indicates the application is stopped
	AppStatusStopped AppStatusType = "stopped"

	// AppStatusStarting indicates the application is starting
	AppStatusStarting AppStatusType = "starting"

	// AppStatusRunning indicates the application is running
	AppStatusRunning AppStatusType = "running"

	// AppStatusFailed indicates the application has failed
	AppStatusFailed AppStatusType = "failed"

	// AppStatusRestarting indicates the application is restarting
	AppStatusRestarting AppStatusType = "restarting"
)

// AppStatus contains detailed status information
type AppStatus struct {
	// App is the application reference
	App *Application `json:"app"`

	// Healthy indicates if the application passed health checks
	Healthy bool `json:"healthy"`

	// Message provides additional status information
	Message string `json:"message,omitempty"`

	// LastHealthCheck is when the last health check was performed
	LastHealthCheck time.Time `json:"last_health_check,omitempty"`

	// ResourceUsage contains current resource usage
	ResourceUsage *ResourceUsage `json:"resource_usage,omitempty"`
}

// Manifest describes an application package
type Manifest struct {
	// Name is the application name
	Name string `yaml:"name" json:"name"`

	// Version is the semantic version
	Version string `yaml:"version" json:"version"`

	// Description is a human-readable description
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// Entrypoint is the main executable path (relative to package)
	Entrypoint string `yaml:"entrypoint" json:"entrypoint"`

	// Args are command-line arguments
	Args []string `yaml:"args,omitempty" json:"args,omitempty"`

	// Env contains environment variables
	Env map[string]string `yaml:"env,omitempty" json:"env,omitempty"`

	// Resources specifies resource limits
	Resources *ResourceLimits `yaml:"resources,omitempty" json:"resources,omitempty"`

	// HealthCheck specifies health check configuration
	HealthCheck *HealthCheckConfig `yaml:"health_check,omitempty" json:"health_check,omitempty"`

	// Hooks contains lifecycle hooks
	Hooks *LifecycleHooks `yaml:"hooks,omitempty" json:"hooks,omitempty"`

	// Dependencies lists other applications this depends on
	Dependencies []string `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`

	// Labels are key-value pairs for organization
	Labels map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
}

// ResourceLimits specifies resource constraints
type ResourceLimits struct {
	// CPUPercent is the CPU limit as a percentage (0-100 per core)
	CPUPercent float64 `yaml:"cpu_percent,omitempty" json:"cpu_percent,omitempty"`

	// MemoryMB is the memory limit in megabytes
	MemoryMB int64 `yaml:"memory_mb,omitempty" json:"memory_mb,omitempty"`
}

// ResourceUsage contains current resource consumption
type ResourceUsage struct {
	// CPUPercent is current CPU usage
	CPUPercent float64 `json:"cpu_percent"`

	// MemoryMB is current memory usage in megabytes
	MemoryMB int64 `json:"memory_mb"`

	// Timestamp is when the measurement was taken
	Timestamp time.Time `json:"timestamp"`
}

// HealthCheckConfig specifies how to check application health
type HealthCheckConfig struct {
	// Type is the health check type: "http", "tcp", or "process"
	Type string `yaml:"type" json:"type"`

	// Endpoint is the URL or address to check (for http/tcp)
	Endpoint string `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`

	// Interval is how often to perform health checks
	Interval time.Duration `yaml:"interval,omitempty" json:"interval,omitempty"`

	// Timeout is the maximum time to wait for a health check
	Timeout time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty"`

	// Retries is the number of consecutive failures before marking unhealthy
	Retries int `yaml:"retries,omitempty" json:"retries,omitempty"`

	// StartPeriod is the initial grace period before starting health checks
	StartPeriod time.Duration `yaml:"start_period,omitempty" json:"start_period,omitempty"`
}

// LifecycleHooks specifies scripts to run at various lifecycle stages
type LifecycleHooks struct {
	// PreStart runs before the application starts
	PreStart string `yaml:"pre_start,omitempty" json:"pre_start,omitempty"`

	// PostStart runs after the application starts
	PostStart string `yaml:"post_start,omitempty" json:"post_start,omitempty"`

	// PreStop runs before the application stops
	PreStop string `yaml:"pre_stop,omitempty" json:"pre_stop,omitempty"`

	// PostStop runs after the application stops
	PostStop string `yaml:"post_stop,omitempty" json:"post_stop,omitempty"`
}

// NodeInfo represents information about a node
type NodeInfo struct {
	// ID is the node's unique identifier
	ID string `json:"id"`

	// Addrs are the node's listening addresses
	Addrs []string `json:"addrs"`

	// Labels are key-value pairs for organization
	Labels map[string]string `json:"labels,omitempty"`

	// Apps are the applications running on this node
	Apps []*Application `json:"apps,omitempty"`

	// LastSeen is when this node was last contacted
	LastSeen time.Time `json:"last_seen"`

	// Version is the daemon version
	Version string `json:"version"`
}

// DeploymentConfig specifies how to deploy an application
type DeploymentConfig struct {
	// Package is the path to the package file
	Package string `json:"package"`

	// Nodes are the target node IDs or addresses
	Nodes []string `json:"nodes,omitempty"`

	// NodeSelector selects nodes by labels
	NodeSelector map[string]string `json:"node_selector,omitempty"`

	// Strategy is the deployment strategy: "immediate", "graceful", "manual"
	Strategy string `json:"strategy,omitempty"`

	// AutoRestart enables automatic restart on failure
	AutoRestart bool `json:"auto_restart,omitempty"`

	// Signature is the package signature (if signed)
	Signature []byte `json:"signature,omitempty"`
}

// TransferProgress reports file transfer progress
type TransferProgress struct {
	// Current is the number of bytes transferred
	Current int64 `json:"current"`

	// Total is the total number of bytes to transfer
	Total int64 `json:"total"`

	// Percentage is the completion percentage (0-100)
	Percentage float64 `json:"percentage"`
}

// LogEntry represents a log entry from an application
type LogEntry struct {
	// Timestamp is when the log was generated
	Timestamp time.Time `json:"timestamp"`

	// AppID is the application ID
	AppID string `json:"app_id"`

	// Level is the log level (stdout/stderr)
	Level string `json:"level"`

	// Message is the log message
	Message string `json:"message"`
}

// VersionInfo represents version information
type VersionInfo struct {
	// Version is the version string
	Version string `json:"version"`

	// Major is the major version number
	Major int `json:"major"`

	// Minor is the minor version number
	Minor int `json:"minor"`

	// Patch is the patch version number
	Patch int `json:"patch"`

	// Prerelease is the prerelease identifier
	Prerelease string `json:"prerelease,omitempty"`

	// Metadata is the build metadata
	Metadata string `json:"metadata,omitempty"`
}

// UpdateStrategy specifies how to handle updates
type UpdateStrategy string

const (
	// UpdateStrategyImmediate updates immediately
	UpdateStrategyImmediate UpdateStrategy = "immediate"

	// UpdateStrategyGraceful waits for the application to stop gracefully
	UpdateStrategyGraceful UpdateStrategy = "graceful"

	// UpdateStrategyManual requires manual approval
	UpdateStrategyManual UpdateStrategy = "manual"
)

// KeyPair represents a cryptographic key pair
type KeyPair struct {
	// PublicKey is the public key
	PublicKey []byte `json:"public_key"`

	// PrivateKey is the private key (should be kept secure)
	PrivateKey []byte `json:"private_key,omitempty"`
}

// PackageInfo contains metadata about a package
type PackageInfo struct {
	// Name is the package name
	Name string `json:"name"`

	// Version is the package version
	Version string `json:"version"`

	// Path is the local path to the package
	Path string `json:"path"`

	// Size is the package size in bytes
	Size int64 `json:"size"`

	// Checksum is the SHA-256 checksum
	Checksum string `json:"checksum"`

	// CreatedAt is when the package was created
	CreatedAt time.Time `json:"created_at"`

	// Signed indicates if the package is signed
	Signed bool `json:"signed"`
}
