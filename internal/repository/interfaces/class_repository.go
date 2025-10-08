package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// ClassQueryParams 职业查询参数
type ClassQueryParams struct {
	Page         int
	PageSize     int
	Tier         string // basic, advanced, elite, legendary, mythic
	IsActive     *bool
	IsVisible    *bool
	SortBy       string // class_name, tier, display_order, created_at
	SortDir      string // ASC, DESC
}

// ClassRepository 职业仓储接口
type ClassRepository interface {
	// GetByID 根据ID获取职业
	GetByID(ctx context.Context, classID string) (*game_config.Class, error)

	// GetByCode 根据职业代码获取职业
	GetByCode(ctx context.Context, classCode string) (*game_config.Class, error)

	// List 获取职业列表（分页）
	List(ctx context.Context, params ClassQueryParams) ([]*game_config.Class, int64, error)

	// Create 创建职业
	Create(ctx context.Context, class *game_config.Class) error

	// Update 更新职业信息
	Update(ctx context.Context, class *game_config.Class) error

	// Delete 软删除职业
	Delete(ctx context.Context, classID string) error

	// Exists 检查职业是否存在
	Exists(ctx context.Context, classID string) (bool, error)
}
