// File: internal/pkg/i18n/middleware.go
package i18n

import (
	"github.com/labstack/echo/v4"
)

// Middleware Echo 中间件 - 从请求头中提取语言偏好并存储到 context
func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 1. 优先从查询参数获取语言 (例如: ?lang=en)
			langCode := c.QueryParam("lang")
			if langCode != "" {
				lang := ParseLanguageCode(langCode)
				ctx := WithLanguage(c.Request().Context(), lang)
				c.SetRequest(c.Request().WithContext(ctx))
				return next(c)
			}

			// 2. 从 Accept-Language 头部获取
			acceptLanguage := c.Request().Header.Get("Accept-Language")
			lang := ParseAcceptLanguage(acceptLanguage)

			// 3. 将语言存储到 context
			ctx := WithLanguage(c.Request().Context(), lang)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
