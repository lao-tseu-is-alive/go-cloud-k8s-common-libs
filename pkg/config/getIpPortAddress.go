package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

var (
	ErrPortInvalid         = errors.New("PORT must be an integer between 1 and 65535")
	ErrListenIpInvalid     = errors.New("SRV_IP must be a valid IP address")
	ErrAllowedIpInvalid    = errors.New("ALLOWED_IP contains invalid IP address")
	ErrAllowedIpEmpty      = errors.New("ALLOWED_IP must contain at least one valid IP")
	ErrAllowedHostsMissing = errors.New("ENV ALLOWED_HOSTS is required")
	ErrAllowedHostsEmpty   = errors.New("ALLOWED_HOSTS must contain at least one valid host")
)

// GetPort returns the listening port from environment variable PORT
// Uses defaultPort if env var is not set. Returns error if not a valid port (1-65535)
func GetPort(defaultPort int) (int, error) {
	val, exist := os.LookupEnv("PORT")
	if !exist {
		if defaultPort < 1 || defaultPort > 65535 {
			return 0, ErrPortInvalid
		}
		return defaultPort, nil
	}
	port, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("PORT must be a valid integer: %w", err)
	}
	if port < 1 || port > 65535 {
		return 0, ErrPortInvalid
	}
	return port, nil
}

// GetListenIp returns the listening IP from environment variable SRV_IP
// Uses defaultSrvIp if env var is not set. Returns error if not a valid IP
func GetListenIp(defaultSrvIp string) (string, error) {
	srvIp := defaultSrvIp
	val, exist := os.LookupEnv("SRV_IP")
	if exist {
		srvIp = val
	}
	if net.ParseIP(srvIp) == nil {
		return "", ErrListenIpInvalid
	}
	return srvIp, nil
}

// GetAllowedIps returns allowed IPs from environment variable ALLOWED_IP
// Uses defaultAllowedIps if env var is not set. Returns error if any IP is invalid
func GetAllowedIps(defaultAllowedIps []string) ([]string, error) {
	allowedIps := defaultAllowedIps
	envValue, exists := os.LookupEnv("ALLOWED_IP")
	if exists {
		allowedIps = []string{}
		ips := strings.Split(envValue, ",")
		for _, ip := range ips {
			trimmedIP := strings.TrimSpace(ip)
			if trimmedIP != "" {
				allowedIps = append(allowedIps, trimmedIP)
			}
		}
	}
	for _, ip := range allowedIps {
		if net.ParseIP(ip) == nil {
			return nil, fmt.Errorf("%w: %s", ErrAllowedIpInvalid, ip)
		}
	}
	if len(allowedIps) == 0 {
		return nil, ErrAllowedIpEmpty
	}
	return allowedIps, nil
}

// GetAllowedHosts returns allowed hosts from environment variable ALLOWED_HOSTS
// Returns error if env var is not set or empty
func GetAllowedHosts() ([]string, error) {
	envValue, exist := os.LookupEnv("ALLOWED_HOSTS")
	if !exist {
		return nil, ErrAllowedHostsMissing
	}
	var allowedHosts []string
	allHosts := strings.Split(envValue, ",")
	for _, hostName := range allHosts {
		trimmedHost := strings.TrimSpace(hostName)
		if trimmedHost != "" {
			allowedHosts = append(allowedHosts, trimmedHost)
		}
	}
	if len(allowedHosts) == 0 {
		return nil, ErrAllowedHostsEmpty
	}
	return allowedHosts, nil
}
