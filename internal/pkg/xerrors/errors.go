// File: internal/pkg/xerrors/errors.go
package xerrors

import "fmt"

// AppError 是我们自定义的、包含丰富上下文的错误结构体。
// 它实现了标准的 `error` 接口。
// 在整个项目中，业务逻辑层应该优先返回这个类型的错误。
type AppError struct {
	// Code 是我们定义的、对机器友好的业务错误码。
	// 前端可以根据这个码来进行特定的UI处理，例如跳转页面或显示特定提示。
	Code int `json:"code"`

	// Message 是对错误的简短、对用户友好的描述。
	// 这个消息可以直接显示给最终用户。
	Message string `json:"message"`

	// Err 是被包裹的、底层的原始Go错误。
	// 在日志中记录这个错误对于调试至关重要，但通常不应该直接暴露给最终用户。
	Err error `json:"-"` // json:"-" 确保在序列化时忽略这个字段
}

// Error 方法实现了标准的 `error` 接口，使得 AppError 可以像普通 error 一样使用。
func (e *AppError) Error() string {
	if e.Err != nil {
		// 返回原始错误信息，便于日志记录和调试
		return e.Err.Error()
	}
	return e.Message
}

// Unwrap 方法用于错误链 (Error Chaining, Go 1.13+)，
// 允许使用 `errors.Is` 和 `errors.As` 来检查被包裹的原始错误。
func (e *AppError) Unwrap() error {
	return e.Err
}

// New 是一个简单的构造函数，用于创建一个新的 AppError。
func New(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

// -----------------------------------------------------------------------------
// 业务错误码统一定义
// 建议按模块或领域对错误码进行分段，便于管理。
// -----------------------------------------------------------------------------
const (
	// --- 通用错误码 (1xxxxx) ---
	OK                = 100000 // 操作成功 (虽然不是错误，但放在这里便于对应)
	Unknown           = 100001 // 未知错误
	Validation        = 100002 // 通用参数验证错误
	Unauthorized      = 100003 // 未经授权
	PermissionDenied  = 100004 // 权限不足
	NotFound          = 100005 // 资源未找到
	RateLimitExceeded = 100006 // 请求频率过高

	// --- 认证/用户相关错误码 (2xxxxx) ---
	InvalidCredentials      = 200001 // 凭证无效 (用户名或密码错误)
	IdentityAlreadyExists   = 200002 // 身份已存在 (邮箱/用户名)
	PasswordPolicyError     = 200003 // 不符合密码策略
	PasswordTooShort        = 200004 // 密码长度不足
	EmailFormatError        = 200005 // 电子邮箱地址格式不正确
	SessionAlreadyAvailable = 200006 // 用户已经登录
	SessionNotFound         = 200007 // 会话不存在或已过期

	// --- 数据库相关错误码 (3xxxxx) ---
	DatabaseError     = 300001 // 数据库通用错误
	DuplicateKeyError = 300002 // 主键或唯一索引冲突

	// ... 在这里为您的其他业务模块定义更多的错误码区间 ...
)

// -----------------------------------------------------------------------------
// 错误构造辅助函数 (可选，但强烈推荐)
// 这些函数可以简化业务代码中创建特定错误的过程。
// -----------------------------------------------------------------------------

func E(code int, message string) *AppError {
	return New(code, message, nil)
}

func E_INVALID_CREDENTIALS(err error) *AppError {
	return New(InvalidCredentials, "用户名或密码不正确", err)
}

func E_IDENTITY_EXISTS(err error) *AppError {
	return New(IdentityAlreadyExists, "该用户已存在", err)
}

func E_NOT_FOUND(resource string, err error) *AppError {
	return New(NotFound, fmt.Sprintf("%s 不存在", resource), err)
}

func E_INTERNAL(err error) *AppError {
	return New(Unknown, "系统内部错误，请稍后重试", err)
}

func E_VALIDATION(err error) *AppError {
	return New(Validation, "请求参数有误", err)
}
