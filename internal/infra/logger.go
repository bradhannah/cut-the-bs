// Package infra provides cross-cutting infrastructure utilities
// including structured logging configuration.
package infra

import (
	"io"
	"log/slog"
	"os"
)

// NewLogger creates a configured slog.Logger with a JSON handler.
// The level parameter controls the minimum log level. If w is nil,
// os.Stderr is used as the output writer.
func NewLogger(level slog.Level, w io.Writer) *slog.Logger {
	if w == nil {
		w = os.Stderr
	}
	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewJSONHandler(w, opts)
	return slog.New(handler)
}

// ComponentLogger returns a child logger with a "component" attribute
// set to the given name. Use this to create scoped loggers for
// specific subsystems (e.g., "sqlite", "pdf", "config").
func ComponentLogger(logger *slog.Logger, component string) *slog.Logger {
	return logger.With(slog.String("component", component))
}

// ParseLogLevel converts a string log level name to a slog.Level.
// Recognized values: "debug", "info", "warn", "error". Defaults to
// slog.LevelInfo for unrecognized values.
func ParseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
