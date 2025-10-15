// File: internal/pkg/trace/trace.go
package trace

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"tsu-self/internal/pkg/ctxkey"
)

// WithTraceID 在 context 中设置 trace ID
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return ctxkey.WithValue(ctx, ctxkey.TraceID, traceID)
}

// GetTraceID 从 context 中获取 trace ID
func GetTraceID(ctx context.Context) string {
	return ctxkey.GetString(ctx, ctxkey.TraceID)
}

// GenerateTraceID 生成新的 trace ID
// 格式: 32 个字符的十六进制字符串 (类似 OpenTelemetry trace ID)
func GenerateTraceID() string {
	b := make([]byte, 16) // 128 bits
	if _, err := rand.Read(b); err != nil {
		// 降级到基于时间的 ID
		return fmt.Sprintf("%032x", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// ExtractFromHeader 从 HTTP 头部提取 trace ID
// 支持多种标准：X-Trace-Id, X-Request-Id, Traceparent (W3C)
func ExtractFromHeader(headers map[string][]string) string {
	// 1. 优先使用 X-Trace-Id
	if traceID := getHeader(headers, "X-Trace-Id"); traceID != "" {
		return traceID
	}

	// 2. 其次使用 X-Request-Id
	if requestID := getHeader(headers, "X-Request-Id"); requestID != "" {
		return requestID
	}

	// 3. 尝试解析 W3C Traceparent 头部
	if traceparent := getHeader(headers, "Traceparent"); traceparent != "" {
		if traceID := parseTraceparent(traceparent); traceID != "" {
			return traceID
		}
	}

	// 4. 如果都没有，生成新的 trace ID
	return GenerateTraceID()
}

// getHeader 从 headers map 中获取指定 key 的值（忽略大小写）
func getHeader(headers map[string][]string, key string) string {
	for k, values := range headers {
		if equalsFold(k, key) && len(values) > 0 {
			return values[0]
		}
	}
	return ""
}

// equalsFold 简单的忽略大小写比较
func equalsFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca := a[i]
		cb := b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

// parseTraceparent 解析 W3C Traceparent 头部
// 格式: "00-<trace-id>-<parent-id>-<flags>"
// 例如: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
func parseTraceparent(traceparent string) string {
	if len(traceparent) < 55 { // 最小长度
		return ""
	}

	// 提取 trace-id 部分 (位置 3-35)
	if traceparent[2] == '-' && len(traceparent) > 35 {
		return traceparent[3:35]
	}

	return ""
}
