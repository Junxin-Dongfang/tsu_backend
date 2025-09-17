package middleware

import (
	"time"

	"tsu-self/internal/pkg/log"

	"github.com/labstack/echo/v4"
)

// LoggingMiddleware 日志中间件
func LoggingMiddleware(logger log.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// 记录请求开始
			logger.InfoContext(c.Request().Context(), "请求开始",
				log.String("method", c.Request().Method),
				log.String("path", c.Request().URL.Path),
				log.String("query", c.Request().URL.RawQuery),
				log.String("user_agent", c.Request().UserAgent()),
				log.String("client_ip", c.RealIP()),
				log.String("trace_id", c.Get("trace_id").(string)),
				log.String("request_id", c.Get("request_id").(string)),
			)

			// 执行下一个处理器
			err := next(c)

			// 计算处理时间
			duration := time.Since(start)
			statusCode := c.Response().Status

			// 记录请求完成
			if err != nil {
				logger.ErrorContext(c.Request().Context(), "请求处理出错",
					log.String("method", c.Request().Method),
					log.String("path", c.Request().URL.Path),
					log.Int("status_code", statusCode),
					log.Duration("duration", duration.Milliseconds()),
					log.Any("error", err),
				)
			} else {
				logLevel := "info"
				if statusCode >= 500 {
					logLevel = "error"
				} else if statusCode >= 400 {
					logLevel = "warn"
				}

				switch logLevel {
				case "error":
					logger.ErrorContext(c.Request().Context(), "请求完成（服务器错误）",
						log.String("method", c.Request().Method),
						log.String("path", c.Request().URL.Path),
						log.Int("status_code", statusCode),
						log.Duration("duration", duration.Milliseconds()),
						log.Int64("response_size", c.Response().Size),
					)
				case "warn":
					logger.WarnContext(c.Request().Context(), "请求完成（客户端错误）",
						log.String("method", c.Request().Method),
						log.String("path", c.Request().URL.Path),
						log.Int("status_code", statusCode),
						log.Duration("duration", duration.Milliseconds()),
						log.Int64("response_size", c.Response().Size),
					)
				default:
					logger.InfoContext(c.Request().Context(), "请求完成",
						log.String("method", c.Request().Method),
						log.String("path", c.Request().URL.Path),
						log.Int("status_code", statusCode),
						log.Duration("duration", duration.Milliseconds()),
						log.Int64("response_size", c.Response().Size),
					)
				}
			}

			return err
		}
	}
}
