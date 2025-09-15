package response

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// EmptyData 是一个用于在 API 成功响应中表示“无数据”的结构体。
// 使用一个具体的空结构体，比直接返回 nil 或 interface{} 更类型安全、意图更明确。
type EmptyData struct{}

// ResponseResult 是一个通用的API响应结构体
type ResponseResult[T any] struct {
	Code      int    `json:"code"`               // 业务响应码
	Message   string `json:"message"`            // 响应消息
	Data      *T     `json:"data,omitempty"`     // 响应数据，成功时返回
	Error     string `json:"error,omitempty"`    // 错误详情，失败时返回
	Timestamp int64  `json:"timestamp"`          // Unix时间戳
	TraceId   string `json:"trace_id,omitempty"` // 请求追踪ID
}

// Success 创建一个成功的响应
func Success[T any](data *T) *ResponseResult[T] {
	return &ResponseResult[T]{
		Code:      100000, // 假设 100000 是标准成功代码
		Message:   "操作成功",
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
}

// Error 创建一个失败的响应
// 注意：对于失败响应，泛型 T 的具体类型不重要，所以 Data 字段将为 nil
func Error[T any](code int, message string, err string) *ResponseResult[T] {
	return &ResponseResult[T]{
		Code:      code,
		Message:   message,
		Error:     err,
		Timestamp: time.Now().Unix(),
	}
}

// JSON 将响应以JSON格式写入 http.ResponseWriter
// 这个辅助函数非常重要，它统一了所有API的输出方式
func JSON[T any](w http.ResponseWriter, statusCode int, resp *ResponseResult[T]) {
	// 尝试从上下文中获取 trace_id (通常由中间件注入)
	// traceId, ok := r.Context().Value("trace_id").(string)
	// if ok {
	//   resp.TraceId = traceId
	// }

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode) // 写入HTTP状态码

	// 将响应结构体序列化为JSON并写入响应体
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// 如果序列化失败，记录日志
		// 注意：此时再写入http.Error会失败，因为header已经写入
		log.Printf("写入JSON响应失败: %v", err)
	}
}
