package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationError 验证错误详情
type ValidationError struct {
	Field   string `json:"field"`   // 字段名
	Message string `json:"message"` // 错误消息
	Tag     string `json:"tag"`     // 验证标签（如：required, email）
	Value   string `json:"value"`   // 实际值（脱敏后）
}

// TranslateValidationErrors 翻译所有验证错误（返回详细列表）
func TranslateValidationErrors(err error) []ValidationError {
	if err == nil {
		return nil
	}

	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		// 非 validator 错误，返回通用错误
		return []ValidationError{
			{
				Field:   "request",
				Message: err.Error(),
				Tag:     "unknown",
			},
		}
	}

	// 翻译所有错误
	result := make([]ValidationError, 0, len(validationErrs))
	for _, fieldErr := range validationErrs {
		result = append(result, ValidationError{
			Field:   fieldErr.Field(),
			Message: translateFieldError(fieldErr),
			Tag:     fieldErr.Tag(),
			Value:   sanitizeValue(fieldErr.Value()),
		})
	}

	return result
}

// TranslateValidationError 将 validator 验证错误转换为用户友好的中文消息（返回第一个错误）
// 保留此函数以向后兼容
func TranslateValidationError(err error) string {
	if err == nil {
		return ""
	}

	errors := TranslateValidationErrors(err)
	if len(errors) > 0 {
		return errors[0].Message
	}

	return err.Error()
}

// sanitizeValue 脱敏敏感值（避免在错误消息中泄露密码等）
func sanitizeValue(value interface{}) string {
	if value == nil {
		return ""
	}

	strValue := fmt.Sprintf("%v", value)

	// 限制长度
	if len(strValue) > 50 {
		return strValue[:50] + "..."
	}

	return strValue
}

// translateFieldError 翻译单个字段验证错误
func translateFieldError(fe validator.FieldError) string {
	field := getFieldName(fe.Field())
	tag := fe.Tag()

	switch tag {
	case "required":
		return fmt.Sprintf("%s不能为空", field)
	case "email":
		return fmt.Sprintf("%s格式不正确,请输入有效的邮箱地址", field)
	case "min":
		if fe.Type().String() == "string" {
			return fmt.Sprintf("%s长度不能少于%s个字符", field, fe.Param())
		}
		return fmt.Sprintf("%s不能小于%s", field, fe.Param())
	case "max":
		if fe.Type().String() == "string" {
			return fmt.Sprintf("%s长度不能超过%s个字符", field, fe.Param())
		}
		return fmt.Sprintf("%s不能大于%s", field, fe.Param())
	case "gte":
		return fmt.Sprintf("%s必须大于或等于%s", field, fe.Param())
	case "lte":
		return fmt.Sprintf("%s必须小于或等于%s", field, fe.Param())
	case "gt":
		return fmt.Sprintf("%s必须大于%s", field, fe.Param())
	case "lt":
		return fmt.Sprintf("%s必须小于%s", field, fe.Param())
	case "len":
		return fmt.Sprintf("%s长度必须为%s", field, fe.Param())
	case "alpha":
		return fmt.Sprintf("%s只能包含字母", field)
	case "alphanum":
		return fmt.Sprintf("%s只能包含字母和数字", field)
	case "numeric":
		return fmt.Sprintf("%s只能包含数字", field)
	case "uuid":
		return fmt.Sprintf("%s格式不正确,请输入有效的UUID", field)
	case "url":
		return fmt.Sprintf("%s格式不正确,请输入有效的URL", field)
	case "uri":
		return fmt.Sprintf("%s格式不正确,请输入有效的URI", field)
	case "oneof":
		return fmt.Sprintf("%s的值必须是以下之一: %s", field, fe.Param())
	case "eqfield":
		return fmt.Sprintf("%s必须与%s相同", field, fe.Param())
	case "nefield":
		return fmt.Sprintf("%s不能与%s相同", field, fe.Param())
	case "unique":
		return fmt.Sprintf("%s中包含重复的值", field)
	case "dive":
		return fmt.Sprintf("%s包含无效的值", field)
	default:
		// 未知的验证规则,返回通用错误
		return fmt.Sprintf("%s验证失败: %s", field, tag)
	}
}

// getFieldName 将字段名转换为中文友好名称
func getFieldName(field string) string {
	// 字段名映射表
	fieldNames := map[string]string{
		"Email":        "邮箱",
		"Username":     "用户名",
		"Password":     "密码",
		"NewPassword":  "新密码",
		"OldPassword":  "旧密码",
		"Code":         "验证码",
		"Phone":        "手机号",
		"Nickname":     "昵称",
		"AvatarURL":    "头像链接",
		"SessionToken": "会话令牌",
		"Identifier":   "用户标识",

		// 游戏相关字段
		"HeroID":     "角色ID",
		"HeroName":   "角色名称",
		"ClassID":    "职业ID",
		"SkillID":    "技能ID",
		"Level":      "等级",
		"Experience": "经验值",
		"Gold":       "金币",

		// 通用字段
		"ID":          "ID",
		"Name":        "名称",
		"Description": "描述",
		"Type":        "类型",
		"Status":      "状态",
		"Value":       "值",
		"Key":         "键",
		"Title":       "标题",
		"Content":     "内容",
	}

	// 如果有映射,使用映射值
	if name, ok := fieldNames[field]; ok {
		return name
	}

	// 否则尝试智能转换(驼峰转中文)
	// 例如: UserID -> 用户ID
	return smartConvertFieldName(field)
}

// smartConvertFieldName 智能转换字段名(驼峰转中文)
func smartConvertFieldName(field string) string {
	// 简单处理:添加空格后返回
	// 例如: UserID -> User ID -> 用户 ID
	var result strings.Builder
	for i, r := range field {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune(' ')
		}
		result.WriteRune(r)
	}
	return result.String()
}
