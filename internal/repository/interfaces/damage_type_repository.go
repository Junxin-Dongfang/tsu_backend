package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// DamageTypeQueryParams 伤害类型查询参数
type DamageTypeQueryParams struct {
	Category *string // 伤害类别（PHYSICAL, MAGICAL等）
	IsActive *bool   // 是否启用
	Limit    int     // 每页数量
	Offset   int     // 偏移量
}

// DamageTypeRepository 伤害类型仓储接口
type DamageTypeRepository interface {
	// GetByID 根据ID获取伤害类型
	GetByID(ctx context.Context, damageTypeID string) (*game_config.DamageType, error)

	// GetByCode 根据代码获取伤害类型
	GetByCode(ctx context.Context, code string) (*game_config.DamageType, error)

	// List 获取伤害类型列表
	List(ctx context.Context, params DamageTypeQueryParams) ([]*game_config.DamageType, int64, error)

	// Create 创建伤害类型
	Create(ctx context.Context, damageType *game_config.DamageType) error

	// Update 更新伤害类型
	Update(ctx context.Context, damageType *game_config.DamageType) error

	// Delete 软删除伤害类型
	Delete(ctx context.Context, damageTypeID string) error

	// Exists 检查代码是否存在
	Exists(ctx context.Context, code string) (bool, error)
}
