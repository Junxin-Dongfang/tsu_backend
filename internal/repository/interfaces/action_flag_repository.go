package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// ActionFlagQueryParams 动作Flag查询参数
type ActionFlagQueryParams struct {
	Category     *string // 分类
	DurationType *string // 持续时间类型
	IsActive     *bool   // 是否启用
	Limit        int     // 每页数量
	Offset       int     // 偏移量
}

// ActionFlagRepository 动作Flag仓储接口
type ActionFlagRepository interface {
	// GetByID 根据ID获取动作Flag
	GetByID(ctx context.Context, flagID string) (*game_config.ActionFlag, error)

	// GetByCode 根据代码获取动作Flag
	GetByCode(ctx context.Context, code string) (*game_config.ActionFlag, error)

	// List 获取动作Flag列表
	List(ctx context.Context, params ActionFlagQueryParams) ([]*game_config.ActionFlag, int64, error)

	// Create 创建动作Flag
	Create(ctx context.Context, flag *game_config.ActionFlag) error

	// Update 更新动作Flag
	Update(ctx context.Context, flag *game_config.ActionFlag) error

	// Delete 软删除动作Flag
	Delete(ctx context.Context, flagID string) error

	// Exists 检查代码是否存在
	Exists(ctx context.Context, code string) (bool, error)
}
