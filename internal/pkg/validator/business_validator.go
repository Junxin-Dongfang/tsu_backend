package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
)

// BusinessValidator 业务规则验证器
type BusinessValidator struct {
	validator *validator.Validate
}

// NewBusinessValidator 创建新的业务验证器
func NewBusinessValidator() *BusinessValidator {
	v := validator.New()

	// 注册自定义验证规则
	v.RegisterValidation("class_code", validateClassCode)
	v.RegisterValidation("attribute_code", validateAttributeCode)
	v.RegisterValidation("chinese_name", validateChineseName)
	v.RegisterValidation("color_hex", validateColorHex)
	v.RegisterValidation("safe_description", validateSafeDescription)
	v.RegisterValidation("positive_number", validatePositiveNumber)
	v.RegisterValidation("tier_range", validateTierRange)
	v.RegisterValidation("display_order", validateDisplayOrder)

	return &BusinessValidator{
		validator: v,
	}
}

// Validate 验证结构体
func (bv *BusinessValidator) Validate(i interface{}) error {
	return bv.validator.Struct(i)
}

// validateClassCode 验证职业代码格式
func validateClassCode(fl validator.FieldLevel) bool {
	code := fl.Field().String()

	// 职业代码规则：
	// 1. 长度 2-32 字符
	// 2. 只能包含大写字母、数字和下划线
	// 3. 必须以字母开头
	// 4. 不能以下划线结尾
	if len(code) < 2 || len(code) > 32 {
		return false
	}

	matched, _ := regexp.MatchString(`^[A-Z][A-Z0-9_]*[A-Z0-9]$|^[A-Z]$`, code)
	return matched
}

// validateAttributeCode 验证属性代码格式
func validateAttributeCode(fl validator.FieldLevel) bool {
	code := fl.Field().String()

	// 属性代码规则：
	// 1. 长度 2-32 字符
	// 2. 只能包含大写字母、数字和下划线
	// 3. 必须以字母开头
	if len(code) < 2 || len(code) > 32 {
		return false
	}

	matched, _ := regexp.MatchString(`^[A-Z][A-Z0-9_]*$`, code)
	return matched
}

// validateChineseName 验证中文名称
func validateChineseName(fl validator.FieldLevel) bool {
	name := fl.Field().String()

	// 中文名称规则：
	// 1. 长度 1-64 字符
	// 2. 可以包含中文、英文、数字、空格、括号、下划线、连字符
	// 3. 不能只包含空格
	if len(strings.TrimSpace(name)) == 0 || utf8.RuneCountInString(name) > 64 {
		return false
	}

	// 检查是否包含危险字符
	dangerousChars := []string{"<", ">", "\"", "'", "&", "script", "javascript", "vbscript"}
	lowerName := strings.ToLower(name)
	for _, char := range dangerousChars {
		if strings.Contains(lowerName, char) {
			return false
		}
	}

	return true
}

// validateColorHex 验证十六进制颜色值
func validateColorHex(fl validator.FieldLevel) bool {
	color := fl.Field().String()

	// 颜色值规则：
	// 1. 必须以 # 开头
	// 2. 后跟 6 位十六进制数字
	matched, _ := regexp.MatchString(`^#[0-9A-Fa-f]{6}$`, color)
	return matched
}

// validateSafeDescription 验证安全的描述文本
func validateSafeDescription(fl validator.FieldLevel) bool {
	desc := fl.Field().String()

	// 描述规则：
	// 1. 长度不超过 1000 字符
	// 2. 不能包含脚本标签和危险内容
	if utf8.RuneCountInString(desc) > 1000 {
		return false
	}

	// 检查危险内容
	dangerousPatterns := []string{
		"<script", "</script>", "javascript:", "vbscript:", "onload=", "onerror=",
		"onclick=", "onmouseover=", "<iframe", "</iframe>", "<object", "</object>",
	}

	lowerDesc := strings.ToLower(desc)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerDesc, pattern) {
			return false
		}
	}

	return true
}

// validatePositiveNumber 验证正数
func validatePositiveNumber(fl validator.FieldLevel) bool {
	switch fl.Field().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fl.Field().Int() > 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fl.Field().Uint() > 0
	case reflect.Float32, reflect.Float64:
		return fl.Field().Float() > 0
	}
	return false
}

// validateTierRange 验证职业等级范围
func validateTierRange(fl validator.FieldLevel) bool {
	tier := fl.Field().Int()
	// 职业等级范围：1-10
	return tier >= 1 && tier <= 10
}

// validateDisplayOrder 验证显示顺序
func validateDisplayOrder(fl validator.FieldLevel) bool {
	order := fl.Field().Int()
	// 显示顺序范围：0-9999
	return order >= 0 && order <= 9999
}

// GetValidationErrorMessage 获取验证错误的友好消息
func GetValidationErrorMessage(err error) string {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			field := fieldError.Field()
			tag := fieldError.Tag()

			switch tag {
			case "required":
				return fmt.Sprintf("字段 %s 是必填项", field)
			case "class_code":
				return fmt.Sprintf("职业代码格式不正确：必须是2-32位大写字母、数字或下划线，以字母开头")
			case "attribute_code":
				return fmt.Sprintf("属性代码格式不正确：必须是2-32位大写字母、数字或下划线，以字母开头")
			case "chinese_name":
				return fmt.Sprintf("名称格式不正确：长度1-64字符，不能包含危险字符")
			case "color_hex":
				return fmt.Sprintf("颜色值格式不正确：必须是#开头的6位十六进制颜色值")
			case "safe_description":
				return fmt.Sprintf("描述内容不安全：长度不超过1000字符，不能包含脚本标签")
			case "positive_number":
				return fmt.Sprintf("字段 %s 必须是正数", field)
			case "tier_range":
				return fmt.Sprintf("职业等级必须在1-10之间")
			case "display_order":
				return fmt.Sprintf("显示顺序必须在0-9999之间")
			case "min":
				return fmt.Sprintf("字段 %s 的值太小", field)
			case "max":
				return fmt.Sprintf("字段 %s 的值太大", field)
			case "email":
				return fmt.Sprintf("邮箱格式不正确")
			case "uuid":
				return fmt.Sprintf("UUID格式不正确")
			default:
				return fmt.Sprintf("字段 %s 验证失败：%s", field, tag)
			}
		}
	}

	return "验证失败：" + err.Error()
}

// ValidateValueRange 验证数值范围的业务逻辑
func ValidateValueRange(minValue, maxValue, defaultValue *float64) error {
	if minValue != nil && maxValue != nil {
		if *minValue >= *maxValue {
			return fmt.Errorf("最小值必须小于最大值")
		}
	}

	if defaultValue != nil {
		if minValue != nil && *defaultValue < *minValue {
			return fmt.Errorf("默认值不能小于最小值")
		}
		if maxValue != nil && *defaultValue > *maxValue {
			return fmt.Errorf("默认值不能大于最大值")
		}
	}

	return nil
}

// ValidateBusinessRules 验证复杂的业务规则
func ValidateBusinessRules(entityType string, data interface{}) error {
	switch entityType {
	case "class":
		return validateClassBusinessRules(data)
	case "attribute_type":
		return validateAttributeTypeBusinessRules(data)
	default:
		return nil
	}
}

// validateClassBusinessRules 验证职业相关的业务规则
func validateClassBusinessRules(data interface{}) error {
	// 这里可以添加职业相关的复杂业务规则
	// 例如：检查职业代码是否与现有职业冲突、验证职业进阶路径等
	return nil
}

// validateAttributeTypeBusinessRules 验证属性类型相关的业务规则
func validateAttributeTypeBusinessRules(data interface{}) error {
	// 这里可以添加属性类型相关的复杂业务规则
	// 例如：检查属性代码唯一性、验证计算公式语法等
	return nil
}
