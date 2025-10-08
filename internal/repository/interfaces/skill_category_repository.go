package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// SkillCategoryQueryParams 技能类别查询参数
type SkillCategoryQueryParams struct {
	IsActive *bool
	Limit    int
	Offset   int
}

// SkillCategoryRepository 技能类别仓储接口
type SkillCategoryRepository interface {
	// GetByID 根据ID获取技能类别
	GetByID(ctx context.Context, categoryID string) (*game_config.SkillCategory, error)

	// GetByCode 根据代码获取技能类别
	GetByCode(ctx context.Context, code string) (*game_config.SkillCategory, error)

	// List 获取技能类别列表
	List(ctx context.Context, params SkillCategoryQueryParams) ([]*game_config.SkillCategory, int64, error)

	// Create 创建技能类别
	Create(ctx context.Context, category *game_config.SkillCategory) error

	// Update 更新技能类别
	Update(ctx context.Context, category *game_config.SkillCategory) error

	// Delete 软删除技能类别
	Delete(ctx context.Context, categoryID string) error

	// Exists 检查代码是否存在
	Exists(ctx context.Context, code string) (bool, error)
}
