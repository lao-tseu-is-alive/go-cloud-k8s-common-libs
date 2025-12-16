package config

import (
	"os"
	"testing"
)

func TestGetJwtDuration(t *testing.T) {
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
		name            string
		envValue        string
		defaultDuration int
		expected        int
		wantErr         bool
	}{
		{"Default duration", "", 60, 60, false},
		{"Valid duration from env", "120", 60, 120, false},
		{"Invalid duration (non-integer)", "abc", 60, 0, true},
		{"Invalid duration (too low)", "0", 60, 0, true},
		{"Invalid duration (too high)", "14401", 60, 0, true},
		{"Valid duration (min)", "1", 60, 1, false},
		{"Valid duration (max)", "1440", 60, 1440, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				setEnv("JWT_DURATION_MINUTES", tt.envValue)
			} else {
				os.Unsetenv("JWT_DURATION_MINUTES")
			}

			result, err := GetJwtDuration(tt.defaultDuration)

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

func TestGetJwtSecret(t *testing.T) {
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
		{"Valid secret", "validSecretLongEnough", "validSecretLongEnough", false},
		{"Missing env variable", "", "", true},
		{"Secret too short", "short", "", true},
		{"Secret exactly minimum length", "a2b4c6t8a2b4c6t8", "a2b4c6t8a2b4c6t8", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				setEnv("JWT_SECRET", tt.envValue)
			} else {
				os.Unsetenv("JWT_SECRET")
			}

			result, err := GetJwtSecret()

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

func TestGetJwtIssuer(t *testing.T) {
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
		{"Valid issuer", "validSecretLongEnough", "validSecretLongEnough", false},
		{"Missing env variable", "", "", true},
		{"Issuer too short", "short", "", true},
		{"Issuer exactly minimum length", "a2b4c6t8a2b4c6t8", "a2b4c6t8a2b4c6t8", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				setEnv("JWT_ISSUER_ID", tt.envValue)
			} else {
				os.Unsetenv("JWT_ISSUER_ID")
			}

			result, err := GetJwtIssuer()

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

func TestGetJwtContextKey(t *testing.T) {
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
		{"Valid context key", "validContextKey", "validContextKey", false},
		{"Missing env variable", "", "", true},
		{"Context key too short", "short", "", true},
		{"Context key exactly minimum length", "abcDef", "abcDef", false},
		{"Context key with numbers", "a2b4c6", "", true},
		{"Special characters", "!@#$'%^&*()", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				setEnv("JWT_CONTEXT_KEY", tt.envValue)
			} else {
				os.Unsetenv("JWT_CONTEXT_KEY")
			}

			result, err := GetJwtContextKey()

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
