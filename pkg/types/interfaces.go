package types

import (
	"context"
	"io"
)

// Host represents a P2P network host
type Host interface {
	// ID returns the host's unique identifier
	ID() string

	// Addrs returns the host's listening addresses
	Addrs() []string

	// Connect establishes a connection to a peer
	Connect(ctx context.Context, addr string) error

	// NewStream creates a new stream to a peer with the given protocol
	NewStream(ctx context.Context, peerID string, protocol string) (Stream, error)

	// SetStreamHandler registers a handler for incoming streams with the given protocol
	SetStreamHandler(protocol string, handler StreamHandler)

	// Close shuts down the host
	Close() error
}

// Stream represents a bidirectional data stream
type Stream interface {
	io.ReadWriteCloser

	// Reset closes the stream abruptly
	Reset() error
}

// StreamHandler handles incoming streams
type StreamHandler func(Stream)

// Runtime manages application lifecycle
type Runtime interface {
	// Start starts an application
	Start(ctx context.Context, app *Application) error

	// Stop stops a running application
	Stop(ctx context.Context, appID string) error

	// Restart restarts an application
	Restart(ctx context.Context, appID string) error

	// Status returns the status of an application
	Status(ctx context.Context, appID string) (*AppStatus, error)

	// Logs returns a stream of application logs
	Logs(ctx context.Context, appID string, follow bool) (io.ReadCloser, error)

	// List returns all managed applications
	List(ctx context.Context) ([]*Application, error)
}

// HealthChecker checks application health
type HealthChecker interface {
	// Check performs a health check on the application
	Check(ctx context.Context, app *Application) error
}

// ResourceLimiter manages resource limits for processes
type ResourceLimiter interface {
	// Apply applies resource limits to a process
	Apply(ctx context.Context, pid int, limits *ResourceLimits) error

	// Release removes resource limits from a process
	Release(ctx context.Context, pid int) error
}

// PackageManager handles application packaging
type PackageManager interface {
	// Pack creates a package from an application directory
	Pack(ctx context.Context, appDir string) (string, error)

	// Unpack extracts a package to a destination directory and returns the manifest
	Unpack(ctx context.Context, pkgPath string, destDir string) (*Manifest, error)

	// Verify verifies a package's integrity and signature
	Verify(ctx context.Context, pkgPath string, signature []byte) error

	// GetManifest reads and parses a manifest file
	GetManifest(ctx context.Context, pkgPath string) (*Manifest, error)
}

// Signer signs and verifies data
type Signer interface {
	// Sign signs data and returns the signature
	Sign(data []byte) ([]byte, error)

	// Verify verifies a signature against data using a public key
	Verify(data []byte, signature []byte, publicKey []byte) error
}

// Authenticator authenticates peers
type Authenticator interface {
	// Authenticate verifies that a peer is authorized
	Authenticate(ctx context.Context, peerID string) error
}

// Storage provides persistent storage
type Storage interface {
	// Save stores data under a key
	Save(ctx context.Context, key string, data []byte) error

	// Load retrieves data by key
	Load(ctx context.Context, key string) ([]byte, error)

	// Delete removes data by key
	Delete(ctx context.Context, key string) error

	// List returns all keys with the given prefix
	List(ctx context.Context, prefix string) ([]string, error)

	// Exists checks if a key exists
	Exists(ctx context.Context, key string) (bool, error)
}

// VersionManager manages application versions
type VersionManager interface {
	// Store stores a new version of an application
	Store(ctx context.Context, app *Application) error

	// Get retrieves a specific version of an application
	Get(ctx context.Context, name, version string) (*Application, error)

	// List returns all versions of an application
	List(ctx context.Context, name string) ([]*Application, error)

	// Delete removes a specific version
	Delete(ctx context.Context, name, version string) error

	// GetLatest returns the latest version of an application
	GetLatest(ctx context.Context, name string) (*Application, error)
}

// TransferManager handles file transfers over P2P network
type TransferManager interface {
	// Send sends a file to a peer
	Send(ctx context.Context, peerID string, filePath string, progress ProgressCallback) error

	// Receive receives a file from a peer
	Receive(ctx context.Context, stream Stream, destPath string, progress ProgressCallback) error
}

// ProgressCallback reports transfer progress
type ProgressCallback func(current, total int64)

// Logger provides structured logging
type Logger interface {
	// Debug logs a debug message
	Debug(msg string, fields ...interface{})

	// Info logs an info message
	Info(msg string, fields ...interface{})

	// Warn logs a warning message
	Warn(msg string, fields ...interface{})

	// Error logs an error message
	Error(msg string, fields ...interface{})

	// With returns a logger with additional fields
	With(fields ...interface{}) Logger
}

// Config represents configuration for the application
type Config interface {
	// GetString retrieves a string configuration value
	GetString(key string) string

	// GetInt retrieves an integer configuration value
	GetInt(key string) int

	// GetBool retrieves a boolean configuration value
	GetBool(key string) bool

	// GetDuration retrieves a duration configuration value
	GetDuration(key string) int64

	// Set sets a configuration value
	Set(key string, value interface{})
}
