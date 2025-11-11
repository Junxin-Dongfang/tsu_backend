// File: internal/pkg/response/response.go
package response

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"tsu-self/internal/pkg/ctxkey"
	"tsu-self/internal/pkg/i18n"
	"tsu-self/internal/pkg/metrics"
	"tsu-self/internal/pkg/xerrors"
)

// ResponseWriter 接口定义（在消费端定义接口）
type Writer interface {
	WriteJSON(ctx context.Context, w http.ResponseWriter, data interface{}, statusCode int) error
	WriteError(ctx context.Context, w http.ResponseWriter, err error) error
	WriteSuccess(ctx context.Context, w http.ResponseWriter, data interface{}) error
}

// Logger 接口定义（与 internal/pkg/log 的 Logger 接口兼容）
type Logger interface {
	InfoContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
	// 为了兼容 internal/pkg/log.Logger,添加其他方法（可选实现）
	// Debug(msg string, args ...any)
	// Info(msg string, args ...any)
	// Warn(msg string, args ...any)
	// Error(msg string, err error, args ...any)
}

// EmptyData 用于在 API 成功响应中表示"无数据"
type EmptyData struct{}

// APIResponse 通用的API响应结构体（完全不包含错误详情）
type APIResponse[T any] struct {
	Code      xerrors.ErrorCode `json:"code"`               // 业务响应码
	Message   string            `json:"message"`            // 面向用户的响应消息
	Data      *T                `json:"data,omitempty"`     // 响应数据，成功时返回
	Timestamp int64             `json:"timestamp"`          // Unix时间戳
	TraceID   string            `json:"trace_id,omitempty"` // 请求追踪ID
}

// Response Swagger 文档用的通用响应结构（非泛型版本，用于 API 文档生成）
type Response struct {
	Code      xerrors.ErrorCode `json:"code" example:"100000"`                    // 业务响应码
	Message   string            `json:"message" example:"操作成功"`                   // 面向用户的响应消息
	Data      interface{}       `json:"data,omitempty"`                           // 响应数据
	Timestamp int64             `json:"timestamp" example:"1759501201"`           // Unix时间戳
	TraceID   string            `json:"trace_id,omitempty" example:"abc-123-xyz"` // 请求追踪ID
}

// ErrorDetail 错误详情（仅在开发环境的特殊端点返回，用于调试）
type ErrorDetail struct {
	Code     xerrors.ErrorCode      `json:"code"`
	Message  string                 `json:"message"`
	Category string                 `json:"category,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Stack    string                 `json:"stack,omitempty"`
	File     string                 `json:"file,omitempty"`
	Line     int                    `json:"line,omitempty"`
}

// ResponseHandler HTTP响应处理器
type ResponseHandler struct {
	logger      Logger
	environment string // "development" | "production"
	metrics     *metrics.ErrorMetrics
}

// NewResponseHandler 创建响应处理器
func NewResponseHandler(logger Logger, environment string) *ResponseHandler {
	return &ResponseHandler{
		logger:      logger,
		environment: environment,
		metrics:     metrics.DefaultErrorMetrics,
	}
}

// NewResponseHandlerWithMetrics 创建带自定义指标的响应处理器
func NewResponseHandlerWithMetrics(logger Logger, environment string, m *metrics.ErrorMetrics) *ResponseHandler {
	return &ResponseHandler{
		logger:      logger,
		environment: environment,
		metrics:     m,
	}
}

// WriteJSON 写入JSON响应
func (h *ResponseHandler) WriteJSON(ctx context.Context, w http.ResponseWriter, data interface{}, statusCode int) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.ErrorContext(ctx, "failed to encode JSON response",
			slog.Any("error", err),
			slog.Int("status_code", statusCode),
		)
		return err
	}

	return nil
}

// WriteSuccess 写入成功响应
func (h *ResponseHandler) WriteSuccess(ctx context.Context, w http.ResponseWriter, data interface{}) error {
	// 获取语言偏好并返回本地化的成功消息
	lang := i18n.GetLanguage(ctx)
	successMsg := i18n.GetErrorMessage(xerrors.CodeSuccess, lang)

	resp := &APIResponse[interface{}]{
		Code:      xerrors.CodeSuccess,
		Message:   successMsg,
		Data:      &data,
		Timestamp: time.Now().Unix(),
		TraceID:   extractTraceID(ctx),
	}

	h.logger.InfoContext(ctx, "successful response",
		slog.Int("code", resp.Code.ToInt()),
		slog.String("message", resp.Message),
	)

	// 记录 Prometheus 指标
	if h.metrics != nil {
		method := extractMethod(ctx)
		h.metrics.RecordHTTPResponse(http.StatusOK, method, metrics.GetServiceName())
	}

	return h.WriteJSON(ctx, w, resp, http.StatusOK)
}

// WriteError 写入错误响应
func (h *ResponseHandler) WriteError(ctx context.Context, w http.ResponseWriter, err error) error {
	startTime := time.Now()

	// 将任意error转换为AppError
	appErr := h.normalizeError(err)

	// 检查是否有验证错误列表（多个验证错误时返回详细信息）
	var responseData interface{} = nil
	if appErr.Context != nil && appErr.Context.Metadata != nil {
		if validationErrors, ok := appErr.Context.Metadata["validation_errors"]; ok {
			// 有验证错误列表，在 data 中返回
			responseData = map[string]interface{}{
				"validation_errors": validationErrors,
			}
		}
	}

	// 构建响应（错误响应通常不包含error详情，但验证错误例外）
	resp := &APIResponse[interface{}]{
		Code:      appErr.Code,
		Message:   h.getUserMessage(ctx, appErr),
		Data:      &responseData,
		Timestamp: time.Now().Unix(),
		TraceID:   extractTraceID(ctx),
	}

	// 记录错误日志（这里才包含完整错误信息）
	h.logError(ctx, appErr)

	// 映射HTTP状态码
	statusCode := xerrors.GetHTTPStatus(appErr.Code)

	// 记录 Prometheus 指标
	if h.metrics != nil {
		method := extractMethod(ctx)
		duration := time.Since(startTime).Seconds()
		h.metrics.RecordError(appErr, statusCode, method, metrics.GetServiceName(), duration)
	}

	return h.WriteJSON(ctx, w, resp, statusCode)
}

// WriteDebugError 写入调试错误响应（仅开发环境使用）
func (h *ResponseHandler) WriteDebugError(ctx context.Context, w http.ResponseWriter, err error) error {
	if h.environment == "production" {
		// 生产环境降级为普通错误响应
		return h.WriteError(ctx, w, err)
	}

	appErr := h.normalizeError(err)

	// 开发环境可以返回详细的错误信息
	detail := &ErrorDetail{
		Code:     appErr.Code,
		Message:  appErr.Message,
		Category: appErr.Category,
		File:     appErr.File,
		Line:     appErr.Line,
		Stack:    appErr.Stack,
	}

	if appErr.Context != nil && appErr.Context.Metadata != nil {
		detail.Metadata = make(map[string]interface{})
		for k, v := range appErr.Context.Metadata {
			detail.Metadata[k] = v
		}
	}

	resp := &APIResponse[*ErrorDetail]{
		Code:      appErr.Code,
		Message:   h.getUserMessage(ctx, appErr),
		Data:      &detail,
		Timestamp: time.Now().Unix(),
		TraceID:   extractTraceID(ctx),
	}

	h.logError(ctx, appErr)
	statusCode := xerrors.GetHTTPStatus(appErr.Code)

	return h.WriteJSON(ctx, w, resp, statusCode)
}

// normalizeError 将任意error转换为AppError
func (h *ResponseHandler) normalizeError(err error) *xerrors.AppError {
	if appErr, ok := err.(*xerrors.AppError); ok {
		return appErr
	}

	// 将标准error包装为AppError
	return xerrors.NewWithError(xerrors.CodeInternalError, "系统内部错误", err)
}

// getUserMessage 获取面向用户的错误消息（支持i18n）
func (h *ResponseHandler) getUserMessage(ctx context.Context, appErr *xerrors.AppError) string {
	// 获取语言偏好
	lang := i18n.GetLanguage(ctx)

	// 在生产环境隐藏敏感错误信息
	if h.environment == "production" && appErr.IsCritical() {
		if lang.String() == "en" || lang.String() == "en-US" {
			return "Service temporarily unavailable, please try again later"
		}
		return "服务暂时不可用，请稍后重试"
	}

	// 优先使用metadata中的详细错误消息
	if appErr.Context != nil && appErr.Context.Metadata != nil {
		// 优先使用用户消息
		if userMsg, ok := appErr.Context.Metadata["user_message"].(string); ok && userMsg != "" {
			return userMsg
		}
		// 其次使用验证消息（用于表单验证错误）
		if validationMsg, ok := appErr.Context.Metadata["validation_message"].(string); ok && validationMsg != "" {
			return validationMsg
		}
		// 最后使用认证消息
		if authMsg, ok := appErr.Context.Metadata["auth_message"].(string); ok && authMsg != "" {
			return authMsg
		}
	}

	// 使用 i18n 获取本地化消息
	localizedMsg := i18n.GetErrorMessage(appErr.Code, lang)
	if localizedMsg != "" {
		return localizedMsg
	}

	// 降级到 AppError 的默认消息
	return appErr.Message
}

// logError 记录错误日志
func (h *ResponseHandler) logError(ctx context.Context, appErr *xerrors.AppError) {
	// 使用AppError的LogValue方法，避免重复序列化逻辑
	if appErr.IsCritical() {
		h.logger.ErrorContext(ctx, "critical error occurred", slog.Any("error", appErr))
	} else if appErr.Level == xerrors.LevelWarn {
		h.logger.WarnContext(ctx, "warning occurred", slog.Any("error", appErr))
	} else {
		h.logger.InfoContext(ctx, "error occurred", slog.Any("error", appErr))
	}
}

// extractTraceID 从context中提取trace_id
func extractTraceID(ctx context.Context) string {
	// 尝试从context中获取trace_id
	if traceID, ok := ctx.Value(ctxkey.TraceID).(string); ok {
		return traceID
	}

	// 或者从OpenTelemetry span中获取
	// if span := trace.SpanFromContext(ctx); span.IsRecording() {
	//     return span.SpanContext().TraceID().String()
	// }

	return ""
}

// extractMethod 从context中提取HTTP方法
func extractMethod(ctx context.Context) string {
	// 尝试从context中获取method
	if method, ok := ctx.Value(ctxkey.HTTPMethod).(string); ok {
		return method
	}
	return "UNKNOWN"
}

// 便捷函数，避免直接操作http.ResponseWriter

// OK 返回成功响应
func OK[T any](ctx context.Context, w http.ResponseWriter, h Writer, data *T) error {
	return h.WriteSuccess(ctx, w, data)
}

// Error 返回错误响应
func Error(ctx context.Context, w http.ResponseWriter, h Writer, err error) error {
	return h.WriteError(ctx, w, err)
}

// 快捷错误响应函数
func BadRequest(ctx context.Context, w http.ResponseWriter, h Writer, message string) error {
	err := xerrors.NewValidationError("request", message)
	return h.WriteError(ctx, w, err)
}

func Unauthorized(ctx context.Context, w http.ResponseWriter, h Writer, message string) error {
	err := xerrors.NewAuthError(message)
	return h.WriteError(ctx, w, err)
}

func Forbidden(ctx context.Context, w http.ResponseWriter, h Writer, resource, action string) error {
	err := xerrors.NewPermissionError(resource, action)
	return h.WriteError(ctx, w, err)
}

func NotFound(ctx context.Context, w http.ResponseWriter, h Writer, resource, identifier string) error {
	err := xerrors.NewNotFoundError(resource, identifier)
	return h.WriteError(ctx, w, err)
}

func InternalServerError(ctx context.Context, w http.ResponseWriter, h Writer, message string) error {
	err := xerrors.NewWithError(xerrors.CodeInternalError, message, nil)
	return h.WriteError(ctx, w, err)
}

// DefaultResponseHandler 创建默认响应处理器
func DefaultResponseHandler() Writer {
	// 简单的默认logger实现
	defaultLogger := &defaultLoggerImpl{
		logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
	}

	environment := os.Getenv("ENV")
	if environment == "" {
		environment = "development"
	}

	return NewResponseHandler(defaultLogger, environment)
}

// defaultLoggerImpl 默认logger实现
type defaultLoggerImpl struct {
	logger *slog.Logger
}

func (l *defaultLoggerImpl) InfoContext(ctx context.Context, msg string, args ...any) {
	l.logger.InfoContext(ctx, msg, args...)
}

func (l *defaultLoggerImpl) WarnContext(ctx context.Context, msg string, args ...any) {
	l.logger.WarnContext(ctx, msg, args...)
}

func (l *defaultLoggerImpl) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.logger.ErrorContext(ctx, msg, args...)
}
