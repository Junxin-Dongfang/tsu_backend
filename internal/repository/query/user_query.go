package query

import "time"

// UserQuery 用户查询参数
type UserQuery struct {
	// 基本过滤
	IDs       []string `json:"ids,omitempty"`
	Usernames []string `json:"usernames,omitempty"`
	Emails    []string `json:"emails,omitempty"`

	// 状态过滤
	IsPremium *bool `json:"is_premium,omitempty"`
	IsBanned  *bool `json:"is_banned,omitempty"`

	// 时间范围过滤
	CreatedAfter  *time.Time `json:"created_after,omitempty"`
	CreatedBefore *time.Time `json:"created_before,omitempty"`

	// 搜索
	SearchTerm string `json:"search_term,omitempty"` // 支持用户名、邮箱、显示名称搜索

	// 排序
	OrderBy   string `json:"order_by,omitempty"`   // created_at, updated_at, username, email
	OrderDesc bool   `json:"order_desc,omitempty"` // 是否降序

	// 分页
	Pagination
}

// LoginHistoryQuery 登录历史查询参数
type LoginHistoryQuery struct {
	// 基本过滤
	UserIDs []string `json:"user_ids,omitempty"`

	// IP 过滤
	IPAddresses []string `json:"ip_addresses,omitempty"`

	// 状态过滤
	Status []string `json:"status,omitempty"` // success, failed, blocked

	// 时间范围过滤
	LoginAfter  *time.Time `json:"login_after,omitempty"`
	LoginBefore *time.Time `json:"login_before,omitempty"`

	// 安全过滤
	IsSuspicious *bool `json:"is_suspicious,omitempty"`

	// 设备类型过滤
	DeviceTypes []string `json:"device_types,omitempty"`

	// 登录方式过滤
	LoginMethods []string `json:"login_methods,omitempty"`

	// 排序
	OrderBy   string `json:"order_by,omitempty"`   // login_time, risk_score
	OrderDesc bool   `json:"order_desc,omitempty"` // 是否降序

	// 分页
	Pagination
}