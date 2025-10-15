// File: internal/pkg/trace/middleware.go
package trace

import (
	"github.com/labstack/echo/v4"
)

// Middleware Echo 中间件 - 自动提取或生成 TraceID 并存储到 context
func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 从请求头中提取 TraceID，如果没有则生成新的
			traceID := ExtractFromHeader(c.Request().Header)

			// 将 TraceID 存储到 context
			ctx := WithTraceID(c.Request().Context(), traceID)
			c.SetRequest(c.Request().WithContext(ctx))

			// 将 TraceID 添加到响应头（方便客户端追踪）
			c.Response().Header().Set("X-Trace-Id", traceID)

			return next(c)
		}
	}
}
