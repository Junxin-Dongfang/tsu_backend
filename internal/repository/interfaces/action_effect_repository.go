package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// ActionEffectRepository 动作效果关联仓储接口
type ActionEffectRepository interface {
	// GetByActionID 获取动作的所有效果
	GetByActionID(ctx context.Context, actionID string) ([]*game_config.ActionEffect, error)

	// Create 创建动作效果关联
	Create(ctx context.Context, actionEffect *game_config.ActionEffect) error

	// Delete 删除动作效果关联
	Delete(ctx context.Context, actionEffectID string) error

	// DeleteByActionAndEffect 删除指定动作和效果的关联
	DeleteByActionAndEffect(ctx context.Context, actionID, effectID string, executionOrder int) error

	// DeleteAllByActionID 删除动作的所有效果关联
	DeleteAllByActionID(ctx context.Context, actionID string) error

	// BatchCreate 批量创建动作效果关联
	BatchCreate(ctx context.Context, actionEffects []*game_config.ActionEffect) error
}
