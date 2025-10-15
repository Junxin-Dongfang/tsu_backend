package security

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// SecurityHeadersConfig 安全头配置
type SecurityHeadersConfig struct {
	XSSProtection         string
	ContentTypeNosniff    string
	XFrameOptions         string
	HSTSMaxAge            int
	ContentSecurityPolicy string
}

// DefaultSecurityHeadersConfig 返回默认的安全头配置
func DefaultSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000, // 1 year
		ContentSecurityPolicy: "default-src 'self'",
	}
}

// SecurityHeadersMiddleware 安全头中间件
func SecurityHeadersMiddleware() echo.MiddlewareFunc {
	return SecurityHeadersMiddlewareWithConfig(DefaultSecurityHeadersConfig())
}

// SecurityHeadersMiddlewareWithConfig 使用自定义配置的安全头中间件
func SecurityHeadersMiddlewareWithConfig(config SecurityHeadersConfig) echo.MiddlewareFunc {
	return middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         config.XSSProtection,
		ContentTypeNosniff:    config.ContentTypeNosniff,
		XFrameOptions:         config.XFrameOptions,
		HSTSMaxAge:            config.HSTSMaxAge,
		ContentSecurityPolicy: config.ContentSecurityPolicy,
	})
}
