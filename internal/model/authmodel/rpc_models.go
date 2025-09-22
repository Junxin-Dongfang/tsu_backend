// internal/model/authmodel/rpc_models.go - 补充完整的 RPC 模型
package authmodel

import "time"

// === 登录相关 ===
type LoginRPCRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
	ClientIP   string `json:"client_ip"`
	UserAgent  string `json:"user_agent"`
}

type LoginRPCResponse struct {
	Success      bool              `json:"success"`
	Token        string            `json:"token"`
	RefreshToken string            `json:"refresh_token,omitempty"`
	UserInfo     *BusinessUserInfo `json:"user_info"`
	ExpiresIn    int64             `json:"expires_in"`
}

// === 注册相关 ===
type RegisterRPCRequest struct {
	Email     string `json:"email"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	ClientIP  string `json:"client_ip"`
	UserAgent string `json:"user_agent"`
}

type RegisterRPCResponse struct {
	Success      bool              `json:"success"`
	IdentityID   string            `json:"identity_id"`
	Token        string            `json:"token"`
	RefreshToken string            `json:"refresh_token,omitempty"`
	UserInfo     *BusinessUserInfo `json:"user_info"`
}

// === Token 验证相关 ===
type ValidateTokenRequest struct {
	Token string `json:"token"`
}

type ValidateTokenResponse struct {
	Valid       bool              `json:"valid"`
	UserID      string            `json:"user_id"`
	UserInfo    *BusinessUserInfo `json:"user_info,omitempty"`
	Permissions []string          `json:"permissions,omitempty"`
	ExpiresAt   int64             `json:"expires_at,omitempty"`
}

// === 登出相关 ===
type LogoutRequest struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
}

type LogoutResponse struct {
	Success bool `json:"success"`
}

// === 权限检查相关 ===
type CheckPermissionRequest struct {
	UserID   string `json:"user_id"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

type CheckPermissionResponse struct {
	Allowed bool `json:"allowed"`
}

// === 用户信息相关 ===
type GetUserInfoRequest struct {
	UserID string `json:"user_id"`
}

type GetUserInfoResponse struct {
	UserInfo *BusinessUserInfo `json:"user_info"`
}

// === 更新用户特征相关 ===
type UpdateUserTraitsRequest struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type UpdateUserTraitsResponse struct {
	Success bool `json:"success"`
}

// === 角色管理相关 ===
type AssignRoleRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

type AssignRoleResponse struct {
	Success bool `json:"success"`
}

type RevokeRoleRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

type RevokeRoleResponse struct {
	Success bool `json:"success"`
}

type CreateRoleRequest struct {
	RoleName    string   `json:"role_name"`
	Permissions []string `json:"permissions"`
}

type CreateRoleResponse struct {
	Success bool `json:"success"`
}

// === 业务用户信息 ===
type BusinessUserInfo struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	IsPremium    bool      `json:"is_premium"`
	DiamondCount int       `json:"diamond_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
