package impl

import (
	"context"
	"database/sql"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/repository/interfaces"

	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

// EquipmentRepositoryImpl 装备Repository实现
type EquipmentRepositoryImpl struct {
	db *sql.DB
}

// NewEquipmentRepository 创建装备Repository
func NewEquipmentRepository(db *sql.DB) interfaces.EquipmentRepository {
	return &EquipmentRepositoryImpl{
		db: db,
	}
}

// GetEquippedItems 查询已装备的物品
// 通过 hero_equipment_slots 表查询英雄已装备的物品
func (r *EquipmentRepositoryImpl) GetEquippedItems(ctx context.Context, heroID string) ([]*game_runtime.PlayerItem, error) {
	// 1. 查询英雄的装备槽位
	slots, err := game_runtime.HeroEquipmentSlots(
		qm.Where("hero_id = ?", heroID),
		qm.Where("equipped_item_id IS NOT NULL"),
	).All(ctx, r.db)
	if err != nil {
		return nil, err
	}

	// 如果没有装备任何物品，返回空列表
	if len(slots) == 0 {
		return []*game_runtime.PlayerItem{}, nil
	}

	// 2. 提取装备实例ID
	itemIDs := make([]interface{}, len(slots))
	for i, slot := range slots {
		itemIDs[i] = slot.EquippedItemID.String
	}

	// 3. 查询装备实例
	items, err := game_runtime.PlayerItems(
		qm.WhereIn("id IN ?", itemIDs...),
		qm.Where("deleted_at IS NULL"),
	).All(ctx, r.db)
	if err != nil {
		return nil, err
	}

	return items, nil
}

