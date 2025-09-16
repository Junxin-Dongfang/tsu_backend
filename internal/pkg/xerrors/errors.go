// File: internal/pkg/xerrors/errors.go
package xerrors

import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"
)

// ErrorLevel 错误级别
type ErrorLevel int

const (
	LevelInfo ErrorLevel = iota
	LevelWarn
	LevelError
	LevelCritical
)

func (l ErrorLevel) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// ErrorContext 错误上下文信息
type ErrorContext struct {
	TraceID   string            `json:"trace_id,omitempty"`
	SpanID    string            `json:"span_id,omitempty"`
	UserID    string            `json:"user_id,omitempty"`
	SessionID string            `json:"session_id,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
	Service   string            `json:"service,omitempty"`
	Method    string            `json:"method,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// AppError 自定义错误结构体
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`

	//错误级别和分类
	Level    ErrorLevel `json:"level,omitempty"`
	Category string     `json:"category,omitempty"`

	// 上下文信息
	Context   *ErrorContext `json:"context,omitempty"`
	Timestamp time.Time     `json:"timestamp,omitempty"`

	//调试信息
	Stack string `json:"stack,omitempty"`
	File  string `json:"file,omitempty"`
	Line  int    `json:"line,omitempty"`

	UserMessage string `json:"user_message,omitempty"`
	Retryable   bool   `json:"retryable,omitempty"`
}

// Error 实现标准 error 接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 实现 errors.Unwrap 接口，返回底层错误
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithContext 添加上下文信息
func (e *AppError) WithContext(ctx *ErrorContext) *AppError {
	newErr := *e // 复制当前错误
	newErr.Context = ctx
	return &newErr
}

// WithTraceID 添加 TraceID
func (e *AppError) WithTraceID(traceID string) *AppError {
	if e.Context == nil {
		e.Context = &ErrorContext{}
	}

	e.Context.TraceID = traceID
	return e
}

// WithUser 添加用户相关信息
func (e *AppError) WithUser(userID, sessionID string) *AppError {
	if e.Context == nil {
		e.Context = &ErrorContext{}
	}

	e.Context.UserID = userID
	e.Context.SessionID = sessionID
	return e
}

// WithService 添加服务和方法信息
func (e *AppError) WithService(service, method string) *AppError {
	if e.Context == nil {
		e.Context = &ErrorContext{}
	}

	e.Context.Service = service
	e.Context.Method = method
	return e
}

// WithMetadata 添加自定义元数据
func (e *AppError) WithMetadata(key, value string) *AppError {
	if e.Context == nil {
		e.Context = &ErrorContext{}
	}
	if e.Context.Metadata == nil {
		e.Context.Metadata = make(map[string]string)
	}
	e.Context.Metadata[key] = value
	return e
}

// WithUserMessage 添加面向用户的错误信息
func (e *AppError) WithUserMessage(userMessage string) *AppError {
	if e.Context == nil {
		e.Context = &ErrorContext{}
	}
	e.UserMessage = userMessage
	return e
}

// WithRetryable 设置错误是否可重试
func (e *AppError) WithRetryable(retryable bool) *AppError {
	if e.Context == nil {
		e.Context = &ErrorContext{}
	}
	e.Retryable = retryable
	return e
}

// IsRetryable 判断是否为可重试错误
func (e *AppError) IsRetryable() bool {
	return e.Retryable
}

// IsCritical 判断是否为严重错误
func (e *AppError) IsCritical() bool {
	return e.Level == LevelCritical
}

// GetUserMessage 获取面向用户的错误信息
func (e *AppError) GetUserMessage() string {
	if e.UserMessage != "" {
		return e.UserMessage
	}
	return e.Message
}

// ToJSON 格式化为 JSON 字符串
func (e *AppError) ToJSON() string {
	data, _ := json.Marshal(e)
	return string(data)
}

// New 创建新的AppError
func New(code int, message string) *AppError {
	return &AppError{
		Code:      code,
		Message:   message,
		Level:     LevelError,
		Category:  getCategoryByCode(code),
		Timestamp: time.Now(),
	}
}

// NewWithError 创建包含原始错误的 AppError
func NewWithError(code int, message string, err error) *AppError {
	appErr := New(code, message)
	appErr.Err = err

	//添加调试信息
	if pc, file, line, ok := runtime.Caller(1); ok {
		appErr.File = file
		appErr.Line = line
		if fn := runtime.FuncForPC(pc); fn != nil {
			appErr.Stack = fn.Name()
		}
	}

	return appErr
}

// FromCode 根据错误码创建 AppError
func FromCode(code int) *AppError {
	msg, ok := codeMessages[code]
	if !ok {
		msg = codeMessages[CodeInternalError]
	}
	return &AppError{
		Code:      code,
		Message:   msg,
		Level:     getLevelByCode(code),
		Category:  getCategoryByCode(code),
		Timestamp: time.Now(),
		Retryable: isRetryableByCode(code),
	}
}

// ErrorBuilder 错误构建器
type ErrorBuilder struct {
	error *AppError
}

// NewBuilder 创建错误构建器
func NewBuilder(code int, message string) *ErrorBuilder {
	return &ErrorBuilder{
		error: New(code, message),
	}
}

func (b *ErrorBuilder) WithError(err error) *ErrorBuilder {
	b.error.Err = err
	return b
}

func (b *ErrorBuilder) WithLevel(level ErrorLevel) *ErrorBuilder {
	b.error.Level = level
	return b
}

func (b *ErrorBuilder) WithCategory(category string) *ErrorBuilder {
	b.error.Category = category
	return b
}

func (b *ErrorBuilder) WithContext(ctx *ErrorContext) *ErrorBuilder {
	b.error.Context = ctx
	return b
}

func (b *ErrorBuilder) WithUserMessage(msg string) *ErrorBuilder {
	b.error.UserMessage = msg
	return b
}

func (b *ErrorBuilder) WithRetryable(retryable bool) *ErrorBuilder {
	b.error.Retryable = retryable
	return b
}

func (b *ErrorBuilder) Build() *AppError {
	return b.error
}

// 快捷构造函数

func NewSystemError(message string) *AppError {
	return FromCode(CodeInternalError).WithUserMessage(message)
}

func NewBusinessError(code int, message string) *AppError {
	return FromCode(code).WithUserMessage(message)
}

func NewValidationError(message string) *AppError {
	return FromCode(CodeInvalidParams).WithUserMessage(message)
}

func NewAuthError(message string) *AppError {
	return FromCode(CodeAuthenticationFailed).WithUserMessage(message)
}

func NewPermissionError(message string) *AppError {
	return FromCode(CodePermissionDenied).WithUserMessage(message)
}

func NewNotFoundError(resource string) *AppError {
	return FromCode(CodeResourceNotFound).
		WithUserMessage(fmt.Sprintf("%s不存在", resource))
}

func NewConflictError(resource string) *AppError {
	return FromCode(CodeDuplicateResource).
		WithUserMessage(fmt.Sprintf("%s已存在", resource))
}
