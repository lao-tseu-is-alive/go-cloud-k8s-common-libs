package goserver

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

// GoHttpServer is a struct type to store information related to all handlers of web server
type GoHttpServer struct {
	listenAddress string
	logger        *log.Logger
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
	srv.ErrorLog.Printf("INFO: 'SIGINT %d interrupt signal received, about to shut down server after max %v seconds...'\n", sig, secondsToWait.Seconds())

	// create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), secondsToWait)
	defer cancel()
	// gracefully shuts down the server without interrupting any active connections
	// as long as the actives connections last less than shutDownTimeout
	// https://pkg.go.dev/net/http#Server.Shutdown
	if err := srv.Shutdown(ctx); err != nil {
		srv.ErrorLog.Printf("💥💥 ERROR: 'Problem doing Shutdown %v'\n", err)
	}
	<-ctx.Done()
	srv.ErrorLog.Println("INFO: 'Server gracefully stopped, will exit'")
	os.Exit(0)
}

// NewGoHttpServer is a constructor that initializes the server,routes and all fields in GoHttpServer type
func NewGoHttpServer(listenAddress string, l *log.Logger, webRootDir string, content embed.FS) *GoHttpServer {
	myServerMux := http.NewServeMux()

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.CORS())
	signingKey, err := config.GetJwtSecretFromEnv()
	JwtSecret := []byte(signingKey)
	if err != nil {
		l.Fatalf("💥💥 ERROR: 'in NewGoHttpServer config.GetJwtSecretFromEnv() got error: %v'\n", err)
	}
	if err != nil {
		l.Fatalf("💥💥 ERROR: 'in NewGoHttpServer config.GetJwtDurationFromEnv() got error: %v'\n", err)
	}
	e.HideBanner = true
	/* will try a better way to handle 404 */
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		l.Printf("TRACE: in customHTTPErrorHandler got error: %v\n", err)
		code := http.StatusInternalServerError
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
		}
		c.Logger().Error(err)
		if code == 404 {
			errorPage := fmt.Sprintf("%s/%d.html", webRootDir, code)
			res, err := content.ReadFile(errorPage)
			if err != nil {
				l.Printf("💥💥 ERROR: 'in  content.ReadFile(%s) got error: %v'\n", errorPage, err)
			}
			if err := c.HTMLBlob(code, res); err != nil {
				l.Printf("💥💥 ERROR: 'in  c.HTMLBlob(%d, %s) got error: %v'\n", code, res, err)
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
	r := e.Group("/api")
	// Configure middleware with the custom claims type
	config := echojwt.Config{
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

			l.Printf("INFO : JWT ParseTokenFunc, Algorithm %v\n", token.Header().Algorithm)
			l.Printf("INFO : JWT ParseTokenFunc, Type      %v\n", token.Header().Type)
			l.Printf("INFO : JWT ParseTokenFunc, Claims    %v\n", string(token.Claims()))
			l.Printf("INFO : JWT ParseTokenFunc, Payload   %v\n", string(token.PayloadPart()))
			l.Printf("INFO : JWT ParseTokenFunc, Token     %v\n", string(token.Bytes()))
			l.Printf("INFO : JWT ParseTokenFunc, ParseTokenFunc : Claims:    %+v\n", string(token.Claims()))
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
				//l.Printf("💥💥 ERROR: 'in  content.ReadFile(%s) got error: %v'\n", errorPage, err)
				return token, nil // ALL IS GOOD HERE
			} else {
				l.Printf("ERROR : JWT ParseTokenFunc,  : IsValidAt(%+v)\n", time.Now())
				return nil, errors.New("token has expired")
			}

		},
	}
	r.Use(echojwt.WithConfig(config))

	myServer := GoHttpServer{
		listenAddress: listenAddress,
		logger:        l,
		r:             r,
		e:             e,
		router:        myServerMux,
		startTime:     time.Now(),
		httpServer: http.Server{
			Addr:         listenAddress,       // configure the bind address
			ErrorLog:     l,                   // set the logger for the server
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
	//s.router.Handle("/", s.getMyDefaultHandler())
	//s.e.GET("/readiness", s.getReadinessHandler())
	// s.e.GET("/health", s.getHealthHandler())
	// the next route is not restricted with jwt token
	s.e.GET(baseURL+urlPath, handler)

}

// StartServer initializes all the handlers paths of this web server, it is called inside the NewGoHttpServer constructor
func (s *GoHttpServer) StartServer() error {

	// Starting the web server in his own goroutine
	go func() {
		s.logger.Printf("INFO: Starting http server listening at %s://localhost%s/", defaultProtocol, s.listenAddress)
		err := s.e.StartServer(&s.httpServer)
		if err != nil && err != http.ErrServerClosed {
			s.logger.Fatalf("💥💥 ERROR: 'Could not listen on %q: %s'\n", s.listenAddress, err)
		}
	}()
	s.logger.Printf("Server listening on : %s PID:[%d]", s.httpServer.Addr, os.Getpid())

	// Graceful Shutdown on SIGINT (interrupt)
	waitForShutdownToExit(&s.httpServer, secondsShutDownTimeout)
	return nil
}
