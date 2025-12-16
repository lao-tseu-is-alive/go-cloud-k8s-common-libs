package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/config"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/goHttpEcho"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/version"
)

const (
	APP                 = "goCloudK8sCommonLibsDemoServer"
	defaultPort         = 8080
	defaultJwtStatusUrl = "/status"
	restrictedUrl       = "/api/v1"
	defaultWebRootDir   = "web/"
	defaultAdminUser    = "goadmin"
	defaultAdminEmail   = "goadmin@yourdomain.org"
	defaultAdminId      = 960901
)

// content holds our static web server content.
//
//go:embed web/*
var content embed.FS

type Service struct {
	Logger golog.MyLogger
	server *goHttpEcho.Server
}

// login is just a trivial stupid example to test this server
func (s Service) login(ctx echo.Context) error {
	goHttpEcho.TraceHttpRequest("login", ctx.Request(), s.Logger)
	login := ctx.FormValue("login")
	passwordHash := ctx.FormValue("hashed")
	s.Logger.Debug("login: %s, hash: %s ", login, passwordHash)
	if len(strings.Trim(login, " ")) < 1 {
		return ctx.JSON(http.StatusUnauthorized, "invalid credentials")
	}

	requestCtx := ctx.Request().Context()
	if s.server.Authenticator.AuthenticateUser(requestCtx, login, passwordHash) {
		userInfo, err := s.server.Authenticator.GetUserInfoFromLogin(requestCtx, login)
		if err != nil {
			errMsg := fmt.Sprintf("Error getting user info from login: %v", err)
			s.Logger.Error(errMsg)
			return ctx.JSON(http.StatusInternalServerError, errMsg)
		}
		token, err := s.server.JwtCheck.GetTokenFromUserInfo(userInfo)
		if err != nil {
			errMsg := fmt.Sprintf("Error getting jwt token from user info: %v", err)
			s.Logger.Error(errMsg)
			return ctx.JSON(http.StatusInternalServerError, errMsg)
		}
		response := map[string]string{"token": token.String()}
		s.Logger.Info("LoginUser(%s) successful login", login)
		return ctx.JSON(http.StatusOK, response)
	}
	return ctx.JSON(http.StatusUnauthorized, "username not found or password invalid")
}

func (s Service) restricted(ctx echo.Context) error {
	goHttpEcho.TraceHttpRequest("restricted", ctx.Request(), s.Logger)
	claims := s.server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	s.Logger.Info("in restricted : currentUserId: %d", currentUserId)
	return ctx.JSON(http.StatusCreated, claims)
}

func main() {
	l, err := golog.NewLogger("simple", os.Stdout, golog.DebugLevel, APP)
	if err != nil {
		log.Fatalf("💥💥 error log.NewLogger error: %v'\n", err)
	}
	l.Info("🚀🚀 Starting:'%s', v%s, rev:%s, build:%v from: %s", APP, version.VERSION, version.REVISION, version.BuildStamp, version.REPOSITORY)

	// Get the ENV JWT_AUTH_URL value
	jwtAuthUrl, err := config.GetJwtAuthUrl()
	if err != nil {
		l.Fatal("💥💥 error getting JWT auth URL: %v", err)
	}
	jwtStatusUrl := config.GetJwtStatusUrl(defaultJwtStatusUrl)

	myVersionReader := goHttpEcho.NewSimpleVersionReader(
		APP,
		version.VERSION,
		version.REPOSITORY,
		version.REVISION,
		version.BuildStamp,
		jwtAuthUrl,
		jwtStatusUrl,
	)

	// Create a new JWT checker using factory function
	myJwt, err := goHttpEcho.GetNewJwtCheckerFromConfig(APP, 60, l)
	if err != nil {
		l.Fatal("💥💥 error creating JWT checker: %v", err)
	}

	// Create a new Authenticator using factory function
	myAuthenticator, err := goHttpEcho.GetSimpleAdminAuthenticatorFromConfig(
		goHttpEcho.AdminDefaults{
			UserId:     defaultAdminId,
			ExternalId: 9999999,
			Login:      defaultAdminUser,
			Email:      defaultAdminEmail,
		},
		myJwt,
	)
	if err != nil {
		l.Fatal("💥💥 error creating authenticator: %v", err)
	}

	server, err := goHttpEcho.CreateNewServerFromEnv(
		defaultPort,
		"0.0.0.0",
		&goHttpEcho.Config{
			ListenAddress: "",
			Authenticator: myAuthenticator,
			JwtCheck:      myJwt,
			VersionReader: myVersionReader,
			Logger:        l,
			WebRootDir:    defaultWebRootDir,
			Content:       content,
			RestrictedUrl: restrictedUrl,
		},
	)
	if err != nil {
		l.Fatal("💥💥 error creating server: %v", err)
	}

	e := server.GetEcho()
	e.GET("/goAppInfo", server.GetAppInfoHandler())
	yourService := Service{
		Logger: l,
		server: server,
	}
	e.POST(jwtAuthUrl, yourService.login)
	r := server.GetRestrictedGroup()
	r.GET(jwtStatusUrl, yourService.restricted)
	err = server.StartServer()
	if err != nil {
		l.Fatal("💥💥 error doing server.StartServer error: %v'\n", err)
	}
}
