package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_runtime"
)

// EquipmentRepository 装备Repository接口（包装PlayerItemRepository的装备相关方法）
type EquipmentRepository interface {
	// GetEquippedItems 查询已装备的物品
	GetEquippedItems(ctx context.Context, heroID string) ([]*game_runtime.PlayerItem, error)
}

