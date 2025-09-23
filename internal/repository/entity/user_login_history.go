package entity

import (
	"database/sql"
	"time"
)

// UserLoginHistory 用户登录历史实体
type UserLoginHistory struct {
	ID     int64  `db:"id" json:"id"`
	UserID string `db:"user_id" json:"user_id"`

	// 登录信息
	LoginTime       time.Time      `db:"login_time" json:"login_time"`
	LogoutTime      sql.NullTime   `db:"logout_time" json:"logout_time,omitempty"`
	SessionDuration sql.NullInt32  `db:"session_duration" json:"session_duration,omitempty"` // 会话时长（秒）

	// 客户端信息
	IPAddress      string         `db:"ip_address" json:"ip_address"`
	UserAgent      sql.NullString `db:"user_agent" json:"user_agent,omitempty"`
	DeviceType     sql.NullString `db:"device_type" json:"device_type,omitempty"`
	BrowserName    sql.NullString `db:"browser_name" json:"browser_name,omitempty"`
	BrowserVersion sql.NullString `db:"browser_version" json:"browser_version,omitempty"`
	OSName         sql.NullString `db:"os_name" json:"os_name,omitempty"`
	OSVersion      sql.NullString `db:"os_version" json:"os_version,omitempty"`

	// 地理位置
	Country sql.NullString `db:"country" json:"country,omitempty"`
	Region  sql.NullString `db:"region" json:"region,omitempty"`
	City    sql.NullString `db:"city" json:"city,omitempty"`

	// 登录方式
	LoginMethod   string         `db:"login_method" json:"login_method"`   // password, oauth, sms, etc.
	OAuthProvider sql.NullString `db:"oauth_provider" json:"oauth_provider,omitempty"` // google, facebook, github, etc.

	// 安全信息
	IsSuspicious bool           `db:"is_suspicious" json:"is_suspicious"`
	RiskScore    sql.NullInt32  `db:"risk_score" json:"risk_score,omitempty"`

	// 状态
	Status string `db:"status" json:"status"` // success, failed, blocked
}

// TableName 返回表名
func (UserLoginHistory) TableName() string {
	return "user_login_history"
}

// IsSuccessful 检查登录是否成功
func (h *UserLoginHistory) IsSuccessful() bool {
	return h.Status == "success"
}

// GetSessionDuration 计算会话时长
func (h *UserLoginHistory) GetSessionDuration() time.Duration {
	if h.LogoutTime.Valid {
		return h.LogoutTime.Time.Sub(h.LoginTime)
	}
	return 0
}

// DeviceTypeEnum 设备类型枚举
type DeviceTypeEnum string

const (
	DeviceTypeMobile  DeviceTypeEnum = "mobile"
	DeviceTypeDesktop DeviceTypeEnum = "desktop"
	DeviceTypeTablet  DeviceTypeEnum = "tablet"
	DeviceTypeOther   DeviceTypeEnum = "other"
)

// LoginMethodEnum 登录方式枚举
type LoginMethodEnum string

const (
	LoginMethodPassword  LoginMethodEnum = "password"
	LoginMethodOAuth     LoginMethodEnum = "oauth"
	LoginMethodSMS       LoginMethodEnum = "sms"
	LoginMethodMagicLink LoginMethodEnum = "magic_link"
	LoginMethodBiometric LoginMethodEnum = "biometric"
)

// LoginStatusEnum 登录状态枚举
type LoginStatusEnum string

const (
	LoginStatusSuccess LoginStatusEnum = "success"
	LoginStatusFailed  LoginStatusEnum = "failed"
	LoginStatusBlocked LoginStatusEnum = "blocked"
)