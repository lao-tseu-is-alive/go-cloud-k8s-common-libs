package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

var (
	ErrTlsModeInvalid = errors.New("TLS_MODE must be one of 'none', 'manual', or 'autocert'")
)

// GetTlsMode returns the configured TLS mode from environment variable TLS_MODE
// Valid values: "none", "manual", "autocert". Defaults to "none" if not set.
func GetTlsMode() (string, error) {
	mode := os.Getenv("TLS_MODE")
	switch mode {
	case "", "none":
		return "none", nil
	case "manual":
		return "manual", nil
	case "autocert":
		return "autocert", nil
	default:
		return "", fmt.Errorf("%w: got '%s'", ErrTlsModeInvalid, mode)
	}
}

// GetTlsCertFile returns the path to the TLS certificate file from TLS_CERT_FILE
func GetTlsCertFile() string {
	return os.Getenv("TLS_CERT_FILE")
}

// GetTlsKeyFile returns the path to the TLS key file from TLS_KEY_FILE
func GetTlsKeyFile() string {
	return os.Getenv("TLS_KEY_FILE")
}

// GetAutocertHosts returns the list of hosts for autocert from AUTOCERT_HOSTS
func GetAutocertHosts() []string {
	hosts := os.Getenv("AUTOCERT_HOSTS")
	if hosts == "" {
		return nil
	}
	return strings.Split(hosts, ",")
}

// GetAutocertDir returns the directory to cache certs from AUTOCERT_DIR
// Defaults to "/certs" if not set
func GetAutocertDir() string {
	dir := os.Getenv("AUTOCERT_DIR")
	if dir == "" {
		return "/certs"
	}
	return dir
}
