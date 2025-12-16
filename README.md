# go-cloud-k8s-common-libs
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=lao-tseu-is-alive_go-cloud-k8s-common-libs&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=lao-tseu-is-alive_go-cloud-k8s-common-libs)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=lao-tseu-is-alive_go-cloud-k8s-common-libs&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=lao-tseu-is-alive_go-cloud-k8s-common-libs)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=lao-tseu-is-alive_go-cloud-k8s-common-libs&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=lao-tseu-is-alive_go-cloud-k8s-common-libs)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=lao-tseu-is-alive_go-cloud-k8s-common-libs&metric=reliability_rating)](https://sonarcloud.io/summary/new_code?id=lao-tseu-is-alive_go-cloud-k8s-common-libs)
[![Go-Test](https://github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/actions/workflows/go.yml/badge.svg)](https://github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/lao-tseu-is-alive/go-cloud-k8s-common-libs/graph/badge.svg?token=8EV2YLEIBA)](https://codecov.io/gh/lao-tseu-is-alive/go-cloud-k8s-common-libs)

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
