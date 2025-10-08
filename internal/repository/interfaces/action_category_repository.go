package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// ActionCategoryQueryParams 动作类别查询参数
type ActionCategoryQueryParams struct {
	IsActive *bool // 是否启用
	Limit    int   // 每页数量
	Offset   int   // 偏移量
}

// ActionCategoryRepository 动作类别仓储接口
type ActionCategoryRepository interface {
	// GetByID 根据ID获取动作类别
	GetByID(ctx context.Context, categoryID string) (*game_config.ActionCategory, error)

	// GetByCode 根据代码获取动作类别
	GetByCode(ctx context.Context, code string) (*game_config.ActionCategory, error)

	// List 获取动作类别列表
	List(ctx context.Context, params ActionCategoryQueryParams) ([]*game_config.ActionCategory, int64, error)

	// Create 创建动作类别
	Create(ctx context.Context, category *game_config.ActionCategory) error

	// Update 更新动作类别
	Update(ctx context.Context, category *game_config.ActionCategory) error

	// Delete 软删除动作类别
	Delete(ctx context.Context, categoryID string) error

	// Exists 检查代码是否存在
	Exists(ctx context.Context, code string) (bool, error)
}
