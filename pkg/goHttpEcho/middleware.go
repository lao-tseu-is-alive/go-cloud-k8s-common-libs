package goHttpEcho

import "github.com/labstack/echo/v4"

func CookieToHeaderMiddleware(cookieName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// If the Authorization header is already present, do nothing.
			if c.Request().Header.Get("Authorization") != "" {
				return next(c)
			}

			cookie, err := c.Cookie(cookieName) // Use the provided cookieName
			if err == nil {
				// If the cookie exists, create the Bearer token header.
				bearerToken := "Bearer " + cookie.Value
				c.Request().Header.Set("Authorization", bearerToken)
			}

			return next(c)
		}
	}
}
