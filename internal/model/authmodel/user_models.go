// internal/model/authmodel/user_models.go - 用户相关的数据模型
package authmodel

import "time"

// === 业务用户信息 ===
type BusinessUserInfo struct {
	ID           string    `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	Email        string    `json:"email" db:"email"`
	IsPremium    bool      `json:"is_premium" db:"is_premium"`
	DiamondCount int       `json:"diamond_count" db:"diamond_count"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
