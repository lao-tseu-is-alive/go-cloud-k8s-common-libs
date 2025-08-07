package goHttpEcho

import (
	"github.com/labstack/echo/v4"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
)

func CookieToHeaderMiddleware(cookieName string, l golog.MyLogger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				if cookie, err := c.Cookie(cookieName); err == nil && cookie != nil {
					// If the cookie exists, create the Bearer token header.
					c.Request().Header.Set("Authorization", "Bearer "+cookie.Value)
				} else {
					l.Warn("empty authorization header and did not receive a cookie")
				}
			}
			// If the Authorization header is already present, do nothing.
			return next(c)
		}
	}
}
