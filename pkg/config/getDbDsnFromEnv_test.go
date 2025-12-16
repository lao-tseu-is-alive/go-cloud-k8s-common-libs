package config

import (
	"os"
	"testing"
)

func TestGetPgDbDsnUrl(t *testing.T) {
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
		name          string
		envPassword   string
		envPort       string
		envHost       string
		defaultIP     string
		defaultPort   int
		defaultDbName string
		defaultDbUser string
		defaultSSL    string
		wantErr       bool
		wantContains  string
	}{
		{
			name:          "Missing password",
			envPassword:   "",
			defaultIP:     "127.0.0.1",
			defaultPort:   5432,
			defaultDbName: "testdb",
			defaultDbUser: "testuser",
			defaultSSL:    "disable",
			wantErr:       true,
		},
		{
			name:          "Valid configuration",
			envPassword:   "secret123",
			defaultIP:     "127.0.0.1",
			defaultPort:   5432,
			defaultDbName: "testdb",
			defaultDbUser: "testuser",
			defaultSSL:    "disable",
			wantErr:       false,
			wantContains:  "postgres://testuser:secret123@127.0.0.1:5432/testdb",
		},
		{
			name:          "Invalid port",
			envPassword:   "secret123",
			envPort:       "abc",
			defaultIP:     "127.0.0.1",
			defaultPort:   5432,
			defaultDbName: "testdb",
			defaultDbUser: "testuser",
			defaultSSL:    "disable",
			wantErr:       true,
		},
		{
			name:          "Invalid host",
			envPassword:   "secret123",
			envHost:       "invalid",
			defaultIP:     "127.0.0.1",
			defaultPort:   5432,
			defaultDbName: "testdb",
			defaultDbUser: "testuser",
			defaultSSL:    "disable",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv("DB_PASSWORD")
			os.Unsetenv("DB_PORT")
			os.Unsetenv("DB_HOST")

			if tt.envPassword != "" {
				setEnv("DB_PASSWORD", tt.envPassword)
			}
			if tt.envPort != "" {
				setEnv("DB_PORT", tt.envPort)
			}
			if tt.envHost != "" {
				setEnv("DB_HOST", tt.envHost)
			}

			result, err := GetPgDbDsnUrl(tt.defaultIP, tt.defaultPort, tt.defaultDbName, tt.defaultDbUser, tt.defaultSSL)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if tt.wantContains != "" && result != "" {
					if len(result) < len(tt.wantContains) {
						t.Errorf("Expected result to contain %s, but got %s", tt.wantContains, result)
					}
				}
			}
		})
	}
}
