package interfaces

import (
	"context"

	"github.com/aarondl/sqlboiler/v4/boil"
	"tsu-self/internal/entity/game_runtime"
)

// TeamWarehouseItemRepository 团队仓库物品仓储接口
type TeamWarehouseItemRepository interface {
	// Create 创建仓库物品记录
	Create(ctx context.Context, execer boil.ContextExecutor, item *game_runtime.TeamWarehouseItem) error

	// GetByID 根据ID获取物品记录
	GetByID(ctx context.Context, itemID string) (*game_runtime.TeamWarehouseItem, error)

	// Update 更新物品记录
	Update(ctx context.Context, execer boil.ContextExecutor, item *game_runtime.TeamWarehouseItem) error

	// Delete 删除物品记录
	Delete(ctx context.Context, execer boil.ContextExecutor, itemID string) error

	// ListByWarehouse 查询仓库物品列表
	ListByWarehouse(ctx context.Context, warehouseID string, limit, offset int) ([]*game_runtime.TeamWarehouseItem, int64, error)

	// AddItem 添加物品（如果已存在则增加数量）
	AddItem(ctx context.Context, execer boil.ContextExecutor, warehouseID, itemID, itemType string, quantity int, sourceDungeonID *string) error

	// DeductItem 扣除物品数量
	DeductItem(ctx context.Context, execer boil.ContextExecutor, warehouseID, itemID string, quantity int) error

	// GetItemCount 查询物品数量
	GetItemCount(ctx context.Context, warehouseID, itemID string) (int, error)

	// CountDistinctItems 统计仓库中不同物品的种类数
	CountDistinctItems(ctx context.Context, warehouseID string) (int64, error)
}

