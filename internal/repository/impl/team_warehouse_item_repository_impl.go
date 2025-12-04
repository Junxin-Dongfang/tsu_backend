package impl

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/google/uuid"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/repository/interfaces"
)

type teamWarehouseItemRepositoryImpl struct {
	db *sql.DB
}

// NewTeamWarehouseItemRepository 创建团队仓库物品仓储实例
func NewTeamWarehouseItemRepository(db *sql.DB) interfaces.TeamWarehouseItemRepository {
	return &teamWarehouseItemRepositoryImpl{db: db}
}

// Create 创建仓库物品记录
func (r *teamWarehouseItemRepositoryImpl) Create(ctx context.Context, execer boil.ContextExecutor, item *game_runtime.TeamWarehouseItem) error {
	// 生成UUID
	if item.ID == "" {
		item.ID = uuid.New().String()
	}

	// 设置时间戳
	item.ObtainedAt = time.Now()

	// 插入数据库
	if err := item.Insert(ctx, execer, boil.Infer()); err != nil {
		return fmt.Errorf("创建仓库物品记录失败: %w", err)
	}

	return nil
}

// GetByID 根据ID获取物品记录
func (r *teamWarehouseItemRepositoryImpl) GetByID(ctx context.Context, itemID string) (*game_runtime.TeamWarehouseItem, error) {
	item, err := game_runtime.TeamWarehouseItems(
		qm.Where("id = ?", itemID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("仓库物品记录不存在: %s", itemID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询仓库物品记录失败: %w", err)
	}

	return item, nil
}

// Update 更新物品记录
func (r *teamWarehouseItemRepositoryImpl) Update(ctx context.Context, execer boil.ContextExecutor, item *game_runtime.TeamWarehouseItem) error {
	if _, err := item.Update(ctx, execer, boil.Infer()); err != nil {
		return fmt.Errorf("更新仓库物品记录失败: %w", err)
	}
	return nil
}

// Delete 删除物品记录
func (r *teamWarehouseItemRepositoryImpl) Delete(ctx context.Context, execer boil.ContextExecutor, itemID string) error {
	item, err := r.GetByID(ctx, itemID)
	if err != nil {
		return err
	}

	if _, err := item.Delete(ctx, execer); err != nil {
		return fmt.Errorf("删除仓库物品记录失败: %w", err)
	}

	return nil
}

// ListByWarehouse 查询仓库物品列表
func (r *teamWarehouseItemRepositoryImpl) ListByWarehouse(ctx context.Context, warehouseID string, limit, offset int) ([]*game_runtime.TeamWarehouseItem, int64, error) {
	// 统计总数
	count, err := game_runtime.TeamWarehouseItems(
		qm.Where("warehouse_id = ?", warehouseID),
	).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("统计仓库物品数量失败: %w", err)
	}

	// 查询列表
	queryMods := []qm.QueryMod{
		qm.Where("warehouse_id = ?", warehouseID),
		qm.OrderBy("obtained_at DESC"),
	}
	if limit > 0 {
		queryMods = append(queryMods, qm.Limit(limit))
	}
	if offset > 0 {
		queryMods = append(queryMods, qm.Offset(offset))
	}

	items, err := game_runtime.TeamWarehouseItems(queryMods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询仓库物品列表失败: %w", err)
	}

	return items, count, nil
}

// AddItem 添加物品
func (r *teamWarehouseItemRepositoryImpl) AddItem(ctx context.Context, execer boil.ContextExecutor, warehouseID, itemID, itemType string, quantity int, sourceDungeonID *string) error {
	// 查找是否已存在相同物品
	existingItem, err := game_runtime.TeamWarehouseItems(
		qm.Where("warehouse_id = ? AND item_id = ?", warehouseID, itemID),
	).One(ctx, execer)

	if err == sql.ErrNoRows {
		// 不存在，创建新记录
		newItem := &game_runtime.TeamWarehouseItem{
			WarehouseID: warehouseID,
			ItemID:      itemID,
			ItemType:    itemType,
			Quantity:    quantity,
		}
		if sourceDungeonID != nil {
			newItem.SourceDungeonID = null.StringFrom(*sourceDungeonID)
		}
		return r.Create(ctx, execer, newItem)
	} else if err != nil {
		return fmt.Errorf("查询物品失败: %w", err)
	}

	// 已存在，增加数量
	existingItem.Quantity += quantity
	return r.Update(ctx, execer, existingItem)
}

// DeductItem 扣除物品数量
func (r *teamWarehouseItemRepositoryImpl) DeductItem(ctx context.Context, execer boil.ContextExecutor, warehouseID, itemID string, quantity int) error {
	item, err := game_runtime.TeamWarehouseItems(
		qm.Where("warehouse_id = ? AND item_id = ?", warehouseID, itemID),
	).One(ctx, execer)

	if err == sql.ErrNoRows {
		return fmt.Errorf("物品不存在")
	}
	if err != nil {
		return fmt.Errorf("查询物品失败: %w", err)
	}

	if item.Quantity < quantity {
		return fmt.Errorf("物品数量不足")
	}

	item.Quantity -= quantity

	// 如果数量为0，删除记录
	if item.Quantity == 0 {
		return r.Delete(ctx, execer, item.ID)
	}

	return r.Update(ctx, execer, item)
}

// GetItemCount 查询物品数量
func (r *teamWarehouseItemRepositoryImpl) GetItemCount(ctx context.Context, warehouseID, itemID string) (int, error) {
	item, err := game_runtime.TeamWarehouseItems(
		qm.Where("warehouse_id = ? AND item_id = ?", warehouseID, itemID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return 0, nil // 物品不存在，返回0
	}
	if err != nil {
		return 0, fmt.Errorf("查询物品数量失败: %w", err)
	}

	return item.Quantity, nil
}

// CountDistinctItems 统计仓库中不同物品的种类数
func (r *teamWarehouseItemRepositoryImpl) CountDistinctItems(ctx context.Context, warehouseID string) (int64, error) {
	var count int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(DISTINCT item_id) FROM game_runtime.team_warehouse_items WHERE warehouse_id = $1", warehouseID).Scan(&count); err != nil {
		return 0, fmt.Errorf("统计物品种类数失败: %w", err)
	}
	return count, nil
}
