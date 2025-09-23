package authmodel

// -------------------------------请求模型-------------------------------

// LoginRequest 登录请求
// @Description 用户登录请求参数
type LoginRequest struct {
	// 可以是用户名或邮箱
	Identifier string `json:"identifier" binding:"required" validate:"required" example:"user@example.com"`
	// 密码
	Password string `json:"password" binding:"required" validate:"required,min=8" example:"password123"`
} // @name LoginRequest

// RegisterRequest 注册请求
// @Description 用户注册请求参数
type RegisterRequest struct {
	// 邮箱
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
	// 用户名
	Username string `json:"username" binding:"required,min=3,max=30" example:"newuser"`
	// 密码
	Password string `json:"password" binding:"required,min=8" example:"password123"`
} // @name RegisterRequest

// RecoveryRequest 恢复请求
// @Description 用户密码恢复请求参数
type RecoveryRequest struct {
	// 邮箱
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
} // @name RecoveryRequest

// -------------------------------响应模型-------------------------------
// LoginResult 登录结果
// @Description 用户登录结果
type LoginResult struct {
	Success       bool                 `json:"success"`
	SessionToken  string               `json:"session_token"`
	SessionCookie string               `json:"session_cookie"`
	UserInfo      *BusinessUserInfo    `json:"user_info,omitempty"`
	ErrorMessage  string               `json:"error_message,omitempty"`
} // @name LoginResult

// RegisterResult 注册结果
// @Description 用户注册结果
type RegisterResult struct {
	Success       bool                 `json:"success"`
	IdentityID    string               `json:"identity_id"`
	SessionToken  string               `json:"session_token"`
	SessionCookie string               `json:"session_cookie"`
	UserInfo      *BusinessUserInfo    `json:"user_info,omitempty"`
	ErrorMessage  string               `json:"error_message,omitempty"`
} // @name RegisterResult
