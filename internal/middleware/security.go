package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// CORSMiddleware CORS 中间件
func CORSMiddleware() echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"}, // 生产环境应该限制具体域名
		AllowMethods: []string{
			echo.GET,
			echo.POST,
			echo.PUT,
			echo.DELETE,
			echo.OPTIONS,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			"X-Trace-ID",
			"X-Request-ID",
			"X-Session-Token",
		},
		ExposeHeaders: []string{
			"X-Trace-ID",
			"X-Request-ID",
		},
		AllowCredentials: true,
	})
}

// SecurityMiddleware 安全中间件
func SecurityMiddleware() echo.MiddlewareFunc {
	return middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000, // 1 year
		ContentSecurityPolicy: "default-src 'self'",
	})
}
