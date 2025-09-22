package middleware

import (
	"tsu-self/internal/pkg/xerrors"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware() echo.MiddlewareFunc {
	config := middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStore(100), // 每秒 100 个请求
		IdentifierExtractor: func(c echo.Context) (string, error) {
			// 使用客户端 IP 作为标识符
			return c.RealIP(), nil
		},
		ErrorHandler: func(context echo.Context, err error) error {
			appErr := xerrors.FromCode(xerrors.CodeRateLimitExceeded).
				WithService("echo-middleware", "rate_limiter").
				WithMetadata("client_ip", context.RealIP())

			return appErr
		},
	}

	return middleware.RateLimiterWithConfig(config)
}
