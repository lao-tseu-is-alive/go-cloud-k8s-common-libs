package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
)

var (
	ErrDbPortInvalid     = errors.New("DB_PORT must be an integer between 1 and 65535")
	ErrDbHostInvalid     = errors.New("DB_HOST must be a valid IP address")
	ErrDbPasswordMissing = errors.New("ENV DB_PASSWORD is required")
)

// GetPgDbDsnUrl returns a PostgreSQL DSN connection string from environment variables
// Uses defaults for missing optional values. Returns error if required values are missing or invalid.
//
// Environment variables:
//   - DB_HOST: IP address (optional, uses defaultIP)
//   - DB_PORT: port number 1-65535 (optional, uses defaultPort)
//   - DB_NAME: database name (optional, uses defaultDbName)
//   - DB_USER: database username (optional, uses defaultDbUser)
//   - DB_PASSWORD: database password (required)
//   - DB_SSL_MODE: SSL mode (optional, uses defaultSSL)
func GetPgDbDsnUrl(defaultIP string, defaultPort int, defaultDbName string, defaultDbUser string, defaultSSL string) (string, error) {
	srvIP := defaultIP
	srvPort := defaultPort
	dbName := defaultDbName
	dbUser := defaultDbUser
	dbSslMode := defaultSSL

	// DB_PORT
	val, exist := os.LookupEnv("DB_PORT")
	if exist {
		port, err := strconv.Atoi(val)
		if err != nil {
			return "", fmt.Errorf("DB_PORT must be a valid integer: %w", err)
		}
		if port < 1 || port > 65535 {
			return "", ErrDbPortInvalid
		}
		srvPort = port
	}

	// DB_HOST
	val, exist = os.LookupEnv("DB_HOST")
	if exist {
		if net.ParseIP(val) == nil {
			return "", ErrDbHostInvalid
		}
		srvIP = val
	}

	// DB_NAME
	val, exist = os.LookupEnv("DB_NAME")
	if exist {
		dbName = val
	}

	// DB_USER
	val, exist = os.LookupEnv("DB_USER")
	if exist {
		dbUser = val
	}

	// DB_PASSWORD (required)
	dbPassword, exist := os.LookupEnv("DB_PASSWORD")
	if !exist {
		return "", ErrDbPasswordMissing
	}

	// DB_SSL_MODE
	val, exist = os.LookupEnv("DB_SSL_MODE")
	if exist {
		dbSslMode = val
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		dbUser, dbPassword, srvIP, srvPort, dbName, dbSslMode), nil
}
