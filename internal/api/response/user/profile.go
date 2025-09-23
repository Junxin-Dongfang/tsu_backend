package user

import "time"

// Profile HTTP API 用户信息响应
// @Description 用户基本信息
type Profile struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	IsPremium    bool      `json:"is_premium"`
	DiamondCount int       `json:"diamond_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
} // @name UserProfile