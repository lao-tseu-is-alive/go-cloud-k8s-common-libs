package config

import (
	"errors"
	"fmt"
	"net/mail"
	"os"
	"strconv"
	"unicode"
	"unicode/utf8"
)

const minUserNameLength = 5
const minUserEmailLength = 12
const minUserPasswordLength = 8

var (
	ErrAdminUserTooShort       = errors.New("ADMIN_USER is too short")
	ErrAdminEmailTooShort      = errors.New("ADMIN_EMAIL is too short")
	ErrAdminEmailInvalid       = errors.New("ADMIN_EMAIL must be a valid email address")
	ErrAdminEmailSpecialChar   = errors.New("ADMIN_EMAIL contains invalid special characters")
	ErrAdminIdInvalid          = errors.New("ADMIN_ID must be a valid integer")
	ErrAdminExternalIdInvalid  = errors.New("ADMIN_EXTERNAL_ID must be a valid integer")
	ErrAdminPasswordMissing    = errors.New("ENV ADMIN_PASSWORD is required")
	ErrAdminPasswordTooShort   = errors.New("ADMIN_PASSWORD is too short")
	ErrAdminPasswordNotComplex = errors.New("ADMIN_PASSWORD must contain lowercase, uppercase, digit, and special character. No whitespace, #, |, or '")
)

// GetAdminUser returns the admin user from environment variable ADMIN_USER
// Uses defaultAdminUser if env var is not set. Returns error if too short (minimum 5 characters)
func GetAdminUser(defaultAdminUser string) (string, error) {
	adminUser := defaultAdminUser
	val, exist := os.LookupEnv("ADMIN_USER")
	if exist {
		adminUser = val
	}
	if utf8.RuneCountInString(adminUser) < minUserNameLength {
		return "", fmt.Errorf("%w: minimum %d characters, got %d", ErrAdminUserTooShort, minUserNameLength, utf8.RuneCountInString(adminUser))
	}
	return adminUser, nil
}

// GetAdminEmail returns the admin email from environment variable ADMIN_EMAIL
// Uses defaultAdminEmail if env var is not set. Returns error if invalid email format
func GetAdminEmail(defaultAdminEmail string) (string, error) {
	adminEmail := defaultAdminEmail
	val, exist := os.LookupEnv("ADMIN_EMAIL")
	if exist {
		adminEmail = val
	}
	if utf8.RuneCountInString(adminEmail) < minUserEmailLength {
		return "", fmt.Errorf("%w: minimum %d characters, got %d", ErrAdminEmailTooShort, minUserEmailLength, utf8.RuneCountInString(adminEmail))
	}
	_, err := mail.ParseAddress(adminEmail)
	if err != nil {
		return "", ErrAdminEmailInvalid
	}
	for _, c := range adminEmail {
		switch {
		case c == '@' || c == '.' || c == '_' || c == '-':
			continue
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			return "", fmt.Errorf("%w: '%c' is not allowed", ErrAdminEmailSpecialChar, c)
		}
	}
	return adminEmail, nil
}

// GetAdminId returns the admin user ID from environment variable ADMIN_ID
// Uses defaultAdminId if env var is not set. Returns error if not a valid integer
func GetAdminId(defaultAdminId int) (int, error) {
	val, exist := os.LookupEnv("ADMIN_ID")
	if !exist {
		return defaultAdminId, nil
	}
	adminId, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrAdminIdInvalid, err)
	}
	return adminId, nil
}

// GetAdminExternalId returns the admin external ID from environment variable ADMIN_EXTERNAL_ID
// Uses defaultAdminExternalId if env var is not set. Returns error if not a valid integer
func GetAdminExternalId(defaultAdminExternalId int) (int, error) {
	val, exist := os.LookupEnv("ADMIN_EXTERNAL_ID")
	if !exist {
		return defaultAdminExternalId, nil
	}
	adminId, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrAdminExternalIdInvalid, err)
	}
	return adminId, nil
}

// GetAdminPassword returns the admin password from environment variable ADMIN_PASSWORD
// Returns error if not set, too short, or doesn't meet complexity requirements
func GetAdminPassword() (string, error) {
	val, exist := os.LookupEnv("ADMIN_PASSWORD")
	if !exist {
		return "", ErrAdminPasswordMissing
	}
	if utf8.RuneCountInString(val) < minUserPasswordLength {
		return "", fmt.Errorf("%w: minimum %d characters, got %d", ErrAdminPasswordTooShort, minUserPasswordLength, utf8.RuneCountInString(val))
	}
	if !VerifyPasswordComplexity(val) {
		return "", ErrAdminPasswordNotComplex
	}
	return val, nil
}

// VerifyPasswordComplexity checks if the password meets the minimum requirements of complexity
// At least one lowercase letter, one uppercase letter, one digit and one special character
// No white space, #, or | or ' character in it
func VerifyPasswordComplexity(s string) bool {
	var hasNumber, hasUpperCase, hasLowercase, hasSpecial bool
	for _, c := range s {
		switch {
		case unicode.IsNumber(c):
			hasNumber = true
		case unicode.IsUpper(c):
			hasUpperCase = true
		case unicode.IsLower(c):
			hasLowercase = true
		case c == '#' || c == '|' || c == '\'' || unicode.IsSpace(c):
			return false
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSpecial = true
		}
	}
	return hasNumber && hasUpperCase && hasLowercase && hasSpecial
}
