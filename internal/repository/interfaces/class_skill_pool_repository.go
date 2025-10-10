package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// ClassSkillPoolQueryParams 职业技能池查询参数
type ClassSkillPoolQueryParams struct {
	ClassID     *string // 职业ID筛选
	SkillID     *string // 技能ID筛选
	SkillTier   *int    // 技能等级筛选
	IsCore      *bool   // 是否核心技能筛选
	IsExclusive *bool   // 是否专属技能筛选
	IsVisible   *bool   // 是否可见筛选
	Limit       int     // 分页限制
	Offset      int     // 分页偏移
}

// ClassSkillPoolRepository 职业技能池仓储接口
type ClassSkillPoolRepository interface {
	// GetClassSkillPools 获取职业技能池列表
	GetClassSkillPools(ctx context.Context, params ClassSkillPoolQueryParams) ([]*game_config.ClassSkillPool, int64, error)

	// GetClassSkillPoolByID 根据ID获取职业技能池
	GetClassSkillPoolByID(ctx context.Context, id string) (*game_config.ClassSkillPool, error)

	// GetClassSkillPoolsByClassID 获取指定职业的所有技能
	GetClassSkillPoolsByClassID(ctx context.Context, classID string) ([]*game_config.ClassSkillPool, error)

	// CreateClassSkillPool 创建职业技能池配置
	CreateClassSkillPool(ctx context.Context, pool *game_config.ClassSkillPool) error

	// UpdateClassSkillPool 更新职业技能池配置
	UpdateClassSkillPool(ctx context.Context, id string, updates map[string]interface{}) error

	// DeleteClassSkillPool 删除职业技能池配置（软删除）
	DeleteClassSkillPool(ctx context.Context, id string) error
}
