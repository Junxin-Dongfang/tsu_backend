package auth

// LoginRequest HTTP API 登录请求
// @Description 用户登录请求参数
type LoginRequest struct {
	// 可以是用户名或邮箱
	Identifier string `json:"identifier" binding:"required" validate:"required" example:"user@example.com"`
	// 密码
	Password string `json:"password" binding:"required" validate:"required,min=8" example:"password123"`
	// 客户端IP (由服务器自动填充)
	ClientIP string `json:"-"`
	// 用户代理 (由服务器自动填充)
	UserAgent string `json:"-"`
} // @name LoginRequest