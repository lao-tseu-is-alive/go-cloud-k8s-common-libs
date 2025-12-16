package golog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"sync"
)

// Define ANSI color codes
const (
	reset                  = "\033[0m"
	cyan                   = "\033[36m"
	whiteHighIntensity     = "\033[1;97m"
	yellowHighIntensity    = "\033[1;93m"
	redBackGroundWhiteText = "\033[1;97;41m" // Red background with white text
)

// ColoredHandler is a slog.Handler that outputs colored, human-friendly logs for development.
// It mimics the SimpleLogger output style with emojis and ANSI colors.
type ColoredHandler struct {
	out    io.Writer
	level  slog.Level
	attrs  []slog.Attr
	groups []string
	mu     *sync.Mutex
}

// NewColoredHandler creates a handler with colored output for development.
func NewColoredHandler(out io.Writer, opts *slog.HandlerOptions) *ColoredHandler {
	level := slog.LevelInfo
	if opts != nil && opts.Level != nil {
		level = opts.Level.Level()
	}
	return &ColoredHandler{out: out, level: level, mu: &sync.Mutex{}}
}

func (h *ColoredHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *ColoredHandler) Handle(_ context.Context, r slog.Record) error {
	// Acquire lock before writing anything
	h.mu.Lock()
	defer h.mu.Unlock()
	// Pick color and prefix based on level
	var color, levelStr string
	switch {
	case r.Level < slog.LevelInfo:
		color, levelStr = cyan, "DEBUG"
	case r.Level < slog.LevelWarn:
		color, levelStr = whiteHighIntensity, "📣 INFO "
	case r.Level < slog.LevelError:
		color, levelStr = yellowHighIntensity, "🚩 WARN "
	default:
		color, levelStr = redBackGroundWhiteText, "⚠️ ⚡ ERROR"
	}

	// Get caller info (mimic log.Lshortfile behavior)
	var file string
	var line int
	if r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		file = filepath.Base(f.File)
		line = f.Line
	}

	// Format timestamp like standard log package
	timestamp := r.Time.Format("2006/01/02 15:04:05")

	// Build the log line
	fmt.Fprintf(h.out, "%s %s:%d %s%s: %s", timestamp, file, line, color, levelStr, r.Message)

	// Append any pre-set attributes from WithAttrs
	for _, a := range h.attrs {
		h.writeAttr(a)
	}

	// Append record attributes (key=value pairs)
	r.Attrs(func(a slog.Attr) bool {
		h.writeAttr(a)
		return true
	})

	fmt.Fprintf(h.out, "%s\n", reset)
	return nil
}

func (h *ColoredHandler) writeAttr(a slog.Attr) {
	// Skip empty attributes
	if a.Equal(slog.Attr{}) {
		return
	}
	// Handle groups by prefixing with group name
	key := a.Key
	for _, g := range h.groups {
		key = g + "." + key
	}
	fmt.Fprintf(h.out, " %s=%v", key, a.Value)
}

func (h *ColoredHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &ColoredHandler{
		out:    h.out,
		level:  h.level,
		attrs:  newAttrs,
		groups: h.groups,
		mu:     h.mu,
	}
}

func (h *ColoredHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name
	return &ColoredHandler{
		out:    h.out,
		level:  h.level,
		attrs:  h.attrs,
		groups: newGroups,
		mu:     h.mu,
	}
}

// SlogLevel converts golog.Level to slog.Level
func SlogLevel(l Level) slog.Level {
	switch l {
	case DebugLevel:
		return slog.LevelDebug
	case InfoLevel:
		return slog.LevelInfo
	case WarnLevel:
		return slog.LevelWarn
	case ErrorLevel, FatalLevel:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// NewSlogLogger creates a *slog.Logger with the specified handler type.
// loggerType can be:
//   - "json": JSON output for production (uses slog.JSONHandler)
//   - "text": Plain text key=value output (uses slog.TextHandler)
//   - "colored" or default: Colored output for development (uses ColoredHandler)
func NewSlogLogger(loggerType string, out io.Writer, level Level) *slog.Logger {
	slogLevel := SlogLevel(level)
	opts := &slog.HandlerOptions{
		Level:     slogLevel,
		AddSource: true,
	}

	var handler slog.Handler
	switch loggerType {
	case "json":
		handler = slog.NewJSONHandler(out, opts)
	case "text":
		handler = slog.NewTextHandler(out, opts)
	default:
		// "colored" or any other value -> development colored output
		handler = NewColoredHandler(out, opts)
	}

	return slog.New(handler)
}

// NewDefaultSlogLogger creates a colored slog logger writing to the provided writer.
// This is a convenience function for development use.
func NewDefaultSlogLogger(out io.Writer, level Level) *slog.Logger {
	return NewSlogLogger("colored", out, level)
}

// SetDefaultSlogLogger sets the provided logger as the default slog logger.
// After calling this, slog.Info(), slog.Debug(), etc. will use this logger.
func SetDefaultSlogLogger(logger *slog.Logger) {
	slog.SetDefault(logger)
}

// Example usage helper - demonstrates how to use the slog logger
func ExampleUsage() {
	// Create colored logger for development
	logger := NewSlogLogger("colored", nil, DebugLevel) // nil defaults to os.Stderr

	// Basic logging with structured key-value pairs
	logger.Info("server started", "port", 8080, "env", "development")
	logger.Debug("processing request", "method", "GET", "path", "/api/users")
	logger.Warn("slow query detected", "duration_ms", 1500, "query", "SELECT *")
	logger.Error("connection failed", "host", "db.example.com", "error", "timeout")

	// Create a child logger with preset attributes
	reqLogger := logger.With("request_id", "abc-123", "user_id", 42)
	reqLogger.Info("handling request") // includes request_id and user_id automatically

	// For production, switch to JSON:
	// logger := NewSlogLogger("json", os.Stdout, InfoLevel)
}
