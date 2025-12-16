package info

import (
	"errors"
	"os"
	"testing"
)

func TestErrorConfig_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		msg      string
		expected string
	}{
		{
			name:     "with error and message",
			err:      errors.New("underlying error"),
			msg:      "Config error",
			expected: "Config error : underlying error",
		},
		{
			name:     "with nil error",
			err:      nil,
			msg:      "No error",
			expected: "No error : <nil>",
		},
		{
			name:     "with empty message",
			err:      errors.New("just error"),
			msg:      "",
			expected: " : just error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ErrorConfig{
				Err: tt.err,
				Msg: tt.msg,
			}
			result := e.Error()
			if result != tt.expected {
				t.Errorf("ErrorConfig.Error() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetKubernetesApiUrlFromEnv(t *testing.T) {
	// Save original env vars to restore later
	origHost := os.Getenv("KUBERNETES_SERVICE_HOST")
	origPort := os.Getenv("KUBERNETES_SERVICE_PORT")
	defer func() {
		if origHost != "" {
			os.Setenv("KUBERNETES_SERVICE_HOST", origHost)
		} else {
			os.Unsetenv("KUBERNETES_SERVICE_HOST")
		}
		if origPort != "" {
			os.Setenv("KUBERNETES_SERVICE_PORT", origPort)
		} else {
			os.Unsetenv("KUBERNETES_SERVICE_PORT")
		}
	}()

	tests := []struct {
		name        string
		host        string
		port        string
		setHost     bool
		setPort     bool
		wantURL     string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid host and port",
			host:    "10.0.0.1",
			port:    "443",
			setHost: true,
			setPort: true,
			wantURL: "https://10.0.0.1:443",
			wantErr: false,
		},
		{
			name:    "valid host without port (default 443)",
			host:    "10.0.0.1",
			setHost: true,
			setPort: false,
			wantURL: "https://10.0.0.1:443",
			wantErr: false,
		},
		{
			name:    "valid host with custom port",
			host:    "kubernetes.default.svc",
			port:    "6443",
			setHost: true,
			setPort: true,
			wantURL: "https://kubernetes.default.svc:6443",
			wantErr: false,
		},
		{
			name:        "missing host env var",
			setHost:     false,
			setPort:     false,
			wantErr:     true,
			errContains: "KUBERNETES_SERVICE_HOST",
		},
		{
			name:        "invalid port - not a number",
			host:        "10.0.0.1",
			port:        "notaport",
			setHost:     true,
			setPort:     true,
			wantErr:     true,
			errContains: "valid integer",
		},
		{
			name:        "invalid port - zero",
			host:        "10.0.0.1",
			port:        "0",
			setHost:     true,
			setPort:     true,
			wantErr:     true,
			errContains: "between 1 and 65535",
		},
		{
			name:        "invalid port - too high",
			host:        "10.0.0.1",
			port:        "70000",
			setHost:     true,
			setPort:     true,
			wantErr:     true,
			errContains: "between 1 and 65535",
		},
		{
			name:        "invalid port - negative",
			host:        "10.0.0.1",
			port:        "-1",
			setHost:     true,
			setPort:     true,
			wantErr:     true,
			errContains: "between 1 and 65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env vars first
			os.Unsetenv("KUBERNETES_SERVICE_HOST")
			os.Unsetenv("KUBERNETES_SERVICE_PORT")

			// Set env vars based on test case
			if tt.setHost {
				os.Setenv("KUBERNETES_SERVICE_HOST", tt.host)
			}
			if tt.setPort {
				os.Setenv("KUBERNETES_SERVICE_PORT", tt.port)
			}

			gotURL, err := GetKubernetesApiUrlFromEnv()

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetKubernetesApiUrlFromEnv() expected error, got nil")
					return
				}
				if tt.errContains != "" {
					errStr := err.Error()
					if !contains(errStr, tt.errContains) {
						t.Errorf("GetKubernetesApiUrlFromEnv() error = %q, should contain %q", errStr, tt.errContains)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("GetKubernetesApiUrlFromEnv() unexpected error: %v", err)
				return
			}

			if gotURL != tt.wantURL {
				t.Errorf("GetKubernetesApiUrlFromEnv() = %q, want %q", gotURL, tt.wantURL)
			}
		})
	}
}

// contains is a helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
