// File: internal/pkg/xerrors/errors.go
package xerrors

import (
	"fmt"
	"log/slog"
	"runtime"
	"time"

	"golang.org/x/text/language"
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
	TraceID   string                 `json:"trace_id,omitempty"`
	SpanID    string                 `json:"span_id,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	Service   string                 `json:"service,omitempty"`
	Operation string                 `json:"operation,omitempty"` // 改为operation，更通用
	Metadata  map[string]interface{} `json:"metadata,omitempty"`  // 支持任意类型
}

// AppError 领域错误
type AppError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Err     error     `json:"-"`

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

// GetLocalizedMessage 获取本地化的错误消息
// 注意: 为避免循环依赖，此方法不直接导入 i18n 包
// 使用者应该通过 i18n.GetErrorMessage(err.Code, lang) 获取本地化消息
func (e *AppError) GetLocalizedMessage(lang language.Tag) string {
	// 此方法预留给未来扩展
	// 目前调用者应使用 i18n.GetErrorMessage(err.Code, lang)
	return e.Message
}

// LogValue 实现 slog.LogValuer 接口，避免重复序列化逻辑
func (e *AppError) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.Int("code", int(e.Code)),
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

// WithMetadata 添加自定义元数据（支持任意类型）
func (e *AppError) WithMetadata(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = &ErrorContext{}
	}
	if e.Context.Metadata == nil {
		e.Context.Metadata = make(map[string]interface{})
	}
	e.Context.Metadata[key] = value
	return e
}

// WithMetadataString 添加字符串类型的元数据（便捷方法，向后兼容）
func (e *AppError) WithMetadataString(key, value string) *AppError {
	return e.WithMetadata(key, value)
}

// WithMetadataInt 添加整数类型的元数据（便捷方法）
func (e *AppError) WithMetadataInt(key string, value int) *AppError {
	return e.WithMetadata(key, value)
}

// WithMetadataStruct 添加结构体类型的元数据（便捷方法）
func (e *AppError) WithMetadataStruct(key string, value interface{}) *AppError {
	return e.WithMetadata(key, value)
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
func New(code ErrorCode, message string) *AppError {
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
func NewWithError(code ErrorCode, message string, err error) *AppError {
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
func FromCode(code ErrorCode) *AppError {
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

func NewInvalidArgumentError(field, message string) *AppError {
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
	var code ErrorCode
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

// 游戏业务错误快捷构造器
func NewHeroNotFoundError(heroID string) *AppError {
	return FromCode(CodeHeroNotFound).
		WithMetadata("hero_id", heroID)
}

func NewHeroNameExistsError(heroName string) *AppError {
	return FromCode(CodeHeroNameExists).
		WithMetadata("hero_name", heroName)
}

func NewSkillNotFoundError(skillID string) *AppError {
	return FromCode(CodeSkillNotFound).
		WithMetadata("skill_id", skillID)
}

func NewSkillCooldownError(skillID string, remainingSeconds int) *AppError {
	return FromCode(CodeSkillCooldown).
		WithMetadata("skill_id", skillID).
		WithMetadata("cooldown_seconds", remainingSeconds) // 现在可以直接存储 int
}

func NewClassNotFoundError(classID string) *AppError {
	return FromCode(CodeClassNotFound).
		WithMetadata("class_id", classID)
}

// 通用错误包装函数
// Wrap 包装标准错误为 AppError(保留堆栈)
func Wrap(err error, code ErrorCode, message string) *AppError {
	if err == nil {
		return nil
	}

	// 如果已经是 AppError,直接返回
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}

	return NewWithError(code, message, err)
}

// WrapWithContext 包装错误并添加上下文
func WrapWithContext(err error, code ErrorCode, message string, ctx *ErrorContext) *AppError {
	appErr := Wrap(err, code, message)
	if appErr != nil {
		appErr.Context = ctx
	}
	return appErr
}

// Must 如果 err 不为 nil 就 panic (用于配置初始化等必须成功的场景)
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// ErrorList 错误列表(批量操作时收集多个错误)
type ErrorList struct {
	Errors []*AppError `json:"errors"`
}

func (e *ErrorList) Error() string {
	if len(e.Errors) == 0 {
		return "no errors"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("%d errors occurred", len(e.Errors))
}

func (e *ErrorList) Add(err *AppError) {
	if err != nil {
		e.Errors = append(e.Errors, err)
	}
}

func (e *ErrorList) HasErrors() bool {
	return len(e.Errors) > 0
}

func (e *ErrorList) First() *AppError {
	if len(e.Errors) > 0 {
		return e.Errors[0]
	}
	return nil
}

// NewErrorList 创建错误列表
func NewErrorList() *ErrorList {
	return &ErrorList{
		Errors: make([]*AppError, 0),
	}
}

// ==================== Kratos 认证服务专用错误 ====================

// NewKratosError 创建 Kratos 服务错误
func NewKratosError(operation string, err error) *AppError {
	appErr := FromCode(CodeKratosError).
		WithMetadata("kratos_operation", operation)
	if err != nil {
		appErr.Err = err
	}
	return appErr
}

// NewKratosAPIError 创建 Kratos API 错误（带状态码）
func NewKratosAPIError(operation string, statusCode int) *AppError {
	return FromCode(CodeKratosError).
		WithMetadata("kratos_operation", operation).
		WithMetadata("status_code", fmt.Sprintf("%d", statusCode))
}

// NewKratosErrorFromMessage 从 Kratos 错误消息创建 AppError（智能翻译）
// 这个函数会自动将 Kratos 的英文错误消息翻译成对应的业务错误码和中文消息
func NewKratosErrorFromMessage(operation string, kratosErrorMsg string, originalErr error) *AppError {
	// 使用智能翻译获取业务错误码和消息
	code, message := TranslateKratosErrorText(kratosErrorMsg)

	appErr := FromCode(code)
	appErr.Message = message // 使用翻译后的中文消息

	if originalErr != nil {
		appErr.Err = originalErr
	}

	return appErr.
		WithMetadata("kratos_operation", operation).
		WithMetadata("kratos_error_text", kratosErrorMsg)
}

// NewKratosErrorFromID 从 Kratos 错误 ID 创建 AppError
func NewKratosErrorFromID(operation string, kratosErrorID int, originalErr error) *AppError {
	// 使用精确映射获取业务错误码和消息
	code, message := TranslateKratosError(kratosErrorID)

	appErr := FromCode(code)
	appErr.Message = message

	if originalErr != nil {
		appErr.Err = originalErr
	}

	return appErr.
		WithMetadata("kratos_operation", operation).
		WithMetadata("kratos_error_id", fmt.Sprintf("%d", kratosErrorID))
}

// NewSessionInvalidError 创建会话无效错误
func NewSessionInvalidError(reason string) *AppError {
	return FromCode(CodeInvalidToken).
		WithMetadata("session_error", reason)
}

// NewSessionExpiredError 创建会话过期错误
func NewSessionExpiredError() *AppError {
	return FromCode(CodeSessionExpired)
}

// NewKratosClientNotInitializedError Kratos 客户端未初始化错误
func NewKratosClientNotInitializedError(clientType string) *AppError {
	return FromCode(CodeInternalError).
		WithMetadata("error_type", "kratos_client_not_initialized").
		WithMetadata("client_type", clientType).
		WithMetadata("hint", "请先调用相应的初始化方法")
}

// NewKratosDataIntegrityError Kratos 数据完整性错误
func NewKratosDataIntegrityError(field string, reason string) *AppError {
	return FromCode(CodeDataIntegrityError).
		WithMetadata("field", field).
		WithMetadata("reason", reason).
		WithMetadata("source", "kratos")
}
