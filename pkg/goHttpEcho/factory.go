package goHttpEcho

import (
	"log/slog"

	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/config"
)

// JwtConfig holds all JWT-related configuration
type JwtConfig struct {
	Secret       string
	Issuer       string
	Audience     string
	ContextKey   string
	DurationMins int
}

// GetJwtConfig retrieves all JWT configuration from environment variables.
// The audience parameter is provided by the caller (typically the app name).
// Returns error if any required environment variable is missing or invalid.
func GetJwtConfig(audience string, defaultDurationMins int) (*JwtConfig, error) {
	secret, err := config.GetJwtSecret()
	if err != nil {
		return nil, err
	}
	issuer, err := config.GetJwtIssuer()
	if err != nil {
		return nil, err
	}
	contextKey, err := config.GetJwtContextKey()
	if err != nil {
		return nil, err
	}
	duration, err := config.GetJwtDuration(defaultDurationMins)
	if err != nil {
		return nil, err
	}
	return &JwtConfig{
		Secret:       secret,
		Issuer:       issuer,
		Audience:     audience,
		ContextKey:   contextKey,
		DurationMins: duration,
	}, nil
}

// GetNewJwtCheckerFromConfig creates a JwtChecker from environment configuration.
// The audience parameter is the application name (used as the JWT subject).
// Returns error if any required JWT environment variable is missing or invalid.
func GetNewJwtCheckerFromConfig(audience string, defaultDurationMins int, logger *slog.Logger) (JwtChecker, error) {
	cfg, err := GetJwtConfig(audience, defaultDurationMins)
	if err != nil {
		return nil, err
	}
	return NewJwtChecker(
		cfg.Secret,
		cfg.Issuer,
		cfg.Audience,
		cfg.ContextKey,
		cfg.DurationMins,
		logger,
	), nil
}

// AdminDefaults provides default values for admin configuration
type AdminDefaults struct {
	UserId     int
	ExternalId int
	Login      string
	Email      string
}

// AdminConfig holds admin user configuration
type AdminConfig struct {
	UserId     int
	ExternalId int
	Login      string
	Email      string
	Password   string
}

// GetAdminConfig retrieves all admin configuration from environment variables.
// Uses defaults where environment variables are not set.
// Returns error if required values are missing or validation fails.
func GetAdminConfig(defaults AdminDefaults) (*AdminConfig, error) {
	login, err := config.GetAdminUser(defaults.Login)
	if err != nil {
		return nil, err
	}
	email, err := config.GetAdminEmail(defaults.Email)
	if err != nil {
		return nil, err
	}
	userId, err := config.GetAdminId(defaults.UserId)
	if err != nil {
		return nil, err
	}
	externalId, err := config.GetAdminExternalId(defaults.ExternalId)
	if err != nil {
		return nil, err
	}
	password, err := config.GetAdminPassword()
	if err != nil {
		return nil, err
	}
	return &AdminConfig{
		UserId:     userId,
		ExternalId: externalId,
		Login:      login,
		Email:      email,
		Password:   password,
	}, nil
}

// GetSimpleAdminAuthenticatorFromConfig creates an Authenticator from environment configuration.
// Returns error if required admin environment variables are missing or invalid.
func GetSimpleAdminAuthenticatorFromConfig(defaults AdminDefaults, jwtChecker JwtChecker) (Authentication, error) {
	cfg, err := GetAdminConfig(defaults)
	if err != nil {
		return nil, err
	}
	return NewSimpleAdminAuthenticator(
		&UserInfo{
			UserId:     cfg.UserId,
			ExternalId: cfg.ExternalId,
			Name:       "Admin",
			Email:      cfg.Email,
			Login:      cfg.Login,
			IsAdmin:    true,
			Groups:     []int{1}, // global_admin group
		},
		cfg.Password,
		jwtChecker,
	), nil
}
