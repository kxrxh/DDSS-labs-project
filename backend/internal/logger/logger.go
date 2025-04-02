package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Log is the global logger instance
	Log *zap.Logger
)

// Config holds logger configuration
type Config struct {
	Level      string `mapstructure:"level"`
	Encoding   string `mapstructure:"encoding"`
	OutputPath string `mapstructure:"output_path"`
}

// Initialize sets up the logger with the given configuration
func Initialize(cfg *Config) error {
	// Set default values if not provided
	if cfg.Level == "" {
		cfg.Level = "info"
	}
	if cfg.Encoding == "" {
		cfg.Encoding = "json"
	}
	if cfg.OutputPath == "" {
		cfg.OutputPath = "stdout"
	}

	// Parse log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return fmt.Errorf("invalid log level %s: %v", cfg.Level, err)
	}

	// Create logger configuration
	zapConfig := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      false,
		Encoding:         cfg.Encoding,
		EncoderConfig:    getEncoderConfig(),
		OutputPaths:      []string{cfg.OutputPath},
		ErrorOutputPaths: []string{cfg.OutputPath},
	}

	// Build the logger
	logger, err := zapConfig.Build()
	if err != nil {
		return fmt.Errorf("failed to build logger: %v", err)
	}

	// Replace global logger
	Log = logger
	return nil
}

// getEncoderConfig returns the encoder configuration for the logger
func getEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// Sync flushes any buffered log entries
func Sync() error {
	if Log != nil {
		return Log.Sync()
	}
	return nil
}

// With creates a child logger with the given fields
func With(fields ...zap.Field) *zap.Logger {
	return Log.With(fields...)
}

// Named creates a child logger with the given name
func Named(name string) *zap.Logger {
	return Log.Named(name)
}
