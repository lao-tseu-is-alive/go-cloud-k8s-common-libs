package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
)

var (
	ErrLogFileTooShort   = errors.New("LOG_FILE name is too short (minimum 5 characters)")
	ErrLogFileOpenFailed = errors.New("failed to open LOG_FILE")
	ErrLogLevelInvalid   = errors.New("invalid LOG_LEVEL (accepted: debug, info, warn, error, fatal or 0-4)")
)

// GetLogWriter returns an io.Writer for logging based on environment variable LOG_FILE
// Uses defaultLogName if env var is not set.
// Special values: "stdout", "stderr", "DISCARD"
// Returns error if file cannot be opened
func GetLogWriter(defaultLogName string) (io.Writer, error) {
	logFileName := defaultLogName
	val, exist := os.LookupEnv("LOG_FILE")
	if exist {
		logFileName = val
	}
	if utf8.RuneCountInString(logFileName) < 5 {
		return nil, fmt.Errorf("%w: got %d characters", ErrLogFileTooShort, utf8.RuneCountInString(logFileName))
	}
	switch logFileName {
	case "stdout":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	case "DISCARD":
		return io.Discard, nil
	default:
		file, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("%w: %s: %v", ErrLogFileOpenFailed, logFileName, err)
		}
		return file, nil
	}
}

// GetLogLevel returns the log level from environment variable LOG_LEVEL
// Uses defaultLevel if env var is not set.
// Accepted values (case-insensitive): debug, info, warn, error, fatal, or 0-4
func GetLogLevel(defaultLevel golog.Level) (golog.Level, error) {
	val, ok := os.LookupEnv("LOG_LEVEL")
	if !ok || strings.TrimSpace(val) == "" {
		return defaultLevel, nil
	}

	v := strings.TrimSpace(strings.ToLower(val))
	switch v {
	case "debug", "0":
		return golog.DebugLevel, nil
	case "info", "1":
		return golog.InfoLevel, nil
	case "warn", "warning", "2":
		return golog.WarnLevel, nil
	case "error", "3":
		return golog.ErrorLevel, nil
	case "fatal", "4":
		return golog.FatalLevel, nil
	default:
		return 0, fmt.Errorf("%w: %q", ErrLogLevelInvalid, val)
	}
}
