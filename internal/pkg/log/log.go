// File: internal/pkg/log/log.go
package log

import (
	"context"
	"log/slog"
	"os"

	"tsu-self/internal/pkg/xerrors"
)

// Logger 接口定义（在消费端定义）
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, err error, args ...any)

	DebugContext(ctx context.Context, msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)

	With(args ...any) Logger
	WithGroup(name string) Logger
}

// StructuredLogger slog的包装器
type StructuredLogger struct {
	logger *slog.Logger
}

// 全局logger实例
var globalLogger Logger

// Init 初始化日志器
func Init(level slog.Level, environment string) {
	var handler slog.Handler

	// 根据环境配置不同的handler
	if environment == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level:     level,
			AddSource: true, // 开发环境显示源码位置
		})
	}

	// 包装context-aware handler
	contextHandler := NewContextHandler(handler)

	logger := slog.New(contextHandler)
	globalLogger = &StructuredLogger{logger: logger}

	// 设置slog的默认logger
	slog.SetDefault(logger)
}

// GetLogger 获取全局logger
func GetLogger() Logger {
	if globalLogger == nil {
		// 如果没有初始化，使用默认配置
		Init(slog.LevelInfo, "development")
	}
	return globalLogger
}

// NewLogger 创建新的logger实例
func NewLogger(handler slog.Handler) Logger {
	return &StructuredLogger{
		logger: slog.New(handler),
	}
}

// StructuredLogger 方法实现

func (l *StructuredLogger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

func (l *StructuredLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *StructuredLogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *StructuredLogger) Error(msg string, err error, args ...any) {
	args = append(args, slog.Any("error", err))
	l.logger.Error(msg, args...)
}

func (l *StructuredLogger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.logger.DebugContext(ctx, msg, args...)
}

func (l *StructuredLogger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.logger.InfoContext(ctx, msg, args...)
}

func (l *StructuredLogger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.logger.WarnContext(ctx, msg, args...)
}

func (l *StructuredLogger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.logger.ErrorContext(ctx, msg, args...)
}

func (l *StructuredLogger) With(args ...any) Logger {
	return &StructuredLogger{
		logger: l.logger.With(args...),
	}
}

func (l *StructuredLogger) WithGroup(name string) Logger {
	return &StructuredLogger{
		logger: l.logger.WithGroup(name),
	}
}

// ContextHandler 上下文感知的handler
type ContextHandler struct {
	next slog.Handler
}

// NewContextHandler 创建上下文handler
func NewContextHandler(next slog.Handler) *ContextHandler {
	return &ContextHandler{next: next}
}

func (h *ContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	// 从context中提取通用字段
	// if traceID, ok := ctx.Value(contextkeys.TraceIDKey).(string); ok && traceID != "" {
	// 	r.AddAttrs(slog.String("trace_id", traceID))
	// }

	// if spanID, ok := ctx.Value(contextkeys.SpanIDKey).(string); ok && spanID != "" {
	// 	r.AddAttrs(slog.String("span_id", spanID))
	// }

	// if userID, ok := ctx.Value(contextkeys.UserIDKey).(string); ok && userID != "" {
	// 	r.AddAttrs(slog.String("user_id", userID))
	// }

	// if requestID, ok := ctx.Value(contextkeys.RequestIDKey).(string); ok && requestID != "" {
	// 	r.AddAttrs(slog.String("request_id", requestID))
	// }

	// 可以添加其他通用字段的提取逻辑

	return h.next.Handle(ctx, r)
}

func (h *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ContextHandler{next: h.next.WithAttrs(attrs)}
}

func (h *ContextHandler) WithGroup(name string) slog.Handler {
	return &ContextHandler{next: h.next.WithGroup(name)}
}

// 便捷函数，使用全局logger

func Debug(msg string, args ...any) {
	GetLogger().Debug(msg, args...)
}

func Info(msg string, args ...any) {
	GetLogger().Info(msg, args...)
}

func Warn(msg string, args ...any) {
	GetLogger().Warn(msg, args...)
}

func Error(msg string, err error, args ...any) {
	GetLogger().Error(msg, err, args...)
}

func DebugContext(ctx context.Context, msg string, args ...any) {
	GetLogger().DebugContext(ctx, msg, args...)
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	GetLogger().InfoContext(ctx, msg, args...)
}

func WarnContext(ctx context.Context, msg string, args ...any) {
	GetLogger().WarnContext(ctx, msg, args...)
}

func ErrorContext(ctx context.Context, msg string, args ...any) {
	GetLogger().ErrorContext(ctx, msg, args...)
}

// 专门的错误记录函数，与xerrors集成

// LogAppError 记录AppError，利用其LogValue方法
func LogAppError(ctx context.Context, msg string, appErr *xerrors.AppError) {
	logger := GetLogger()

	// 使用AppError的LogValue方法获取结构化数据
	switch appErr.Level {
	case xerrors.LevelCritical:
		logger.ErrorContext(ctx, msg, slog.Any("app_error", appErr))
	case xerrors.LevelError:
		logger.ErrorContext(ctx, msg, slog.Any("app_error", appErr))
	case xerrors.LevelWarn:
		logger.WarnContext(ctx, msg, slog.Any("app_error", appErr))
	default:
		logger.InfoContext(ctx, msg, slog.Any("app_error", appErr))
	}
}

// LogOperation 记录操作日志
func LogOperation(ctx context.Context, operation string, result string, duration int64, metadata map[string]interface{}) {
	args := []any{
		slog.String("operation", operation),
		slog.String("result", result),
		slog.Int64("duration_ms", duration),
	}

	if metadata != nil {
		args = append(args, slog.Any("metadata", metadata))
	}

	GetLogger().InfoContext(ctx, "operation completed", args...)
}

// LogHTTPRequest 记录HTTP请求日志
func LogHTTPRequest(ctx context.Context, method, path string, statusCode int, duration int64, clientIP string) {
	args := []any{
		slog.String("method", method),
		slog.String("path", path),
		slog.Int("status_code", statusCode),
		slog.Int64("duration_ms", duration),
		slog.String("client_ip", clientIP),
	}

	// 根据状态码决定日志级别
	if statusCode >= 500 {
		GetLogger().ErrorContext(ctx, "HTTP request completed with server error", args...)
	} else if statusCode >= 400 {
		GetLogger().WarnContext(ctx, "HTTP request completed with client error", args...)
	} else {
		GetLogger().InfoContext(ctx, "HTTP request completed", args...)
	}
}

// LogDatabaseOperation 记录数据库操作日志
func LogDatabaseOperation(ctx context.Context, operation, table string, duration int64, rowsAffected int64, err error) {
	args := []any{
		slog.String("db_operation", operation),
		slog.String("table", table),
		slog.Int64("duration_ms", duration),
		slog.Int64("rows_affected", rowsAffected),
	}

	if err != nil {
		args = append(args, slog.Any("error", err))
		GetLogger().ErrorContext(ctx, "database operation failed", args...)
	} else {
		GetLogger().DebugContext(ctx, "database operation completed", args...)
	}
}

// LogBusinessEvent 记录业务事件
func LogBusinessEvent(ctx context.Context, event string, entityType, entityID string, metadata map[string]interface{}) {
	args := []any{
		slog.String("event", event),
		slog.String("entity_type", entityType),
		slog.String("entity_id", entityID),
	}

	if metadata != nil {
		args = append(args, slog.Any("metadata", metadata))
	}

	GetLogger().InfoContext(ctx, "business event occurred", args...)
}

// 性能优化相关函数

// IsDebugEnabled 检查是否启用了debug级别（避免昂贵的调试数据计算）
func IsDebugEnabled(ctx context.Context) bool {
	return GetLogger().(*StructuredLogger).logger.Enabled(ctx, slog.LevelDebug)
}

// ConditionalDebug 条件性debug日志（仅在debug启用时执行昂贵操作）
func ConditionalDebug(ctx context.Context, msg string, expensiveFunc func() []any) {
	if IsDebugEnabled(ctx) {
		args := expensiveFunc()
		GetLogger().DebugContext(ctx, msg, args...)
	}
}

// 专门的middleware日志记录器

// MiddlewareLogger 中间件专用logger
type MiddlewareLogger struct {
	logger Logger
}

// NewMiddlewareLogger 创建中间件logger
func NewMiddlewareLogger(logger Logger) *MiddlewareLogger {
	return &MiddlewareLogger{logger: logger}
}

// LogRequest 记录请求开始
func (ml *MiddlewareLogger) LogRequest(ctx context.Context, method, path, userAgent, clientIP string) {
	ml.logger.InfoContext(ctx, "request started",
		slog.String("method", method),
		slog.String("path", path),
		slog.String("user_agent", userAgent),
		slog.String("client_ip", clientIP),
	)
}

// LogResponse 记录响应完成
func (ml *MiddlewareLogger) LogResponse(ctx context.Context, statusCode int, duration int64, responseSize int64) {
	level := "info"
	if statusCode >= 500 {
		level = "error"
	} else if statusCode >= 400 {
		level = "warn"
	}

	args := []any{
		slog.Int("status_code", statusCode),
		slog.Int64("duration_ms", duration),
		slog.Int64("response_size_bytes", responseSize),
	}

	switch level {
	case "error":
		ml.logger.ErrorContext(ctx, "request completed with error", args...)
	case "warn":
		ml.logger.WarnContext(ctx, "request completed with warning", args...)
	default:
		ml.logger.InfoContext(ctx, "request completed", args...)
	}
}

// 结构化日志辅助函数

// Attrs 便捷的属性构造函数
func Attrs(keyvals ...interface{}) []any {
	var attrs []any
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			key := keyvals[i].(string)
			value := keyvals[i+1]
			attrs = append(attrs, slog.Any(key, value))
		}
	}
	return attrs
}

// Group 便捷的分组属性构造函数
func Group(name string, keyvals ...interface{}) slog.Attr {
	return slog.Group(name, Attrs(keyvals...)...)
}

// String 字符串属性
func String(key, value string) slog.Attr {
	return slog.String(key, value)
}

// Int 整数属性
func Int(key string, value int) slog.Attr {
	return slog.Int(key, value)
}

func Int64(key string, value int64) slog.Attr {
	return slog.Int64(key, value)
}

// Float64 浮点数属性
func Float64(key string, value float64) slog.Attr {
	return slog.Float64(key, value)
}

// Bool 布尔属性
func Bool(key string, value bool) slog.Attr {
	return slog.Bool(key, value)
}

// Any 任意类型属性
func Any(key string, value interface{}) slog.Attr {
	return slog.Any(key, value)
}

// Duration 时间间隔属性（以毫秒为单位）
func Duration(key string, duration int64) slog.Attr {
	return slog.Int64(key+"_ms", duration)
}
