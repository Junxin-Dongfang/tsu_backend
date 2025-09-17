package middleware

import (
	"fmt"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/xerrors"

	"github.com/labstack/echo/v4"
)

// RecoveryMiddleware 恢复中间件
func RecoveryMiddleware(respWriter response.Writer, logger log.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					ctx := c.Request().Context()

					// 记录 panic 信息
					logger.ErrorContext(ctx, "应用程序 panic",
						log.Any("panic_value", r),
						log.String("path", c.Request().URL.Path),
						log.String("method", c.Request().Method),
					)

					// 创建系统错误
					appErr := xerrors.FromCode(xerrors.CodeInternalError).
						WithService("echo-middleware", "recovery").
						WithMetadata("panic_value", fmt.Sprintf("%v", r))

					// 发送错误响应
					respWriter.WriteError(ctx, c.Response().Writer, appErr)
				}
			}()

			return next(c)
		}
	}
}
