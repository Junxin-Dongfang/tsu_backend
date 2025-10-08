package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// EffectQueryParams 效果查询参数
type EffectQueryParams struct {
	EffectType *string // 效果类型
	IsActive   *bool   // 是否启用
	Limit      int     // 每页数量
	Offset     int     // 偏移量
}

// EffectRepository 效果仓储接口
type EffectRepository interface {
	// GetByID 根据ID获取效果
	GetByID(ctx context.Context, effectID string) (*game_config.Effect, error)

	// GetByCode 根据代码获取效果
	GetByCode(ctx context.Context, code string) (*game_config.Effect, error)

	// List 获取效果列表
	List(ctx context.Context, params EffectQueryParams) ([]*game_config.Effect, int64, error)

	// Create 创建效果
	Create(ctx context.Context, effect *game_config.Effect) error

	// Update 更新效果
	Update(ctx context.Context, effect *game_config.Effect) error

	// Delete 软删除效果
	Delete(ctx context.Context, effectID string) error

	// Exists 检查代码是否存在
	Exists(ctx context.Context, code string) (bool, error)
}
