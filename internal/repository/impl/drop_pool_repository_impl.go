package impl

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/interfaces"
)

type dropPoolRepositoryImpl struct {
	db *sql.DB
}

// NewDropPoolRepository 创建掉落池仓储实例
func NewDropPoolRepository(db *sql.DB) interfaces.DropPoolRepository {
	return &dropPoolRepositoryImpl{db: db}
}

// GetByID 根据ID获取掉落池
func (r *dropPoolRepositoryImpl) GetByID(ctx context.Context, poolID string) (*game_config.DropPool, error) {
	pool, err := game_config.DropPools(
		qm.Where("id = ? AND deleted_at IS NULL", poolID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("掉落池不存在: %s", poolID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询掉落池失败: %w", err)
	}

	return pool, nil
}

// GetByCode 根据代码获取掉落池
func (r *dropPoolRepositoryImpl) GetByCode(ctx context.Context, poolCode string) (*game_config.DropPool, error) {
	pool, err := game_config.DropPools(
		qm.Where("pool_code = ? AND deleted_at IS NULL", poolCode),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("掉落池不存在: %s", poolCode)
	}
	if err != nil {
		return nil, fmt.Errorf("查询掉落池失败: %w", err)
	}

	return pool, nil
}

// GetByType 根据类型获取掉落池列表
func (r *dropPoolRepositoryImpl) GetByType(ctx context.Context, poolType string) ([]*game_config.DropPool, error) {
	pools, err := game_config.DropPools(
		qm.Where("pool_type = ? AND deleted_at IS NULL", poolType),
		qm.Where("is_active = ?", true),
		qm.OrderBy("created_at DESC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询掉落池列表失败: %w", err)
	}

	return pools, nil
}

// GetPoolItems 获取掉落池中的所有物品
func (r *dropPoolRepositoryImpl) GetPoolItems(ctx context.Context, poolID string) ([]*game_config.DropPoolItem, error) {
	items, err := game_config.DropPoolItems(
		qm.Where("drop_pool_id = ? AND deleted_at IS NULL", poolID),
		qm.Where("is_active = ?", true),
		qm.OrderBy("drop_weight DESC, drop_rate DESC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询掉落池物品失败: %w", err)
	}

	return items, nil
}

// GetPoolItemsByLevel 获取掉落池中符合等级要求的物品
func (r *dropPoolRepositoryImpl) GetPoolItemsByLevel(ctx context.Context, poolID string, playerLevel int) ([]*game_config.DropPoolItem, error) {
	items, err := game_config.DropPoolItems(
		qm.Where("drop_pool_id = ? AND deleted_at IS NULL", poolID),
		qm.Where("is_active = ?", true),
		qm.Where("(min_level IS NULL OR min_level <= ?)", playerLevel),
		qm.Where("(max_level IS NULL OR max_level >= ?)", playerLevel),
		qm.OrderBy("drop_weight DESC, drop_rate DESC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询掉落池物品失败: %w", err)
	}

	return items, nil
}

// List 查询掉落池列表
func (r *dropPoolRepositoryImpl) List(ctx context.Context, params interfaces.ListDropPoolParams) ([]*game_config.DropPool, int64, error) {
	// 构建查询条件
	mods := []qm.QueryMod{
		qm.Where("deleted_at IS NULL"),
	}

	// 添加筛选条件
	if params.PoolType != nil {
		mods = append(mods, qm.Where("pool_type = ?", *params.PoolType))
	}
	if params.IsActive != nil {
		mods = append(mods, qm.Where("is_active = ?", *params.IsActive))
	}
	if params.Keyword != nil && *params.Keyword != "" {
		keyword := "%" + *params.Keyword + "%"
		mods = append(mods, qm.Where("(pool_code ILIKE ? OR pool_name ILIKE ?)", keyword, keyword))
	}

	// 查询总数
	count, err := game_config.DropPools(mods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询掉落池总数失败: %w", err)
	}

	// 添加分页和排序
	if params.PageSize > 0 {
		offset := (params.Page - 1) * params.PageSize
		mods = append(mods, qm.Limit(params.PageSize), qm.Offset(offset))
	}

	// 排序
	sortBy := params.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortOrder := params.SortOrder
	if sortOrder == "" {
		sortOrder = "DESC"
	}
	mods = append(mods, qm.OrderBy(fmt.Sprintf("%s %s", sortBy, sortOrder)))

	// 查询列表
	pools, err := game_config.DropPools(mods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询掉落池列表失败: %w", err)
	}

	return pools, count, nil
}

// Create 创建掉落池
func (r *dropPoolRepositoryImpl) Create(ctx context.Context, pool *game_config.DropPool) error {
	return pool.Insert(ctx, r.db, boil.Infer())
}

// Update 更新掉落池
func (r *dropPoolRepositoryImpl) Update(ctx context.Context, pool *game_config.DropPool) error {
	_, err := pool.Update(ctx, r.db, boil.Infer())
	return err
}

// Delete 删除掉落池（软删除）
func (r *dropPoolRepositoryImpl) Delete(ctx context.Context, poolID string) error {
	pool, err := r.GetByID(ctx, poolID)
	if err != nil {
		return err
	}

	_, err = pool.Delete(ctx, r.db, true) // soft delete
	return err
}

// Count 统计掉落池数量
func (r *dropPoolRepositoryImpl) Count(ctx context.Context, params interfaces.ListDropPoolParams) (int64, error) {
	mods := []qm.QueryMod{
		qm.Where("deleted_at IS NULL"),
	}

	if params.PoolType != nil {
		mods = append(mods, qm.Where("pool_type = ?", *params.PoolType))
	}
	if params.IsActive != nil {
		mods = append(mods, qm.Where("is_active = ?", *params.IsActive))
	}
	if params.Keyword != nil && *params.Keyword != "" {
		keyword := "%" + *params.Keyword + "%"
		mods = append(mods, qm.Where("(pool_code ILIKE ? OR pool_name ILIKE ?)", keyword, keyword))
	}

	return game_config.DropPools(mods...).Count(ctx, r.db)
}

// CreatePoolItem 添加掉落池物品
func (r *dropPoolRepositoryImpl) CreatePoolItem(ctx context.Context, item *game_config.DropPoolItem) error {
	return item.Insert(ctx, r.db, boil.Infer())
}

// GetPoolItemByID 根据ID获取掉落池物品
func (r *dropPoolRepositoryImpl) GetPoolItemByID(ctx context.Context, poolID, itemID string) (*game_config.DropPoolItem, error) {
	item, err := game_config.DropPoolItems(
		qm.Where("drop_pool_id = ? AND item_id = ? AND deleted_at IS NULL", poolID, itemID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("掉落池物品不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询掉落池物品失败: %w", err)
	}

	return item, nil
}

// UpdatePoolItem 更新掉落池物品
func (r *dropPoolRepositoryImpl) UpdatePoolItem(ctx context.Context, item *game_config.DropPoolItem) error {
	_, err := item.Update(ctx, r.db, boil.Infer())
	return err
}

// DeletePoolItem 删除掉落池物品（软删除）
func (r *dropPoolRepositoryImpl) DeletePoolItem(ctx context.Context, poolID, itemID string) error {
	item, err := r.GetPoolItemByID(ctx, poolID, itemID)
	if err != nil {
		return err
	}

	_, err = item.Delete(ctx, r.db, true) // soft delete
	return err
}

// ListPoolItems 查询掉落池物品列表
func (r *dropPoolRepositoryImpl) ListPoolItems(ctx context.Context, params interfaces.ListDropPoolItemParams) ([]*game_config.DropPoolItem, int64, error) {
	mods := []qm.QueryMod{
		qm.Where("drop_pool_id = ? AND deleted_at IS NULL", params.DropPoolID),
	}

	if params.IsActive != nil {
		mods = append(mods, qm.Where("is_active = ?", *params.IsActive))
	}
	if params.MinLevel != nil {
		mods = append(mods, qm.Where("(max_level IS NULL OR max_level >= ?)", *params.MinLevel))
	}
	if params.MaxLevel != nil {
		mods = append(mods, qm.Where("(min_level IS NULL OR min_level <= ?)", *params.MaxLevel))
	}

	// 查询总数
	count, err := game_config.DropPoolItems(mods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询掉落池物品总数失败: %w", err)
	}

	// 添加分页
	if params.PageSize > 0 {
		offset := (params.Page - 1) * params.PageSize
		mods = append(mods, qm.Limit(params.PageSize), qm.Offset(offset))
	}

	// 排序
	mods = append(mods, qm.OrderBy("drop_weight DESC, drop_rate DESC"))

	// 查询列表
	items, err := game_config.DropPoolItems(mods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询掉落池物品列表失败: %w", err)
	}

	return items, count, nil
}

// CountPoolItems 统计掉落池物品数量
func (r *dropPoolRepositoryImpl) CountPoolItems(ctx context.Context, params interfaces.ListDropPoolItemParams) (int64, error) {
	mods := []qm.QueryMod{
		qm.Where("drop_pool_id = ? AND deleted_at IS NULL", params.DropPoolID),
	}

	if params.IsActive != nil {
		mods = append(mods, qm.Where("is_active = ?", *params.IsActive))
	}
	if params.MinLevel != nil {
		mods = append(mods, qm.Where("(max_level IS NULL OR max_level >= ?)", *params.MinLevel))
	}
	if params.MaxLevel != nil {
		mods = append(mods, qm.Where("(min_level IS NULL OR min_level <= ?)", *params.MaxLevel))
	}

	return game_config.DropPoolItems(mods...).Count(ctx, r.db)
}

