// File: internal/pkg/response/echo.go
package response

import (
	"tsu-self/internal/pkg/xerrors"

	"github.com/labstack/echo/v4"
)

// Echo 框架适配器 - 简化 Echo Handler 中的响应处理

// EchoOK Echo 成功响应
func EchoOK[T any](c echo.Context, h Writer, data T) error {
	return h.WriteSuccess(c.Request().Context(), c.Response().Writer, data)
}

// EchoError Echo 错误响应
func EchoError(c echo.Context, h Writer, err error) error {
	return h.WriteError(c.Request().Context(), c.Response().Writer, err)
}

// EchoBadRequest Echo 400 错误响应
func EchoBadRequest(c echo.Context, h Writer, message string) error {
	err := xerrors.NewValidationError("request", message)
	return h.WriteError(c.Request().Context(), c.Response().Writer, err)
}

// EchoUnauthorized Echo 401 错误响应
func EchoUnauthorized(c echo.Context, h Writer, message string) error {
	err := xerrors.NewAuthError(message)
	return h.WriteError(c.Request().Context(), c.Response().Writer, err)
}

// EchoForbidden Echo 403 错误响应
func EchoForbidden(c echo.Context, h Writer, resource, action string) error {
	err := xerrors.NewPermissionError(resource, action)
	return h.WriteError(c.Request().Context(), c.Response().Writer, err)
}

// EchoNotFound Echo 404 错误响应
func EchoNotFound(c echo.Context, h Writer, resource, identifier string) error {
	err := xerrors.NewNotFoundError(resource, identifier)
	return h.WriteError(c.Request().Context(), c.Response().Writer, err)
}

// EchoInternalServerError Echo 500 错误响应
func EchoInternalServerError(c echo.Context, h Writer, message string) error {
	err := xerrors.NewWithError(xerrors.CodeInternalError, message, nil)
	return h.WriteError(c.Request().Context(), c.Response().Writer, err)
}

// EchoJSON Echo 直接返回 JSON 响应(跳过 APIResponse 包装)
func EchoJSON(c echo.Context, h Writer, data any, statusCode int) error {
	return h.WriteJSON(c.Request().Context(), c.Response().Writer, data, statusCode)
}
