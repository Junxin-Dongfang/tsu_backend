// File: internal/app/login-shim/middleware/middleware.go
package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"tsu-self/internal/pkg/contextkeys"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/response"
)

// Config 中间件配置
type Config struct {
	Logger     log.Logger
	RespWriter response.Writer
	Skipper    middleware.Skipper
}

// TraceMiddleware 链路追踪中间件
func TraceMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 生成或获取 trace ID
			traceID := c.Request().Header.Get("X-Trace-ID")
			if traceID == "" {
				traceID = uuid.New().String()
			}

			requestID := uuid.New().String()

			// 设置到 Echo context
			c.Set("trace_id", traceID)
			c.Set("request_id", requestID)

			// 设置响应头
			c.Response().Header().Set("X-Trace-ID", traceID)
			c.Response().Header().Set("X-Request-ID", requestID)

			// 更新 request context
			ctx := c.Request().Context()
			ctx = contextkeys.WithTraceID(ctx, traceID)
			ctx = contextkeys.WithRequestID(ctx, requestID)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
