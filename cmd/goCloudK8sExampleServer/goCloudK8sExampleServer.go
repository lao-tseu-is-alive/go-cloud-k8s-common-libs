package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/config"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/goHttpEcho"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/metadata"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/tools"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/version"
)

const (
	APP                        = "goCloudK8sCommonLibsDemoServer"
	defaultPort                = 8080
	defaultLogName             = "stderr"
	defaultDBPort              = 5432
	defaultDBIp                = "127.0.0.1"
	defaultDBSslMode           = "prefer"
	defaultJwtStatusUrl        = "/status"
	defaultReadTimeout         = 10 * time.Second // max time to read request from the client
	defaultWebRootDir          = "goCloudK8sExampleFront/dist/"
	defaultSqlDbMigrationsPath = "db/migrations"
	defaultAdminUser           = "goadmin"
	defaultAdminEmail          = "goadmin@yourdomain.org"
	defaultAdminId             = 960901
	charsetUTF8                = "charset=UTF-8"
	MIMEHtml                   = "text/html"
	MIMEHtmlCharsetUTF8        = MIMEHtml + "; " + charsetUTF8
)

// content holds our static web server content.
//
//go:embed goCloudK8sExampleFront/dist/*
var content embed.FS

// sqlMigrations holds our db migrations sql files using https://github.com/golang-migrate/migrate
// in the line above you SHOULD have the same path  as const defaultSqlDbMigrationsPath
//
//go:embed db/migrations/*.sql
var sqlMigrations embed.FS

type Service struct {
	Logger *slog.Logger
	//Store       Storage
	dbConn database.DB
	server *goHttpEcho.Server
}

// login is just a trivial stupid example to test this server
// you should use the jwt token returned from LoginUser  in github.com/lao-tseu-is-alive/go-cloud-k8s-user-group'
// and share the same secret with the above component
func (s Service) login(ctx echo.Context) error {
	goHttpEcho.TraceHttpRequest("login", ctx.Request(), s.Logger)
	login := ctx.FormValue("login")
	passwordHash := ctx.FormValue("hashed")
	s.Logger.Debug("login attempt", "login", login)
	// maybe it was not a form but a fetch data post
	if len(strings.Trim(login, " ")) < 1 {
		return ctx.JSON(http.StatusUnauthorized, "invalid credentials")
	}

	requestCtx := ctx.Request().Context()
	if s.server.Authenticator.AuthenticateUser(requestCtx, login, passwordHash) {
		userInfo, err := s.server.Authenticator.GetUserInfoFromLogin(requestCtx, login)
		if err != nil {
			myErrMsg := fmt.Sprintf("Error getting user info from login: %v", err)
			s.Logger.Error(myErrMsg)
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"jwtStatus": myErrMsg, "token": ""})
		}
		token, err := s.server.JwtCheck.GetTokenFromUserInfo(userInfo)
		if err != nil {
			myErrMsg := fmt.Sprintf("Error getting jwt token from user info: %v", err)
			s.Logger.Error(myErrMsg)
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"jwtStatus": myErrMsg, "token": ""})
		}
		// Prepare the response
		response := map[string]string{
			"jwtStatus": "success",
			"token":     token.String(),
		}
		s.Logger.Info("LoginUser successful login", "login", login)
		return ctx.JSON(http.StatusOK, response)
	} else {
		myErrMsg := "username not found or password invalid"
		s.Logger.Warn(myErrMsg)
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"jwtStatus": myErrMsg, "token": ""})

	}
}

func (s Service) restricted(ctx echo.Context) error {
	goHttpEcho.TraceHttpRequest("restricted", ctx.Request(), s.Logger)
	// get the current user from JWT TOKEN
	claims := s.server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	s.Logger.Info("in restricted", "currentUserId", currentUserId)
	// you can check if the user is not active anymore and RETURN 401 Unauthorized
	//if !s.Store.IsUserActive(currentUserId) {
	//	return echo.NewHTTPError(http.StatusUnauthorized, "current calling user is not active anymore")
	//}
	return ctx.JSON(http.StatusCreated, claims)
}

func checkHealthy(info string) bool {
	// you decide what makes you ready, may be it is the connection to the database
	//if !stillConnectedToDB {
	//	return false
	//}
	return true
}

func (s Service) helloHandler(c echo.Context) error {
	handlerName := "helloHandler"
	s.Logger.Debug("initial call to handler", "handler", handlerName)
	// Create an instance of PageData with dynamic content
	app := s.server.VersionReader.GetAppInfo()
	appTitle := fmt.Sprintf("%s, v%s", app.App, app.Version)
	pageData := goHttpEcho.PageData{
		Title:       appTitle,
		Description: "A generic page with dynamic content blocks.",
		Theme:       "blue",
		Content: []goHttpEcho.ContentBlock{
			{Type: "heading", Value: fmt.Sprintf("Welcome to %s", appTitle)},
			{Type: "paragraph", Value: "This is the first paragraph. You can add as many as you need."},
			{Type: "paragraph", Value: "Here's another paragraph, demonstrating how the template can handle a list of content blocks."},
			{Type: "heading", Value: "Another Section"},
			{Type: "paragraph", Value: "This is a new section with a different heading."},
		},
	}

	html, err := goHttpEcho.GetHtmlPage(pageData)
	if err != nil {
		return c.HTML(http.StatusInternalServerError, goHttpEcho.GetHtmlError("💥 💥 failed to generate page"))
	}
	return c.HTML(http.StatusOK, html)

}

func main() {
	logWriter, err := config.GetLogWriter(defaultLogName)
	if err != nil {
		log.Fatalf("💥💥 error getting log writer: %v'\n", err)
	}
	logLevel, err := config.GetLogLevel(golog.InfoLevel)
	if err != nil {
		log.Fatalf("💥💥 error getting log level: %v'\n", err)
	}
	l := golog.NewLogger("simple", logWriter, logLevel, APP)
	l.Info("🚀 Starting", "app", APP, "version", version.VERSION, "revision", version.REVISION, "build", version.BuildStamp, "repository", version.REPOSITORY)

	dbDsn, err := config.GetPgDbDsnUrl(defaultDBIp, defaultDBPort, tools.ToSnakeCase(version.APP), version.AppSnake, defaultDBSslMode)
	if err != nil {
		l.Error("💥💥 error getting database DSN", "error", err)
		os.Exit(1)
	}
	dbConnCtx, dbConnCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbConnCancel()
	db, err := database.GetInstance(dbConnCtx, "pgx", dbDsn, runtime.NumCPU(), l)
	if err != nil {
		l.Error("💥💥 error doing database.GetInstance", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	dbVersion, err := db.GetVersion(context.Background())
	if err != nil {
		l.Error("💥💥 error doing dbConn.GetVersion", "error", err)
		os.Exit(1)
	}
	l.Info("connected to db", "version", dbVersion)

	// checking metadata information
	metadataService := metadata.Service{Log: l, Db: db}
	metadataService.CreateMetadataTableOrFail(context.Background())
	found, ver := metadataService.GetServiceVersionOrFail(context.Background(), version.APP)
	if found {
		l.Info("service was found in metadata", "service", version.APP, "version", ver)
	} else {
		l.Info("service was not found in metadata", "service", version.APP)
	}
	metadataService.SetServiceVersionOrFail(context.Background(), version.APP, version.VERSION)

	// https://github.com/golang-migrate/migrate
	d, err := iofs.New(sqlMigrations, defaultSqlDbMigrationsPath)
	if err != nil {
		l.Error("💥💥 error doing iofs.New for db migrations", "error", err)
		os.Exit(1)
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, strings.Replace(dbDsn, "postgres", "pgx5", 1))
	if err != nil {
		l.Error("💥💥 error doing migrate.NewWithSourceInstance", "dbURL", dbDsn, "error", err)
		os.Exit(1)
	}

	err = m.Up()
	if err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			l.Error("💥💥 error doing migrate.Up", "error", err)
			os.Exit(1)
		}
	}

	// Get the ENV JWT_AUTH_URL value
	jwtAuthUrl, err := config.GetJwtAuthUrl()
	if err != nil {
		l.Error("💥💥 error getting JWT auth URL", "error", err)
		os.Exit(1)
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
		l.Error("💥💥 error creating JWT checker", "error", err)
		os.Exit(1)
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
		l.Error("💥💥 error creating authenticator", "error", err)
		os.Exit(1)
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
			RestrictedUrl: "/api/v1",
		},
	)
	if err != nil {
		l.Error("💥💥 error creating server", "error", err)
		os.Exit(1)
	}

	e := server.GetEcho()
	e.Use(middleware.RequestLogger()) // Automatically logs requests
	e.GET("/readiness", server.GetReadinessHandler(func(info string) bool {
		ver, err := db.GetVersion(context.Background())
		if err != nil {
			l.Error("Error getting db version", "error", err)
			return false
		}
		l.Info("Connected to DB", "version", ver)
		return true
	}, "Connection to DB"))
	e.GET("/health", server.GetHealthHandler(checkHealthy, "Connection to DB"))
	yourService := Service{
		Logger: l,
		dbConn: db,
		server: server,
	}
	e.GET("/goAppInfo", server.GetAppInfoHandler())
	e.POST(jwtAuthUrl, yourService.login)
	r := server.GetRestrictedGroup()
	r.GET(jwtStatusUrl, yourService.restricted)

	e.GET("/hello", yourService.helloHandler)

	err = server.StartServer()
	if err != nil {
		log.Fatalf("💥💥 error doing server.StartServer error: %v'\n", err)
	}
}
