// Package domain 定义了login-shim服务内部使用的领域模型
// 这些结构体在controller、adapter等不同层之间共享，避免重复定义
package domain

// LoginRequest 定义了登录接口接收的JSON结构体
// 在controller层接收请求，在adapter层处理业务逻辑时都会使用此结构
type LoginRequest struct {
	Identifier string `json:"identifier" validate:"required"`
	Password   string `json:"password"   validate:"required,min=8"`
}

// RegisterRequest 定义了注册接口接收的JSON结构体
// 在controller层接收请求，在adapter层处理业务逻辑时都会使用此结构
type RegisterRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Phone    string `json:"phone"    validate:"omitempty,e164"`
	UserName string `json:"username" validate:"required,alphanum,min=3,max=30"`
}

// AuthResponse 定义了认证相关接口的通用响应结构
// 可以根据需要扩展，比如返回用户信息、token等
type AuthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	// 可以根据需要添加更多字段，比如：
	// UserID   string `json:"user_id,omitempty"`
	// Token    string `json:"token,omitempty"`
}

// Session 定义了登录成功后返回的会话信息。
// 目前我们只关心 cookie，但未来可能会有 token 等。
type Session struct {
	Cookie string
}
