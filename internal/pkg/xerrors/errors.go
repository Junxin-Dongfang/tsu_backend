// File: internal/pkg/xerrors/errors.go
package xerrors

import (
	"fmt"
	"log/slog"
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

// ErrorContext 错误上下文信息（精简版，移除HTTP特定字段）
type ErrorContext struct {
	TraceID   string            `json:"trace_id,omitempty"`
	SpanID    string            `json:"span_id,omitempty"`
	UserID    string            `json:"user_id,omitempty"`
	SessionID string            `json:"session_id,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
	Service   string            `json:"service,omitempty"`
	Operation string            `json:"operation,omitempty"` // 改为operation，更通用
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// AppError 领域错误
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`

	// 错误分类和级别
	Level    ErrorLevel `json:"level,omitempty"`
	Category string     `json:"category,omitempty"`

	// 业务上下文
	Context   *ErrorContext `json:"context,omitempty"`
	Timestamp time.Time     `json:"timestamp,omitempty"`

	// 调试信息
	Stack string `json:"stack,omitempty"`
	File  string `json:"file,omitempty"`
	Line  int    `json:"line,omitempty"`

	// 业务属性
	Retryable   bool `json:"retryable,omitempty"`
	Recoverable bool `json:"recoverable,omitempty"` // 新增：是否可恢复
}

// Error 实现标准 error 接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 实现 errors.Unwrap 接口
func (e *AppError) Unwrap() error {
	return e.Err
}

// LogValue 实现 slog.LogValuer 接口，避免重复序列化逻辑
func (e *AppError) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.Int("code", e.Code),
		slog.String("message", e.Message),
		slog.String("level", e.Level.String()),
		slog.String("category", e.Category),
		slog.Bool("retryable", e.Retryable),
		slog.Bool("recoverable", e.Recoverable),
	}

	if e.Context != nil {
		if e.Context.TraceID != "" {
			attrs = append(attrs, slog.String("trace_id", e.Context.TraceID))
		}
		if e.Context.UserID != "" {
			attrs = append(attrs, slog.String("user_id", e.Context.UserID))
		}
		if e.Context.Service != "" {
			attrs = append(attrs, slog.String("service", e.Context.Service))
		}
		if e.Context.Operation != "" {
			attrs = append(attrs, slog.String("operation", e.Context.Operation))
		}
	}

	if e.Err != nil {
		attrs = append(attrs, slog.Any("underlying_error", e.Err))
	}

	return slog.GroupValue(attrs...)
}

// WithContext 添加上下文信息
func (e *AppError) WithContext(ctx *ErrorContext) *AppError {
	newErr := *e
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

// WithService 添加服务和操作信息
func (e *AppError) WithService(service, operation string) *AppError {
	if e.Context == nil {
		e.Context = &ErrorContext{}
	}
	e.Context.Service = service
	e.Context.Operation = operation
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

// WithRecoverable 设置错误是否可恢复
func (e *AppError) WithRecoverable(recoverable bool) *AppError {
	e.Recoverable = recoverable
	return e
}

// IsRetryable 判断是否为可重试错误
func (e *AppError) IsRetryable() bool {
	return e.Retryable
}

// IsRecoverable 判断是否为可恢复错误
func (e *AppError) IsRecoverable() bool {
	return e.Recoverable
}

// IsCritical 判断是否为严重错误
func (e *AppError) IsCritical() bool {
	return e.Level == LevelCritical
}

// New 创建新的AppError
func New(code int, message string) *AppError {
	return &AppError{
		Code:      code,
		Message:   message,
		Level:     LevelError,
		Category:  getCategoryByCode(code),
		Timestamp: time.Now(),
		Retryable: isRetryableByCode(code),
	}
}

// NewWithError 创建包含原始错误的 AppError
func NewWithError(code int, message string, err error) *AppError {
	appErr := New(code, message)
	appErr.Err = err

	// 添加调试信息
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

// 快捷构造函数（专注于业务错误，使用你的错误码）
func NewValidationError(field, message string) *AppError {
	return FromCode(CodeInvalidParams).
		WithMetadata("field", field).
		WithMetadata("validation_message", message)
}

func NewAuthError(message string) *AppError {
	return FromCode(CodeAuthenticationFailed).
		WithMetadata("auth_message", message)
}

func NewPermissionError(resource, action string) *AppError {
	return FromCode(CodePermissionDenied).
		WithMetadata("resource", resource).
		WithMetadata("action", action)
}

func NewNotFoundError(resource, identifier string) *AppError {
	return FromCode(CodeResourceNotFound).
		WithMetadata("resource", resource).
		WithMetadata("identifier", identifier)
}

func NewConflictError(resource, reason string) *AppError {
	return FromCode(CodeDuplicateResource).
		WithMetadata("resource", resource).
		WithMetadata("conflict_reason", reason)
}

// 新增：基于你的错误码系统的便捷函数
func NewUserNotFoundError(userID string) *AppError {
	return FromCode(CodeUserNotFound).
		WithMetadata("user_id", userID)
}

func NewUserExistsError(field, value string) *AppError {
	var code int
	switch field {
	case "email":
		code = CodeEmailExists
	case "username":
		code = CodeUsernameExists
	case "phone":
		code = CodePhoneExists
	default:
		code = CodeUserAlreadyExists
	}
	return FromCode(code).
		WithMetadata("field", field).
		WithMetadata("value", value)
}

func NewTokenError(tokenType string) *AppError {
	return FromCode(CodeInvalidToken).
		WithMetadata("token_type", tokenType)
}

func NewTokenExpiredError(tokenType string) *AppError {
	return FromCode(CodeTokenExpired).
		WithMetadata("token_type", tokenType)
}

func NewRoleError(roleID string) *AppError {
	return FromCode(CodeRoleNotFound).
		WithMetadata("role_id", roleID)
}

func NewExternalServiceError(service string, err error) *AppError {
	appErr := FromCode(CodeExternalServiceError).
		WithMetadata("external_service", service)
	if err != nil {
		appErr.Err = err
	}
	return appErr
}

func NewDatabaseError(operation, table string, err error) *AppError {
	appErr := FromCode(CodeDatabaseError).
		WithMetadata("db_operation", operation).
		WithMetadata("table", table)
	if err != nil {
		appErr.Err = err
	}
	return appErr
}
