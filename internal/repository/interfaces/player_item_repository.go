package interfaces

import (
	"context"
	"database/sql"

	"github.com/aarondl/sqlboiler/v4/boil"
	"tsu-self/internal/entity/game_runtime"
)

// PlayerItemRepository 玩家装备实例仓储接口
type PlayerItemRepository interface {
	// Create 创建装备实例
	Create(ctx context.Context, execer boil.ContextExecutor, item *game_runtime.PlayerItem) error

	// GetByID 根据ID获取装备实例
	GetByID(ctx context.Context, itemInstanceID string) (*game_runtime.PlayerItem, error)

	// GetByIDForUpdate 根据ID获取装备实例（带行锁）
	GetByIDForUpdate(ctx context.Context, tx *sql.Tx, itemInstanceID string) (*game_runtime.PlayerItem, error)

	// GetByOwner 查询玩家的装备实例列表
	GetByOwner(ctx context.Context, ownerID string, location *string) ([]*game_runtime.PlayerItem, error)

	// GetByOwnerPaginated 分页查询玩家的装备实例列表
	GetByOwnerPaginated(ctx context.Context, params GetPlayerItemsParams) ([]*game_runtime.PlayerItem, int64, error)

	// Update 更新装备实例
	Update(ctx context.Context, execer boil.ContextExecutor, item *game_runtime.PlayerItem) error

	// UpdateLocation 更新装备位置
	UpdateLocation(ctx context.Context, execer boil.ContextExecutor, itemInstanceID string, location string) error

	// UpdateDurability 更新装备耐久度
	UpdateDurability(ctx context.Context, execer boil.ContextExecutor, itemInstanceID string, durability int) error

	// Delete 删除装备实例（软删除）
	Delete(ctx context.Context, itemInstanceID string) error

	// BatchUpdateDurability 批量更新装备耐久度
	BatchUpdateDurability(ctx context.Context, execer boil.ContextExecutor, updates []DurabilityUpdate) error

	// GetEquippedItems 查询已装备的物品
	GetEquippedItems(ctx context.Context, ownerID string) ([]*game_runtime.PlayerItem, error)
}

// GetPlayerItemsParams 查询玩家装备实例列表参数
type GetPlayerItemsParams struct {
	OwnerID      string  // 所有者ID
	ItemLocation *string // 物品位置
	ItemType     *string // 物品类型(需要join items表)
	ItemQuality  *string // 物品品质(需要join items表)
	Page         int     // 页码(从1开始)
	PageSize     int     // 每页数量
}

// DurabilityUpdate 耐久度更新参数
type DurabilityUpdate struct {
	ItemInstanceID string // 装备实例ID
	Durability     int    // 新的耐久度
}

