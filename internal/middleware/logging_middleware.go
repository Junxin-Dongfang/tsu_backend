package middleware

import (
	"bytes"
	"io"
	"strings"
	"time"

	"tsu-self/internal/pkg/ctxkey"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/trace"

	"github.com/labstack/echo/v4"
)

// LoggingConfig 日志配置
type LoggingConfig struct {
	// SkipPaths 跳过日志记录的路径
	SkipPaths []string

	// DetailedLog 是否记录详细日志（请求头、请求体）
	DetailedLog bool

	// LogRequestBody 是否记录请求体
	LogRequestBody bool

	// LogResponseBody 是否记录响应体（仅开发环境）
	LogResponseBody bool

	// MaxBodySize 最大记录的 body 大小（字节）
	MaxBodySize int64

	// SensitiveHeaders 需要脱敏的 Header
	SensitiveHeaders []string
}

// DefaultLoggingConfig 默认日志配置
func DefaultLoggingConfig() *LoggingConfig {
	return &LoggingConfig{
		SkipPaths: []string{
			"/health",
			"/metrics",
			"/favicon.ico",
		},
		DetailedLog:      false,
		LogRequestBody:   false,
		LogResponseBody:  false,
		MaxBodySize:      10 * 1024, // 10KB
		SensitiveHeaders: []string{
			"Authorization",
			"Cookie",
			"X-Session-Token",
			"X-Api-Key",
		},
	}
}

// LoggingMiddleware 日志中间件
func LoggingMiddleware(logger log.Logger) echo.MiddlewareFunc {
	return LoggingMiddlewareWithConfig(logger, DefaultLoggingConfig())
}

// LoggingMiddlewareWithConfig 带配置的日志中间件
func LoggingMiddlewareWithConfig(logger log.Logger, config *LoggingConfig) echo.MiddlewareFunc {
	if config == nil {
		config = DefaultLoggingConfig()
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 检查是否跳过
			if shouldSkip(c.Request().URL.Path, config.SkipPaths) {
				return next(c)
			}

			start := time.Now()
			ctx := c.Request().Context()

			// 获取 traceID
			traceID := trace.GetTraceID(ctx)

			// 基础日志字段
			baseFields := []any{
				log.String("method", c.Request().Method),
				log.String("path", c.Request().URL.Path),
				log.String("client_ip", c.RealIP()),
				log.String("trace_id", traceID),
			}

			// 详细日志
			if config.DetailedLog {
				if c.Request().URL.RawQuery != "" {
					baseFields = append(baseFields, log.String("query", c.Request().URL.RawQuery))
				}
				baseFields = append(baseFields,
					log.String("user_agent", c.Request().UserAgent()),
					log.String("referer", c.Request().Referer()),
				)

				// 记录请求头（脱敏）
				if headers := sanitizeHeaders(c.Request().Header, config.SensitiveHeaders); len(headers) > 0 {
					baseFields = append(baseFields, log.Any("headers", headers))
				}

				// 记录请求体
				if config.LogRequestBody && c.Request().Body != nil {
					if body := readAndRestoreBody(c, config.MaxBodySize); body != "" {
						baseFields = append(baseFields, log.String("request_body", body))
					}
				}
			}

			// 记录请求开始
			logger.InfoContext(ctx, "请求开始", baseFields...)

			// 执行下一个处理器
			err := next(c)

			// 计算处理时间
			duration := time.Since(start)
			statusCode := c.Response().Status

			// 响应日志字段
			responseFields := []any{
				log.String("method", c.Request().Method),
				log.String("path", c.Request().URL.Path),
				log.Int("status_code", statusCode),
				log.Duration("duration_ms", duration.Milliseconds()),
				log.Int64("response_size", c.Response().Size),
				log.String("trace_id", traceID),
			}

			// 获取用户ID（如果存在）
			if userID := ctxkey.GetString(ctx, ctxkey.UserID); userID != "" {
				responseFields = append(responseFields, log.String("user_id", userID))
			}

			// 记录请求完成
			if err != nil {
				responseFields = append(responseFields, log.Any("error", err))
				logger.ErrorContext(ctx, "请求处理出错", responseFields...)
			} else {
				// 根据状态码选择日志级别
				switch {
				case statusCode >= 500:
					logger.ErrorContext(ctx, "请求完成（服务器错误）", responseFields...)
				case statusCode >= 400:
					logger.WarnContext(ctx, "请求完成（客户端错误）", responseFields...)
				default:
					logger.InfoContext(ctx, "请求完成", responseFields...)
				}
			}

			return err
		}
	}
}

// shouldSkip 检查是否应该跳过日志记录
func shouldSkip(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// sanitizeHeaders 脱敏敏感 Header
func sanitizeHeaders(headers map[string][]string, sensitiveHeaders []string) map[string]string {
	result := make(map[string]string)
	for k, v := range headers {
		if len(v) == 0 {
			continue
		}

		// 检查是否是敏感 header
		isSensitive := false
		for _, sensitive := range sensitiveHeaders {
			if strings.EqualFold(k, sensitive) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			result[k] = "***REDACTED***"
		} else {
			result[k] = v[0]
		}
	}
	return result
}

// readAndRestoreBody 读取并恢复请求体
func readAndRestoreBody(c echo.Context, maxSize int64) string {
	if c.Request().Body == nil {
		return ""
	}

	// 限制读取大小
	limitedReader := io.LimitReader(c.Request().Body, maxSize)
	bodyBytes, err := io.ReadAll(limitedReader)
	if err != nil {
		return ""
	}

	// 恢复 body（让后续处理器可以读取）
	c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	body := string(bodyBytes)
	if len(bodyBytes) >= int(maxSize) {
		body += "... (truncated)"
	}

	return body
}
