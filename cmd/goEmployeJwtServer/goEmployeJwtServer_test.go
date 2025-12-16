package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/config"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/gohttpclient"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	"github.com/stretchr/testify/assert"
)

const (
	DEBUG                           = true
	assertCorrectStatusCodeExpected = "expected status code should be returned"
)

type testStruct struct {
	name           string
	contentType    string
	wantStatusCode int
	wantBody       string
	paramKeyValues map[string]string
	httpMethod     string
	url            string
	body           string
	headers        map[string]string
}

// isPortAvailable checks if a port is available for binding
func isPortAvailable(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// TestMainExec is instantiating the "real" main code using the env variable
// This test requires a running PostgreSQL database with proper env vars
func TestMainExec(t *testing.T) {
	l := golog.NewLogger("simple", os.Stdout, golog.DebugLevel, "TestGoEmployeJwtServer")
	listenPort, err := config.GetPort(defaultPort)
	if err != nil {
		t.Fatalf("error getting port: %v", err)
	}

	// Skip if port is already in use (another integration test is running)
	if !isPortAvailable(listenPort) {
		t.Skipf("Skipping test: port %d is already in use by another test", listenPort)
	}

	listenAddr := fmt.Sprintf("http://localhost:%d", listenPort)
	fmt.Printf("INFO: 'goEmployeJwtServer will start HTTP server listening on port %s'\n", listenAddr)

	newRequest := func(method, url string, body string, headers map[string]string) *http.Request {
		fmt.Printf("INFO: 💥💥'newRequest %s on %s ##BODY : %+v'\n", method, url, body)
		r, err := http.NewRequest(method, url, strings.NewReader(body))
		if err != nil {
			t.Fatalf("### ERROR http.NewRequest %s on [%s] error is :%v\n", method, url, err)
		}
		if method == http.MethodPost {
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
		for k, v := range headers {
			r.Header.Set(k, v)
		}
		return r
	}

	// Get the ENV JWT_AUTH_URL value
	jwtAuthUrl, err := config.GetJwtAuthUrl()
	if err != nil {
		t.Fatalf("error getting JWT auth URL: %v", err)
	}
	formLogin := make(url.Values)
	mainAdminUser, err := config.GetAdminUser(defaultAdminUser)
	if err != nil {
		t.Fatalf("error getting admin user: %v", err)
	}
	mainAdminPassword, err := config.GetAdminPassword()
	if err != nil {
		t.Fatalf("error getting admin password: %v", err)
	}
	h := sha256.New()
	h.Write([]byte(mainAdminPassword))
	mainAdminPasswordHash := fmt.Sprintf("%x", h.Sum(nil))
	fmt.Printf("## mainAdminUserLogin: %s\n", mainAdminUser)
	formLogin.Set("login", mainAdminUser)
	formLogin.Set("hashed", mainAdminPasswordHash)

	tests := []testStruct{
		{
			name:           "1: Get on / should return HTML content",
			wantStatusCode: http.StatusOK,
			contentType:    "text/html",
			wantBody:       "<html",
			paramKeyValues: make(map[string]string, 0),
			httpMethod:     http.MethodGet,
			url:            "/",
			body:           "",
		},
		{
			name:           "2: Get on /health should return OK",
			wantStatusCode: http.StatusOK,
			contentType:    "application/json",
			wantBody:       "true",
			paramKeyValues: make(map[string]string, 0),
			httpMethod:     http.MethodGet,
			url:            "/health",
			body:           "",
		},
		{
			name:           "3: Get on /readiness should return OK",
			wantStatusCode: http.StatusOK,
			contentType:    "application/json",
			wantBody:       "true",
			paramKeyValues: make(map[string]string, 0),
			httpMethod:     http.MethodGet,
			url:            "/readiness",
			body:           "",
		},
		{
			name:           "4: Get on /goAppInfo should return app info",
			wantStatusCode: http.StatusOK,
			contentType:    "application/json",
			wantBody:       "app",
			paramKeyValues: make(map[string]string, 0),
			httpMethod:     http.MethodGet,
			url:            "/goAppInfo",
			body:           "",
		},
		{
			name:           "5: Post on / should return method not allowed",
			wantStatusCode: http.StatusMethodNotAllowed,
			contentType:    "text/html",
			wantBody:       "Method Not Allowed",
			paramKeyValues: make(map[string]string, 0),
			httpMethod:     http.MethodPost,
			url:            "/",
			body:           `{"junk":"test with junk text"}`,
		},
		{
			name:           "6: Get on nonexistent route should return 404",
			wantStatusCode: http.StatusNotFound,
			contentType:    "text/html",
			wantBody:       "not found",
			paramKeyValues: make(map[string]string, 0),
			httpMethod:     http.MethodGet,
			url:            "/aroutethatwillneverexisthere",
			body:           "",
		},
		{
			name:           "7: POST to login with valid admin credentials should return JWT token",
			wantStatusCode: http.StatusOK,
			contentType:    "application/json",
			wantBody:       "token",
			paramKeyValues: make(map[string]string, 0),
			httpMethod:     http.MethodPost,
			url:            jwtAuthUrl,
			body:           formLogin.Encode(),
		},
		{
			name:           "8: POST to login with empty credentials should return error",
			wantStatusCode: http.StatusInternalServerError,
			contentType:    "application/json",
			wantBody:       "error",
			paramKeyValues: make(map[string]string, 0),
			httpMethod:     http.MethodPost,
			url:            jwtAuthUrl,
			body:           "",
		},
		{
			name:           "9: GET on login URL without UserId header should return 401",
			wantStatusCode: http.StatusUnauthorized,
			contentType:    "application/json",
			wantBody:       "UserId",
			paramKeyValues: make(map[string]string, 0),
			httpMethod:     http.MethodGet,
			url:            jwtAuthUrl,
			body:           "",
		},
	}

	// starting main in its own go routine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		main()
	}()
	gohttpclient.WaitForHttpServer(listenAddr, 1*time.Second, 15, l)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newRequest(tt.httpMethod, listenAddr+tt.url, tt.body, tt.headers)
			resp, err := http.DefaultClient.Do(r)
			if DEBUG {
				fmt.Printf("### %s : %s on %s\n", tt.name, r.Method, r.URL)
			}
			if err != nil {
				fmt.Printf("### GOT ERROR : %s\n%+v", err, resp)
				t.Fatal(err)
			}
			defer resp.Body.Close()
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode, assertCorrectStatusCodeExpected)
			receivedJson, _ := io.ReadAll(resp.Body)

			if DEBUG {
				fmt.Printf("WANTED   :%T - %#v\n", tt.wantBody, tt.wantBody)
				fmt.Printf("RECEIVED :%T - %#v\n", receivedJson, string(receivedJson))
			}
			assert.Contains(t, string(receivedJson), tt.wantBody, "Response should contain what was expected.")
		})
	}
}
