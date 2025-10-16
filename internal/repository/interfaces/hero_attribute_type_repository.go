package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// HeroAttributeTypeQueryParams 属性类型查询参数
type HeroAttributeTypeQueryParams struct {
	Category  *string // 属性类别 (BASE, COMBAT, DERIVED等)
	IsActive  *bool   // 是否启用
	IsVisible *bool   // 是否可见
	Limit     int     // 每页数量
	Offset    int     // 偏移量
}

// HeroAttributeTypeRepository 属性类型仓储接口
type HeroAttributeTypeRepository interface {
	// GetByID 根据ID获取属性类型
	GetByID(ctx context.Context, attributeTypeID string) (*game_config.HeroAttributeType, error)

	// GetByCode 根据代码获取属性类型
	GetByCode(ctx context.Context, code string) (*game_config.HeroAttributeType, error)

	// List 获取属性类型列表
	List(ctx context.Context, params HeroAttributeTypeQueryParams) ([]*game_config.HeroAttributeType, int64, error)

	// Create 创建属性类型
	Create(ctx context.Context, attributeType *game_config.HeroAttributeType) error

	// Update 更新属性类型
	Update(ctx context.Context, attributeType *game_config.HeroAttributeType) error

	// Delete 软删除属性类型
	Delete(ctx context.Context, attributeTypeID string) error

	// Exists 检查代码是否存在
	Exists(ctx context.Context, code string) (bool, error)

	// ListByCategory 按分类获取所有活跃的属性类型
	ListByCategory(ctx context.Context, category string) ([]*game_config.HeroAttributeType, error)
}
