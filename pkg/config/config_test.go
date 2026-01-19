package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/asjdf/p2p-playground-lite/pkg/config"
)

func TestNewConfig(t *testing.T) {
	cfg := config.New()
	if cfg == nil {
		t.Fatal("expected config, got nil")
	}
}

func TestConfigSetGet(t *testing.T) {
	cfg := config.New()

	tests := []struct {
		name     string
		setKey   string
		setValue interface{}
		getFunc  func(string) interface{}
		want     interface{}
	}{
		{
			name:     "string value",
			setKey:   "test.string",
			setValue: "hello",
			getFunc:  func(k string) interface{} { return cfg.GetString(k) },
			want:     "hello",
		},
		{
			name:     "int value",
			setKey:   "test.int",
			setValue: 42,
			getFunc:  func(k string) interface{} { return cfg.GetInt(k) },
			want:     42,
		},
		{
			name:     "bool value",
			setKey:   "test.bool",
			setValue: true,
			getFunc:  func(k string) interface{} { return cfg.GetBool(k) },
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg.Set(tt.setKey, tt.setValue)
			got := tt.getFunc(tt.setKey)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigDefaults(t *testing.T) {
	cfg := config.New()
	cfg.SetDefault("test.default", "default_value")

	got := cfg.GetString("test.default")
	if got != "default_value" {
		t.Errorf("got %v, want 'default_value'", got)
	}
}

func TestLoadDaemonConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "daemon.yaml")

	configContent := `
node:
  listen_addrs:
    - /ip4/0.0.0.0/tcp/9000
  enable_mdns: true
  labels:
    env: test
    region: us-west

storage:
  data_dir: /tmp/test-data

runtime:
  max_apps: 5
  log_retention_days: 3

logging:
  level: debug
  format: json

security:
  enable_auth: true
  auth_method: psk
  psk: test-secret-key
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Load config
	cfg, err := config.LoadDaemonConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Verify values
	if !cfg.Node.EnableMDNS {
		t.Error("expected enable_mdns to be true")
	}

	if cfg.Node.Labels["env"] != "test" {
		t.Errorf("got label env=%v, want 'test'", cfg.Node.Labels["env"])
	}

	if cfg.Storage.DataDir != "/tmp/test-data" {
		t.Errorf("got data_dir=%v, want '/tmp/test-data'", cfg.Storage.DataDir)
	}

	if cfg.Runtime.MaxApps != 5 {
		t.Errorf("got max_apps=%v, want 5", cfg.Runtime.MaxApps)
	}

	if cfg.Logging.Level != "debug" {
		t.Errorf("got level=%v, want 'debug'", cfg.Logging.Level)
	}

	if cfg.Security.PSK != "test-secret-key" {
		t.Errorf("got psk=%v, want 'test-secret-key'", cfg.Security.PSK)
	}
}

func TestLoadControllerConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "controller.yaml")

	configContent := `
node:
  listen_addrs:
    - /ip4/0.0.0.0/tcp/9001
  enable_mdns: true

deployment:
  default_strategy: graceful
  timeout: 10m
  retry_attempts: 5
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Load config
	cfg, err := config.LoadControllerConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Verify values
	if !cfg.Node.EnableMDNS {
		t.Error("expected enable_mdns to be true")
	}

	if cfg.Deployment.DefaultStrategy != "graceful" {
		t.Errorf("got strategy=%v, want 'graceful'", cfg.Deployment.DefaultStrategy)
	}

	if cfg.Deployment.RetryAttempts != 5 {
		t.Errorf("got retry_attempts=%v, want 5", cfg.Deployment.RetryAttempts)
	}
}

func TestLoadConfigWithDefaults(t *testing.T) {
	// Load with empty path to use defaults
	cfg, err := config.LoadDaemonConfig("")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Verify defaults are set
	if len(cfg.Node.ListenAddrs) == 0 {
		t.Error("expected default listen_addrs to be set")
	}

	if !cfg.Node.EnableMDNS {
		t.Error("expected default enable_mdns to be true")
	}

	if cfg.Runtime.MaxApps != 10 {
		t.Errorf("got max_apps=%v, want default 10", cfg.Runtime.MaxApps)
	}

	if cfg.Logging.Level != "info" {
		t.Errorf("got level=%v, want default 'info'", cfg.Logging.Level)
	}
}
