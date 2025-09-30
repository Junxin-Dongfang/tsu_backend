package auth

// RecoveryRequest HTTP API 密码恢复请求
// @Description 用户密码恢复请求参数
type RecoveryRequest struct {
	// 邮箱
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
} // @name RecoveryRequest