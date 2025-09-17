package middleware

import (
	"fmt"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/xerrors"

	"github.com/labstack/echo/v4"
)

// ErrorMiddleware 统一错误处理中间件
func ErrorMiddleware(respWriter response.Writer, logger log.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err == nil {
				return nil
			}

			ctx := c.Request().Context()

			// 处理不同类型的错误
			switch e := err.(type) {
			case *xerrors.AppError:
				// 业务错误，直接使用 response writer 处理
				return respWriter.WriteError(ctx, c.Response().Writer, e)

			case *echo.HTTPError:
				// Echo HTTP 错误，转换为业务错误
				appErr := convertEchoError(e)
				return respWriter.WriteError(ctx, c.Response().Writer, appErr)

			default:
				// 其他未知错误，包装为系统错误
				appErr := xerrors.NewWithError(
					xerrors.CodeInternalError,
					"系统内部错误",
					err,
				).WithService("echo-middleware", "error_handler")

				logger.ErrorContext(ctx, "未处理的错误",
					log.Any("original_error", err),
					log.String("error_type", fmt.Sprintf("%T", err)),
				)

				return respWriter.WriteError(ctx, c.Response().Writer, appErr)
			}
		}
	}
}

// convertEchoError 将 Echo 错误转换为业务错误
func convertEchoError(echoErr *echo.HTTPError) *xerrors.AppError {
	switch echoErr.Code {
	case 400:
		return xerrors.FromCode(xerrors.CodeInvalidParams).
			WithMetadata("echo_message", fmt.Sprintf("%v", echoErr.Message))
	case 401:
		return xerrors.FromCode(xerrors.CodeAuthenticationFailed).
			WithMetadata("echo_message", fmt.Sprintf("%v", echoErr.Message))
	case 403:
		return xerrors.FromCode(xerrors.CodePermissionDenied).
			WithMetadata("echo_message", fmt.Sprintf("%v", echoErr.Message))
	case 404:
		return xerrors.FromCode(xerrors.CodeResourceNotFound).
			WithMetadata("echo_message", fmt.Sprintf("%v", echoErr.Message))
	case 409:
		return xerrors.FromCode(xerrors.CodeDuplicateResource).
			WithMetadata("echo_message", fmt.Sprintf("%v", echoErr.Message))
	case 429:
		return xerrors.FromCode(xerrors.CodeRateLimitExceeded).
			WithMetadata("echo_message", fmt.Sprintf("%v", echoErr.Message))
	default:
		return xerrors.FromCode(xerrors.CodeInternalError).
			WithMetadata("echo_code", fmt.Sprintf("%d", echoErr.Code)).
			WithMetadata("echo_message", fmt.Sprintf("%v", echoErr.Message))
	}
}
