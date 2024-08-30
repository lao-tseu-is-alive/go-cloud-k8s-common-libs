package goHttpEcho

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

func (s *Server) GetReadinessHandler(readyFunc FuncAreWeReady, msg string) echo.HandlerFunc {
	handlerName := "GetReadinessHandler"
	s.logger.Debug(initCallMsg, handlerName)
	return echo.HandlerFunc(func(ctx echo.Context) error {
		ready := readyFunc(msg)
		r := ctx.Request()
		s.logger.Debug(formatTraceRequest, handlerName, r.Method, r.URL.Path, r.RemoteAddr, msg, ready)
		if ready {
			msgOK := fmt.Sprintf("GetReadinessHandler: (%s) is ready: %#v ", msg, ready)
			return ctx.JSON(http.StatusOK, msgOK)
		} else {
			msgErr := fmt.Sprintf("GetReadinessHandler: (%s) is not ready: %#v ", msg, ready)
			return echo.NewHTTPError(http.StatusInternalServerError, msgErr)
		}
	})
}
func (s *Server) GetHealthHandler(healthyFunc FuncAreWeHealthy, msg string) echo.HandlerFunc {
	handlerName := "GetHealthHandler"
	s.logger.Debug(initCallMsg, handlerName)
	return echo.HandlerFunc(func(ctx echo.Context) error {
		healthy := healthyFunc(msg)
		r := ctx.Request()
		s.logger.Debug(formatTraceRequest, handlerName, r.Method, r.URL.Path, r.RemoteAddr, msg, healthy)
		if healthy {
			msgOK := fmt.Sprintf("GetHealthHandler: (%s) is healthy: %#v ", msg, healthy)
			return ctx.JSON(http.StatusOK, msgOK)
		} else {
			msgErr := fmt.Sprintf("GetHealthHandler: (%s) is not healthy: %#v ", msg, healthy)
			return echo.NewHTTPError(http.StatusInternalServerError, msgErr)
		}
	})
}
