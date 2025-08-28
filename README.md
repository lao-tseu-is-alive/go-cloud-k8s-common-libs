# go-cloud-k8s-common-libs
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=lao-tseu-is-alive_go-cloud-k8s-common-libs&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=lao-tseu-is-alive_go-cloud-k8s-common-libs)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=lao-tseu-is-alive_go-cloud-k8s-common-libs&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=lao-tseu-is-alive_go-cloud-k8s-common-libs)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=lao-tseu-is-alive_go-cloud-k8s-common-libs&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=lao-tseu-is-alive_go-cloud-k8s-common-libs)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=lao-tseu-is-alive_go-cloud-k8s-common-libs&metric=reliability_rating)](https://sonarcloud.io/summary/new_code?id=lao-tseu-is-alive_go-cloud-k8s-common-libs)
[![Go-Test](https://github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/actions/workflows/go.yml/badge.svg)](https://github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/actions/workflows/go.yml)


**Common Golang packages for MicroServices in the Goeland team (using the Echo framework)**

This repository provides a set of reusable Go libraries designed for building microservices with the Echo framework. It includes utilities for configuration, database access, JWT authentication, and more, along with an example server (`goCloudK8sExampleServer`) to demonstrate practical usage.

## Table of Contents
- [Overview](#overview)
- [Features](#features)
- [Getting Started](#getting-started)
    - [Prerequisites](#prerequisites)
    - [Installation](#installation)
    - [Running the Example Server](#running-the-example-server)
- [Usage](#usage)
- [Project Structure](#project-structure)
- [Contributing](#contributing)
- [License](#license)
- [Contact](#contact)

## Overview
`go-cloud-k8s-common-libs` aims to streamline microservice development by offering pre-built, tested packages that handle common functionalities like environment configuration, PostgreSQL database interactions, and secure authentication with JWT. Whether you're part of the Goeland team or an external developer, these libraries can save time and ensure consistency across projects.

## Features
- **Configuration Management**: Easily load environment variables for service setup.
- **Database Support**: Simplified PostgreSQL interactions with connection pooling and migration support using `go-migrate`.
- **HTTP Framework**: Built on the Echo framework for fast and scalable web services.
- **JWT Authentication**: Secure endpoints with JSON Web Tokens.
- **Example Server**: A fully functional server (`goCloudK8sExampleServer`) to showcase library usage.
- **CI/CD Integration**: Automated testing with GitHub Actions and quality checks via SonarCloud.

## Getting Started

### Prerequisites
Before you begin, ensure you have the following installed:
- Go (version 1.24 or later)
- PostgreSQL (for database operations and testing)
- Git (to clone the repository)

### Installation
1. Clone the repository to your local machine:
   ```bash
   git clone https://github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs.git
   cd go-cloud-k8s-common-libs
   ```
2. Install the dependencies:
   ```bash
   go mod tidy
   go mod download
   ```

### Running the Example Server
The `goCloudK8sExampleServer` is a practical way to see the libraries in action. Follow these steps to run it locally:
1. Copy the sample environment file and adjust the settings:
   ```bash
   cp .env_sample .env
   ```
   Edit `.env` to set your database credentials, JWT secret, and other configurations as needed.
2. Set up a local PostgreSQL database (or use the provided script):
   ```bash
   ./scripts/createLocalDBAndUser.sh
   ```
3. Run the server using the provided script:
   ```bash
   ./scripts/GoRunWithEnv.sh cmd/goCloudK8sExampleServer/main.go .env
   ```
4. Access the server at `http://localhost:3333` (or the port specified in your `.env` file). You can interact with the frontend or use API endpoints like `/login` for JWT authentication.

## Usage
To use these libraries in your own microservice project, import the desired packages and configure them based on your environment. Here's a quick example of setting up a basic Echo server with JWT authentication:

here is the code for minimal server example with jwt authentication in cmd/minServer directory :

```go
package main
import (
	"embed"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/config"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/goHttpEcho"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/version"
	"log"
	"net/http"
	"strings"
)
const (
	APP                 = "goCloudK8sCommonLibsDemoServer"
	defaultPort         = 8080
	defaultJwtStatusUrl = "/status"
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
	//Store       Storage
	//dbConn database.DB
	server *goHttpEcho.Server
}

// login is just a trivial stupid example to test this server
// you should use the jwt token returned from LoginUser  in github.com/lao-tseu-is-alive/go-cloud-k8s-user-group'
// and share the same secret with the above component
func (s Service) login(ctx echo.Context) error {
	s.Logger.TraceHttpRequest("login", ctx.Request())
	login := ctx.FormValue("login")
	passwordHash := ctx.FormValue("hashed")
	s.Logger.Debug("login: %s, hash: %s ", login, passwordHash)
	// maybe it was not a form but a fetch data post
	if len(strings.Trim(login, " ")) < 1 {
		return ctx.JSON(http.StatusUnauthorized, "invalid credentials")
	}

	if s.server.Authenticator.AuthenticateUser(login, passwordHash) {
		userInfo, err := s.server.Authenticator.GetUserInfoFromLogin(login)
		if err != nil {
			errGetUInfFromLogin := fmt.Sprintf("Error getting user info from login: %v", err)
			s.Logger.Error(errGetUInfFromLogin)
			return ctx.JSON(http.StatusInternalServerError, errGetUInfFromLogin)
		}
		token, err := s.server.JwtCheck.GetTokenFromUserInfo(userInfo)
		if err != nil {
			errGetUInfFromLogin := fmt.Sprintf("Error getting jwt token from user info: %v", err)
			s.Logger.Error(errGetUInfFromLogin)
			return ctx.JSON(http.StatusInternalServerError, errGetUInfFromLogin)
		}
		// Prepare the response
		response := map[string]string{
			"token": token.String(),
		}
		s.Logger.Info("LoginUser(%s) successful login", login)
		return ctx.JSON(http.StatusOK, response)
	} else {
		return ctx.JSON(http.StatusUnauthorized, "username not found or password invalid")
	}
}

func (s Service) restricted(ctx echo.Context) error {
	s.Logger.TraceHttpRequest("restricted", ctx.Request())
	// get the current user from JWT TOKEN
	claims := s.server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	s.Logger.Info("in restricted : currentUserId: %d", currentUserId)
	return ctx.JSON(http.StatusCreated, claims)
}

func main() {
	l, err := golog.NewLogger("zap", golog.DebugLevel, APP)
	if err != nil {
		log.Fatalf("💥💥 error log.NewLogger error: %v'\n", err)
	}
	l.Info("🚀🚀 Starting:'%s', v%s, rev:%s, build:%v from: %s", APP, version.VERSION, version.REVISION, version.BuildStamp, version.REPOSITORY)

	// Get the ENV JWT_AUTH_URL value
	jwtAuthUrl := config.GetJwtAuthUrlFromEnvOrPanic()
	jwtStatusUrl := config.GetJwtStatusUrlFromEnv(defaultJwtStatusUrl)

	myVersionReader := goHttpEcho.NewSimpleVersionReader(
		APP,
		version.VERSION,
		version.REPOSITORY,
		version.REVISION,
		version.BuildStamp,
		jwtAuthUrl,
		jwtStatusUrl,
	)
	// Create a new JWT checker
	myJwt := goHttpEcho.NewJwtChecker(
		config.GetJwtSecretFromEnvOrPanic(),
		config.GetJwtIssuerFromEnvOrPanic(),
		APP,
		config.GetJwtContextKeyFromEnvOrPanic(),
		config.GetJwtDurationFromEnvOrPanic(60),
		l)
	// Create a new Authenticator with a simple admin user
	myAuthenticator := goHttpEcho.NewSimpleAdminAuthenticator(&goHttpEcho.UserInfo{
		UserId:     config.GetAdminIdFromEnvOrPanic(defaultAdminId),
		ExternalId: config.GetAdminExternalIdFromEnvOrPanic(9999999),
		Name:       "NewSimpleAdminAuthenticator_Admin",
		Email:      config.GetAdminEmailFromEnvOrPanic(defaultAdminEmail),
		Login:      config.GetAdminUserFromEnvOrPanic(defaultAdminUser),
		IsAdmin:    false,
		Groups:     []int{1}, // this is the group id of the global_admin group
	},

		config.GetAdminPasswordFromEnvOrPanic(),
		myJwt)

	server := goHttpEcho.CreateNewServerFromEnvOrFail(
		defaultPort,
		"0.0.0.0", // defaultServerIp,
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
```

For detailed usage of specific packages like `database` or `jwt`, refer to the inline documentation or example server code in `cmd/goCloudK8sExampleServer`.

## Project Structure
Here's an overview of the key directories and files:
- `pkg/`: Core library packages (e.g., `config`, `database`, `goHttpEcho`).
- `cmd/minServer/`: Minimal Example server implementation with JWT helper.
- `cmd/goCloudK8sExampleServer/`: Example server implementation using databse and JWT the libraries.
- `scripts/`: Utility scripts for development, testing, and deployment.
- `.github/workflows/`: CI/CD configurations for automated testing.
- `db/migrations/`: SQL migration scripts for database schema management.

## Contributing
We welcome contributions from the community! Please read our [CONTRIBUTING.md](CONTRIBUTING.md) (coming soon) for guidelines on coding standards, testing, and submitting pull requests. To get started:
1. Fork the repository.
2. Make your changes in a feature branch.
3. Submit a pull request with a clear description of your changes.

## License
This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

## Contact
For questions, suggestions, or issues, feel free to open a GitHub issue or reach out to the Goeland team via the repository's [discussions](https://github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/discussions).

---
