package validator

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// CustomValidator wraps go-playground validator for Echo
type CustomValidator struct {
	validator *validator.Validate
}

// Validate implements echo.Validator interface
// 直接返回原始验证错误,让 Handler 决定如何处理
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

// New creates a new custom validator instance
func New() echo.Validator {
	return &CustomValidator{
		validator: validator.New(),
	}
}
