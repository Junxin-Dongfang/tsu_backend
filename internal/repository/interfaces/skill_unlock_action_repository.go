package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// SkillUnlockActionRepository 技能解锁动作仓储接口
type SkillUnlockActionRepository interface {
	// GetBySkillID 获取技能的所有解锁动作
	GetBySkillID(ctx context.Context, skillID string) ([]*game_config.SkillUnlockAction, error)

	// Create 创建技能解锁动作关联
	Create(ctx context.Context, unlockAction *game_config.SkillUnlockAction) error

	// Delete 删除技能解锁动作关联
	Delete(ctx context.Context, unlockActionID string) error

	// DeleteBySkillAndAction 删除指定技能和动作的关联
	DeleteBySkillAndAction(ctx context.Context, skillID, actionID string) error

	// DeleteAllBySkillID 删除技能的所有解锁动作
	DeleteAllBySkillID(ctx context.Context, skillID string) error

	// BatchCreate 批量创建技能解锁动作
	BatchCreate(ctx context.Context, unlockActions []*game_config.SkillUnlockAction) error
}
