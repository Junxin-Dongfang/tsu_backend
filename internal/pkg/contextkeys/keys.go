// File: internal/pkg/contextkeys/keys.go
package contextkeys

// 定义所有context key的类型，避免包间冲突
type contextKey string

const (
	TraceIDKey   contextKey = "trace_id"
	SpanIDKey    contextKey = "span_id"
	UserIDKey    contextKey = "user_id"
	SessionIDKey contextKey = "session_id"
	RequestIDKey contextKey = "request_id"

	// 可以添加更多键
	CorrelationIDKey contextKey = "correlation_id"
	TenantIDKey      contextKey = "tenant_id"
	ClientIPKey      contextKey = "client_ip"
	UserAgentKey     contextKey = "user_agent"
)
