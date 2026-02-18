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

// Info logs at LevelInfo
func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

// Error logs at LevelError
func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

// Debug logs at LevelDebug
func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

// Warn logs at LevelWarn
func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

// With returns a new logger with the given attributes
func With(args ...any) *slog.Logger {
	return defaultLogger.With(args...)
}

// TraceIDContextKey is the key for the trace ID in the context
type TraceIDContextKey struct{}

// FromContext returns a logger with a trace_id from the context if it exists
func FromContext(ctx context.Context) *slog.Logger {
	if traceID, ok := ctx.Value(TraceIDContextKey{}).(string); ok {
		return defaultLogger.With("trace_id", traceID)
	}
	return defaultLogger
}
