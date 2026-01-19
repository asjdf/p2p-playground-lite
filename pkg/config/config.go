package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config wraps viper for configuration management
type Config struct {
	v *viper.Viper
}

// New creates a new Config instance
func New() *Config {
	return &Config{
		v: viper.New(),
	}
}

// LoadFromFile loads configuration from a file
func (c *Config) LoadFromFile(path string) error {
	c.v.SetConfigFile(path)
	c.v.SetConfigType("yaml") // Explicitly set config type
	if err := c.v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	return nil
}

// GetString retrieves a string configuration value
func (c *Config) GetString(key string) string {
	return c.v.GetString(key)
}

// GetInt retrieves an integer configuration value
func (c *Config) GetInt(key string) int {
	return c.v.GetInt(key)
}

// GetBool retrieves a boolean configuration value
func (c *Config) GetBool(key string) bool {
	return c.v.GetBool(key)
}

// GetDuration retrieves a duration configuration value (in milliseconds)
func (c *Config) GetDuration(key string) int64 {
	return c.v.GetDuration(key).Milliseconds()
}

// GetStringSlice retrieves a string slice configuration value
func (c *Config) GetStringSlice(key string) []string {
	return c.v.GetStringSlice(key)
}

// GetStringMap retrieves a string map configuration value
func (c *Config) GetStringMap(key string) map[string]interface{} {
	return c.v.GetStringMap(key)
}

// GetStringMapString retrieves a string-to-string map configuration value
func (c *Config) GetStringMapString(key string) map[string]string {
	return c.v.GetStringMapString(key)
}

// Set sets a configuration value
func (c *Config) Set(key string, value interface{}) {
	c.v.Set(key, value)
}

// SetDefault sets a default value for a key
func (c *Config) SetDefault(key string, value interface{}) {
	c.v.SetDefault(key, value)
}

// BindEnv binds a configuration key to an environment variable
func (c *Config) BindEnv(key string, envVar string) error {
	return c.v.BindEnv(key, envVar)
}

// AutomaticEnv enables automatic environment variable binding
func (c *Config) AutomaticEnv() {
	c.v.AutomaticEnv()
}

// SetEnvPrefix sets the environment variable prefix
func (c *Config) SetEnvPrefix(prefix string) {
	c.v.SetEnvPrefix(prefix)
}

// GetViper returns the underlying viper instance
func (c *Config) GetViper() *viper.Viper {
	return c.v
}

// DaemonConfig contains daemon-specific configuration
type DaemonConfig struct {
	// Node contains P2P node configuration
	Node NodeConfig `yaml:"node" mapstructure:"node"`

	// Storage contains storage configuration
	Storage StorageConfig `yaml:"storage" mapstructure:"storage"`

	// Runtime contains runtime configuration
	Runtime RuntimeConfig `yaml:"runtime" mapstructure:"runtime"`

	// Logging contains logging configuration
	Logging LoggingConfig `yaml:"logging" mapstructure:"logging"`

	// Security contains security configuration
	Security SecurityConfig `yaml:"security" mapstructure:"security"`
}

// NodeConfig contains P2P node configuration
type NodeConfig struct {
	// ListenAddrs are the addresses to listen on
	ListenAddrs []string `yaml:"listen_addrs" mapstructure:"listen_addrs"`

	// BootstrapPeers are initial peers to connect to
	BootstrapPeers []string `yaml:"bootstrap_peers" mapstructure:"bootstrap_peers"`

	// EnableMDNS enables mDNS discovery
	EnableMDNS bool `yaml:"enable_mdns" mapstructure:"enable_mdns"`

	// Labels are node labels for organization
	Labels map[string]string `yaml:"labels" mapstructure:"labels"`

	// ID is the node ID (optional, auto-generated if not provided)
	ID string `yaml:"id" mapstructure:"id"`
}

// StorageConfig contains storage configuration
type StorageConfig struct {
	// DataDir is the base directory for all data
	DataDir string `yaml:"data_dir" mapstructure:"data_dir"`

	// PackagesDir is where packages are stored
	PackagesDir string `yaml:"packages_dir" mapstructure:"packages_dir"`

	// AppsDir is where applications are deployed
	AppsDir string `yaml:"apps_dir" mapstructure:"apps_dir"`

	// KeysDir is where cryptographic keys are stored
	KeysDir string `yaml:"keys_dir" mapstructure:"keys_dir"`
}

// RuntimeConfig contains runtime configuration
type RuntimeConfig struct {
	// MaxApps is the maximum number of concurrent applications
	MaxApps int `yaml:"max_apps" mapstructure:"max_apps"`

	// LogRetentionDays is how long to keep logs
	LogRetentionDays int `yaml:"log_retention_days" mapstructure:"log_retention_days"`

	// LogMaxSizeMB is the maximum size of a single log file
	LogMaxSizeMB int `yaml:"log_max_size_mb" mapstructure:"log_max_size_mb"`

	// LogMaxFiles is the maximum number of log files to keep
	LogMaxFiles int `yaml:"log_max_files" mapstructure:"log_max_files"`

	// EnableResourceLimits enables resource limiting
	EnableResourceLimits bool `yaml:"enable_resource_limits" mapstructure:"enable_resource_limits"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	// Level is the log level (debug, info, warn, error)
	Level string `yaml:"level" mapstructure:"level"`

	// Format is the log format (json, console)
	Format string `yaml:"format" mapstructure:"format"`

	// OutputPath is where to write logs
	OutputPath string `yaml:"output_path" mapstructure:"output_path"`

	// ErrorOutputPath is where to write error logs
	ErrorOutputPath string `yaml:"error_output_path" mapstructure:"error_output_path"`
}

// SecurityConfig contains security configuration
type SecurityConfig struct {
	// EnableAuth enables authentication
	EnableAuth bool `yaml:"enable_auth" mapstructure:"enable_auth"`

	// AuthMethod is the authentication method (psk, cert)
	AuthMethod string `yaml:"auth_method" mapstructure:"auth_method"`

	// PSK is the pre-shared key (for psk auth)
	PSK string `yaml:"psk" mapstructure:"psk"`

	// TrustedPeers are the trusted peer IDs
	TrustedPeers []string `yaml:"trusted_peers" mapstructure:"trusted_peers"`

	// AllowUnsignedPackages allows deploying packages without signatures
	AllowUnsignedPackages bool `yaml:"allow_unsigned_packages" mapstructure:"allow_unsigned_packages"`

	// PublicKeysDir is where public keys for verification are stored
	PublicKeysDir string `yaml:"public_keys_dir" mapstructure:"public_keys_dir"`
}

// ControllerConfig contains controller-specific configuration
type ControllerConfig struct {
	// Node contains P2P node configuration
	Node NodeConfig `yaml:"node" mapstructure:"node"`

	// Storage contains storage configuration
	Storage StorageConfig `yaml:"storage" mapstructure:"storage"`

	// Logging contains logging configuration
	Logging LoggingConfig `yaml:"logging" mapstructure:"logging"`

	// Security contains security configuration
	Security SecurityConfig `yaml:"security" mapstructure:"security"`

	// Deployment contains deployment defaults
	Deployment DeploymentConfig `yaml:"deployment" mapstructure:"deployment"`
}

// DeploymentConfig contains deployment configuration
type DeploymentConfig struct {
	// DefaultStrategy is the default deployment strategy
	DefaultStrategy string `yaml:"default_strategy" mapstructure:"default_strategy"`

	// Timeout is the deployment timeout
	Timeout time.Duration `yaml:"timeout" mapstructure:"timeout"`

	// RetryAttempts is the number of retry attempts
	RetryAttempts int `yaml:"retry_attempts" mapstructure:"retry_attempts"`

	// RetryDelay is the delay between retries
	RetryDelay time.Duration `yaml:"retry_delay" mapstructure:"retry_delay"`
}

// LoadDaemonConfig loads daemon configuration from a file
func LoadDaemonConfig(path string) (*DaemonConfig, error) {
	cfg := New()

	// Load from file first if provided
	if path != "" {
		if err := cfg.LoadFromFile(path); err != nil {
			return nil, err
		}
	}

	var daemonCfg DaemonConfig
	if err := cfg.GetViper().Unmarshal(&daemonCfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Apply defaults for any missing fields (only if no file was loaded)
	if path == "" {
		applyDaemonDefaults(&daemonCfg)
	}

	return &daemonCfg, nil
}

// LoadControllerConfig loads controller configuration from a file
func LoadControllerConfig(path string) (*ControllerConfig, error) {
	cfg := New()

	// Load from file first if provided
	if path != "" {
		if err := cfg.LoadFromFile(path); err != nil {
			return nil, err
		}
	}

	var controllerCfg ControllerConfig
	if err := cfg.GetViper().Unmarshal(&controllerCfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Apply defaults for any missing fields (only if no file was loaded)
	if path == "" {
		applyControllerDefaults(&controllerCfg)
	}

	return &controllerCfg, nil
}

// applyDaemonDefaults applies default values to daemon config after unmarshaling
func applyDaemonDefaults(cfg *DaemonConfig) {
	if len(cfg.Node.ListenAddrs) == 0 {
		cfg.Node.ListenAddrs = []string{"/ip4/0.0.0.0/tcp/9000", "/ip4/0.0.0.0/udp/9000/quic"}
	}
	// Always set EnableMDNS to true when applying defaults
	cfg.Node.EnableMDNS = true

	if cfg.Storage.DataDir == "" {
		cfg.Storage.DataDir = "~/.p2p-playground"
	}
	if cfg.Storage.PackagesDir == "" {
		cfg.Storage.PackagesDir = "~/.p2p-playground/packages"
	}
	if cfg.Storage.AppsDir == "" {
		cfg.Storage.AppsDir = "~/.p2p-playground/apps"
	}
	if cfg.Storage.KeysDir == "" {
		cfg.Storage.KeysDir = "~/.p2p-playground/keys"
	}

	if cfg.Runtime.MaxApps == 0 {
		cfg.Runtime.MaxApps = 10
	}
	if cfg.Runtime.LogRetentionDays == 0 {
		cfg.Runtime.LogRetentionDays = 7
	}
	if cfg.Runtime.LogMaxSizeMB == 0 {
		cfg.Runtime.LogMaxSizeMB = 10
	}
	if cfg.Runtime.LogMaxFiles == 0 {
		cfg.Runtime.LogMaxFiles = 5
	}
	// Always set EnableResourceLimits to true when applying defaults
	cfg.Runtime.EnableResourceLimits = true

	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "console"
	}
	if cfg.Logging.OutputPath == "" {
		cfg.Logging.OutputPath = "stdout"
	}
	if cfg.Logging.ErrorOutputPath == "" {
		cfg.Logging.ErrorOutputPath = "stderr"
	}

	if cfg.Security.AuthMethod == "" {
		cfg.Security.AuthMethod = "psk"
	}
}

// applyControllerDefaults applies default values to controller config after unmarshaling
func applyControllerDefaults(cfg *ControllerConfig) {
	if len(cfg.Node.ListenAddrs) == 0 {
		cfg.Node.ListenAddrs = []string{"/ip4/0.0.0.0/tcp/9001", "/ip4/0.0.0.0/udp/9001/quic"}
	}
	// Always set EnableMDNS to true when applying defaults
	cfg.Node.EnableMDNS = true

	if cfg.Storage.DataDir == "" {
		cfg.Storage.DataDir = "~/.p2p-playground-controller"
	}
	if cfg.Storage.PackagesDir == "" {
		cfg.Storage.PackagesDir = "~/.p2p-playground-controller/packages"
	}
	if cfg.Storage.KeysDir == "" {
		cfg.Storage.KeysDir = "~/.p2p-playground-controller/keys"
	}

	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "console"
	}
	if cfg.Logging.OutputPath == "" {
		cfg.Logging.OutputPath = "stdout"
	}
	if cfg.Logging.ErrorOutputPath == "" {
		cfg.Logging.ErrorOutputPath = "stderr"
	}

	if cfg.Security.AuthMethod == "" {
		cfg.Security.AuthMethod = "psk"
	}

	if cfg.Deployment.DefaultStrategy == "" {
		cfg.Deployment.DefaultStrategy = "immediate"
	}
	if cfg.Deployment.Timeout == 0 {
		cfg.Deployment.Timeout = 5 * time.Minute
	}
	if cfg.Deployment.RetryAttempts == 0 {
		cfg.Deployment.RetryAttempts = 3
	}
	if cfg.Deployment.RetryDelay == 0 {
		cfg.Deployment.RetryDelay = 10 * time.Second
	}
}
