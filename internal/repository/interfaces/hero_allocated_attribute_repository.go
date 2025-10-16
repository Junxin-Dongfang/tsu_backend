package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_runtime"
)

// HeroAllocatedAttributeRepository 英雄已分配属性仓储接口
type HeroAllocatedAttributeRepository interface {
	// GetByHeroAndCode 根据英雄ID和属性代码获取
	GetByHeroAndCode(ctx context.Context, heroID, attributeCode string) (*game_runtime.HeroAllocatedAttribute, error)

	// GetByHeroID 获取英雄的所有已分配属性
	GetByHeroID(ctx context.Context, heroID string) ([]*game_runtime.HeroAllocatedAttribute, error)

	// Create 创建新的已分配属性
	Create(ctx context.Context, executor interface{}, attr *game_runtime.HeroAllocatedAttribute) error

	// Update 更新已分配属性
	Update(ctx context.Context, executor interface{}, attr *game_runtime.HeroAllocatedAttribute) error

	// Delete 软删除已分配属性
	Delete(ctx context.Context, executor interface{}, heroID, attributeCode string) error

	// BatchCreateForHero 为英雄批量创建初始属性
	BatchCreateForHero(ctx context.Context, executor interface{}, heroID string, attrs []*game_runtime.HeroAllocatedAttribute) error

	// GetByHeroIDForUpdate 根据英雄ID获取属性（带锁用于事务）
	GetByHeroIDForUpdate(ctx context.Context, executor interface{}, heroID string) ([]*game_runtime.HeroAllocatedAttribute, error)
}
