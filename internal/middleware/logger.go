// File: internal/middleware/logger.go
package middleware

import (
	"net/http"
	"time"

	"tsu-self/internal/pkg/log" // 引入我们统一的 log 包
)

// responseWriter 包装器，用于捕获由下游处理程序写入的 HTTP 响应状态码。
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	// 默认状态码为 200 OK，如果下游没有显式写入状态码，这就是最终的状态。
	return &responseWriter{w, http.StatusOK}
}

// WriteHeader 捕获状态码并调用原始的 WriteHeader 方法。
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logger 是一个 HTTP 中间件，它在每个请求处理完成后，记录一条包含丰富信息的结构化日志。
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 包装原始的 ResponseWriter 以便捕获状态码。
		rw := newResponseWriter(w)

		// 调用链中的下一个中间件或最终的 HTTP 处理程序。
		// 我们将包装后的 rw 传递下去。
		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		// 使用我们统一的 log 包来记录日志。
		// log.Info 函数会自动从请求的 context 中提取 trace_id。
		log.InfoWithCtx(r.Context(), "http request processed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.statusCode,
			"duration_ms", duration.Milliseconds(),
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)
	})
}
