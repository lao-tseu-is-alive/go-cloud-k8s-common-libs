package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"unicode/utf8"
)

const (
	minSecretLength     = 16
	minContextKeyLength = 6
)

var (
	ErrJwtSecretMissing      = errors.New("ENV JWT_SECRET is required")
	ErrJwtSecretTooShort     = errors.New("ENV JWT_SECRET is too short")
	ErrJwtIssuerMissing      = errors.New("ENV JWT_ISSUER_ID is required")
	ErrJwtIssuerTooShort     = errors.New("ENV JWT_ISSUER_ID is too short")
	ErrJwtContextKeyMissing  = errors.New("ENV JWT_CONTEXT_KEY is required")
	ErrJwtContextKeyTooShort = errors.New("ENV JWT_CONTEXT_KEY is too short")
	ErrJwtContextKeyInvalid  = errors.New("ENV JWT_CONTEXT_KEY must contain only letters (a-z, A-Z)")
	ErrJwtAuthUrlMissing     = errors.New("ENV JWT_AUTH_URL is required")
	ErrJwtAuthUrlInvalid     = errors.New("ENV JWT_AUTH_URL must be a valid URL")
	ErrJwtDurationInvalid    = errors.New("ENV JWT_DURATION_MINUTES must be between 1 and 1440")
)

// GetJwtSecret returns the JWT secret from environment variable JWT_SECRET
// Returns error if not set or too short (minimum 16 characters)
func GetJwtSecret() (string, error) {
	val, exist := os.LookupEnv("JWT_SECRET")
	if !exist {
		return "", ErrJwtSecretMissing
	}
	if utf8.RuneCountInString(val) < minSecretLength {
		return "", fmt.Errorf("%w: minimum %d characters, got %d", ErrJwtSecretTooShort, minSecretLength, utf8.RuneCountInString(val))
	}
	return val, nil
}

// GetJwtIssuer returns the JWT issuer from environment variable JWT_ISSUER_ID
// Returns error if not set or too short (minimum 16 characters)
func GetJwtIssuer() (string, error) {
	val, exist := os.LookupEnv("JWT_ISSUER_ID")
	if !exist {
		return "", ErrJwtIssuerMissing
	}
	if utf8.RuneCountInString(val) < minSecretLength {
		return "", fmt.Errorf("%w: minimum %d characters, got %d", ErrJwtIssuerTooShort, minSecretLength, utf8.RuneCountInString(val))
	}
	return val, nil
}

// GetJwtContextKey returns the JWT context key from environment variable JWT_CONTEXT_KEY
// Returns error if not set, too short (minimum 6 characters), or contains non-letter characters
func GetJwtContextKey() (string, error) {
	val, exist := os.LookupEnv("JWT_CONTEXT_KEY")
	if !exist {
		return "", ErrJwtContextKeyMissing
	}
	if utf8.RuneCountInString(val) < minContextKeyLength {
		return "", fmt.Errorf("%w: minimum %d characters, got %d", ErrJwtContextKeyTooShort, minContextKeyLength, utf8.RuneCountInString(val))
	}
	match, _ := regexp.MatchString("^[a-zA-Z]+$", val)
	if !match {
		return "", ErrJwtContextKeyInvalid
	}
	return val, nil
}

// GetJwtAuthUrl returns the JWT authentication URL from environment variable JWT_AUTH_URL
// Returns error if not set or not a valid URL
func GetJwtAuthUrl() (string, error) {
	val, exist := os.LookupEnv("JWT_AUTH_URL")
	if !exist {
		return "", ErrJwtAuthUrlMissing
	}
	match, _ := regexp.MatchString("^(?:(?:https?|ftp):\\/\\/(?:[^@]+@)?[^:\\/?#]+(?::\\d+)?(?:\\/[^?#]*)?)|\\/[^?#]*$", val)
	if !match {
		return "", ErrJwtAuthUrlInvalid
	}
	return val, nil
}

// GetJwtDuration returns the JWT duration in minutes from environment variable JWT_DURATION_MINUTES
// Uses defaultDuration if env var is not set. Returns error if value is invalid (must be 1-1440)
func GetJwtDuration(defaultDuration int) (int, error) {
	val, exist := os.LookupEnv("JWT_DURATION_MINUTES")
	if !exist {
		if defaultDuration < 1 || defaultDuration > 1440 {
			return 0, ErrJwtDurationInvalid
		}
		return defaultDuration, nil
	}
	duration, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("ENV JWT_DURATION_MINUTES must be a valid integer: %w", err)
	}
	if duration < 1 || duration > 1440 {
		return 0, ErrJwtDurationInvalid
	}
	return duration, nil
}

// GetJwtCookieName returns the JWT cookie name from environment variable JWT_COOKIE_NAME
// Uses defaultName if env var is not set
func GetJwtCookieName(defaultName string) string {
	val, exist := os.LookupEnv("JWT_COOKIE_NAME")
	if !exist {
		return defaultName
	}
	return val
}

// GetJwtStatusUrl returns the JWT status URL from environment variable JWT_STATUS_URL
// Uses defaultName if env var is not set
func GetJwtStatusUrl(defaultName string) string {
	val, exist := os.LookupEnv("JWT_STATUS_URL")
	if !exist {
		return defaultName
	}
	return val
}
