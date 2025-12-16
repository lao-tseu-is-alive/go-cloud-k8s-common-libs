package config

import (
	"os"
	"testing"
)

func TestGetPort(t *testing.T) {
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
		name        string
		envValue    string
		defaultPort int
		expected    int
		wantErr     bool
	}{
		{"Default port", "", 8080, 8080, false},
		{"Valid port from env", "9000", 8080, 9000, false},
		{"Invalid port (non-integer)", "abc", 8080, 0, true},
		{"Invalid port (too low)", "0", 8080, 0, true},
		{"Invalid port (too high)", "65536", 8080, 0, true},
		{"Valid port (min)", "1", 8080, 1, false},
		{"Valid port (max)", "65535", 8080, 65535, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				setEnv("PORT", tt.envValue)
			} else {
				os.Unsetenv("PORT")
			}

			result, err := GetPort(tt.defaultPort)

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

func TestGetListenIp(t *testing.T) {
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
		name         string
		envValue     string
		defaultSrvIp string
		expected     string
		wantErr      bool
	}{
		{"Default IP", "", "0.0.0.0", "0.0.0.0", false},
		{"Valid IP from env", "127.0.0.1", "0.0.0.0", "127.0.0.1", false},
		{"Invalid IP", "invalid", "0.0.0.0", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				setEnv("SRV_IP", tt.envValue)
			} else {
				os.Unsetenv("SRV_IP")
			}

			result, err := GetListenIp(tt.defaultSrvIp)

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

func TestGetAllowedIps(t *testing.T) {
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
		defaultAllowedIps []string
		expectedLen       int
		wantErr           bool
	}{
		{"Default IPs", "", []string{"127.0.0.1"}, 1, false},
		{"Valid IPs from env", "192.168.1.1,10.0.0.1", []string{"127.0.0.1"}, 2, false},
		{"Invalid IP in list", "192.168.1.1,invalid", []string{"127.0.0.1"}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				setEnv("ALLOWED_IP", tt.envValue)
			} else {
				os.Unsetenv("ALLOWED_IP")
			}

			result, err := GetAllowedIps(tt.defaultAllowedIps)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(result) != tt.expectedLen {
					t.Errorf("Expected %d IPs, but got %d", tt.expectedLen, len(result))
				}
			}
		})
	}
}

func TestGetAllowedHosts(t *testing.T) {
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
		name        string
		envValue    string
		expectedLen int
		wantErr     bool
	}{
		{"Missing env", "", 0, true},
		{"Valid hosts", "example.com,localhost", 2, false},
		{"Single host", "example.com", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				setEnv("ALLOWED_HOSTS", tt.envValue)
			} else {
				os.Unsetenv("ALLOWED_HOSTS")
			}

			result, err := GetAllowedHosts()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(result) != tt.expectedLen {
					t.Errorf("Expected %d hosts, but got %d", tt.expectedLen, len(result))
				}
			}
		})
	}
}
