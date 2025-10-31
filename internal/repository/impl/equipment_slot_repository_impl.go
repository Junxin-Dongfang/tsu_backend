package impl

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/repository/interfaces"
)

type equipmentSlotRepositoryImpl struct {
	db *sql.DB
}

// NewEquipmentSlotRepository 创建装备槽位仓储实例
func NewEquipmentSlotRepository(db *sql.DB) interfaces.EquipmentSlotRepository {
	return &equipmentSlotRepositoryImpl{db: db}
}

// InitializeSlots 初始化英雄槽位(根据职业配置)
func (r *equipmentSlotRepositoryImpl) InitializeSlots(ctx context.Context, tx *sql.Tx, heroID string, classID string) error {
	// 查询职业的槽位配置
	configs, err := r.GetSlotConfigs(ctx, classID)
	if err != nil {
		return fmt.Errorf("查询职业槽位配置失败: %w", err)
	}

	// 为每个槽位配置创建槽位实例
	for _, config := range configs {
		// 创建默认数量的槽位
		for i := 0; i < int(config.DefaultCount); i++ {
			slot := &game_runtime.HeroEquipmentSlot{
				HeroID:      heroID,
				SlotType:    config.SlotType,
				SlotIndex:   int16(i),
				IsUnlocked:  true,
				UnlockLevel: config.UnlockLevel,
			}

			if err := slot.Insert(ctx, tx, boil.Infer()); err != nil {
				return fmt.Errorf("创建槽位失败: %w", err)
			}
		}
	}

	return nil
}

// GetSlots 查询英雄所有槽位
func (r *equipmentSlotRepositoryImpl) GetSlots(ctx context.Context, heroID string) ([]*game_runtime.HeroEquipmentSlot, error) {
	slots, err := game_runtime.HeroEquipmentSlots(
		qm.Where("hero_id = ?", heroID),
		qm.OrderBy("slot_type, slot_index"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询英雄槽位失败: %w", err)
	}

	return slots, nil
}

// GetSlotByType 查询特定类型的槽位
func (r *equipmentSlotRepositoryImpl) GetSlotByType(ctx context.Context, heroID string, slotType string) ([]*game_runtime.HeroEquipmentSlot, error) {
	slots, err := game_runtime.HeroEquipmentSlots(
		qm.Where("hero_id = ? AND slot_type = ?", heroID, slotType),
		qm.OrderBy("slot_index"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询特定类型槽位失败: %w", err)
	}

	return slots, nil
}

// GetSlotByID 根据ID获取槽位
func (r *equipmentSlotRepositoryImpl) GetSlotByID(ctx context.Context, slotID string) (*game_runtime.HeroEquipmentSlot, error) {
	slot, err := game_runtime.HeroEquipmentSlots(
		qm.Where("id = ?", slotID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("槽位不存在: %s", slotID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询槽位失败: %w", err)
	}

	return slot, nil
}

// GetSlotByIDForUpdate 根据ID获取槽位（带行锁）
func (r *equipmentSlotRepositoryImpl) GetSlotByIDForUpdate(ctx context.Context, tx *sql.Tx, slotID string) (*game_runtime.HeroEquipmentSlot, error) {
	slot, err := game_runtime.HeroEquipmentSlots(
		qm.Where("id = ?", slotID),
		qm.For("UPDATE"),
	).One(ctx, tx)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("槽位不存在: %s", slotID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询槽位失败（带锁）: %w", err)
	}

	return slot, nil
}

// UpdateSlot 更新槽位
func (r *equipmentSlotRepositoryImpl) UpdateSlot(ctx context.Context, execer boil.ContextExecutor, slot *game_runtime.HeroEquipmentSlot) error {
	if _, err := slot.Update(ctx, execer, boil.Infer()); err != nil {
		return fmt.Errorf("更新槽位失败: %w", err)
	}
	return nil
}

// AddSlot 增加槽位(装备效果)
func (r *equipmentSlotRepositoryImpl) AddSlot(ctx context.Context, execer boil.ContextExecutor, slot *game_runtime.HeroEquipmentSlot) error {
	if err := slot.Insert(ctx, execer, boil.Infer()); err != nil {
		return fmt.Errorf("增加槽位失败: %w", err)
	}
	return nil
}

// DeleteSlot 删除槽位(卸下增加槽位的装备时)
func (r *equipmentSlotRepositoryImpl) DeleteSlot(ctx context.Context, execer boil.ContextExecutor, slotID string) error {
	slot, err := r.GetSlotByID(ctx, slotID)
	if err != nil {
		return err
	}

	if _, err := slot.Delete(ctx, execer); err != nil {
		return fmt.Errorf("删除槽位失败: %w", err)
	}

	return nil
}

// GetSlotConfigs 查询职业的槽位配置
func (r *equipmentSlotRepositoryImpl) GetSlotConfigs(ctx context.Context, classID string) ([]*game_config.EquipmentSlotConfig, error) {
	configs, err := game_config.EquipmentSlotConfigs(
		qm.Where("class_id = ?", classID),
		qm.OrderBy("slot_type"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询职业槽位配置失败: %w", err)
	}

	return configs, nil
}

// FindAvailableSlot 查找可用的槽位(未装备且已解锁)
func (r *equipmentSlotRepositoryImpl) FindAvailableSlot(ctx context.Context, heroID string, slotType string) (*game_runtime.HeroEquipmentSlot, error) {
	slot, err := game_runtime.HeroEquipmentSlots(
		qm.Where("hero_id = ? AND slot_type = ?", heroID, slotType),
		qm.Where("equipped_item_id IS NULL"),
		qm.Where("is_unlocked = ?", true),
		qm.OrderBy("slot_index"),
		qm.Limit(1),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("没有可用的%s槽位", slotType)
	}
	if err != nil {
		return nil, fmt.Errorf("查找可用槽位失败: %w", err)
	}

	return slot, nil
}

// GetSlotsAddedByItem 查询由特定装备增加的槽位
func (r *equipmentSlotRepositoryImpl) GetSlotsAddedByItem(ctx context.Context, itemInstanceID string) ([]*game_runtime.HeroEquipmentSlot, error) {
	slots, err := game_runtime.HeroEquipmentSlots(
		qm.Where("added_by_item_id = ?", itemInstanceID),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询由装备增加的槽位失败: %w", err)
	}

	return slots, nil
}

