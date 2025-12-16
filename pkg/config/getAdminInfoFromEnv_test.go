package config

import (
	"os"
	"testing"
)

func TestGetAdminUser(t *testing.T) {
	setEnv := func(key, value string) {
		oldValue, exists := os.LookupEnv(key)
		os.Setenv(key, value)
		t.Cleanup(func() {
			if exists {
				os.Setenv(key, oldValue)
			} else {
				os.Unsetenv(key)
			}
		})
	}

	tests := []struct {
		name             string
		envValue         string
		defaultAdminUser string
		expected         string
		wantErr          bool
	}{
		{"Default admin user", "", "goadmin", "goadmin", false},
		{"Valid admin user from env", "newadmin", "goadmin", "newadmin", false},
		{"Admin user too short", "ab", "", "", true},
		{"Admin user exactly min length", "admin", "goadmin", "admin", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				setEnv("ADMIN_USER", tt.envValue)
			} else {
				os.Unsetenv("ADMIN_USER")
			}

			result, err := GetAdminUser(tt.defaultAdminUser)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %s, but got %s", tt.expected, result)
				}
			}
		})
	}
}

func TestGetAdminEmail(t *testing.T) {
	setEnv := func(key, value string) {
		oldValue, exists := os.LookupEnv(key)
		os.Setenv(key, value)
		t.Cleanup(func() {
			if exists {
				os.Setenv(key, oldValue)
			} else {
				os.Unsetenv(key)
			}
		})
	}

	tests := []struct {
		name              string
		envValue          string
		defaultAdminEmail string
		expected          string
		wantErr           bool
	}{
		{"Default admin email", "", "admin@example.com", "admin@example.com", false},
		{"Valid admin email from env", "new@example.com", "admin@example.com", "new@example.com", false},
		{"Invalid email format", "invalidemail", "", "", true},
		{"Email too short", "a@b.c", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				setEnv("ADMIN_EMAIL", tt.envValue)
			} else {
				os.Unsetenv("ADMIN_EMAIL")
			}

			result, err := GetAdminEmail(tt.defaultAdminEmail)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %s, but got %s", tt.expected, result)
				}
			}
		})
	}
}

func TestGetAdminId(t *testing.T) {
	setEnv := func(key, value string) {
		oldValue, exists := os.LookupEnv(key)
		os.Setenv(key, value)
		t.Cleanup(func() {
			if exists {
				os.Setenv(key, oldValue)
			} else {
				os.Unsetenv(key)
			}
		})
	}

	tests := []struct {
		name           string
		envValue       string
		defaultAdminId int
		expected       int
		wantErr        bool
	}{
		{"Default admin id", "", 1, 1, false},
		{"Valid admin id from env", "42", 1, 42, false},
		{"Invalid admin id (non-integer)", "abc", 1, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				setEnv("ADMIN_ID", tt.envValue)
			} else {
				os.Unsetenv("ADMIN_ID")
			}

			result, err := GetAdminId(tt.defaultAdminId)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %d, but got %d", tt.expected, result)
				}
			}
		})
	}
}

func TestGetAdminPassword(t *testing.T) {
	setEnv := func(key, value string) {
		oldValue, exists := os.LookupEnv(key)
		os.Setenv(key, value)
		t.Cleanup(func() {
			if exists {
				os.Setenv(key, oldValue)
			} else {
				os.Unsetenv(key)
			}
		})
	}

	tests := []struct {
		name     string
		envValue string
		expected string
		wantErr  bool
	}{
		{"Missing password", "", "", true},
		{"Password too short", "Ab1!", "", true},
		{"Valid password", "SecureP@ss1", "SecureP@ss1", false},
		{"Password without complexity", "simplepassword", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				setEnv("ADMIN_PASSWORD", tt.envValue)
			} else {
				os.Unsetenv("ADMIN_PASSWORD")
			}

			result, err := GetAdminPassword()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %s, but got %s", tt.expected, result)
				}
			}
		})
	}
}
