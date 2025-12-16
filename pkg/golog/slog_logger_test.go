package golog

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestColoredHandler_Levels(t *testing.T) {
	var buf bytes.Buffer
	logger := NewSlogLogger("colored", &buf, DebugLevel)

	logger.Debug("debug message", "key", "value")
	logger.Info("info message", "count", 42)
	logger.Warn("warn message", "threshold", 100)
	logger.Error("error message", "err", "something failed")

	output := buf.String()

	// Check all levels appear
	if !strings.Contains(output, "DEBUG") {
		t.Error("expected DEBUG in output")
	}
	if !strings.Contains(output, "INFO") {
		t.Error("expected INFO in output")
	}
	if !strings.Contains(output, "WARN") {
		t.Error("expected WARN in output")
	}
	if !strings.Contains(output, "ERROR") {
		t.Error("expected ERROR in output")
	}

	// Check structured attributes appear
	if !strings.Contains(output, "key=value") {
		t.Error("expected key=value in output")
	}
	if !strings.Contains(output, "count=42") {
		t.Error("expected count=42 in output")
	}
}

func TestColoredHandler_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	// Set to WarnLevel - should filter out Debug and Info
	logger := NewSlogLogger("colored", &buf, WarnLevel)

	logger.Debug("should not appear")
	logger.Info("should not appear")
	logger.Warn("should appear")
	logger.Error("should appear")

	output := buf.String()

	if strings.Contains(output, "should not appear") {
		t.Error("debug/info messages should be filtered out")
	}
	if !strings.Contains(output, "should appear") {
		t.Error("warn/error messages should appear")
	}
}

func TestColoredHandler_WithAttrs(t *testing.T) {
	var buf bytes.Buffer
	logger := NewSlogLogger("colored", &buf, DebugLevel)

	// Create child logger with preset attributes
	childLogger := logger.With("request_id", "abc-123")
	childLogger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "request_id=abc-123") {
		t.Errorf("expected preset attribute in output, got: %s", output)
	}
}

func TestColoredHandler_WithGroup(t *testing.T) {
	var buf bytes.Buffer
	logger := NewSlogLogger("colored", &buf, DebugLevel)

	// Create grouped logger
	groupedLogger := logger.WithGroup("http")
	groupedLogger.Info("request", "method", "GET")

	output := buf.String()
	if !strings.Contains(output, "http.method=GET") {
		t.Errorf("expected grouped attribute in output, got: %s", output)
	}
}

func TestNewSlogLogger_Types(t *testing.T) {
	tests := []struct {
		loggerType string
		contains   string
	}{
		{"json", `"msg"`},   // JSON format has quoted keys
		{"text", "msg="},    // Text format has key=value
		{"colored", "INFO"}, // Colored has level prefix
		{"", "INFO"},        // Default is colored
	}

	for _, tt := range tests {
		t.Run(tt.loggerType, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewSlogLogger(tt.loggerType, &buf, InfoLevel)
			logger.Info("test")

			if !strings.Contains(buf.String(), tt.contains) {
				t.Errorf("expected %q in output for type %q, got: %s", tt.contains, tt.loggerType, buf.String())
			}
		})
	}
}

func TestSlogLevel_Mapping(t *testing.T) {
	tests := []struct {
		in       Level
		expected slog.Level
	}{
		{DebugLevel, slog.LevelDebug},
		{InfoLevel, slog.LevelInfo},
		{WarnLevel, slog.LevelWarn},
		{ErrorLevel, slog.LevelError},
		{FatalLevel, slog.LevelError}, // Fatal maps to Error in slog
	}

	for _, tt := range tests {
		if got := SlogLevel(tt.in); got != tt.expected {
			t.Errorf("SlogLevel(%d) = %v, want %v", tt.in, got, tt.expected)
		}
	}
}
