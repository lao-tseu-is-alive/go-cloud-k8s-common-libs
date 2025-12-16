package goHttpEcho

import (
	"log/slog"

	"github.com/labstack/echo/v4"
)

func CookieToHeaderMiddleware(cookieName string, l *slog.Logger) echo.MiddlewareFunc {
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
