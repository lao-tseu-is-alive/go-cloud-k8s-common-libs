package golog_test

import (
	"os"

	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
)

func Example_coloredOutput() {
	// Development: colored output with emojis
	logger := golog.NewSlogLogger("colored", os.Stdout, golog.DebugLevel)

	logger.Debug("processing request", "method", "GET", "path", "/api/users")
	logger.Info("server started", "port", 8080, "env", "development")
	logger.Warn("slow query detected", "duration_ms", 1500)
	logger.Error("connection failed", "host", "db.example.com")

	// Child logger with preset attributes
	reqLogger := logger.With("request_id", "abc-123")
	reqLogger.Info("handling authenticated request", "user_id", 42)
}

func Example_jsonOutput() {
	// Production: JSON output for log aggregation
	logger := golog.NewSlogLogger("json", os.Stdout, golog.InfoLevel)

	logger.Info("server started", "port", 8080, "env", "production")
	logger.Error("database connection failed", "host", "db.prod.internal", "retry", 3)
}
