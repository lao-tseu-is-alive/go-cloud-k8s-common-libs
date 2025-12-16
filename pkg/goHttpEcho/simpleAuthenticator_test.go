package goHttpEcho

import (
	"context"
	"log/slog"
	"os"
	"testing"
)

func TestNewSimpleAdminAuthenticator(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	jwtChecker := NewJwtChecker(
		"test-secret-key-for-testing-purposes-only-must-be-long",
		"test-issuer",
		"test-subject",
		"test-context-key",
		30,
		logger,
	)

	adminUser := &UserInfo{
		UserId:     1,
		ExternalId: 100,
		Name:       "Admin",
		Email:      "admin@example.com",
		Login:      "admin",
		IsAdmin:    true,
		Groups:     []int{1},
	}

	auth := NewSimpleAdminAuthenticator(adminUser, "secret-password", jwtChecker)

	if auth == nil {
		t.Fatal("NewSimpleAdminAuthenticator returned nil")
	}
}

func TestSimpleAdminAuthenticator_AuthenticateUser(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	jwtChecker := NewJwtChecker(
		"test-secret-key-for-testing-purposes-only-must-be-long",
		"test-issuer",
		"test-subject",
		"test-context-key",
		30,
		logger,
	)

	adminUser := &UserInfo{
		UserId:     1,
		ExternalId: 100,
		Name:       "Admin",
		Email:      "admin@example.com",
		Login:      "admin",
		IsAdmin:    true,
		Groups:     []int{1},
	}

	// Password will be hashed with SHA-256 internally
	password := "test-password"
	auth := NewSimpleAdminAuthenticator(adminUser, password, jwtChecker)

	tests := []struct {
		name         string
		login        string
		passwordHash string
		want         bool
	}{
		{
			name:         "valid admin credentials with correct hash",
			login:        "admin",
			passwordHash: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08", // SHA-256 of "test"
			want:         false,                                                              // Hash doesn't match
		},
		{
			name:         "wrong login",
			login:        "wronguser",
			passwordHash: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
			want:         false,
		},
		{
			name:         "empty login",
			login:        "",
			passwordHash: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
			want:         false,
		},
		{
			name:         "empty password hash",
			login:        "admin",
			passwordHash: "",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auth.AuthenticateUser(context.Background(), tt.login, tt.passwordHash)
			if result != tt.want {
				t.Errorf("AuthenticateUser(%q, %q) = %v, want %v", tt.login, tt.passwordHash, result, tt.want)
			}
		})
	}
}

func TestSimpleAdminAuthenticator_GetUserInfoFromLogin(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	jwtChecker := NewJwtChecker(
		"test-secret-key-for-testing-purposes-only-must-be-long",
		"test-issuer",
		"test-subject",
		"test-context-key",
		30,
		logger,
	)

	adminUser := &UserInfo{
		UserId:     1,
		ExternalId: 100,
		Name:       "Admin User",
		Email:      "admin@example.com",
		Login:      "admin",
		IsAdmin:    true,
		Groups:     []int{1},
	}

	auth := NewSimpleAdminAuthenticator(adminUser, "test-password", jwtChecker)

	t.Run("get admin user info", func(t *testing.T) {
		userInfo, err := auth.GetUserInfoFromLogin(context.Background(), "admin")
		if err != nil {
			t.Fatalf("GetUserInfoFromLogin() unexpected error: %v", err)
		}

		if userInfo.Login != "admin" {
			t.Errorf("Login = %q, want %q", userInfo.Login, "admin")
		}
		if userInfo.UserId != 1 {
			t.Errorf("UserId = %d, want %d", userInfo.UserId, 1)
		}
		if userInfo.Email != "admin@example.com" {
			t.Errorf("Email = %q, want %q", userInfo.Email, "admin@example.com")
		}
		if !userInfo.IsAdmin {
			t.Error("IsAdmin = false, want true")
		}
	})

	// SimpleAdminAuthenticator always returns admin user info for any login
	// (it's a simple authenticator that only knows about one admin user)
	t.Run("any login returns admin info", func(t *testing.T) {
		userInfo, err := auth.GetUserInfoFromLogin(context.Background(), "anyuser")
		if err != nil {
			t.Fatalf("GetUserInfoFromLogin() unexpected error: %v", err)
		}
		// The login field is set to the requested login, not the admin login
		if userInfo.Login != "anyuser" {
			t.Errorf("Login = %q, want %q", userInfo.Login, "anyuser")
		}
		// But all other fields are from the admin user
		if userInfo.UserId != 1 {
			t.Errorf("UserId = %d, want %d", userInfo.UserId, 1)
		}
	})
}
