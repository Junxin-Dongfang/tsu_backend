// File: internal/middleware/middleware.go (汇总中间件)
package middleware

import (
	"log"
	"net/http"
	"runtime"

	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/xerrors"
)

// Recovery panic 恢复中间件
func Recovery() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// 获取调用栈信息
					stack := make([]byte, 1024*8)
					length := runtime.Stack(stack, false)

					log.Printf("[PANIC RECOVERED] %v\nStack: %s", err, stack[:length])

					// 检查响应是否已经写入
					if w.Header().Get("Content-Type") == "" {
						appErr := xerrors.NewSystemError("系统内部错误，请稍后重试")
						if traceId := r.Context().Value("trace_id"); traceId != nil {
							appErr.WithTraceID(traceId.(string))
						}

						response.Error[response.EmptyData](w, r, appErr)
					}
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
