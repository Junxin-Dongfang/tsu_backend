package impl

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/repository/interfaces"
)

type playerItemRepositoryImpl struct {
	db *sql.DB
}

// NewPlayerItemRepository 创建玩家装备实例仓储实例
func NewPlayerItemRepository(db *sql.DB) interfaces.PlayerItemRepository {
	return &playerItemRepositoryImpl{db: db}
}

// Create 创建装备实例
func (r *playerItemRepositoryImpl) Create(ctx context.Context, execer boil.ContextExecutor, item *game_runtime.PlayerItem) error {
	if err := item.Insert(ctx, execer, boil.Infer()); err != nil {
		return fmt.Errorf("创建装备实例失败: %w", err)
	}
	return nil
}

// GetByID 根据ID获取装备实例
func (r *playerItemRepositoryImpl) GetByID(ctx context.Context, itemInstanceID string) (*game_runtime.PlayerItem, error) {
	item, err := game_runtime.PlayerItems(
		qm.Where("id = ? AND deleted_at IS NULL", itemInstanceID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("装备实例不存在: %s", itemInstanceID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询装备实例失败: %w", err)
	}

	return item, nil
}

// GetByIDForUpdate 根据ID获取装备实例（带行锁）
func (r *playerItemRepositoryImpl) GetByIDForUpdate(ctx context.Context, tx *sql.Tx, itemInstanceID string) (*game_runtime.PlayerItem, error) {
	item, err := game_runtime.PlayerItems(
		qm.Where("id = ? AND deleted_at IS NULL", itemInstanceID),
		qm.For("UPDATE"),
	).One(ctx, tx)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("装备实例不存在: %s", itemInstanceID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询装备实例失败（带锁）: %w", err)
	}

	return item, nil
}

// GetByOwner 查询玩家的装备实例列表
func (r *playerItemRepositoryImpl) GetByOwner(ctx context.Context, ownerID string, location *string) ([]*game_runtime.PlayerItem, error) {
	mods := []qm.QueryMod{
		qm.Where("owner_id = ? AND deleted_at IS NULL", ownerID),
	}

	if location != nil {
		mods = append(mods, qm.Where("item_location = ?", *location))
	}

	mods = append(mods, qm.OrderBy("created_at DESC"))

	items, err := game_runtime.PlayerItems(mods...).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("查询玩家装备实例列表失败: %w", err)
	}

	return items, nil
}

// GetByOwnerPaginated 分页查询玩家的装备实例列表
func (r *playerItemRepositoryImpl) GetByOwnerPaginated(ctx context.Context, params interfaces.GetPlayerItemsParams) ([]*game_runtime.PlayerItem, int64, error) {
	// 构建查询条件
	mods := []qm.QueryMod{
		qm.Where("owner_id = ? AND deleted_at IS NULL", params.OwnerID),
	}

	// 添加筛选条件
	if params.ItemLocation != nil {
		mods = append(mods, qm.Where("item_location = ?", *params.ItemLocation))
	}

	// 如果需要按物品类型或品质筛选,需要join items表
	if params.ItemType != nil || params.ItemQuality != nil {
		mods = append(mods, qm.InnerJoin("game_config.items ON game_runtime.player_items.item_id = game_config.items.id"))
		if params.ItemType != nil {
			mods = append(mods, qm.Where("game_config.items.item_type = ?", *params.ItemType))
		}
		if params.ItemQuality != nil {
			mods = append(mods, qm.Where("game_config.items.item_quality = ?", *params.ItemQuality))
		}
	}

	// 查询总数
	count, err := game_runtime.PlayerItems(mods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询装备实例总数失败: %w", err)
	}

	// 添加分页和排序
	if params.PageSize > 0 {
		offset := (params.Page - 1) * params.PageSize
		mods = append(mods, qm.Limit(params.PageSize), qm.Offset(offset))
	}
	mods = append(mods, qm.OrderBy("created_at DESC"))

	// 查询列表
	items, err := game_runtime.PlayerItems(mods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询装备实例列表失败: %w", err)
	}

	return items, count, nil
}

// Update 更新装备实例
func (r *playerItemRepositoryImpl) Update(ctx context.Context, execer boil.ContextExecutor, item *game_runtime.PlayerItem) error {
	if _, err := item.Update(ctx, execer, boil.Infer()); err != nil {
		return fmt.Errorf("更新装备实例失败: %w", err)
	}
	return nil
}

// UpdateLocation 更新装备位置
func (r *playerItemRepositoryImpl) UpdateLocation(ctx context.Context, execer boil.ContextExecutor, itemInstanceID string, location string) error {
	item, err := r.GetByID(ctx, itemInstanceID)
	if err != nil {
		return err
	}

	item.ItemLocation = location

	if _, err := item.Update(ctx, execer, boil.Whitelist("item_location", "updated_at")); err != nil {
		return fmt.Errorf("更新装备位置失败: %w", err)
	}

	return nil
}

// UpdateDurability 更新装备耐久度
func (r *playerItemRepositoryImpl) UpdateDurability(ctx context.Context, execer boil.ContextExecutor, itemInstanceID string, durability int) error {
	item, err := r.GetByID(ctx, itemInstanceID)
	if err != nil {
		return err
	}

	item.CurrentDurability = null.IntFrom(durability)

	if _, err := item.Update(ctx, execer, boil.Whitelist("current_durability", "updated_at")); err != nil {
		return fmt.Errorf("更新装备耐久度失败: %w", err)
	}

	return nil
}

// Delete 删除装备实例（软删除）
func (r *playerItemRepositoryImpl) Delete(ctx context.Context, itemInstanceID string) error {
	item, err := r.GetByID(ctx, itemInstanceID)
	if err != nil {
		return err
	}

	item.DeletedAt = null.TimeFromPtr(nullTimeNow())

	if _, err := item.Update(ctx, r.db, boil.Whitelist("deleted_at", "updated_at")); err != nil {
		return fmt.Errorf("软删除装备实例失败: %w", err)
	}

	return nil
}

// BatchUpdateDurability 批量更新装备耐久度
func (r *playerItemRepositoryImpl) BatchUpdateDurability(ctx context.Context, execer boil.ContextExecutor, updates []interfaces.DurabilityUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	// 使用事务批量更新
	for _, update := range updates {
		if err := r.UpdateDurability(ctx, execer, update.ItemInstanceID, update.Durability); err != nil {
			return fmt.Errorf("批量更新装备耐久度失败: %w", err)
		}
	}

	return nil
}

// GetEquippedItems 查询已装备的物品
func (r *playerItemRepositoryImpl) GetEquippedItems(ctx context.Context, ownerID string) ([]*game_runtime.PlayerItem, error) {
	location := "equipped"
	return r.GetByOwner(ctx, ownerID, &location)
}

