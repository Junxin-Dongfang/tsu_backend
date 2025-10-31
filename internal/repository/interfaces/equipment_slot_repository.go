package interfaces

import (
	"context"
	"database/sql"

	"github.com/aarondl/sqlboiler/v4/boil"
	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/entity/game_runtime"
)

// EquipmentSlotRepository 装备槽位仓储接口
type EquipmentSlotRepository interface {
	// InitializeSlots 初始化英雄槽位(根据职业配置)
	InitializeSlots(ctx context.Context, tx *sql.Tx, heroID string, classID string) error

	// GetSlots 查询英雄所有槽位
	GetSlots(ctx context.Context, heroID string) ([]*game_runtime.HeroEquipmentSlot, error)

	// GetSlotByType 查询特定类型的槽位
	GetSlotByType(ctx context.Context, heroID string, slotType string) ([]*game_runtime.HeroEquipmentSlot, error)

	// GetSlotByID 根据ID获取槽位
	GetSlotByID(ctx context.Context, slotID string) (*game_runtime.HeroEquipmentSlot, error)

	// GetSlotByIDForUpdate 根据ID获取槽位（带行锁）
	GetSlotByIDForUpdate(ctx context.Context, tx *sql.Tx, slotID string) (*game_runtime.HeroEquipmentSlot, error)

	// UpdateSlot 更新槽位
	UpdateSlot(ctx context.Context, execer boil.ContextExecutor, slot *game_runtime.HeroEquipmentSlot) error

	// AddSlot 增加槽位(装备效果)
	AddSlot(ctx context.Context, execer boil.ContextExecutor, slot *game_runtime.HeroEquipmentSlot) error

	// DeleteSlot 删除槽位(卸下增加槽位的装备时)
	DeleteSlot(ctx context.Context, execer boil.ContextExecutor, slotID string) error

	// GetSlotConfigs 查询职业的槽位配置
	GetSlotConfigs(ctx context.Context, classID string) ([]*game_config.EquipmentSlotConfig, error)

	// FindAvailableSlot 查找可用的槽位(未装备且已解锁)
	FindAvailableSlot(ctx context.Context, heroID string, slotType string) (*game_runtime.HeroEquipmentSlot, error)

	// GetSlotsAddedByItem 查询由特定装备增加的槽位
	GetSlotsAddedByItem(ctx context.Context, itemInstanceID string) ([]*game_runtime.HeroEquipmentSlot, error)
}

