package interfaces

import (
	"context"

	"tsu-self/internal/entity/auth"
)

// UserQueryParams 用户查询参数
type UserQueryParams struct {
	Page     int
	PageSize int
	Keyword  string
	IsBanned *bool
	SortBy   string
	SortDir  string
}

// UserRepository 用户仓储接口
type UserRepository interface {
	// GetByID 根据ID获取用户
	GetByID(ctx context.Context, userID string) (*auth.User, error)

	// GetByUsername 根据用户名获取用户
	GetByUsername(ctx context.Context, username string) (*auth.User, error)

	// GetByEmail 根据邮箱获取用户
	GetByEmail(ctx context.Context, email string) (*auth.User, error)

	// List 获取用户列表（分页）
	List(ctx context.Context, params UserQueryParams) ([]*auth.User, int64, error)

	// Create 创建用户
	Create(ctx context.Context, user *auth.User) error

	// Update 更新用户信息
	Update(ctx context.Context, user *auth.User) error

	// BanUser 封禁用户
	BanUser(ctx context.Context, userID string, banUntil *string, banReason string) error

	// UnbanUser 解禁用户
	UnbanUser(ctx context.Context, userID string) error

	// UpdateLoginInfo 更新登录信息
	UpdateLoginInfo(ctx context.Context, userID string, loginIP string) error

	// Exists 检查用户是否存在
	Exists(ctx context.Context, userID string) (bool, error)
}
