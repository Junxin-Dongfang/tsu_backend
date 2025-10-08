package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// BuffEffectRepository Buff效果关联仓储接口
type BuffEffectRepository interface {
	// GetByBuffID 获取Buff的所有效果
	GetByBuffID(ctx context.Context, buffID string) ([]*game_config.BuffEffect, error)

	// Create 创建Buff效果关联
	Create(ctx context.Context, buffEffect *game_config.BuffEffect) error

	// Delete 删除Buff效果关联
	Delete(ctx context.Context, buffEffectID string) error

	// DeleteByBuffAndEffect 删除指定Buff和效果的关联
	DeleteByBuffAndEffect(ctx context.Context, buffID, effectID, triggerTiming string) error

	// DeleteAllByBuffID 删除Buff的所有效果关联
	DeleteAllByBuffID(ctx context.Context, buffID string) error

	// BatchCreate 批量创建Buff效果关联
	BatchCreate(ctx context.Context, buffEffects []*game_config.BuffEffect) error
}
