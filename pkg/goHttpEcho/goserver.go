package goHttpEcho

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cristalhq/jwt/v4"
	echojwt "github.com/labstack/echo-jwt/v4"
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

const (
	defaultProtocol        = "http"
	secondsShutDownTimeout = 5 * time.Second  // maximum number of second to wait before closing server
	defaultReadTimeout     = 10 * time.Second // max time to read request from the client
	defaultWriteTimeout    = 10 * time.Second // max time to write response to the client
	defaultIdleTimeout     = 2 * time.Minute  // max time for connections using TCP Keep-Alive
	initCallMsg            = "INITIAL CALL TO %s()"
	formatTraceRequest     = "TRACE: [%s] %s  path:'%s', RemoteAddrIP: [%s], msg: %s, val: %v"
)

// JwtCustomClaims are custom claims extending default ones.
type JwtCustomClaims struct {
	jwt.RegisteredClaims
	Id       int32  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
}

type FuncAreWeReady func(msg string) bool

type FuncAreWeHealthy func(msg string) bool

// GoHttpServer is a struct type to store information related to all handlers of web server
type GoHttpServer struct {
	listenAddress string
	log           golog.MyLogger
	e             *echo.Echo
	r             *echo.Group // // Restricted group
	router        *http.ServeMux
	startTime     time.Time
	httpServer    http.Server
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

// NewGoHttpServer is a constructor that initializes the server,routes and all fields in GoHttpServer type
func NewGoHttpServer(listenAddress string, l golog.MyLogger, webRootDir string, content embed.FS, restrictedUrl string) *GoHttpServer {
	myServerMux := http.NewServeMux()

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.CORS())
	signingKey := config.GetJwtSecretFromEnvOrPanic()
	JwtSecret := []byte(signingKey)
	e.HideBanner = true
	/* will try a better way to handle 404 */
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		l.Debug("in customHTTPErrorHandler got error: %v", err)
		re := c.Request()
		l.Debug("customHTTPErrorHandler original failed request: %+v", re)
		code := http.StatusInternalServerError
		if he, ok := err.(*echo.HTTPError); ok {
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
	// Configure middleware with the custom claims type
	configJwt := echojwt.Config{
		ContextKey: "jwtdata",
		SigningKey: JwtSecret,

		ParseTokenFunc: func(c echo.Context, auth string) (interface{}, error) {
			verifier, _ := jwt.NewVerifierHS(jwt.HS512, JwtSecret)
			// claims are of type `jwt.MapClaims` when token is created with `jwt.Parse`
			token, err := jwt.Parse([]byte(auth), verifier)
			if err != nil {
				return nil, err
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
				// TODO: find a way to call this function (in User microservice)
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

	var defaultHttpLogger *log.Logger
	defaultHttpLogger, err := l.GetDefaultLogger()
	if err != nil {
		// in case we cannot get a valid log.Logger for http let's create a reasonable one
		defaultHttpLogger = log.New(os.Stderr, "NewGoHttpServer::defaultHttpLogger", log.Ldate|log.Ltime|log.Lshortfile)
	}

	myServer := GoHttpServer{
		listenAddress: listenAddress,
		log:           l,
		r:             r,
		e:             e,
		router:        myServerMux,
		startTime:     time.Now(),
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

// GetEcho  returns a pointer to the Echo reference
func (s *GoHttpServer) GetEcho() *echo.Echo {
	return s.e
}

// GetRestrictedGroup  adds a handler for this web server
func (s *GoHttpServer) GetRestrictedGroup() *echo.Group {
	return s.r
}

// AddGetRoute  adds a handler for this web server
func (s *GoHttpServer) AddGetRoute(baseURL string, urlPath string, handler echo.HandlerFunc) {
	// the next route is not restricted with jwt token
	s.e.GET(baseURL+urlPath, handler)

}

//############# BEGIN K8S standard Echo HANDLERS

func (s *GoHttpServer) GetReadinessHandler(readyFunc FuncAreWeReady, msg string) echo.HandlerFunc {
	handlerName := "GetReadinessHandler"
	s.log.Debug(initCallMsg, handlerName)
	return echo.HandlerFunc(func(ctx echo.Context) error {
		ready := readyFunc(msg)
		r := ctx.Request()
		s.log.Debug(formatTraceRequest, handlerName, r.Method, r.URL.Path, r.RemoteAddr, msg, ready)
		if ready {
			msgOK := fmt.Sprintf("GetReadinessHandler: (%s) is ready: %#v ", msg, ready)
			return ctx.JSON(http.StatusOK, msgOK)
		} else {
			msgErr := fmt.Sprintf("GetReadinessHandler: (%s) is not ready: %#v ", msg, ready)
			return echo.NewHTTPError(http.StatusInternalServerError, msgErr)
		}
	})
}
func (s *GoHttpServer) GetHealthHandler(healthyFunc FuncAreWeHealthy, msg string) echo.HandlerFunc {
	handlerName := "GetHealthHandler"
	s.log.Debug(initCallMsg, handlerName)
	return echo.HandlerFunc(func(ctx echo.Context) error {
		healthy := healthyFunc(msg)
		r := ctx.Request()
		s.log.Debug(formatTraceRequest, handlerName, r.Method, r.URL.Path, r.RemoteAddr, msg, healthy)
		if healthy {
			msgOK := fmt.Sprintf("GetHealthHandler: (%s) is healthy: %#v ", msg, healthy)
			return ctx.JSON(http.StatusOK, msgOK)
		} else {
			msgErr := fmt.Sprintf("GetHealthHandler: (%s) is not healthy: %#v ", msg, healthy)
			return echo.NewHTTPError(http.StatusInternalServerError, msgErr)
		}
	})
}

// StartServer initializes all the handlers paths of this web server, it is called inside the NewGoHttpServer constructor
func (s *GoHttpServer) StartServer() error {

	// Starting the web server in his own goroutine
	go func() {
		s.log.Info("starting http server listening at %s://localhost%s/", defaultProtocol, s.listenAddress)
		err := s.e.StartServer(&s.httpServer)
		if err != nil && err != http.ErrServerClosed {
			s.log.Fatal("💥💥 error could not listen on tcp port %q. error: %s", s.listenAddress, err)
		}
	}()
	s.log.Debug("Server listening on : %s PID:[%d]", s.httpServer.Addr, os.Getpid())

	// Graceful Shutdown on SIGINT (interrupt)
	waitForShutdownToExit(&s.httpServer, secondsShutDownTimeout)
	return nil
}
