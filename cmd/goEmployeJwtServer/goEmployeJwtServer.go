package main

import (
	"context"
	"crypto/sha256"
	"embed"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/config"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/f5"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/goHttpEcho"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/metadata"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/tools"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/version"
)

const (
	defaultPort                  = 8080
	defaultDBPort                = 5432
	defaultDBIp                  = "127.0.0.1"
	defaultDBSslMode             = "prefer"
	defaultRestrictedUrlBasePath = "/goapi/v1"
	defaultJwtCookieName         = "goJWT_token"
	defaultJwtStatusUrl          = "/status"
	defaultWebRootDir            = "goEmployeJwtFront/dist/"
	defaultAdminUser             = "goadmin"
	defaultAdminEmail            = "goadmin@yourdomain.org"
	defaultAdminId               = 960901
)

// content holds our static web server content.
//
//go:embed goEmployeJwtFront/dist/*
var content embed.FS

// UserLogin defines model for UserLogin.
type UserLogin struct {
	PasswordHash string `json:"password_hash"`
	Username     string `json:"username"`
}
type Service struct {
	// AllowedHostnames is a list of strings which will be matched to the client
	// requesting for a connection upgrade to a websocket connection
	AllowedHostnames []string
	Logger           golog.MyLogger
	Store            f5.Storage
	dbConn           database.DB
	server           *goHttpEcho.Server
	auth             f5.Authentication
	jwtCookieName    string
}

func validateHostAllowed(r *http.Request, allowedHostnames []string, l golog.MyLogger) error {
	requesterHostname := r.Host
	l.Info("validateHostAllowed(remote host: %s)", requesterHostname)
	if slices.Contains(allowedHostnames, "*") {
		return nil
	}
	if strings.Contains(requesterHostname, ":") {
		requesterHostname = strings.Split(requesterHostname, ":")[0]
	}
	if slices.Contains(allowedHostnames, "localhost") {
		if requesterHostname == "127.0.0.1" || requesterHostname == "::1" {
			return nil
		}
	}
	for _, allowedHostname := range allowedHostnames {
		if requesterHostname == allowedHostname {
			return nil
		}
	}
	msgErr := fmt.Sprintf("failed to find '%s' in the list of allowed hostnames", requesterHostname)
	l.Warn(msgErr)
	return errors.New(msgErr)
}

func (s Service) getJwtCookieFromF5(ctx echo.Context) error {
	goHttpEcho.TraceHttpRequest("getJwtCookieFromF5", ctx.Request(), s.Logger)
	err := validateHostAllowed(ctx.Request(), s.AllowedHostnames, s.Logger)
	if err != nil {
		errMsg := fmt.Sprintf("error validating host: %v", err)
		s.Logger.Error(errMsg)
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"status": "error", "message": errMsg})
	}
	// get the user from the F5 Header UserId
	login := strings.TrimSpace(ctx.Request().Header.Get("UserId"))
	if login == "" {
		errMsg := "getJwtCookieFromF5 failed to get login because UserId F5 header is missing"
		s.Logger.Warn(errMsg)
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"status": "error", "message": errMsg})
	} else {
		s.Logger.Debug("About to check username: %s ", login)
		err := f5.ValidateLogin(login)
		if err != nil {
			errMsg := fmt.Sprintf("error validating user login: %v", err)
			s.Logger.Error(errMsg)
			return ctx.JSON(http.StatusBadRequest, map[string]string{"status": "error", "message": errMsg})
		}
		h := sha256.New()
		h.Write([]byte(version.APP))
		// just to get a valid hash, not used with F5
		appPasswordHash := fmt.Sprintf("%x", h.Sum(nil))
		requestCtx := ctx.Request().Context()
		if s.auth.AuthenticateUser(requestCtx, login, appPasswordHash) {
			userInfo, err := s.server.Authenticator.GetUserInfoFromLogin(requestCtx, login)
			if err != nil {
				errMsg := fmt.Sprintf("getJwtCookieFromF5 failed to get user info from login: %v", err)
				s.Logger.Error(errMsg)
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"status": "error", "message": errMsg})
			}
			token, err := s.server.JwtCheck.GetTokenFromUserInfo(userInfo)
			if err != nil {
				errMsg := fmt.Sprintf("getJwtCookieFromF5 failed to get jwt token from user info: %v", err)
				s.Logger.Error(errMsg)
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"status": "error", "message": errMsg})
			}
			// Prepare the http only cookie for jwt token
			cookie := new(http.Cookie)
			cookie.Name = s.jwtCookieName
			cookie.Path = "/"
			cookie.Value = token.String()
			cookie.Expires = time.Now().Add(24 * time.Hour) // Set expiration
			cookie.HttpOnly = true                          // ⭐ Most important part: prevents JS access
			cookie.Secure = true                            // Only send over HTTPS
			cookie.SameSite = http.SameSiteLaxMode          // CSRF protection
			ctx.SetCookie(cookie)
			myMsg := fmt.Sprintf("getJwtCookieFromF5(%s) successful, token set in HTTP-Only cookie.", login)
			s.Logger.Info(myMsg)
			return ctx.JSON(http.StatusOK, map[string]string{"status": "success", "message": myMsg})
		} else {
			errMsg := fmt.Sprintf("getJwtCookieFromF5 failed to get jwt token user: %s, does not exist in DB", login)
			s.Logger.Warn(errMsg)
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"status": "error", "message": errMsg})
		}
	}
}

// login is just a trivial stupid example to test this server
// you should use the jwt token returned from LoginUser  in github.com/lao-tseu-is-alive/go-cloud-k8s-user-group'
// and share the same secret with the above component
func (s Service) login(ctx echo.Context) error {
	goHttpEcho.TraceHttpRequest("login", ctx.Request(), s.Logger)
	err := validateHostAllowed(ctx.Request(), s.AllowedHostnames, s.Logger)
	if err != nil {
		errMsg := fmt.Sprintf("error validating host: %v", err)
		s.Logger.Error(errMsg)
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"status": "error", "message": errMsg})
	}
	uLogin := new(UserLogin)
	login := ctx.FormValue("login")
	passwordHash := ctx.FormValue("hashed")
	s.Logger.Debug("login: %s, hash: %s ", login, passwordHash)
	// maybe it was not a form but a fetch data post
	if len(strings.Trim(login, " ")) < 1 {
		if err := ctx.Bind(uLogin); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"status": "error", "message": "invalid user login or json format in request body"})
		}
	} else {
		uLogin.Username = login
		uLogin.PasswordHash = passwordHash
	}
	err = f5.ValidateLogin(uLogin.Username)
	if err != nil {
		errMsg := fmt.Sprintf("error validating user login: %v", err)
		s.Logger.Error(errMsg)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"status": "error", "message": errMsg})
	}
	err = f5.ValidatePasswordHash(uLogin.PasswordHash)
	if err != nil {
		errMsg := fmt.Sprintf("error validating password hash: %v", err)
		s.Logger.Error(errMsg)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"status": "error", "message": errMsg})
	}
	s.Logger.Debug("About to check username: %s , password: %s", uLogin.Username, uLogin.PasswordHash)
	requestCtx := ctx.Request().Context()
	if s.server.Authenticator.AuthenticateUser(requestCtx, uLogin.Username, uLogin.PasswordHash) {
		userInfo, err := s.server.Authenticator.GetUserInfoFromLogin(requestCtx, login)
		if err != nil {
			errGetUInfFromLogin := fmt.Sprintf("Error getting user info from login: %v", err)
			s.Logger.Error(errGetUInfFromLogin)
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"status": "error", "message": errGetUInfFromLogin})
		}
		token, err := s.server.JwtCheck.GetTokenFromUserInfo(userInfo)
		if err != nil {
			errGetUInfFromLogin := fmt.Sprintf("Error getting jwt token from user info: %v", err)
			s.Logger.Error(errGetUInfFromLogin)
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"status": "error", "message": errGetUInfFromLogin})
		}
		// Prepare the response
		response := map[string]interface{}{
			"status": "success",
			"token":  token.String(),
		}
		s.Logger.Info("LoginUser(%s) successful login", login)
		return ctx.JSON(http.StatusOK, response)
	} else {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"status": "error", "message": "username not found or password invalid"})
	}
}

func (s Service) GetStatus(ctx echo.Context) error {
	goHttpEcho.TraceHttpRequest("GetStatus", ctx.Request(), s.Logger)
	err := validateHostAllowed(ctx.Request(), s.AllowedHostnames, s.Logger)
	if err != nil {
		errMsg := fmt.Sprintf("error validating host: %v", err)
		s.Logger.Error(errMsg)
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"status": "error", "message": errMsg})
	}
	// get the current user from JWT TOKEN
	claims := s.server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	currentUserLogin := claims.User.Login
	s.Logger.Info("in GetStatus : currentUserId: %d", currentUserId)
	requestCtx := ctx.Request().Context()
	// you can check if the user is not active anymore and RETURN 401 Unauthorized
	if !s.Store.Exist(requestCtx, currentUserLogin) {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"status": "error", "message": "current calling user does not exist"})
	}
	return ctx.JSON(http.StatusOK, map[string]interface{}{"status": "success", "claims": claims})
}

func main() {
	l, err := golog.NewLogger("simple", os.Stdout, golog.DebugLevel, version.APP)
	if err != nil {
		log.Fatalf("💥💥 error log.NewLogger error: %v'\n", err)
	}
	l.Info("🚀🚀 Starting:'%s', v%s, rev:%s, build:%v from: %s", version.APP, version.VERSION, version.REVISION, version.BuildStamp, version.REPOSITORY)

	dbDsn, err := config.GetPgDbDsnUrl(defaultDBIp, defaultDBPort, tools.ToSnakeCase(version.APP), version.AppSnake, defaultDBSslMode)
	if err != nil {
		l.Fatal("💥💥 error getting database DSN: %v", err)
	}
	dbConnCtx, dbConnCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbConnCancel()
	db, err := database.GetInstance(dbConnCtx, "pgx", dbDsn, runtime.NumCPU(), l)
	if err != nil {
		l.Fatal("💥💥 error doing database.GetInstance(pgx ...) error: %v", err)
	}
	defer db.Close()

	dbVersion, err := db.GetVersion(context.Background())
	if err != nil {
		l.Fatal("💥💥 error doing dbConn.GetVersion() error: %v", err)
	}
	l.Info("connected to db version : %s", dbVersion)

	// checking metadata information
	metadataService := metadata.Service{Log: l, Db: db}
	metadataService.CreateMetadataTableOrFail(context.Background())
	found, ver := metadataService.GetServiceVersionOrFail(context.Background(), version.APP)
	if found {
		l.Info("service %s was found in metadata with version: %s", version.APP, ver)
	} else {
		l.Info("service %s was not found in metadata", version.APP)
	}
	metadataService.SetServiceVersionOrFail(context.Background(), version.APP, version.VERSION)

	// Get the ENV JWT_AUTH_URL value
	jwtAuthUrl, err := config.GetJwtAuthUrl()
	if err != nil {
		l.Fatal("💥💥 error getting JWT auth URL: %v", err)
	}
	jwtStatusUrl := config.GetJwtStatusUrl(defaultJwtStatusUrl)

	myVersionReader := goHttpEcho.NewSimpleVersionReader(
		version.APP,
		version.VERSION,
		version.REPOSITORY,
		version.REVISION,
		version.BuildStamp,
		jwtAuthUrl,
		jwtStatusUrl,
	)

	// Create a new JWT checker using factory function
	myJwt, err := goHttpEcho.GetNewJwtCheckerFromConfig(version.APP, 60, l)
	if err != nil {
		l.Fatal("💥💥 error creating JWT checker: %v", err)
	}

	allowedHosts, err := config.GetAllowedHosts()
	if err != nil {
		l.Fatal("💥💥 error getting allowed hosts: %v", err)
	}
	myF5Store := f5.GetStorageInstanceOrPanic("pgx", db, l)

	// Get admin config
	adminId, err := config.GetAdminId(defaultAdminId)
	if err != nil {
		l.Fatal("💥💥 error getting admin ID: %v", err)
	}
	adminExternalId, err := config.GetAdminExternalId(99999)
	if err != nil {
		l.Fatal("💥💥 error getting admin external ID: %v", err)
	}
	adminEmail, err := config.GetAdminEmail(defaultAdminEmail)
	if err != nil {
		l.Fatal("💥💥 error getting admin email: %v", err)
	}
	adminUser, err := config.GetAdminUser(defaultAdminUser)
	if err != nil {
		l.Fatal("💥💥 error getting admin user: %v", err)
	}
	adminPassword, err := config.GetAdminPassword()
	if err != nil {
		l.Fatal("💥💥 error getting admin password: %v", err)
	}

	// Create a new Authenticator with F5
	myAuthenticator := f5.NewF5Authenticator(
		&goHttpEcho.UserInfo{
			UserId:     adminId,
			ExternalId: adminExternalId,
			Name:       "NewSimpleAdminAuthenticator_Admin",
			Email:      adminEmail,
			Login:      adminUser,
			IsAdmin:    false,
			Groups:     []int{1}, // this is the group id of the global_admin group
		},
		adminPassword,
		myJwt,
		myF5Store,
	)

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
			RestrictedUrl: defaultRestrictedUrlBasePath,
		},
	)
	if err != nil {
		l.Fatal("💥💥 error creating server: %v", err)
	}
	cookieNameForJWT := config.GetJwtCookieName(defaultJwtCookieName)
	myF5Service := Service{
		AllowedHostnames: allowedHosts,
		Logger:           l,
		Store:            myF5Store,
		dbConn:           db,
		server:           server,
		auth:             myAuthenticator,
		jwtCookieName:    cookieNameForJWT,
	}

	e := server.GetEcho()
	e.Use(goHttpEcho.CookieToHeaderMiddleware(myF5Service.jwtCookieName, l))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"https://golux.lausanne.ch", "http://localhost:3000"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))
	e.GET("/readiness", server.GetReadinessHandler(func(info string) bool {
		ver, err := db.GetVersion(context.Background())
		if err != nil {
			l.Error("Error getting db version : %v", err)
			return false
		}
		l.Info("Connected to DB version : %s", ver)
		return true
	}, "Connection to DB"))
	e.GET("/health", server.GetHealthHandler(func(info string) bool {
		// you decide what makes you ready, may be it is the connection to the database
		getVersion, err := db.GetVersion(context.Background())
		if err != nil {
			l.Error("Error getting db version : %v", err)
			return false
		}
		l.Info("%s DB version : %s", info, getVersion)
		return true
	}, "Connection to DB"))

	e.GET("/goAppInfo", server.GetAppInfoHandler())
	e.POST(jwtAuthUrl, myF5Service.login)
	//curl -v -H "UserId: YOUR_F5_USER" -c cookies.txt http://localhost:8787/goLogin
	//curl -v -b cookies.txt http://localhost:8787/goapi/v1/status|jq
	// or if you have a token stored in $TOKEN
	//curl -v -b "yourOwnJwtCookieName=${TOKEN}"  http://localhost:8787/goapi/v1/status
	e.GET(jwtAuthUrl, myF5Service.getJwtCookieFromF5)
	r := server.GetRestrictedGroup()
	r.GET(jwtStatusUrl, myF5Service.GetStatus)

	err = server.StartServer()
	if err != nil {
		l.Fatal("💥💥 error doing server.StartServer error: %v'\n", err)
	}
}
