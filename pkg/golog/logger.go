package golog

import (
	"io"
	"log/slog"
	"os"
)

// Level represents the logging verbosity threshold.
// Lower values are more verbose (Debug) and higher values are more severe (Fatal).
type Level int8

// Supported logging levels in increasing order of severity.
const (
	// DebugLevel enables verbose diagnostic logs useful during development.
	DebugLevel Level = iota // most verbose
	// InfoLevel enables general informational logs for normal operation.
	InfoLevel
	// WarnLevel enables logs for unexpected but non-fatal conditions.
	WarnLevel
	// ErrorLevel enables logs for failures that require attention.
	ErrorLevel
	// FatalLevel logs a critical error and typically terminates the process.
	FatalLevel // most severe
)

// NewLogger creates a *slog.Logger with the specified handler type.
// loggerType can be:
//   - "json": JSON output for production (uses slog.JSONHandler)
//   - "text": Plain text key=value output (uses slog.TextHandler)
//   - "colored" or "simple" or default: Colored output for development (uses ColoredHandler)
//
// The prefix parameter is added as a "prefix" attribute to all log entries when using json/text handlers.
// For colored output, the prefix is currently ignored (but could be shown in the output format).
func NewLogger(loggerType string, out io.Writer, logLevel Level, prefix string) *slog.Logger {
	if out == nil {
		out = os.Stderr
	}

	slogLevel := SlogLevel(logLevel)
	opts := &slog.HandlerOptions{
		Level:     slogLevel,
		AddSource: true,
	}

	var handler slog.Handler
	switch loggerType {
	case "json", "zap": // "zap" for backwards compatibility
		handler = slog.NewJSONHandler(out, opts)
	case "text":
		handler = slog.NewTextHandler(out, opts)
	default:
		// "colored", "simple", or any other value -> development colored output
		handler = NewColoredHandler(out, opts)
	}

	logger := slog.New(handler)

	// Add prefix as an attribute if provided (for json/text handlers)
	if prefix != "" && (loggerType == "json" || loggerType == "zap" || loggerType == "text") {
		logger = logger.With("prefix", prefix)
	}

	return logger
}
