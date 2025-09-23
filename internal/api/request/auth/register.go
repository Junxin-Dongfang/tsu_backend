package auth

// RegisterRequest HTTP API 注册请求
// @Description 用户注册请求参数
type RegisterRequest struct {
	// 邮箱
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
	// 用户名
	Username string `json:"username" binding:"required,min=3,max=30" example:"newuser"`
	// 密码
	Password string `json:"password" binding:"required,min=8" example:"password123"`
	// 客户端IP (由服务器自动填充)
	ClientIP string `json:"-"`
	// 用户代理 (由服务器自动填充)
	UserAgent string `json:"-"`
} // @name RegisterRequest