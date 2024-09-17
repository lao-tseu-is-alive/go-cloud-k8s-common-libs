package golog

import (
	"log"
	"net/http"
)

type MyLogger interface {
	Trace(msg string, v ...interface{})
	Debug(msg string, v ...interface{})
	Info(msg string, v ...interface{})
	Warn(msg string, v ...interface{})
	Error(msg string, v ...interface{})
	Fatal(msg string, v ...interface{})
	GetDefaultLogger() (*log.Logger, error)
	TraceHttpRequest(handlerName string, r *http.Request)
}

type Level int8

const (
	// TraceLevel defines trace log level.
	TraceLevel Level = iota
	// DebugLevel defines debug log level.
	DebugLevel
	// InfoLevel defines info log level.
	InfoLevel
	// WarnLevel defines warn log level.
	WarnLevel
	// ErrorLevel defines error log level.
	ErrorLevel
	// FatalLevel defines fatal log level.
	FatalLevel
)

func NewLogger(loggerType string, logLevel Level, prefix string) (MyLogger, error) {
	var (
		logger MyLogger
		err    error
	)

	switch loggerType {
	case "production":
		// here we can handle structured  json log
	default:
		logger, err = NewSimpleLogger(logLevel, prefix)
		if err != nil {
			return nil, err
		}
	}

	return logger, nil
}
