// Package logger provides structured logging for GOSP using the standard log/slog package.
// It defaults to JSON output for compatibility with log aggregation systems like ELK and Grafana.
package logger

import (
	"context"
	"log/slog"
	"os"
)

var defaultLogger *slog.Logger

func init() {
	// Default to JSON logging for industry-standard observability (ELK/Grafana)
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	defaultLogger = slog.New(handler)
}

// Info logs a message at LevelInfo with optional key-value attributes.
func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

// Error logs a message at LevelError with optional key-value attributes.
func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

// Debug logs a message at LevelDebug with optional key-value attributes.
func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

// Warn logs a message at LevelWarn with optional key-value attributes.
func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

// With returns a new logger with the given attributes.
func With(args ...any) *slog.Logger {
	return defaultLogger.With(args...)
}

// TraceIDContextKey is the key for the trace ID in the context.
type TraceIDContextKey struct{}

// FromContext returns a logger with a trace_id from the context if it exists.
func FromContext(ctx context.Context) *slog.Logger {
	if traceID, ok := ctx.Value(TraceIDContextKey{}).(string); ok {
		return defaultLogger.With("trace_id", traceID)
	}
	return defaultLogger
}
