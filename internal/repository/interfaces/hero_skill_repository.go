package interfaces

import (
	"context"

	"github.com/aarondl/sqlboiler/v4/boil"
	"tsu-self/internal/entity/game_runtime"
)

// HeroSkillRepository 英雄技能仓储接口
type HeroSkillRepository interface {
	// Create 创建英雄技能记录
	Create(ctx context.Context, execer boil.ContextExecutor, heroSkill *game_runtime.HeroSkill) error

	// GetByID 根据ID获取技能
	GetByID(ctx context.Context, heroSkillID string) (*game_runtime.HeroSkill, error)

	// GetByIDForUpdate 根据ID获取技能（带行锁）
	GetByIDForUpdate(ctx context.Context, execer boil.ContextExecutor, heroSkillID string) (*game_runtime.HeroSkill, error)

	// GetByHeroID 获取英雄的所有技能
	GetByHeroID(ctx context.Context, heroID string) ([]*game_runtime.HeroSkill, error)

	// GetByHeroAndSkillID 获取英雄的特定技能
	GetByHeroAndSkillID(ctx context.Context, heroID, skillID string) (*game_runtime.HeroSkill, error)

	// Update 更新技能信息
	Update(ctx context.Context, execer boil.ContextExecutor, heroSkill *game_runtime.HeroSkill) error

	// Delete 删除技能记录
	Delete(ctx context.Context, execer boil.ContextExecutor, heroSkillID string) error

	// DeleteAllByHeroID 删除英雄的所有技能（重生时使用）
	DeleteAllByHeroID(ctx context.Context, execer boil.ContextExecutor, heroID string) error
}

