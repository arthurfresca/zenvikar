package logger

import (
	"log/slog"
	"testing"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"INFO", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"WARN", slog.LevelWarn},
		{"error", slog.LevelError},
		{"ERROR", slog.LevelError},
		{"", slog.LevelInfo},
		{"unknown", slog.LevelInfo},
		{" debug ", slog.LevelDebug},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseLevel(tt.input)
			if got != tt.expected {
				t.Errorf("parseLevel(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestNew(t *testing.T) {
	t.Setenv("LOG_LEVEL", "debug")
	logger := New()
	if logger == nil {
		t.Fatal("New() returned nil")
	}
	if !logger.Enabled(nil, slog.LevelDebug) {
		t.Error("expected debug level to be enabled")
	}
}

func TestNew_DefaultLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "")
	logger := New()
	if logger == nil {
		t.Fatal("New() returned nil")
	}
	if logger.Enabled(nil, slog.LevelDebug) {
		t.Error("expected debug level to be disabled at default info level")
	}
	if !logger.Enabled(nil, slog.LevelInfo) {
		t.Error("expected info level to be enabled")
	}
}
