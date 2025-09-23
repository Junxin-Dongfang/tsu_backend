package validator

import (
	"regexp"
	"unicode"

	"github.com/go-playground/validator/v10"
)

// RegisterAuthValidators 注册认证相关的自定义验证器
func RegisterAuthValidators(v *validator.Validate) {
	v.RegisterValidation("strong_password", validateStrongPassword)
	v.RegisterValidation("username_format", validateUsernameFormat)
}

// validateStrongPassword 验证强密码
// 要求：至少8位，包含大小写字母、数字和特殊字符
func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}

// validateUsernameFormat 验证用户名格式
// 要求：3-30位，只能包含字母、数字和下划线，不能以数字开头
func validateUsernameFormat(fl validator.FieldLevel) bool {
	username := fl.Field().String()

	// 检查长度
	if len(username) < 3 || len(username) > 30 {
		return false
	}

	// 检查格式：只能包含字母、数字和下划线，不能以数字开头
	matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, username)
	return matched
}