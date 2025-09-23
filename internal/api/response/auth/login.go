package auth

import "tsu-self/internal/api/response/user"

// LoginResult HTTP API 登录响应
// @Description 用户登录结果
type LoginResult struct {
	Success       bool              `json:"success"`
	SessionToken  string            `json:"session_token"`
	SessionCookie string            `json:"session_cookie"`
	UserInfo      *user.Profile     `json:"user_info,omitempty"`
	ErrorMessage  string            `json:"error_message,omitempty"`
} // @name LoginResult