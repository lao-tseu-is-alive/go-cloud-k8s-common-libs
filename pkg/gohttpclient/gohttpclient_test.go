package gohttpclient

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestWaitForHttpServer_Success(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	// This should succeed without panicking
	WaitForHttpServer(server.URL, 100*time.Millisecond, 3, logger)
}

func TestWaitForHttpServer_Panic(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Use an invalid URL that will never respond
	invalidURL := "http://localhost:59999"

	defer func() {
		r := recover()
		if r == nil {
			t.Error("WaitForHttpServer should panic when server is not available")
		}
	}()

	// This should panic after retries are exhausted
	WaitForHttpServer(invalidURL, 10*time.Millisecond, 2, logger)
}

func TestGetJsonFromUrlWithBearerAuth_InvalidURL(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Test with invalid URL
	_, err := GetJsonFromUrlWithBearerAuth(
		"not-a-valid-url",
		"test-token",
		[]byte{},
		true,
		5*time.Second,
		logger,
	)

	if err == nil {
		t.Error("GetJsonFromUrlWithBearerAuth should return error for invalid URL")
	}
}

func TestGetJsonFromUrlWithBearerAuth_Success(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	expectedResponse := `{"status": "ok", "data": "test"}`
	var receivedAuthHeader string

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuthHeader = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedResponse))
	}))
	defer server.Close()

	result, err := GetJsonFromUrlWithBearerAuth(
		server.URL,
		"my-test-token",
		[]byte{},
		true, // allow insecure for test server
		5*time.Second,
		logger,
	)

	if err != nil {
		t.Fatalf("GetJsonFromUrlWithBearerAuth() unexpected error: %v", err)
	}

	if result != expectedResponse {
		t.Errorf("GetJsonFromUrlWithBearerAuth() = %q, want %q", result, expectedResponse)
	}

	// Verify Bearer token was sent
	expectedAuth := "Bearer my-test-token"
	if receivedAuthHeader != expectedAuth {
		t.Errorf("Authorization header = %q, want %q", receivedAuthHeader, expectedAuth)
	}
}

func TestGetJsonFromUrlWithBearerAuth_NonOKStatus(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create a test server that returns 401
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
	}))
	defer server.Close()

	_, err := GetJsonFromUrlWithBearerAuth(
		server.URL,
		"invalid-token",
		[]byte{},
		true,
		5*time.Second,
		logger,
	)

	if err == nil {
		t.Error("GetJsonFromUrlWithBearerAuth should return error for non-OK status")
	}
}

func TestGetJsonFromUrlWithBearerAuth_ServerNotFound(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	_, err := GetJsonFromUrlWithBearerAuth(
		"http://localhost:59998/nonexistent",
		"test-token",
		[]byte{},
		true,
		1*time.Second,
		logger,
	)

	if err == nil {
		t.Error("GetJsonFromUrlWithBearerAuth should return error when server is not found")
	}
}
