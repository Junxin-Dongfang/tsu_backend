package middleware

import (
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/pkg/response"
)

// UUID 正则表达式
var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// UUIDValidationMiddleware 验证路径参数中的 UUID 格式
// 适用于包含 :id 参数的路由
func UUIDValidationMiddleware(respWriter response.Writer) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 获取所有路径参数
			params := c.ParamNames()

			for _, paramName := range params {
				// 检查参数名是否是 ID 类型（包含 id 后缀或就是 id）
				if paramName == "id" || strings.HasSuffix(paramName, "_id") {
					paramValue := c.Param(paramName)

					// 如果参数非空且不是有效的 UUID，返回 404
					if paramValue != "" && !uuidRegex.MatchString(paramValue) {
						return response.EchoNotFound(c, respWriter, "资源", paramValue)
					}
				}
			}

			return next(c)
		}
	}
}
