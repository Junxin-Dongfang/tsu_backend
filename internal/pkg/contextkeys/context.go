// File: internal/pkg/contextkeys/context.go
package contextkeys

import "context"

// WithTraceID 向 context 中添加 trace ID
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// WithRequestID 向 context 中添加 request ID
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// WithUserID 向 context 中添加 user ID
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// GetTraceID 从 context 中获取 trace ID
func GetTraceID(ctx context.Context) (string, bool) {
	traceID, ok := ctx.Value(TraceIDKey).(string)
	return traceID, ok
}

// GetRequestID 从 context 中获取 request ID
func GetRequestID(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(RequestIDKey).(string)
	return requestID, ok
}

// GetUserID 从 context 中获取 user ID
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}
