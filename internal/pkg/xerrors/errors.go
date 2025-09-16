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

// 辅助函数
// getCategoryByCode 根据错误码获取分类
func getCategoryByCode(code int) string {
	switch {
	case code >= 100000 && code < 200000:
		return "system"
	case code >= 200000 && code < 300000:
		return "authentication"
	case code >= 300000 && code < 400000:
		return "authorization"
	case code >= 400000 && code < 500000:
		return "user"
	case code >= 500000 && code < 600000:
		return "role"
	case code >= 600000 && code < 700000:
		return "business"
	case code >= 700000 && code < 800000:
		return "external"
	case code >= 800000 && code < 900000:
		return "game"
	default:
		return "unknown"
	}
}

// getLevelByCode 根据错误码获取级别
func getLevelByCode(code int) ErrorLevel {
	switch {
	case code == CodeSuccess:
		return LevelInfo
	case code >= 100001 && code <= 100003: // 参数错误等
		return LevelWarn
	case code >= 700001: // 外部服务错误
		return LevelCritical
	default:
		return LevelError
	}
}

// isRetryableByCode 根据错误码判断是否可重试
func isRetryableByCode(code int) bool {
	retryableCodes := map[int]bool{
		CodeInternalError:        true,
		CodeExternalServiceError: true,
		CodeKratosError:          true,
		CodeDatabaseError:        true,
		CodeCacheError:           true,
		CodeMessageQueueError:    true,
		CodeRateLimitExceeded:    true,
	}
	return retryableCodes[code]
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

// -----------------------------------------------------------------------------
// HTTP 状态码常量定义
// -----------------------------------------------------------------------------
const (
	HTTPStatusOK        = 200 // 请求成功
	HTTPStatusCreated   = 201 // 资源已创建
	HTTPStatusAccepted  = 202 // 请求已接受但未处理
	HTTPStatusNoContent = 204 // 请求成功但无内容返回

	HTTPStatusBadRequest          = 400 // 错误请求
	HTTPStatusUnauthorized        = 401 // 未经授权
	HTTPStatusForbidden           = 403 // 禁止访问
	HTTPStatusNotFound            = 404 // 资源未找到
	HTTPStatusMethodNotAllowed    = 405 // 方法不被允许
	HTTPStatusConflict            = 409 // 资源冲突
	HTTPStatusUnprocessableEntity = 422 // 无法处理的实体
	HTTPStatusTooManyRequests     = 429 // 请求过多

	HTTPStatusInternalServerError = 500 // 内部服务器错误
	HTTPStatusNotImplemented      = 501 // 未实现
	HTTPStatusBadGateway          = 502 // 错误网关
	HTTPStatusServiceUnavailable  = 503 // 服务不可用
	HTTPStatusGatewayTimeout      = 504 // 网关超时
)

func GetHTTPStatus(code int) int {
	switch {
	case code == CodeSuccess:
		return HTTPStatusOK
	case code >= 200000 && code < 300000:
		if code == CodeAuthenticationFailed || code == CodeInvalidToken || code == CodeTokenExpired || code == CodeInvalidCredentials {
			return HTTPStatusUnauthorized
		}
		return HTTPStatusForbidden
	case code >= 300000 && code < 400000:
		return HTTPStatusForbidden
	case code >= 400000 && code < 500000:
		if code == CodeUserNotFound {
			return HTTPStatusNotFound
		}
		if code == CodeUserAlreadyExists || code == CodeUsernameExists || code == CodeEmailExists || code == CodePhoneExists {
			return HTTPStatusConflict
		}
		return HTTPStatusBadRequest
	case code == CodeResourceNotFound:
		return HTTPStatusNotFound
	case code == CodeDuplicateResource:
		return HTTPStatusConflict
	case code == CodeInvalidParams || code == CodeInvalidRequest:
		return HTTPStatusBadRequest
	case code == CodeRateLimitExceeded:
		return HTTPStatusTooManyRequests
	case code >= 500000 && code < 600000:
		return HTTPStatusBadRequest
	case code >= 600000 && code < 700000:
		return HTTPStatusBadRequest
	case code >= 700000:
		return HTTPStatusServiceUnavailable
	default:
		return HTTPStatusInternalServerError
	}
}

// -----------------------------------------------------------------------------
// 业务错误码统一定义
// 按模块或领域对错误码进行分段，便于管理。
// -----------------------------------------------------------------------------
const (
	// 1xxxxx: 通用错误码
	CodeSuccess           = 100000 // 操作成功
	CodeInternalError     = 100001 // 内部服务错误
	CodeInvalidParams     = 100002 // 参数错误
	CodeInvalidRequest    = 100003 // 请求格式错误
	CodeResourceNotFound  = 100404 // 资源不存在
	CodeDuplicateResource = 100409 // 资源已存在
	CodeRateLimitExceeded = 100429 // 请求频率限制

	// 2xxxxx: 认证相关错误码
	CodeAuthenticationFailed = 200001 // 认证失败
	CodeInvalidToken         = 200002 // 无效令牌
	CodeTokenExpired         = 200003 // 令牌过期
	CodeInvalidCredentials   = 200004 // 凭据无效
	CodeAccountLocked        = 200005 // 账户被锁定
	CodeAccountBanned        = 200006 // 账户被封禁
	CodeSessionExpired       = 200007 // 会话过期

	// 3xxxxx: 权限相关错误码
	CodePermissionDenied       = 300001 // 权限不足
	CodeInsufficientPrivileges = 300002 // 权限级别不够
	CodeRoleNotAssigned        = 300003 // 角色未分配
	CodePermissionNotExists    = 300004 // 权限不存在

	// 4xxxxx: 用户管理错误码
	CodeUserNotFound      = 400001 // 用户不存在
	CodeUserAlreadyExists = 400002 // 用户已存在
	CodeUsernameExists    = 400003 // 用户名已存在
	CodeEmailExists       = 400004 // 邮箱已存在
	CodePhoneExists       = 400005 // 手机号已存在
	CodeInvalidUserStatus = 400006 // 用户状态无效
	CodeUserBanned        = 400007 // 用户被封禁

	// 5xxxxx: 角色权限管理错误码
	CodeRoleNotFound           = 500001 // 角色不存在
	CodeRoleAlreadyExists      = 500002 // 角色已存在
	CodeRoleInUse              = 500003 // 角色正在使用中
	CodeSystemRoleProtected    = 500004 // 系统角色受保护
	CodePermissionAssignFailed = 500005 // 权限分配失败

	// 6xxxxx: 业务逻辑错误码
	CodeBusinessLogicError  = 600001 // 业务逻辑错误
	CodeDataIntegrityError  = 600002 // 数据完整性错误
	CodeOperationNotAllowed = 600003 // 操作不被允许
	CodeResourceLocked      = 600004 // 资源被锁定
	CodeQuotaExceeded       = 600005 // 配额超限

	// 7xxxxx: 外部服务错误码
	CodeExternalServiceError = 700001 // 外部服务错误
	CodeKratosError          = 700002 // Kratos服务错误
	CodeDatabaseError        = 700003 // 数据库错误
	CodeCacheError           = 700004 // 缓存服务错误
	CodeMessageQueueError    = 700005 // 消息队列错误
)

// -----------------------------------------------------------------------------
// 错误消息映射
// -----------------------------------------------------------------------------
var codeMessages = map[int]string{
	CodeSuccess:           "操作成功",
	CodeInternalError:     "内部服务错误",
	CodeInvalidParams:     "参数错误",
	CodeInvalidRequest:    "请求格式错误",
	CodeResourceNotFound:  "资源不存在",
	CodeDuplicateResource: "资源已存在",
	CodeRateLimitExceeded: "请求频率限制",

	CodeAuthenticationFailed: "认证失败",
	CodeInvalidToken:         "无效令牌",
	CodeTokenExpired:         "令牌过期",
	CodeInvalidCredentials:   "凭据无效",
	CodeAccountLocked:        "账户被锁定",
	CodeAccountBanned:        "账户被封禁",
	CodeSessionExpired:       "会话过期",

	CodePermissionDenied:       "权限不足",
	CodeInsufficientPrivileges: "权限级别不够",
	CodeRoleNotAssigned:        "角色未分配",
	CodePermissionNotExists:    "权限不存在",

	CodeUserNotFound:      "用户不存在",
	CodeUserAlreadyExists: "用户已存在",
	CodeUsernameExists:    "用户名已存在",
	CodeEmailExists:       "邮箱已存在",
	CodePhoneExists:       "手机号已存在",
	CodeInvalidUserStatus: "用户状态无效",
	CodeUserBanned:        "用户被封禁",

	CodeRoleNotFound:           "角色不存在",
	CodeRoleAlreadyExists:      "角色已存在",
	CodeRoleInUse:              "角色正在使用中",
	CodeSystemRoleProtected:    "系统角色受保护",
	CodePermissionAssignFailed: "权限分配失败",

	CodeBusinessLogicError:  "业务逻辑错误",
	CodeDataIntegrityError:  "数据完整性错误",
	CodeOperationNotAllowed: "操作不被允许",
	CodeResourceLocked:      "资源被锁定",
	CodeQuotaExceeded:       "配额超限",

	CodeExternalServiceError: "外部服务错误",
	CodeKratosError:          "Kratos服务错误",
	CodeDatabaseError:        "数据库错误",
	CodeCacheError:           "缓存服务错误",
	CodeMessageQueueError:    "消息队列错误",
}
