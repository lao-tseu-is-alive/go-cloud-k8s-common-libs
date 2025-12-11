package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
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
}

// TestMainExec is instantiating the "real" main code using the env variable (in your .env files if you use the Makefile rule)
func TestMainExec(t *testing.T) {
	l, err := golog.NewLogger("simple", os.Stdout, golog.DebugLevel, "TestMainExec")
	if err != nil {
		log.Fatalf("💥💥 error log.NewLogger error: %v'\n", err)
	}
	listenPort := config.GetPortFromEnvOrPanic(defaultPort)
	listenAddr := fmt.Sprintf("http://localhost:%d", listenPort)
	fmt.Printf("INFO: 'Will start HTTP server listening on port %s'\n", listenAddr)

	newRequest := func(method, url string, body string) *http.Request {
		fmt.Printf("INFO: 💥💥'newRequest %s on %s ##BODY : %+v'\n", method, url, body)
		r, err := http.NewRequest(method, url, strings.NewReader(body))
		if err != nil {
			t.Fatalf("### ERROR http.NewRequest %s on [%s] error is :%v\n", method, url, err)
		}
		if method == http.MethodPost {
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
		return r
	}

	// Get the ENV JWT_AUTH_URL value
	jwtAuthUrl := config.GetJwtAuthUrlFromEnvOrPanic()
	formLogin := make(url.Values)
	mainAdminUser := config.GetAdminUserFromEnvOrPanic(defaultAdminUser)
	mainAdminPassword := config.GetAdminPasswordFromEnvOrPanic()
	h := sha256.New()
	h.Write([]byte(mainAdminPassword))
	mainAdminPasswordHash := fmt.Sprintf("%x", h.Sum(nil))
	fmt.Printf("## mainAdminUserLogin: %s", mainAdminUser)
	//fmt.Printf("## mainAdminPasswordHash: %s", mainAdminPasswordHash)
	formLogin.Set("login", mainAdminUser)
	formLogin.Set("hashed", mainAdminPasswordHash)

	tests := []testStruct{
		{
			name:           "1: Get on default get handler should contain html tag",
			wantStatusCode: http.StatusOK,
			contentType:    MIMEHtmlCharsetUTF8,
			wantBody:       "<html",
			paramKeyValues: make(map[string]string, 0),
			httpMethod:     http.MethodGet,
			url:            "/",
			body:           "",
		},
		{
			name:           "2: Post on default get handler should return an http error method not allowed ",
			wantStatusCode: http.StatusMethodNotAllowed,
			contentType:    MIMEHtmlCharsetUTF8,
			wantBody:       "Method Not Allowed",
			paramKeyValues: make(map[string]string, 0),
			httpMethod:     http.MethodPost,
			url:            "/",
			body:           `{"junk":"test with junk text"}`,
		},
		{
			name:           "3: Get on nonexistent route should return an http error not found ",
			wantStatusCode: http.StatusNotFound,
			contentType:    MIMEHtmlCharsetUTF8,
			wantBody:       "page not found",
			paramKeyValues: make(map[string]string, 0),
			httpMethod:     http.MethodGet,
			url:            "/aroutethatwillneverexisthere",
			body:           "",
		},
		{
			name:           "4: POST to login with valid credential should return a JWT token ",
			wantStatusCode: http.StatusOK,
			contentType:    MIMEHtmlCharsetUTF8,
			wantBody:       "token",
			paramKeyValues: make(map[string]string, 0),
			httpMethod:     http.MethodPost,
			url:            jwtAuthUrl,
			body:           formLogin.Encode(),
		},
	}

	// starting main in his own go routine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		main()
	}()
	gohttpclient.WaitForHttpServer(listenAddr, 1*time.Second, 10, l)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := newRequest(tt.httpMethod, listenAddr+tt.url, tt.body)
			//r.Header.Set(HeaderContentType, tt.contentType)
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
			// check that receivedJson contains the specified tt.wantBody substring . https://pkg.go.dev/github.com/stretchr/testify/assert#Contains
			assert.Contains(t, string(receivedJson), tt.wantBody, "Response should contain what was expected.")
		})
	}
}
