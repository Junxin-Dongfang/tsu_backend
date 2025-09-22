package middleware

import (
	"context"
	"net/http"

	"tsu-self/internal/pkg/contextkeys" // 引入我们统一的 "钥匙保管处"

	"github.com/google/uuid"
)

// TraceID 是一个 HTTP 中间件，它确保每个请求都有一个唯一的 Trace ID。
// 它首先会尝试从 "X-Request-ID" 请求头中获取，如果不存在，则会生成一个新的UUID。
// 然后，它会将这个 Trace ID 存入请求的 context 中，并设置到响应头中。
func TraceID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. 尝试从请求头中获取 Trace ID，这对于保持跨服务调用的追踪链至关重要。
		traceID := r.Header.Get("X-Request-ID")

		// 2. 如果请求头中没有，说明这是链路的第一个请求，我们为它生成一个新的唯一ID。
		if traceID == "" {
			traceID = uuid.NewString()
		}

		// 3. 使用从 contextkeys 包导入的唯一键，将 Trace ID 存入请求的 context 中。
		// 这样，后续的所有处理程序都可以安全地从中获取。
		ctx := context.WithValue(r.Context(), contextkeys.TraceIDKey, traceID)
		r = r.WithContext(ctx)

		// 4. 将 Trace ID 设置到响应头中，方便前端或调用方进行问题追踪。
		w.Header().Set("X-Request-ID", traceID)

		// 5. 调用链中的下一个中间件或最终的 HTTP 处理程序。
		next.ServeHTTP(w, r)
	})
}
