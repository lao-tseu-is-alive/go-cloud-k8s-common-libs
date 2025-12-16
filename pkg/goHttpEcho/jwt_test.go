package goHttpEcho

import (
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestNewJwtChecker(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	jwtChecker := NewJwtChecker(
		"test-secret-key-for-testing-purposes-only",
		"test-issuer",
		"test-subject",
		"test-context-key",
		30,
		logger,
	)

	if jwtChecker == nil {
		t.Fatal("NewJwtChecker returned nil")
	}

	if jwtChecker.GetIssuerId() != "test-issuer" {
		t.Errorf("GetIssuerId() = %q, want %q", jwtChecker.GetIssuerId(), "test-issuer")
	}

	if jwtChecker.GetJwtDuration() != 30 {
		t.Errorf("GetJwtDuration() = %d, want %d", jwtChecker.GetJwtDuration(), 30)
	}

	if jwtChecker.GetLogger() != logger {
		t.Error("GetLogger() did not return the expected logger")
	}
}

func TestGetTokenFromUserInfo(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	jwtChecker := NewJwtChecker(
		"test-secret-key-for-testing-purposes-only-must-be-long",
		"test-issuer",
		"test-subject",
		"test-context-key",
		30,
		logger,
	)

	userInfo := &UserInfo{
		UserId:     1,
		ExternalId: 100,
		Name:       "Test User",
		Email:      "test@example.com",
		Login:      "testuser",
		IsAdmin:    true,
		Groups:     []int{1, 2, 3},
	}

	token, err := jwtChecker.GetTokenFromUserInfo(userInfo)
	if err != nil {
		t.Fatalf("GetTokenFromUserInfo() error = %v", err)
	}

	if token == nil {
		t.Fatal("GetTokenFromUserInfo() returned nil token")
	}

	// Token should have content
	tokenString := token.String()
	if len(tokenString) < 50 {
		t.Errorf("Token string seems too short: %q", tokenString)
	}
}

func TestParseToken_Valid(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	secret := "test-secret-key-for-testing-purposes-only-must-be-long"
	jwtChecker := NewJwtChecker(
		secret,
		"test-issuer",
		"test-subject",
		"test-context-key",
		30,
		logger,
	)

	userInfo := &UserInfo{
		UserId:     42,
		ExternalId: 142,
		Name:       "John Doe",
		Email:      "john@example.com",
		Login:      "johndoe",
		IsAdmin:    false,
		Groups:     []int{5, 10},
	}

	// Generate a token
	token, err := jwtChecker.GetTokenFromUserInfo(userInfo)
	if err != nil {
		t.Fatalf("GetTokenFromUserInfo() error = %v", err)
	}

	// Parse the token back
	claims, err := jwtChecker.ParseToken(token.String())
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}

	// Verify the claims
	if claims.User.UserId != 42 {
		t.Errorf("claims.User.UserId = %d, want %d", claims.User.UserId, 42)
	}
	if claims.User.Login != "johndoe" {
		t.Errorf("claims.User.Login = %q, want %q", claims.User.Login, "johndoe")
	}
	if claims.User.Email != "john@example.com" {
		t.Errorf("claims.User.Email = %q, want %q", claims.User.Email, "john@example.com")
	}
	if claims.User.IsAdmin != false {
		t.Errorf("claims.User.IsAdmin = %v, want %v", claims.User.IsAdmin, false)
	}
}

func TestParseToken_InvalidToken(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	jwtChecker := NewJwtChecker(
		"test-secret-key-for-testing-purposes-only-must-be-long",
		"test-issuer",
		"test-subject",
		"test-context-key",
		30,
		logger,
	)

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "garbage token",
			token: "not-a-valid-jwt-token",
		},
		{
			name:  "wrong format",
			token: "header.payload",
		},
		{
			name:  "invalid base64",
			token: "!!!.@@@.###",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := jwtChecker.ParseToken(tt.token)
			if err == nil {
				t.Errorf("ParseToken(%q) expected error, got nil", tt.token)
			}
		})
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create token with one secret
	jwtChecker1 := NewJwtChecker(
		"secret-key-one-for-testing-purposes-only-must-be-long",
		"test-issuer",
		"test-subject",
		"test-context-key",
		30,
		logger,
	)

	userInfo := &UserInfo{
		UserId: 1,
		Login:  "testuser",
	}

	token, err := jwtChecker1.GetTokenFromUserInfo(userInfo)
	if err != nil {
		t.Fatalf("GetTokenFromUserInfo() error = %v", err)
	}

	// Try to parse with different secret
	jwtChecker2 := NewJwtChecker(
		"secret-key-two-for-testing-purposes-only-must-be-long",
		"test-issuer",
		"test-subject",
		"test-context-key",
		30,
		logger,
	)

	_, err = jwtChecker2.ParseToken(token.String())
	if err == nil {
		t.Error("ParseToken() with wrong secret expected error, got nil")
	}
}

func TestUserInfoFields(t *testing.T) {
	userInfo := &UserInfo{
		UserId:     123,
		ExternalId: 456,
		Name:       "Test User",
		Email:      "test@example.com",
		Login:      "testlogin",
		IsAdmin:    true,
		Groups:     []int{1, 2, 3},
	}

	if userInfo.UserId != 123 {
		t.Errorf("UserId = %d, want %d", userInfo.UserId, 123)
	}
	if userInfo.ExternalId != 456 {
		t.Errorf("ExternalId = %d, want %d", userInfo.ExternalId, 456)
	}
	if userInfo.Name != "Test User" {
		t.Errorf("Name = %q, want %q", userInfo.Name, "Test User")
	}
	if userInfo.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", userInfo.Email, "test@example.com")
	}
	if userInfo.Login != "testlogin" {
		t.Errorf("Login = %q, want %q", userInfo.Login, "testlogin")
	}
	if !userInfo.IsAdmin {
		t.Error("IsAdmin = false, want true")
	}
	if len(userInfo.Groups) != 3 {
		t.Errorf("len(Groups) = %d, want %d", len(userInfo.Groups), 3)
	}
}

func TestJwtInfo_TokenRoundTrip(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	jwtChecker := NewJwtChecker(
		"test-secret-key-for-testing-purposes-only-must-be-long",
		"my-issuer-id",
		"my-subject",
		"my-context-key",
		60,
		logger,
	)

	// Test multiple users
	users := []*UserInfo{
		{UserId: 1, Login: "admin", IsAdmin: true, Groups: []int{1}},
		{UserId: 2, Login: "user", IsAdmin: false, Groups: []int{2, 3}},
		{UserId: 100, Login: "power_user", IsAdmin: false, Groups: []int{1, 2, 3, 4, 5}},
	}

	for _, user := range users {
		t.Run(user.Login, func(t *testing.T) {
			token, err := jwtChecker.GetTokenFromUserInfo(user)
			if err != nil {
				t.Fatalf("GetTokenFromUserInfo() error = %v", err)
			}

			claims, err := jwtChecker.ParseToken(token.String())
			if err != nil {
				t.Fatalf("ParseToken() error = %v", err)
			}

			if claims.User.UserId != user.UserId {
				t.Errorf("UserId mismatch: got %d, want %d", claims.User.UserId, user.UserId)
			}
			if claims.User.Login != user.Login {
				t.Errorf("Login mismatch: got %q, want %q", claims.User.Login, user.Login)
			}
			if claims.User.IsAdmin != user.IsAdmin {
				t.Errorf("IsAdmin mismatch: got %v, want %v", claims.User.IsAdmin, user.IsAdmin)
			}

			// Verify registered claims
			if !claims.IsValidAt(time.Now()) {
				t.Error("Token should be valid at current time")
			}
		})
	}
}
