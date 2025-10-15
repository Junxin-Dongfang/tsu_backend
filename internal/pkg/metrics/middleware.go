// File: internal/pkg/metrics/middleware.go
package metrics

import (
	"net/http"

	"tsu-self/internal/pkg/ctxkey"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Middleware Echo 中间件 - 将HTTP方法存储到 context 中（用于 Prometheus 指标）
func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 将 HTTP 方法存储到 context
			ctx := c.Request().Context()
			ctx = ctxkey.WithValue(ctx, ctxkey.HTTPMethod, c.Request().Method)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

// Handler 返回 Prometheus metrics HTTP 处理器
// 用于暴露 /metrics 端点
func Handler() http.Handler {
	return promhttp.Handler()
}

// EchoHandler Echo 框架的 Prometheus metrics 处理器
func EchoHandler() echo.HandlerFunc {
	h := promhttp.Handler()
	return func(c echo.Context) error {
		h.ServeHTTP(c.Response().Writer, c.Request())
		return nil
	}
}
