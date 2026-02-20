package logger

import (
	"context"
	"testing"
)

func TestLogging(t *testing.T) {
	// These just exercise the code paths since they write to stdout
	Info("test info", "key", "value")
	Error("test error", "err", "something failed")
	Debug("test debug")
	Warn("test warn")
}

func TestWith(t *testing.T) {
	l := With("module", "test")
	if l == nil {
		t.Fatal("Expected logger, got nil")
	}
}

func TestFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), TraceIDContextKey{}, "12345")
	l := FromContext(ctx)
	if l == nil {
		t.Fatal("Expected logger from context")
	}
}
