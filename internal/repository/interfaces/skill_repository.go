package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// SkillQueryParams 技能查询参数
type SkillQueryParams struct {
	SkillType  *string // 技能类型
	CategoryID *string // 类别ID
	IsActive   *bool   // 是否启用
	Limit      int     // 每页数量
	Offset     int     // 偏移量
}

// SkillRepository 技能仓储接口
type SkillRepository interface {
	// GetByID 根据ID获取技能
	GetByID(ctx context.Context, skillID string) (*game_config.Skill, error)

	// GetByCode 根据代码获取技能
	GetByCode(ctx context.Context, code string) (*game_config.Skill, error)

	// List 获取技能列表
	List(ctx context.Context, params SkillQueryParams) ([]*game_config.Skill, int64, error)

	// Create 创建技能
	Create(ctx context.Context, skill *game_config.Skill) error

	// Update 更新技能
	Update(ctx context.Context, skill *game_config.Skill) error

	// Delete 软删除技能
	Delete(ctx context.Context, skillID string) error

	// Exists 检查代码是否存在
	Exists(ctx context.Context, code string) (bool, error)
}
