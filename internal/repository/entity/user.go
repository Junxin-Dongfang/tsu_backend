package entity

import (
	"database/sql"
	"time"
)

// User 数据库用户实体
type User struct {
	// 主键：与 Kratos identity ID 对应
	ID string `db:"id" json:"id"`

	// 业务核心字段
	IsPremium    bool `db:"is_premium" json:"is_premium"`
	DiamondCount int  `db:"diamond_count" json:"diamond_count"`

	// 用户信息
	Username    string         `db:"username" json:"username"`
	Nickname    sql.NullString `db:"nickname" json:"nickname,omitempty"`
	Email       string         `db:"email" json:"email"`
	PhoneNumber sql.NullString `db:"phone_number" json:"phone_number,omitempty"`

	// 用户状态管理
	IsBanned  bool           `db:"is_banned" json:"is_banned"`
	BanUntil  sql.NullTime   `db:"ban_until" json:"ban_until,omitempty"`
	BanReason sql.NullString `db:"ban_reason" json:"ban_reason,omitempty"`

	// 个人资料
	AvatarURL   sql.NullString `db:"avatar_url" json:"avatar_url,omitempty"`
	Bio         sql.NullString `db:"bio" json:"bio,omitempty"`
	DisplayName sql.NullString `db:"display_name" json:"display_name,omitempty"`
	BirthDate   sql.NullTime   `db:"birth_date" json:"birth_date,omitempty"`
	Gender      sql.NullString `db:"gender" json:"gender,omitempty"`
	Timezone    string         `db:"timezone" json:"timezone"`
	Language    string         `db:"language" json:"language"`

	// 业务统计
	TotalSpent    float64        `db:"total_spent" json:"total_spent"`
	ReferralCode  sql.NullString `db:"referral_code" json:"referral_code,omitempty"`
	ReferredBy    sql.NullString `db:"referred_by" json:"referred_by,omitempty"`
	ReferralCount int            `db:"referral_count" json:"referral_count"`

	// 登录追踪
	LastLoginAt sql.NullTime   `db:"last_login_at" json:"last_login_at,omitempty"`
	LastLoginIP sql.NullString `db:"last_login_ip" json:"last_login_ip,omitempty"`
	LoginCount  int            `db:"login_count" json:"login_count"`

	// 时间戳
	CreatedAt time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt time.Time    `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at" json:"deleted_at,omitempty"`
}

// TableName 返回表名
func (User) TableName() string {
	return "users"
}

// IsDeleted 检查用户是否被软删除
func (u *User) IsDeleted() bool {
	return u.DeletedAt.Valid
}

// IsActive 检查用户是否活跃（未被删除且未被封禁）
func (u *User) IsActive() bool {
	if u.IsDeleted() || u.IsBanned {
		return false
	}

	// 检查封禁是否已过期
	if u.BanUntil.Valid && u.BanUntil.Time.After(time.Now()) {
		return false
	}

	return true
}
