package goHttpEcho

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/config"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type FuncAreWeReady func(msg string) bool

type FuncAreWeHealthy func(msg string) bool

// Server is a struct type to store information related to all handlers of web server
type Server struct {
	listenAddress string
	logger        golog.MyLogger
	e             *echo.Echo
	r             *echo.Group // // Restricted group
	router        *http.ServeMux
	startTime     time.Time
	Authenticator Authentication
	JwtCheck      JwtChecker
	VersionReader VersionReader
	httpServer    http.Server
}

type Config struct {
	ListenAddress string
	Authenticator Authentication
	JwtCheck      JwtChecker
	VersionReader VersionReader
	Logger        golog.MyLogger
	WebRootDir    string
	Content       embed.FS
	RestrictedUrl string
}

// NewGoHttpServer is a constructor that initializes the server,routes and all fields in Server type
func NewGoHttpServer(serverConfig *Config) *Server {
	l := serverConfig.Logger
	listenAddress := serverConfig.ListenAddress
	Auth := serverConfig.Authenticator
	JwtCheck := serverConfig.JwtCheck
	Ver := serverConfig.VersionReader
	webRootDir := serverConfig.WebRootDir
	content := serverConfig.Content
	restrictedUrl := serverConfig.RestrictedUrl
	myServerMux := http.NewServeMux()
	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		l.Debug("in customHTTPErrorHandler got error: %v", err)
		code := http.StatusInternalServerError
		var he *echo.HTTPError
		if errors.As(err, &he) {
			code = he.Code
		}
		l.TraceHttpRequest(fmt.Sprintf("⚠️ customHTTPErrorHandler http status:%d", code), c.Request())
		response := GetStandardResponse("error", err.Error(), false, nil, err.Error())
		c.JSON(code, response)
	}
	var contentHandler = echo.WrapHandler(http.FileServer(http.FS(content)))

	// The embedded files will all be in the '/yourWebRootDir/dist/' folder so need to rewrite the request (could also do this with fs.Sub)
	var contentRewrite = middleware.Rewrite(map[string]string{"/*": fmt.Sprintf("/%s$1", webRootDir)})
	e.GET("/*", contentHandler, contentRewrite, func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err != nil && errors.Is(err, os.ErrNotExist) {
				l.Warn("Static file not found: %v, path: %s", err, c.Request().URL.Path)
				return c.JSON(http.StatusNotFound, GetStandardResponse("error", "File not found", false, nil, "Requested resource not found"))
			}
			return err
		}
	})

	// Restricted group definition : we decide to only all authenticated calls to the URL /api
	r := e.Group(restrictedUrl)
	r.Use(JwtCheck.JwtMiddleware)

	var defaultHttpLogger *log.Logger
	defaultHttpLogger, err := l.GetDefaultLogger()
	if err != nil {
		// in case we cannot get a valid logger.Logger for http let's create a reasonable one
		defaultHttpLogger = log.New(os.Stderr, "NewGoHttpServer::defaultHttpLogger", log.Ldate|log.Ltime|log.Lshortfile)
	}

	myServer := Server{
		listenAddress: listenAddress,
		logger:        l,
		r:             r,
		e:             e,
		router:        myServerMux,
		startTime:     time.Now(),
		Authenticator: Auth,
		JwtCheck:      JwtCheck,
		VersionReader: Ver,
		httpServer: http.Server{
			Addr:         listenAddress,       // configure the bind address
			ErrorLog:     defaultHttpLogger,   // set the logger for the server
			ReadTimeout:  defaultReadTimeout,  // max time to read request from the client
			WriteTimeout: defaultWriteTimeout, // max time to write response to the client
			IdleTimeout:  defaultIdleTimeout,  // max time for connections using TCP Keep-Alive
		},
	}

	return &myServer
}

// CreateNewServerFromEnvOrFail creates a new server from environment variables or fails
func CreateNewServerFromEnvOrFail(defaultPort int, defaultServerIp string, srvConfig *Config) *Server {
	listenPort := config.GetPortFromEnvOrPanic(defaultPort)
	listenIP := config.GetListenIpFromEnvOrPanic(defaultServerIp)
	listenAddr := fmt.Sprintf("%s:%d", listenIP, listenPort)

	server := NewGoHttpServer(&Config{
		ListenAddress: listenAddr,
		Authenticator: srvConfig.Authenticator,
		JwtCheck:      srvConfig.JwtCheck,
		VersionReader: srvConfig.VersionReader,
		Logger:        srvConfig.Logger,
		WebRootDir:    srvConfig.WebRootDir,
		Content:       srvConfig.Content,
		RestrictedUrl: srvConfig.RestrictedUrl,
	})
	return server

}

// GetEcho  returns a pointer to the Echo reference
func (s *Server) GetEcho() *echo.Echo {
	return s.e
}

// GetRestrictedGroup  adds a handler for this web server
func (s *Server) GetRestrictedGroup() *echo.Group {
	return s.r
}

// AddGetRoute  adds a handler for this web server
func (s *Server) AddGetRoute(baseURL string, urlPath string, handler echo.HandlerFunc) {
	// the next route is not restricted with jwt token
	s.e.GET(baseURL+urlPath, handler)
}

// GetRouter returns the ServeMux of this web server
func (s *Server) GetRouter() *http.ServeMux {
	return s.router
}

// GetListenAddress returns the listen address of this web server
func (s *Server) GetListenAddress() string {
	return s.listenAddress
}

// GetLog returns the log of this web server
func (s *Server) GetLog() golog.MyLogger {
	return s.logger
}

// GetStartTime returns the start time of this web server
func (s *Server) GetStartTime() time.Time {
	return s.startTime
}

// StartServer initializes all the handlers paths of this web server, it is called inside the NewGoHttpServer constructor
func (s *Server) StartServer() error {
	// Starting the web server in his own goroutine
	go func() {
		s.logger.Info("starting http server listening at %s://%s/", defaultProtocol, s.listenAddress)
		err := s.e.StartServer(&s.httpServer)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Fatal("💥💥 error starting server on %q. error: %s", s.listenAddress, err)
		}
	}()
	s.logger.Debug("Server listening on : %s PID:[%d]", s.httpServer.Addr, os.Getpid())

	// Graceful Shutdown on SIGINT (interrupt)
	waitForShutdownToExit(&s.httpServer, secondsShutDownTimeout)
	return nil
}

// waitForShutdownToExit will wait for interrupt signal SIGINT or SIGTERM and gracefully shutdown the server after secondsToWait seconds.
func waitForShutdownToExit(srv *http.Server, secondsToWait time.Duration) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received.
	// wait for SIGINT (interrupt) 	: ctrl + C keypress, or in a shell : kill -SIGINT processId
	sig := <-interruptChan
	srv.ErrorLog.Printf("INFO: 'SIGINT %d interrupt signal received, about to shut down server after max %v seconds...'", sig, secondsToWait.Seconds())

	// create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), secondsToWait)
	defer cancel()
	// gracefully shuts down the server without interrupting any active connections
	// as long as the actives connections last less than shutDownTimeout
	// https://pkg.go.dev/net/http#Server.Shutdown
	if err := srv.Shutdown(ctx); err != nil {
		srv.ErrorLog.Printf("💥💥 ERROR: 'Problem doing Shutdown %v'", err)
	}
	<-ctx.Done()
	srv.ErrorLog.Println("INFO: 'Server gracefully stopped, will exit'")
	os.Exit(0)
}
