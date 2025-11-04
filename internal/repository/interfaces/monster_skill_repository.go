package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// MonsterSkillRepository 怪物技能仓储接口
type MonsterSkillRepository interface {
	// Create 创建怪物技能关联
	Create(ctx context.Context, monsterSkill *game_config.MonsterSkill) error

	// BatchCreate 批量创建怪物技能关联
	BatchCreate(ctx context.Context, monsterSkills []*game_config.MonsterSkill) error

	// GetByMonsterID 获取怪物的所有技能
	GetByMonsterID(ctx context.Context, monsterID string) ([]*game_config.MonsterSkill, error)

	// GetByMonsterAndSkill 获取怪物的特定技能
	GetByMonsterAndSkill(ctx context.Context, monsterID, skillID string) (*game_config.MonsterSkill, error)

	// Update 更新怪物技能配置
	Update(ctx context.Context, monsterSkill *game_config.MonsterSkill) error

	// Delete 软删除怪物技能关联
	Delete(ctx context.Context, monsterID, skillID string) error

	// DeleteByMonsterID 删除怪物的所有技能（软删除）
	DeleteByMonsterID(ctx context.Context, monsterID string) error

	// Exists 检查怪物技能关联是否存在
	Exists(ctx context.Context, monsterID, skillID string) (bool, error)
}

