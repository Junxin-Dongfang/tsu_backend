package auth

import "tsu-self/internal/api/response/user"

// RegisterResult HTTP API 注册响应
// @Description 用户注册结果
type RegisterResult struct {
	Success       bool              `json:"success"`
	IdentityID    string            `json:"identity_id"`
	SessionToken  string            `json:"session_token"`
	SessionCookie string            `json:"session_cookie"`
	UserInfo      *user.Profile     `json:"user_info,omitempty"`
	ErrorMessage  string            `json:"error_message,omitempty"`
} // @name RegisterResult