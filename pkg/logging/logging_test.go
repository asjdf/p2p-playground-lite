package logging_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/asjdf/p2p-playground-lite/pkg/config"
	"github.com/asjdf/p2p-playground-lite/pkg/logging"
	"github.com/asjdf/p2p-playground-lite/pkg/types"
)

func TestNewLogger(t *testing.T) {
	cfg := &config.LoggingConfig{
		Level:  "info",
		Format: "console",
	}

	logger, err := logging.New(cfg)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	if logger == nil {
		t.Fatal("expected logger, got nil")
	}
}

func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		name          string
		level         string
		logFunc       func(types.Logger)
		shouldOutput  bool
	}{
		{
			name:  "debug level logs debug",
			level: "debug",
			logFunc: func(l types.Logger) {
				l.Debug("debug message")
			},
			shouldOutput: true,
		},
		{
			name:  "info level skips debug",
			level: "info",
			logFunc: func(l types.Logger) {
				l.Debug("debug message")
			},
			shouldOutput: false,
		},
		{
			name:  "info level logs info",
			level: "info",
			logFunc: func(l types.Logger) {
				l.Info("info message")
			},
			shouldOutput: true,
		},
		{
			name:  "warn level skips info",
			level: "warn",
			logFunc: func(l types.Logger) {
				l.Info("info message")
			},
			shouldOutput: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cfg := &config.LoggingConfig{
				Level:  tt.level,
				Format: "json",
			}

			logger, err := logging.NewWithOutput(cfg, &buf)
			if err != nil {
				t.Fatalf("failed to create logger: %v", err)
			}

			tt.logFunc(logger)

			output := buf.String()
			if tt.shouldOutput && output == "" {
				t.Error("expected log output, got empty string")
			}
			if !tt.shouldOutput && output != "" {
				t.Errorf("expected no log output, got: %s", output)
			}
		})
	}
}

func TestLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer
	cfg := &config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}

	logger, err := logging.NewWithOutput(cfg, &buf)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	logger = logger.With("app_id", "test-app", "version", "1.0.0")
	logger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "test-app") {
		t.Errorf("expected output to contain 'test-app', got: %s", output)
	}
	if !strings.Contains(output, "1.0.0") {
		t.Errorf("expected output to contain '1.0.0', got: %s", output)
	}
}

func TestLoggerFormats(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{"json format", "json"},
		{"console format", "console"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cfg := &config.LoggingConfig{
				Level:  "info",
				Format: tt.format,
			}

			logger, err := logging.NewWithOutput(cfg, &buf)
			if err != nil {
				t.Fatalf("failed to create logger: %v", err)
			}

			logger.Info("test message")

			output := buf.String()
			if output == "" {
				t.Error("expected log output, got empty string")
			}

			// JSON format should contain JSON structure
			if tt.format == "json" && !strings.Contains(output, "{") {
				t.Errorf("expected JSON format, got: %s", output)
			}
		})
	}
}

func TestLoggerError(t *testing.T) {
	var buf bytes.Buffer
	cfg := &config.LoggingConfig{
		Level:  "error",
		Format: "json",
	}

	logger, err := logging.NewWithOutput(cfg, &buf)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	logger.Error("error message", "error", "test error")

	output := buf.String()
	if !strings.Contains(output, "error message") {
		t.Errorf("expected output to contain 'error message', got: %s", output)
	}
}

func TestInvalidLogLevel(t *testing.T) {
	cfg := &config.LoggingConfig{
		Level:  "invalid",
		Format: "json",
	}

	_, err := logging.New(cfg)
	if err == nil {
		t.Error("expected error for invalid log level, got nil")
	}
}
