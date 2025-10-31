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

type itemRepositoryImpl struct {
	db *sql.DB
}

// NewItemRepository 创建装备配置仓储实例
func NewItemRepository(db *sql.DB) interfaces.ItemRepository {
	return &itemRepositoryImpl{db: db}
}

// GetByID 根据ID获取装备配置
func (r *itemRepositoryImpl) GetByID(ctx context.Context, itemID string) (*game_config.Item, error) {
	item, err := game_config.Items(
		qm.Where("id = ? AND deleted_at IS NULL", itemID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("装备配置不存在: %s", itemID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询装备配置失败: %w", err)
	}

	return item, nil
}

// GetByCode 根据代码获取装备配置
func (r *itemRepositoryImpl) GetByCode(ctx context.Context, itemCode string) (*game_config.Item, error) {
	item, err := game_config.Items(
		qm.Where("item_code = ? AND deleted_at IS NULL", itemCode),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("装备配置不存在: %s", itemCode)
	}
	if err != nil {
		return nil, fmt.Errorf("查询装备配置失败: %w", err)
	}

	return item, nil
}

// List 查询装备配置列表
func (r *itemRepositoryImpl) List(ctx context.Context, params interfaces.ListItemParams) ([]*game_config.Item, int64, error) {
	// 构建查询条件
	mods := []qm.QueryMod{
		qm.Where("game_config.items.deleted_at IS NULL"),
	}

	// 添加筛选条件
	if params.ItemType != nil {
		mods = append(mods, qm.Where("game_config.items.item_type = ?", *params.ItemType))
	}
	if params.ItemQuality != nil {
		mods = append(mods, qm.Where("game_config.items.item_quality = ?", *params.ItemQuality))
	}
	if params.EquipSlot != nil {
		mods = append(mods, qm.Where("game_config.items.equip_slot = ?", *params.EquipSlot))
	}
	if params.MinLevel != nil {
		mods = append(mods, qm.Where("game_config.items.item_level >= ?", *params.MinLevel))
	}
	if params.MaxLevel != nil {
		mods = append(mods, qm.Where("game_config.items.item_level <= ?", *params.MaxLevel))
	}
	if params.IsActive != nil {
		mods = append(mods, qm.Where("game_config.items.is_active = ?", *params.IsActive))
	}

	// 关键词搜索
	if params.Keyword != nil && *params.Keyword != "" {
		keyword := "%" + *params.Keyword + "%"
		mods = append(mods, qm.Where("(game_config.items.item_code LIKE ? OR game_config.items.item_name LIKE ?)", keyword, keyword))
	}

	// 按标签筛选
	if len(params.TagIDs) > 0 {
		// 使用子查询筛选包含所有指定标签的物品
		mods = append(mods, qm.InnerJoin("game_config.tags_relations ON game_config.items.id = game_config.tags_relations.entity_id"))
		mods = append(mods, qm.Where("game_config.tags_relations.entity_type = ?", "item"))
		mods = append(mods, qm.WhereIn("game_config.tags_relations.tag_id IN ?", toInterfaceSlice(params.TagIDs)...))
		mods = append(mods, qm.GroupBy("game_config.items.id"))
		mods = append(mods, qm.Having("COUNT(DISTINCT game_config.tags_relations.tag_id) = ?", len(params.TagIDs)))
	}

	// 查询总数
	count, err := game_config.Items(mods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询装备配置总数失败: %w", err)
	}

	// 添加分页和排序
	if params.PageSize > 0 {
		offset := (params.Page - 1) * params.PageSize
		mods = append(mods, qm.Limit(params.PageSize), qm.Offset(offset))
	}
	mods = append(mods, qm.OrderBy("game_config.items.item_level DESC, game_config.items.item_quality DESC, game_config.items.created_at DESC"))

	// 查询列表
	items, err := game_config.Items(mods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询装备配置列表失败: %w", err)
	}

	return items, count, nil
}

// GetByType 根据类型获取装备配置列表
func (r *itemRepositoryImpl) GetByType(ctx context.Context, itemType string) ([]*game_config.Item, error) {
	items, err := game_config.Items(
		qm.Where("item_type = ? AND deleted_at IS NULL", itemType),
		qm.Where("is_active = ?", true),
		qm.OrderBy("item_level DESC, item_quality DESC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询装备配置列表失败: %w", err)
	}

	return items, nil
}

// GetByIDs 根据ID列表批量获取装备配置
func (r *itemRepositoryImpl) GetByIDs(ctx context.Context, itemIDs []string) ([]*game_config.Item, error) {
	if len(itemIDs) == 0 {
		return []*game_config.Item{}, nil
	}

	items, err := game_config.Items(
		qm.WhereIn("id IN ?", toInterfaceSlice(itemIDs)...),
		qm.Where("deleted_at IS NULL"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("批量查询装备配置失败: %w", err)
	}

	return items, nil
}

// Create 创建装备配置
func (r *itemRepositoryImpl) Create(ctx context.Context, item *game_config.Item) error {
	err := item.Insert(ctx, r.db, boil.Infer())
	if err != nil {
		return fmt.Errorf("创建装备配置失败: %w", err)
	}
	return nil
}

// Update 更新装备配置
func (r *itemRepositoryImpl) Update(ctx context.Context, item *game_config.Item) error {
	_, err := item.Update(ctx, r.db, boil.Infer())
	if err != nil {
		return fmt.Errorf("更新装备配置失败: %w", err)
	}
	return nil
}

// Delete 删除装备配置(软删除)
func (r *itemRepositoryImpl) Delete(ctx context.Context, itemID string) error {
	item, err := r.GetByID(ctx, itemID)
	if err != nil {
		return err
	}

	_, err = item.Delete(ctx, r.db, true) // true表示软删除
	if err != nil {
		return fmt.Errorf("删除装备配置失败: %w", err)
	}
	return nil
}

// CheckCodeExists 检查物品代码是否已存在
func (r *itemRepositoryImpl) CheckCodeExists(ctx context.Context, itemCode string, excludeID *string) (bool, error) {
	mods := []qm.QueryMod{
		qm.Where("item_code = ? AND deleted_at IS NULL", itemCode),
	}

	if excludeID != nil {
		mods = append(mods, qm.Where("id != ?", *excludeID))
	}

	count, err := game_config.Items(mods...).Count(ctx, r.db)
	if err != nil {
		return false, fmt.Errorf("检查物品代码是否存在失败: %w", err)
	}

	return count > 0, nil
}

// GetUnassignedItems 查询未关联套装的装备
func (r *itemRepositoryImpl) GetUnassignedItems(ctx context.Context, params interfaces.ListItemParams) ([]*game_config.Item, int64, error) {
	// 构建查询条件
	mods := []qm.QueryMod{
		qm.Where("game_config.items.deleted_at IS NULL"),
		qm.Where("game_config.items.set_id IS NULL"), // 未关联套装
		qm.Where("game_config.items.item_type = ?", "equipment"), // 只查询装备类型
	}

	// 添加筛选条件
	if params.ItemQuality != nil {
		mods = append(mods, qm.Where("game_config.items.item_quality = ?", *params.ItemQuality))
	}
	if params.EquipSlot != nil {
		mods = append(mods, qm.Where("game_config.items.equip_slot = ?", *params.EquipSlot))
	}
	if params.Keyword != nil {
		keyword := "%" + *params.Keyword + "%"
		mods = append(mods, qm.Where("(game_config.items.item_code ILIKE ? OR game_config.items.item_name ILIKE ?)", keyword, keyword))
	}

	// 查询总数
	countMods := make([]qm.QueryMod, len(mods))
	copy(countMods, mods)
	total, err := game_config.Items(countMods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("统计未关联套装的装备数量失败: %w", err)
	}

	// 分页参数
	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	// 添加排序和分页
	mods = append(mods, qm.OrderBy("game_config.items.item_code"))
	mods = append(mods, qm.Limit(pageSize))
	mods = append(mods, qm.Offset(offset))

	// 查询列表
	items, err := game_config.Items(mods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询未关联套装的装备列表失败: %w", err)
	}

	return items, total, nil
}

