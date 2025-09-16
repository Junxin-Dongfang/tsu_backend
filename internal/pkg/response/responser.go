package response

import (
	"encoding/json"
	"net/http"
	"time"
	"tsu-self/internal/pkg/xerrors"
)

// EmptyData 用于在 API 成功响应中表示“无数据”结构体
type EmptyData struct{}

// ResponseResult 通用的API响应结构体
type ResponseResult[T any] struct {
	Code      int               `json:"code"`               // 业务响应码
	Message   string            `json:"message"`            // 响应消息
	Data      *T                `json:"data,omitempty"`     // 响应数据，成功时返回
	Error     *xerrors.AppError `json:"error,omitempty"`    // 错误详情，失败时返回
	Timestamp int64             `json:"timestamp"`          // Unix时间戳
	TraceId   string            `json:"trace_id,omitempty"` // 请求追踪ID
}

// Success 创建一个成功的响应
func Success[T any](data *T) *ResponseResult[T] {
	return &ResponseResult[T]{
		Code:      xerrors.CodeSuccess,
		Message:   "操作成功",
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
}

// SuccessWithMessage 创建一个成功的响应，并允许自定义消息
func SuccessWithMessage[T any](data *T, message string) *ResponseResult[T] {
	return &ResponseResult[T]{
		Code:      xerrors.CodeSuccess,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
}

// ErrorFromAppError 创建一个失败的响应，基于 AppError
func ErrorFromAppError[T any](appErr *xerrors.AppError) *ResponseResult[T] {
	resp := &ResponseResult[T]{
		Code:      appErr.Code,
		Message:   appErr.GetUserMessage(),
		Error:     appErr,
		Timestamp: time.Now().Unix(),
	}

	if appErr.Context != nil {
		resp.TraceId = appErr.Context.TraceID
	}
	return resp
}

// JSON 将响应以JSON格式写入 http.ResponseWriter
func JSON[T any](w http.ResponseWriter, r *http.Request, resp *ResponseResult[T]) {
	// 尝试从上下文中获取 trace_id (通常由中间件注入)
	if resp.TraceId == "" {
		if traceID, ok := r.Context().Value("trace_id").(string); ok {
			resp.TraceId = traceID
		}
	}

	if isProduction() && resp.Error != nil {
		// 在生产环境中，隐藏详细的错误信息
		resp.Error = &xerrors.AppError{
			Code:    resp.Error.Code,
			Message: "服务器内部错误",
			Level:   resp.Error.Level,
		}
	}

	statusCode := xerrors.GetHTTPStatus(resp.Code)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// 序列化失败的兜底处理
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// isProduction 检查是否为生产环境
func isProduction() bool {
	// 根据你的环境变量或配置来判断
	// return os.Getenv("ENV") == "production"
	return false // 这里先返回 false，你可以根据实际情况修改
}

// OK 返回成功响应
func OK[T any](w http.ResponseWriter, r *http.Request, data *T) {
	JSON(w, r, Success(data))
}

// OKWithMessage 返回带消息的成功响应
func OKWithMessage[T any](w http.ResponseWriter, r *http.Request, data *T, message string) {
	JSON(w, r, SuccessWithMessage(data, message))
}

// Error 返回错误响应
func Error[T any](w http.ResponseWriter, r *http.Request, appErr *xerrors.AppError) {
	JSON(w, r, ErrorFromAppError[T](appErr))
}

// 以下是常用错误的便捷函数

// BadRequest 返回参数错误响应
func BadRequest(w http.ResponseWriter, r *http.Request, message string) {
	appErr := xerrors.NewValidationError(message)
	Error[EmptyData](w, r, appErr)
}

// Unauthorized 返回未授权响应
func Unauthorized(w http.ResponseWriter, r *http.Request, message string) {
	appErr := xerrors.NewAuthError(message)
	Error[EmptyData](w, r, appErr)
}

// Forbidden 返回权限不足响应
func Forbidden(w http.ResponseWriter, r *http.Request, message string) {
	appErr := xerrors.NewPermissionError(message)
	Error[EmptyData](w, r, appErr)
}

// NotFound 返回资源不存在响应
func NotFound(w http.ResponseWriter, r *http.Request, resource string) {
	appErr := xerrors.NewNotFoundError(resource)
	Error[EmptyData](w, r, appErr)
}

// InternalServerError 返回内部服务错误响应
func InternalServerError(w http.ResponseWriter, r *http.Request, message string) {
	appErr := xerrors.NewSystemError(message)
	Error[EmptyData](w, r, appErr)
}
