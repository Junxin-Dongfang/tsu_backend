package entity

import "time"

// UserSettings 用户设置实体
type UserSettings struct {
	ID     int64  `db:"id" json:"id"`
	UserID string `db:"user_id" json:"user_id"`

	// 通知设置
	EmailNotifications bool `db:"email_notifications" json:"email_notifications"`
	SMSNotifications   bool `db:"sms_notifications" json:"sms_notifications"`
	PushNotifications  bool `db:"push_notifications" json:"push_notifications"`
	MarketingEmails    bool `db:"marketing_emails" json:"marketing_emails"`

	// 隐私设置
	ProfileVisibility   string `db:"profile_visibility" json:"profile_visibility"`   // public, friends, private
	ShowOnlineStatus    bool   `db:"show_online_status" json:"show_online_status"`
	AllowFriendRequests bool   `db:"allow_friend_requests" json:"allow_friend_requests"`

	// 安全设置
	TwoFactorEnabled bool `db:"two_factor_enabled" json:"two_factor_enabled"`
	LoginAlerts      bool `db:"login_alerts" json:"login_alerts"`

	// 其他设置
	Theme string `db:"theme" json:"theme"` // light, dark, auto

	// 时间戳
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// TableName 返回表名
func (UserSettings) TableName() string {
	return "user_settings"
}

// ProfileVisibilityEnum 个人资料可见性枚举
type ProfileVisibilityEnum string

const (
	ProfileVisibilityPublic  ProfileVisibilityEnum = "public"
	ProfileVisibilityFriends ProfileVisibilityEnum = "friends"
	ProfileVisibilityPrivate ProfileVisibilityEnum = "private"
)

// ThemeEnum 主题枚举
type ThemeEnum string

const (
	ThemeLight ThemeEnum = "light"
	ThemeDark  ThemeEnum = "dark"
	ThemeAuto  ThemeEnum = "auto"
)

// DefaultUserSettings 创建默认用户设置
func DefaultUserSettings(userID string) *UserSettings {
	return &UserSettings{
		UserID:              userID,
		EmailNotifications:  true,
		SMSNotifications:    false,
		PushNotifications:   true,
		MarketingEmails:     false,
		ProfileVisibility:   string(ProfileVisibilityPublic),
		ShowOnlineStatus:    true,
		AllowFriendRequests: true,
		TwoFactorEnabled:    false,
		LoginAlerts:         true,
		Theme:               string(ThemeLight),
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
}