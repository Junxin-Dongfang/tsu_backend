package user

import "time"

// Profile HTTP API 用户信息响应
// @Description 用户基本信息
type Profile struct {
	ID           string    `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Username     string    `json:"username" example:"johndoe"`
	Nickname     string    `json:"nickname" example:"Johnny"`
	Phone        string    `json:"phone" example:"11234567890"`
	Email        string    `json:"email" example:"johndoe@example.com"`
	IsPremium    bool      `json:"is_premium" example:"true"`
	DiamondCount int       `json:"diamond_count" example:"100"`
	CreatedAt    time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt    time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z"`
} // @name UserProfile
