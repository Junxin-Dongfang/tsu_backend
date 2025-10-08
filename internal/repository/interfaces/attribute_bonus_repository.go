package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// AttributeBonusRepository 职业属性加成仓储接口
type AttributeBonusRepository interface {
	// GetByClassID 获取职业的所有属性加成
	GetByClassID(ctx context.Context, classID string) ([]*game_config.ClassAttributeBonuse, error)

	// GetByID 根据ID获取属性加成
	GetByID(ctx context.Context, bonusID string) (*game_config.ClassAttributeBonuse, error)

	// Create 创建属性加成
	Create(ctx context.Context, bonus *game_config.ClassAttributeBonuse) error

	// Update 更新属性加成
	Update(ctx context.Context, bonus *game_config.ClassAttributeBonuse) error

	// Delete 删除属性加成
	Delete(ctx context.Context, bonusID string) error

	// BatchCreate 批量创建属性加成
	BatchCreate(ctx context.Context, bonuses []*game_config.ClassAttributeBonuse) error

	// DeleteByClassID 删除职业的所有属性加成
	DeleteByClassID(ctx context.Context, classID string) error

	// Exists 检查职业-属性组合是否已存在
	Exists(ctx context.Context, classID, attributeID string) (bool, error)
}
