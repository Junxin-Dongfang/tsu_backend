// File: internal/pkg/ctxkey/ctxkey.go
package ctxkey

import "context"

// ContextKey 统一的 context key 类型
type ContextKey string

const (
	// Language 语言偏好
	Language ContextKey = "language"

	// TraceID 请求追踪 ID
	TraceID ContextKey = "trace_id"

	// HTTPMethod HTTP 请求方法
	HTTPMethod ContextKey = "http_method"

	// UserID 用户 ID (从认证中间件设置)
	UserID ContextKey = "user_id"

	// SessionID 会话 ID
	SessionID ContextKey = "session_id"

	// RequestID 请求 ID (可选，与 TraceID 不同)
	RequestID ContextKey = "request_id"

	// CurrentUser 当前用户完整信息（存储在 Echo Context 中）
	CurrentUser ContextKey = "current_user"

	// HeroID 当前操作英雄 ID (从英雄中间件设置)
	HeroID ContextKey = "hero_id"
)

// WithValue 在 context 中设置指定 key 的值
func WithValue(ctx context.Context, key ContextKey, value interface{}) context.Context {
	return context.WithValue(ctx, key, value)
}

// GetString 从 context 中获取字符串类型的值
func GetString(ctx context.Context, key ContextKey) string {
	if value, ok := ctx.Value(key).(string); ok {
		return value
	}
	return ""
}
