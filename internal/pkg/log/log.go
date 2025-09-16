// File: internal/pkg/log/log.go
package log

import (
	"context"
	"log/slog"
	"os"

	"tsu-self/internal/pkg/contextkeys" // 再次引入 "钥匙保管处"
)

// logger 是一个包级别的私有变量，用于持有我们全局配置的 slog 实例。
var logger *slog.Logger

// Init 初始化日志器。这个函数应该在每个服务的 main.go 的最开始被调用一次。
func Init(level slog.Level) {
	// 使用 slog 的 JSON Handler，输出结构化的 JSON 日志到标准输出。
	// 在生产环境中，这便于被日志收集系统（如 Fluentd, Logstash）采集。
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		// 设置日志级别。低于此级别的日志将不会被输出。
		Level: level,
		// (可选) AddSource: true, // 如果需要，可以添加日志源文件和行号，但会轻微影响性能。
	})

	logger = slog.New(handler)

	// (可选) 您可以设置 slog 的全局默认日志器，但这通常不推荐，因为它不携带 context。
	// 我们鼓励在项目中使用本包提供的 Info, Warn, Error 等函数。
	// slog.SetDefault(logger)
}

// fromContext 是一个内部辅助函数，它从 context 中提取通用字段（如 trace_id），
// 并返回一个已经附带了这些字段的新的 slog.Logger 实例。
func fromContext(ctx context.Context) *slog.Logger {
	l := logger // 从包级别的日志器开始

	// 安全检查：如果 logger 还没有被初始化，使用默认的 logger
	if l == nil {
		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
		l = slog.New(handler)
	}

	// 使用从 contextkeys 包导入的唯一键，安全地从 context 中取值。
	if traceID, ok := ctx.Value(contextkeys.TraceIDKey).(string); ok {
		// 如果 context 中有 trace_id，就把它添加到这条日志的所有属性中。
		l = l.With("trace_id", traceID)
	}

	// 未来如果需要，您可以在这里添加其他想从 context 提取并添加到每条日志的通用字段。
	// if userID, ok := ctx.Value(contextkeys.UserIDKey).(string); ok {
	// 	l = l.With("user_id", userID)
	// }

	return l
}

// Info 记录一条 info 级别的日志。它会自动包含 context 中的 trace_id。
func InfoWithCtx(ctx context.Context, msg string, args ...any) {
	fromContext(ctx).Info(msg, args...)
}

// Warn 记录一条 warn 级别的日志。
func WarnWithCtx(ctx context.Context, msg string, args ...any) {
	fromContext(ctx).Warn(msg, args...)
}

// Error 记录一条 error 级别的日志，并自动将 Go 的 error 对象作为一个结构化的字段。
func ErrorWithCtx(ctx context.Context, msg string, err error, args ...any) {
	// 将 Go 的 error 对象作为一个结构化的属性 "error"，而不是简单的字符串拼接。
	// 这使得在日志系统中按错误类型进行过滤和告警成为可能。
	args = append(args, slog.Any("error", err))
	fromContext(ctx).Error(msg, args...)
}

// Debug 记录一条 debug 级别的日志。
func DebugWithCtx(ctx context.Context, msg string, args ...any) {
	fromContext(ctx).Debug(msg, args...)
}

// Info 记录一条 info 级别的日志。它会自动包含 context 中的 trace_id。
func Info(msg string, args ...any) {
	fromContext(context.Background()).Info(msg, args...)
}

// Warn 记录一条 warn 级别的日志。
func Warn(msg string, args ...any) {
	fromContext(context.Background()).Warn(msg, args...)
}

// Error 记录一条 error 级别的日志，并自动将 Go 的 error 对象作为一个结构化的字段。
func Error(msg string, err error, args ...any) {
	// 将 Go 的 error 对象作为一个结构化的属性 "error"，而不是简单的字符串拼接。
	// 这使得在日志系统中按错误类型进行过滤和告警成为可能。
	args = append(args, slog.Any("error", err))
	fromContext(context.Background()).Error(msg, args...)
}

// Debug 记录一条 debug 级别的日志。
func Debug(msg string, args ...any) {
	fromContext(context.Background()).Debug(msg, args...)
}
