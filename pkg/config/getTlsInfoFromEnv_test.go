package config

import (
	"os"
	"testing"
)

func TestGetTlsMode(t *testing.T) {
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
		{"Default (empty)", "", "none", false},
		{"Explicit none", "none", "none", false},
		{"Manual mode", "manual", "manual", false},
		{"Autocert mode", "autocert", "autocert", false},
		{"Invalid mode", "invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				setEnv("TLS_MODE", tt.envValue)
			} else {
				os.Unsetenv("TLS_MODE")
			}

			result, err := GetTlsMode()

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
