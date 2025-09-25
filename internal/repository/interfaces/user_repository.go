package interfaces

import (
	"context"
	"tsu-self/internal/entity"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	// 基础CRUD
	GetByID(ctx context.Context, id string) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id string) error

	// 业务查询
	GetActiveUsers(ctx context.Context, limit, offset int) ([]*entity.User, error)
	GetPremiumUsers(ctx context.Context) ([]*entity.User, error)
	GetBannedUsers(ctx context.Context) ([]*entity.User, error)

	// 统计查询
	CountActiveUsers(ctx context.Context) (int, error)
	CountPremiumUsers(ctx context.Context) (int, error)
}