package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// SkillLevelConfigQueryParams 技能等级配置查询参数
type SkillLevelConfigQueryParams struct {
	SkillID     *string // 技能ID
	LevelNumber *int    // 等级
	Limit       int     // 每页数量
	Offset      int     // 偏移量
}

// SkillLevelConfigRepository 技能等级配置仓储接口
type SkillLevelConfigRepository interface {
	// GetByID 根据ID获取配置
	GetByID(ctx context.Context, id string) (*game_config.SkillLevelConfig, error)

	// GetBySkillIDAndLevel 根据技能ID和等级获取配置
	GetBySkillIDAndLevel(ctx context.Context, skillID string, levelNumber int) (*game_config.SkillLevelConfig, error)

	// ListBySkillID 根据技能ID获取所有等级配置
	ListBySkillID(ctx context.Context, skillID string) ([]*game_config.SkillLevelConfig, error)

	// List 获取配置列表
	List(ctx context.Context, params SkillLevelConfigQueryParams) ([]*game_config.SkillLevelConfig, int64, error)

	// Create 创建配置
	Create(ctx context.Context, config *game_config.SkillLevelConfig) error

	// Update 更新配置
	Update(ctx context.Context, config *game_config.SkillLevelConfig) error

	// Delete 软删除配置
	Delete(ctx context.Context, id string) error

	// DeleteBySkillID 删除技能的所有等级配置
	DeleteBySkillID(ctx context.Context, skillID string) error
}
