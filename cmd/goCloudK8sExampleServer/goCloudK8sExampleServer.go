package main

import (
	"embed"
	"fmt"
	"github.com/cristalhq/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/config"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/goserver"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/tools"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/version"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

const (
	defaultPort            = 8080
	defaultDBPort          = 5432
	defaultDBIp            = "127.0.0.1"
	defaultDBSslMode       = "prefer"
	defaultWebRootDir      = "goCloudK8sExampleFront/dist/"
	defaultUsername        = "bill"
	defaultFakeStupidPass  = "board"
	charsetUTF8            = "charset=UTF-8"
	MIMEAppJSON            = "application/json"
	MIMEHtml               = "text/html"
	MIMEAppJSONCharsetUTF8 = MIMEAppJSON + "; " + charsetUTF8
	MIMEHtmlCharsetUTF8    = MIMEHtml + "; " + charsetUTF8
	HeaderContentType      = "Content-Type"
)

// content holds our static web server content.
//
//go:embed goCloudK8sExampleFront/dist/*
var content embed.FS

type Service struct {
	Log *log.Logger
	//Store       Storage
	dbConn      database.DB
	JwtSecret   []byte
	JwtDuration int
}

// login is just a trivial stupid example to test this server
// you should use the jwt token returned from LoginUser  in github.com/lao-tseu-is-alive/go-cloud-k8s-user-group'
// and share the same secret with the above component
func (s Service) login(ctx echo.Context) error {
	s.Log.Printf("TRACE: entering login() \n##request: %+v \n", ctx.Request())
	username := ctx.FormValue("login")
	fakePassword := ctx.FormValue("pass")

	// Throws unauthorized error
	if username != defaultUsername || fakePassword != defaultFakeStupidPass {
		return ctx.JSON(http.StatusUnauthorized, "username not found or password invalid")
	}

	// Set custom claims
	claims := &goserver.JwtCustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "",
			Audience:  nil,
			Issuer:    "",
			Subject:   "",
			ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(time.Minute * time.Duration(s.JwtDuration))},
			IssuedAt:  &jwt.NumericDate{Time: time.Now()},
			NotBefore: nil,
		},
		Id:       999,
		Name:     "Bill Whatever",
		Email:    "bill@whatever.com",
		Username: defaultUsername,
		IsAdmin:  false,
	}

	// Create token with claims
	signer, _ := jwt.NewSignerHS(jwt.HS512, s.JwtSecret)
	builder := jwt.NewBuilder(signer)
	token, err := builder.Build(claims)
	if err != nil {
		return err
	}
	msg := fmt.Sprintf("LoginUser(%s) succesfull login for user id (%d)", claims.Username, claims.Id)
	s.Log.Printf(msg)
	return ctx.JSON(http.StatusOK, echo.Map{
		"token": token.String(),
	})
}

func (s Service) restricted(ctx echo.Context) error {
	s.Log.Println("TRACE: entering restricted() ")
	// get the current user from JWT TOKEN
	u := ctx.Get("jwtdata").(*jwt.Token)
	claims := goserver.JwtCustomClaims{}
	err := u.DecodeClaims(&claims)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	//callerUserId := claims.Id
	// you can check if the user is not active anymore and RETURN 401 Unauthorized
	//if !s.Store.IsUserActive(currentUserId) {
	//	return echo.NewHTTPError(http.StatusUnauthorized, "current calling user is not active anymore")
	//}
	return ctx.JSON(http.StatusCreated, claims)
}

func checkReady(info string) bool {
	// you decide what makes you ready, may be it is the connection to the database
	//if !connectedToDB {
	//	return false
	//}
	return true
}

func checkHealthy(info string) bool {
	// you decide what makes you ready, may be it is the connection to the database
	//if !stillConnectedToDB {
	//	return false
	//}
	return true
}

func main() {
	l := log.New(os.Stdout, fmt.Sprintf("%s ", version.APP), log.Ldate|log.Ltime|log.Lshortfile)
	l.Printf("INFO: 'Starting %s v:%s  rev:%s  build: %s'", version.APP, version.VERSION, version.REVISION, version.BuildStamp)
	l.Printf("INFO: 'Repository url: https://%s'", version.REPOSITORY)
	secret, err := config.GetJwtSecretFromEnv()
	if err != nil {
		l.Fatalf("💥💥 ERROR: 'in NewGoHttpServer config.GetJwtSecretFromEnv() got error: %v'\n", err)
	}
	tokenDuration, err := config.GetJwtDurationFromEnv(60)
	if err != nil {
		l.Fatalf("💥💥 ERROR: 'in NewGoHttpServer config.GetJwtDurationFromEnv() got error: %v'\n", err)
	}
	dbDsn, err := config.GetPgDbDsnUrlFromEnv(defaultDBIp, defaultDBPort,
		tools.ToSnakeCase(version.APP), version.AppSnake, defaultDBSslMode)
	if err != nil {
		l.Fatalf("💥💥 error doing config.GetPgDbDsnUrlFromEnv. error: %v\n", err)
	}
	dbConn, err := database.GetInstance("pgx", dbDsn, runtime.NumCPU(), l)
	if err != nil {
		l.Fatalf("💥💥 error doing database.GetInstance(pgx, dbDsn  : %v\n", err)
	}
	defer dbConn.Close()

	dbVersion, err := dbConn.GetVersion()
	if err != nil {
		l.Fatalf("💥💥 error doing dbConn.GetVersion() : %v\n", err)
	}
	l.Printf("INFO: connected to DB version : %s", dbVersion)

	yourService := Service{
		Log:         l,
		dbConn:      dbConn,
		JwtSecret:   []byte(secret),
		JwtDuration: tokenDuration,
	}

	listenAddr, err := config.GetPortFromEnv(defaultPort)
	if err != nil {
		l.Fatalf("💥💥 ERROR: 'calling GetPortFromEnv got error: %v'\n", err)
	}
	l.Printf("INFO: 'Will start HTTP server listening on port %s'", listenAddr)
	server := goserver.NewGoHttpServer(listenAddr, l, defaultWebRootDir, content, "/api")
	e := server.GetEcho()

	e.GET("/readiness", server.GetReadinessHandler(checkReady, "Connection to DB"))
	e.GET("/health", server.GetHealthHandler(checkHealthy, "Connection to DB"))
	// Login route
	e.POST("/login", yourService.login)
	r := server.GetRestrictedGroup()
	// now with restricted group reference you can here the routes defined in OpenApi users.yaml are registered
	// yourModelEntityFromOpenApi.RegisterHandlers(r, &yourModelService)
	r.GET("/secret", yourService.restricted)
	loginExample := fmt.Sprintf("curl -v -X POST -d 'login=%s' -d 'pass=%s' http://localhost%s/login", defaultUsername, defaultFakeStupidPass, listenAddr)
	getSecretExample := fmt.Sprintf(" curl -v  -H \"Authorization: Bearer ${TOKEN}\" http://localhost%s/api/secret |jq\n", listenAddr)
	l.Printf("INFO: from another terminal just try :\n %s", loginExample)
	l.Printf("INFO: then type export TOKEN=your_token_above_goes_here   \n %s", getSecretExample)

	err = server.StartServer()
	if err != nil {
		l.Fatalf("💥💥 ERROR: 'calling echo.Start(%s) got error: %v'\n", listenAddr, err)
	}

}
