// Package security 提供通用的安全相关中间件
package security

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// CORSConfig CORS 配置
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
}

// DefaultCORSConfig 返回默认的 CORS 配置
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
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
	}
}

// CORSMiddleware CORS 中间件
func CORSMiddleware() echo.MiddlewareFunc {
	return CORSMiddlewareWithConfig(DefaultCORSConfig())
}

// CORSMiddlewareWithConfig 使用自定义配置的 CORS 中间件
func CORSMiddlewareWithConfig(config CORSConfig) echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     config.AllowOrigins,
		AllowMethods:     config.AllowMethods,
		AllowHeaders:     config.AllowHeaders,
		ExposeHeaders:    config.ExposeHeaders,
		AllowCredentials: config.AllowCredentials,
	})
}
