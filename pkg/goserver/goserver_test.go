package goserver

import (
	"embed"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestGoHttpServer_AddGetRoute(t *testing.T) {
	type fields struct {
		listenAddress string
		logger        *log.Logger
		e             *echo.Echo
		r             *echo.Group
		router        *http.ServeMux
		startTime     time.Time
		httpServer    http.Server
	}
	type args struct {
		baseURL string
		urlPath string
		handler echo.HandlerFunc
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GoHttpServer{
				listenAddress: tt.fields.listenAddress,
				logger:        tt.fields.logger,
				e:             tt.fields.e,
				r:             tt.fields.r,
				router:        tt.fields.router,
				startTime:     tt.fields.startTime,
				httpServer:    tt.fields.httpServer,
			}
			s.AddGetRoute(tt.args.baseURL, tt.args.urlPath, tt.args.handler)
		})
	}
}

func TestGoHttpServer_GetEcho(t *testing.T) {
	type fields struct {
		listenAddress string
		logger        *log.Logger
		e             *echo.Echo
		r             *echo.Group
		router        *http.ServeMux
		startTime     time.Time
		httpServer    http.Server
	}
	tests := []struct {
		name   string
		fields fields
		want   *echo.Echo
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GoHttpServer{
				listenAddress: tt.fields.listenAddress,
				logger:        tt.fields.logger,
				e:             tt.fields.e,
				r:             tt.fields.r,
				router:        tt.fields.router,
				startTime:     tt.fields.startTime,
				httpServer:    tt.fields.httpServer,
			}
			if got := s.GetEcho(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEcho() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGoHttpServer_GetHealthHandler(t *testing.T) {
	type fields struct {
		listenAddress string
		logger        *log.Logger
		e             *echo.Echo
		r             *echo.Group
		router        *http.ServeMux
		startTime     time.Time
		httpServer    http.Server
	}
	type args struct {
		healthyFunc FuncAreWeHealthy
		msg         string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   echo.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GoHttpServer{
				listenAddress: tt.fields.listenAddress,
				logger:        tt.fields.logger,
				e:             tt.fields.e,
				r:             tt.fields.r,
				router:        tt.fields.router,
				startTime:     tt.fields.startTime,
				httpServer:    tt.fields.httpServer,
			}
			if got := s.GetHealthHandler(tt.args.healthyFunc, tt.args.msg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetHealthHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGoHttpServer_GetReadinessHandler(t *testing.T) {
	type fields struct {
		listenAddress string
		logger        *log.Logger
		e             *echo.Echo
		r             *echo.Group
		router        *http.ServeMux
		startTime     time.Time
		httpServer    http.Server
	}
	type args struct {
		readyFunc FuncAreWeReady
		msg       string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   echo.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GoHttpServer{
				listenAddress: tt.fields.listenAddress,
				logger:        tt.fields.logger,
				e:             tt.fields.e,
				r:             tt.fields.r,
				router:        tt.fields.router,
				startTime:     tt.fields.startTime,
				httpServer:    tt.fields.httpServer,
			}
			if got := s.GetReadinessHandler(tt.args.readyFunc, tt.args.msg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetReadinessHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGoHttpServer_GetRestrictedGroup(t *testing.T) {
	type fields struct {
		listenAddress string
		logger        *log.Logger
		e             *echo.Echo
		r             *echo.Group
		router        *http.ServeMux
		startTime     time.Time
		httpServer    http.Server
	}
	tests := []struct {
		name   string
		fields fields
		want   *echo.Group
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GoHttpServer{
				listenAddress: tt.fields.listenAddress,
				logger:        tt.fields.logger,
				e:             tt.fields.e,
				r:             tt.fields.r,
				router:        tt.fields.router,
				startTime:     tt.fields.startTime,
				httpServer:    tt.fields.httpServer,
			}
			if got := s.GetRestrictedGroup(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRestrictedGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGoHttpServer_StartServer(t *testing.T) {
	type fields struct {
		listenAddress string
		logger        *log.Logger
		e             *echo.Echo
		r             *echo.Group
		router        *http.ServeMux
		startTime     time.Time
		httpServer    http.Server
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GoHttpServer{
				listenAddress: tt.fields.listenAddress,
				logger:        tt.fields.logger,
				e:             tt.fields.e,
				r:             tt.fields.r,
				router:        tt.fields.router,
				startTime:     tt.fields.startTime,
				httpServer:    tt.fields.httpServer,
			}
			if err := s.StartServer(); (err != nil) != tt.wantErr {
				t.Errorf("StartServer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewGoHttpServer(t *testing.T) {
	type args struct {
		listenAddress string
		l             *log.Logger
		webRootDir    string
		content       embed.FS
		restrictedUrl string
	}
	tests := []struct {
		name string
		args args
		want *GoHttpServer
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewGoHttpServer(tt.args.listenAddress, tt.args.l, tt.args.webRootDir, tt.args.content, tt.args.restrictedUrl); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGoHttpServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_waitForShutdownToExit(t *testing.T) {
	type args struct {
		srv           *http.Server
		secondsToWait time.Duration
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			waitForShutdownToExit(tt.args.srv, tt.args.secondsToWait)
		})
	}
}
