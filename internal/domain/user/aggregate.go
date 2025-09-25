package user

import (
	"time"
	"tsu-self/internal/entity"
)

// UserAggregate 用户聚合根 - 组合用户基础信息和财务信息
type UserAggregate struct {
	*entity.User
	Finance *entity.UserFinance `json:"finance,omitempty"`
}

// IsPremium 检查用户是否为会员
func (u *UserAggregate) IsPremium() bool {
	if u.Finance == nil {
		return false
	}
	if !u.Finance.PremiumExpiry.Valid {
		return false
	}
	return u.Finance.PremiumExpiry.Time.After(time.Now())
}

// GetDiamondCount 获取钻石数量
func (u *UserAggregate) GetDiamondCount() int {
	if u.Finance == nil {
		return 0
	}
	if !u.Finance.CurrentDiamonds.Valid {
		return 0
	}
	return int(u.Finance.CurrentDiamonds.Int64)
}

// GetTotalDiamonds 获取总钻石数量
func (u *UserAggregate) GetTotalDiamonds() int {
	if u.Finance == nil {
		return 0
	}
	if !u.Finance.TotalDiamonds.Valid {
		return 0
	}
	return int(u.Finance.TotalDiamonds.Int64)
}