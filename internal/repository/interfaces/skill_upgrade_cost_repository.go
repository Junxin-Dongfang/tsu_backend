package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// SkillUpgradeCostRepository 技能升级消耗仓储接口
type SkillUpgradeCostRepository interface {
	// Create 创建升级消耗配置
	Create(ctx context.Context, cost *game_config.SkillUpgradeCost) error

	// GetByID 根据ID获取
	GetByID(ctx context.Context, id string) (*game_config.SkillUpgradeCost, error)

	// GetByLevel 根据等级获取
	GetByLevel(ctx context.Context, level int) (*game_config.SkillUpgradeCost, error)

	// List 获取所有升级消耗配置
	List(ctx context.Context) ([]*game_config.SkillUpgradeCost, error)

	// Update 更新升级消耗配置
	Update(ctx context.Context, id string, updates map[string]interface{}) error

	// Delete 删除升级消耗配置
	Delete(ctx context.Context, id string) error
}
