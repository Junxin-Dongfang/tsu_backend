// File: internal/pkg/metrics/middleware.go
package metrics

import (
	"net/http"
	"time"

	"tsu-self/internal/pkg/ctxkey"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// pathLimitTracker 全局路径限制追踪器（限制为 100 个不同路径）
var pathLimitTracker = NewPathLimitTracker(100)

// Middleware Echo 中间件 - 自动收集 HTTP 性能指标
func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 获取路由模板（使用 c.Path() 而非 c.Request().URL.Path）
			path := c.Path()

			// 跳过健康检查端点
			if IsHealthCheckEndpoint(path) {
				return next(c)
			}

			// 标签基数控制
			path = pathLimitTracker.TrackPath(path)

			service := GetServiceName()

			// 记录请求开始时间
			start := time.Now()

			// 增加当前进行中的请求数
			DefaultHTTPMetrics.IncInProgress(service)
			defer DefaultHTTPMetrics.DecInProgress(service)

			// 将 HTTP 方法存储到 context（保持兼容性）
			ctx := c.Request().Context()
			ctx = ctxkey.WithValue(ctx, ctxkey.HTTPMethod, c.Request().Method)
			c.SetRequest(c.Request().WithContext(ctx))

			// 执行请求处理
			err := next(c)

			// 记录指标
			duration := time.Since(start)
			method := c.Request().Method
			statusCode := c.Response().Status

			DefaultHTTPMetrics.RecordRequest(service, path, method, statusCode, duration)

			// 如果接近路径限制，记录警告
			if warning := pathLimitTracker.LogWarning(); warning != "" {
				// TODO: 使用日志系统记录警告
				// log.Warn(warning)
			}

			return err
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
