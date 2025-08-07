package goHttpEcho

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

const (
	ReadinessOKMsg  = "(%s) is ready"
	ReadinessErrMsg = "(%s) is not ready"
	HealthOKMsg     = "(%s) is healthy"
	HealthErrMsg    = "(%s) is not healthy"
)

type StandardResponse struct {
	Status string      `json:"status"`
	Msg    string      `json:"msg"`
	IsOk   bool        `json:"isOk"`
	Data   interface{} `json:"data,omitempty"`   // Optional: for including response data
	Errors []string    `json:"errors,omitempty"` // Optional: for detailed error messages
}

func GetStandardResponse(statusMsg, msg string, state bool, data interface{}, errors []string) StandardResponse {
	return StandardResponse{
		Status: statusMsg,
		Msg:    msg,
		IsOk:   state,
		Data:   data,
		Errors: errors,
	}
}

func (s *Server) SendJSONResponse(ctx echo.Context, statusCode int, status, msg string, isOk bool, data interface{}, errors ...string) error {
	response := GetStandardResponse(status, msg, isOk, data, errors)
	if statusCode >= 400 {
		return echo.NewHTTPError(statusCode, response)
	}
	return ctx.JSON(statusCode, response)
}

func (s *Server) GetReadinessHandler(readyFunc FuncAreWeReady, msg string) echo.HandlerFunc {
	handlerName := "GetReadinessHandler"
	s.logger.Debug(initCallMsg, handlerName)
	return func(ctx echo.Context) error {
		ready := readyFunc(msg)
		s.logger.TraceHttpRequest(handlerName, ctx.Request())
		if ready {
			msgOK := fmt.Sprintf(ReadinessOKMsg, msg)
			return s.SendJSONResponse(ctx, http.StatusOK, "ready", msgOK, ready, nil)
		} else {
			msgErr := fmt.Sprintf(ReadinessErrMsg, msg)
			return s.SendJSONResponse(ctx, http.StatusServiceUnavailable, "error", msgErr, ready, nil)
		}
	}
}
func (s *Server) GetHealthHandler(healthyFunc FuncAreWeHealthy, msg string) echo.HandlerFunc {
	handlerName := "GetHealthHandler"
	s.logger.Debug(initCallMsg, handlerName)
	return func(ctx echo.Context) error {
		healthy := healthyFunc(msg)
		s.logger.TraceHttpRequest(handlerName, ctx.Request())
		if healthy {
			msgOK := fmt.Sprintf(HealthOKMsg, msg)
			return s.SendJSONResponse(ctx, http.StatusOK, "healthy", msgOK, healthy, nil)
		} else {
			msgErr := fmt.Sprintf(HealthErrMsg, msg)
			return s.SendJSONResponse(ctx, http.StatusServiceUnavailable, "error", msgErr, healthy, nil)
		}
	}
}

func (s *Server) GetAppInfoHandler() echo.HandlerFunc {
	handlerName := "GetAppInfoHandler"
	s.logger.Debug(initCallMsg, handlerName)
	return func(ctx echo.Context) error {
		s.logger.TraceHttpRequest(handlerName, ctx.Request())
		return ctx.JSON(http.StatusOK, s.VersionReader.GetAppInfo())
	}
}
