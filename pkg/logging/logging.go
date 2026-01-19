package logging

import (
	"fmt"
	"io"
	"os"

	"github.com/asjdf/p2p-playground-lite/pkg/config"
	"github.com/asjdf/p2p-playground-lite/pkg/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// logger wraps zap.Logger to implement types.Logger
type logger struct {
	zap *zap.Logger
}

// New creates a new logger from configuration
func New(cfg *config.LoggingConfig) (types.Logger, error) {
	// Parse log level
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	// Create encoder config
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Set up output paths
	outputPath := cfg.OutputPath
	if outputPath == "" || outputPath == "stdout" {
		outputPath = "stdout"
	}

	errorOutputPath := cfg.ErrorOutputPath
	if errorOutputPath == "" || errorOutputPath == "stderr" {
		errorOutputPath = "stderr"
	}

	// Create zap config
	zapCfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      false,
		Encoding:         cfg.Format,
		EncoderConfig:    encoderCfg,
		OutputPaths:      []string{outputPath},
		ErrorOutputPaths: []string{errorOutputPath},
	}

	// Build logger
	zapLogger, err := zapCfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return &logger{zap: zapLogger}, nil
}

// NewWithOutput creates a logger with a custom output writer (for testing)
func NewWithOutput(cfg *config.LoggingConfig, output io.Writer) (types.Logger, error) {
	// Parse log level
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	// Create encoder config
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create encoder based on format
	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	}

	// Create writer syncer
	writeSyncer := zapcore.AddSync(output)

	// Create core
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// Build logger
	zapLogger := zap.New(core)

	return &logger{zap: zapLogger}, nil
}

// Debug logs a debug message
func (l *logger) Debug(msg string, fields ...interface{}) {
	l.zap.Debug(msg, convertFields(fields)...)
}

// Info logs an info message
func (l *logger) Info(msg string, fields ...interface{}) {
	l.zap.Info(msg, convertFields(fields)...)
}

// Warn logs a warning message
func (l *logger) Warn(msg string, fields ...interface{}) {
	l.zap.Warn(msg, convertFields(fields)...)
}

// Error logs an error message
func (l *logger) Error(msg string, fields ...interface{}) {
	l.zap.Error(msg, convertFields(fields)...)
}

// With returns a logger with additional fields
func (l *logger) With(fields ...interface{}) types.Logger {
	return &logger{zap: l.zap.With(convertFields(fields)...)}
}

// Sync flushes any buffered log entries
func (l *logger) Sync() error {
	return l.zap.Sync()
}

// parseLevel parses a log level string
func parseLevel(level string) (zapcore.Level, error) {
	switch level {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn", "warning":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	case "dpanic":
		return zapcore.DPanicLevel, nil
	case "panic":
		return zapcore.PanicLevel, nil
	case "fatal":
		return zapcore.FatalLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("invalid log level: %s", level)
	}
}

// convertFields converts a variadic list of key-value pairs to zap.Field
func convertFields(fields []interface{}) []zap.Field {
	if len(fields) == 0 {
		return nil
	}

	zapFields := make([]zap.Field, 0, len(fields)/2)
	for i := 0; i < len(fields)-1; i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			continue
		}
		zapFields = append(zapFields, zap.Any(key, fields[i+1]))
	}
	return zapFields
}

// NewNopLogger creates a no-op logger that discards all logs
func NewNopLogger() types.Logger {
	return &logger{zap: zap.NewNop()}
}

// NewDevelopmentLogger creates a logger suitable for development
func NewDevelopmentLogger() (types.Logger, error) {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	return &logger{zap: zapLogger}, nil
}

// NewProductionLogger creates a logger suitable for production
func NewProductionLogger() (types.Logger, error) {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return &logger{zap: zapLogger}, nil
}

// MustNew creates a logger or panics on error
func MustNew(cfg *config.LoggingConfig) types.Logger {
	logger, err := New(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}
	return logger
}

// Global logger instance
var globalLogger types.Logger = &logger{zap: zap.NewNop()}

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(l types.Logger) {
	globalLogger = l
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() types.Logger {
	return globalLogger
}

// Debug logs a debug message using the global logger
func Debug(msg string, fields ...interface{}) {
	globalLogger.Debug(msg, fields...)
}

// Info logs an info message using the global logger
func Info(msg string, fields ...interface{}) {
	globalLogger.Info(msg, fields...)
}

// Warn logs a warning message using the global logger
func Warn(msg string, fields ...interface{}) {
	globalLogger.Warn(msg, fields...)
}

// Error logs an error message using the global logger
func Error(msg string, fields ...interface{}) {
	globalLogger.Error(msg, fields...)
}

// InitFromConfig initializes the global logger from configuration
func InitFromConfig(cfg *config.LoggingConfig) error {
	logger, err := New(cfg)
	if err != nil {
		return err
	}
	SetGlobalLogger(logger)
	return nil
}

// InitForCLI initializes logging for CLI usage
func InitForCLI(level string) error {
	cfg := &config.LoggingConfig{
		Level:           level,
		Format:          "console",
		OutputPath:      "stdout",
		ErrorOutputPath: "stderr",
	}
	return InitFromConfig(cfg)
}

// RedirectStdLog redirects standard library log to zap
func RedirectStdLog(l types.Logger) func() {
	zapLogger := l.(*logger).zap
	return zap.RedirectStdLog(zapLogger)
}

// DefaultLogger creates a logger with default settings
func DefaultLogger() types.Logger {
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	writer := zapcore.Lock(os.Stdout)
	core := zapcore.NewCore(encoder, writer, zapcore.InfoLevel)
	zapLogger := zap.New(core)
	return &logger{zap: zapLogger}
}
