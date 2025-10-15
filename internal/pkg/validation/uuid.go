// Package validation 提供通用的验证工具和中间件
package validation

import (
	"regexp"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/pkg/response"
)

// UUID 正则表达式
var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// IsValidUUID 检查字符串是否是有效的 UUID
func IsValidUUID(s string) bool {
	return uuidRegex.MatchString(s)
}

// UUIDValidationMiddleware 验证路径参数中的 UUID 格式
// 使用白名单机制，只验证明确应该是 UUID 的参数
func UUIDValidationMiddleware(respWriter response.Writer) echo.MiddlewareFunc {
	// UUID 参数白名单（这些参数必须是 UUID）
	uuidParams := map[string]bool{
		"id":             true, // 通用 ID
		"user_id":        true, // 用户 ID
		"role_id":        true, // 角色 ID
		"class_id":       true, // 职业 ID
		"skill_id":       true, // 技能 ID
		"action_id":      true, // 动作 ID
		"effect_id":      true, // 效果 ID
		"buff_id":        true, // Buff ID
		"tag_id":         true, // 标签 ID
		"bonus_id":       true, // 属性加成 ID
		"entity_id":      true, // 实体 ID
		"permission_id":  true, // 权限 ID
		"category_id":    true, // 类别 ID
		"damage_type_id": true, // 伤害类型 ID
		"flag_id":        true, // Flag ID
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 获取所有路径参数
			params := c.ParamNames()

			for _, paramName := range params {
				// 只验证白名单中的参数
				if uuidParams[paramName] {
					paramValue := c.Param(paramName)

					// 如果参数非空且不是有效的 UUID，返回 404
					if paramValue != "" && !IsValidUUID(paramValue) {
						return response.EchoNotFound(c, respWriter, "资源", paramValue)
					}
				}
			}

			return next(c)
		}
	}
}
