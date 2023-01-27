package config

import (
	"errors"
	"fmt"
	"os"
)

// GetAdminUserFromFromEnv returns the DB driver based on the value of environment variables :
//
//	ADMIN_USER : string containing the username to use for the administrative account
func GetAdminUserFromFromEnv(defaultAdminUser string) string {
	adminUser := defaultAdminUser
	val, exist := os.LookupEnv("ADMIN_USER")
	if exist {
		adminUser = val
	}
	return fmt.Sprintf("%s", adminUser)
}

// GetAdminPasswordFromFromEnv returns the DB driver based on the value of environment variables :
//
//	ADMIN_PASSWORD : string containing the password to use for the administrative account
func GetAdminPasswordFromFromEnv() (string, error) {
	adminPassword := ""
	val, exist :=
		os.LookupEnv("ADMIN_PASSWORD")
	if exist {
		adminPassword = val
	}
	if len(adminPassword) > 0 {
		return fmt.Sprintf("%s", adminPassword), nil
	} else {
		return "", errors.New("env variable ADMIN_PASSWORD should be set to a non-empty string")
	}
}
