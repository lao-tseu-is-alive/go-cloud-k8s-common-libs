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
	"github.com/rs/xid"
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

// NewGoHttpServer is a constructor that initializes the server,routes and all fields in Server type
func NewGoHttpServer(listenAddress string, Auth Authentication, JwtCheck JwtChecker, Ver VersionReader, l golog.MyLogger, webRootDir string, content embed.FS, restrictedUrl string) *Server {
	myServerMux := http.NewServeMux()

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.CORS())
	e.HideBanner = true
	/* will try a better way to handle 404 */
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		l.Debug("in customHTTPErrorHandler got error: %v", err)
		re := c.Request()
		TraceRequest("customHTTPErrorHandler", re, l)
		code := http.StatusInternalServerError
		var he *echo.HTTPError
		if errors.As(err, &he) {
			code = he.Code
		}
		if code == 404 {
			errorPage := fmt.Sprintf("%s/%d.html", webRootDir, code)
			res, err := content.ReadFile(errorPage)
			if err != nil {
				l.Error("in  content.ReadFile(%s) got error: %v", errorPage, err)
			}
			if err := c.HTMLBlob(code, res); err != nil {
				l.Error("in  c.HTMLBlob(%d, %s) got error: %v", code, res, err)
				c.Logger().Error(err)
			}
		} else {
			c.JSON(code, err)
		}
	}
	var contentHandler = echo.WrapHandler(http.FileServer(http.FS(content)))

	// The embedded files will all be in the '/goCloudK8sUserGroupFront/dist/' folder so need to rewrite the request (could also do this with fs.Sub)
	var contentRewrite = middleware.Rewrite(map[string]string{"/*": fmt.Sprintf("/%s$1", webRootDir)})
	e.GET("/*", contentHandler, contentRewrite)

	// Restricted group definition : we decide to only all authenticated calls to the URL /api
	r := e.Group(restrictedUrl)
	/*
		// Configure middleware with the custom claims type
		configJwt := echojwt.Config{
			ContextKey: "jwtdata",
			SigningKey: JwtSecret,

			ParseTokenFunc: func(c echo.Context, auth string) (interface{}, error) {
				//verifier, _ := jwt.NewVerifierHS(jwt.HS512, JwtSecret)
				verifier, err := jwt.NewVerifierHS(jwt.HS512, []byte(JwtSecret))
				if err != nil {
					return nil, errors.New(fmt.Sprintf("error in ParseToken creating verifier: %s", err))
				}
				// claims are of type `jwt.MapClaims` when token is created with `jwt.Parse`
				token, err := jwt.Parse([]byte(auth), verifier)
				if err != nil {
					return nil, errors.New(fmt.Sprintf("error in ParseToken parsing token: %s", err))
				}
				// get REGISTERED claims
				var newClaims jwt.RegisteredClaims
				err = json.Unmarshal(token.Claims(), &newClaims)
				if err != nil {
					return nil, err
				}

				l.Debug("JWT ParseTokenFunc, Algorithm %v", token.Header().Algorithm)
				l.Debug("JWT ParseTokenFunc, Type      %v", token.Header().Type)
				l.Debug("JWT ParseTokenFunc, Claims    %v", string(token.Claims()))
				l.Debug("JWT ParseTokenFunc, Payload   %v", string(token.PayloadPart()))
				l.Debug("JWT ParseTokenFunc, Token     %v", string(token.Bytes()))
				l.Debug("JWT ParseTokenFunc, ParseTokenFunc : Claims:    %+v", string(token.Claims()))
				if newClaims.IsValidAt(time.Now()) {
					claims := JwtCustomClaims{}
					err := token.DecodeClaims(&claims)
					if err != nil {
						return nil, errors.New("token cannot be parsed")
					}
					// IF USER IS DEACTIVATED  token should be invalidated RETURN 401 Unauthorized
					// find a way to call this function (in User microservice)
					//currentUserId := claims.Id
					//if store.IsUserActive(currentUserId) {
					//	return token, nil // ALL IS GOOD HERE
					//} else {
					// return nil, errors.New("token invalid because user account has been deactivated")
					//}
					//l.Printf("💥💥 ERROR: 'in  content.ReadFile(%s) got error: %v'", errorPage, err)
					return token, nil // ALL IS GOOD HERE
				} else {
					l.Error("JWT ParseTokenFunc,  : IsValidAt(%+v)", time.Now())
					return nil, errors.New("token has expired")
				}

			},
		}
		r.Use(echojwt.WithConfig(configJwt))
	*/
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
func CreateNewServerFromEnvOrFail(
	defaultPort int,
	defaultServerIp string,
	myAuthenticator Authentication,
	myJwt JwtChecker,
	myVersionReader VersionReader,
	l golog.MyLogger,
	webRootDir string,
	content embed.FS,
	restrictedUrl string,
) *Server {
	listenPort := config.GetPortFromEnvOrPanic(defaultPort)
	listenAddr := fmt.Sprintf("%s:%d", defaultServerIp, listenPort)
	l.Info("HTTP server will listen : %s", listenAddr)

	server := NewGoHttpServer(listenAddr, myAuthenticator, myJwt, myVersionReader, l, webRootDir, content, restrictedUrl)
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
			s.logger.Fatal("💥💥 error could not listen on tcp port %q. error: %s", s.listenAddress, err)
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

func getHtmlHeader(title string, description string) string {
	return fmt.Sprintf("%s<meta name=\"description\" content=\"%s\"><title>%s</title></head>", htmlHeaderStart, description, title)
}

func getHtmlPage(title string, description string) string {
	return getHtmlHeader(title, description) +
		fmt.Sprintf("\n<body><div class=\"container\"><h4>%s</h4></div></body></html>", title)
}
func TraceRequest(handlerName string, r *http.Request, l golog.MyLogger) {
	const formatTraceRequest = "TraceRequest:[%s] %s '%s', RemoteIP: [%s],id:%s\n"
	remoteIp := r.RemoteAddr // ip address of the original request or the last proxy
	requestedUrlPath := r.URL.Path
	guid := xid.New()
	l.Debug(formatTraceRequest, handlerName, r.Method, requestedUrlPath, remoteIp, guid.String())
}
