package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// ActionQueryParams 动作查询参数
type ActionQueryParams struct {
	ActionType *string // 动作类型
	CategoryID *string // 分类ID
	IsActive   *bool   // 是否启用
	Limit      int     // 每页数量
	Offset     int     // 偏移量
}

// ActionRepository 动作仓储接口
type ActionRepository interface {
	// GetByID 根据ID获取动作
	GetByID(ctx context.Context, actionID string) (*game_config.Action, error)

	// GetByIDs 批量根据ID获取动作
	GetByIDs(ctx context.Context, actionIDs []string) ([]*game_config.Action, error)

	// GetByCode 根据代码获取动作
	GetByCode(ctx context.Context, code string) (*game_config.Action, error)

	// List 获取动作列表
	List(ctx context.Context, params ActionQueryParams) ([]*game_config.Action, int64, error)

	// Create 创建动作
	Create(ctx context.Context, action *game_config.Action) error

	// Update 更新动作
	Update(ctx context.Context, action *game_config.Action) error

	// Delete 软删除动作
	Delete(ctx context.Context, actionID string) error

	// Exists 检查代码是否存在
	Exists(ctx context.Context, code string) (bool, error)
}
