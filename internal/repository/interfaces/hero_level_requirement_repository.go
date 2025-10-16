package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// HeroLevelRequirementRepository 英雄等级需求配置仓储接口
type HeroLevelRequirementRepository interface {
	// GetByLevel 根据等级获取需求配置
	GetByLevel(ctx context.Context, level int) (*game_config.HeroLevelRequirement, error)

	// GetNextLevelRequirement 获取下一级需求
	GetNextLevelRequirement(ctx context.Context, currentLevel int) (*game_config.HeroLevelRequirement, error)

	// GetAll 获取所有等级需求
	GetAll(ctx context.Context) ([]*game_config.HeroLevelRequirement, error)

	// CheckCanLevelUp 检查是否可以升级（返回可升到的最高等级）
	CheckCanLevelUp(ctx context.Context, experienceTotal int, currentLevel int) (canLevelUp bool, targetLevel int, error error)
}

