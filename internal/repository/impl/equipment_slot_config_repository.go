package impl

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/interfaces"
)

// EquipmentSlotConfigRepositoryImpl 装备槽位配置Repository实现
type EquipmentSlotConfigRepositoryImpl struct {
	db *sql.DB
}

// NewEquipmentSlotConfigRepository 创建装备槽位配置Repository
func NewEquipmentSlotConfigRepository(db *sql.DB) interfaces.EquipmentSlotConfigRepository {
	return &EquipmentSlotConfigRepositoryImpl{db: db}
}

// Create 创建槽位配置
func (r *EquipmentSlotConfigRepositoryImpl) Create(ctx context.Context, slot *game_config.EquipmentSlot) error {
	return slot.Insert(ctx, r.db, boil.Infer())
}

// GetByID 根据ID获取槽位配置
func (r *EquipmentSlotConfigRepositoryImpl) GetByID(ctx context.Context, id string) (*game_config.EquipmentSlot, error) {
	return game_config.EquipmentSlots(
		game_config.EquipmentSlotWhere.ID.EQ(id),
		game_config.EquipmentSlotWhere.DeletedAt.IsNull(),
	).One(ctx, r.db)
}

// GetByCode 根据槽位代码获取槽位配置
func (r *EquipmentSlotConfigRepositoryImpl) GetByCode(ctx context.Context, code string) (*game_config.EquipmentSlot, error) {
	return game_config.EquipmentSlots(
		game_config.EquipmentSlotWhere.SlotCode.EQ(code),
		game_config.EquipmentSlotWhere.DeletedAt.IsNull(),
	).One(ctx, r.db)
}

// List 查询槽位列表（支持分页和筛选）
func (r *EquipmentSlotConfigRepositoryImpl) List(ctx context.Context, params interfaces.ListSlotConfigParams) ([]*game_config.EquipmentSlot, error) {
	// 构建查询条件
	mods := []qm.QueryMod{
		game_config.EquipmentSlotWhere.DeletedAt.IsNull(),
	}

	// 槽位类型筛选
	if params.SlotType != nil {
		mods = append(mods, game_config.EquipmentSlotWhere.SlotType.EQ(*params.SlotType))
	}

	// 激活状态筛选
	if params.IsActive != nil {
		mods = append(mods, game_config.EquipmentSlotWhere.IsActive.EQ(*params.IsActive))
	}

	// 关键词搜索
	if params.Keyword != nil && *params.Keyword != "" {
		keyword := "%" + *params.Keyword + "%"
		mods = append(mods, qm.Where("slot_code ILIKE ? OR slot_name ILIKE ?", keyword, keyword))
	}

	// 排序
	sortBy := params.SortBy
	if sortBy == "" {
		sortBy = "display_order"
	}
	sortOrder := params.SortOrder
	if sortOrder == "" {
		sortOrder = "asc"
	}
	orderClause := fmt.Sprintf("%s %s", sortBy, sortOrder)
	mods = append(mods, qm.OrderBy(orderClause))

	// 分页
	if params.PageSize > 0 {
		offset := (params.Page - 1) * params.PageSize
		mods = append(mods, qm.Limit(params.PageSize), qm.Offset(offset))
	}

	return game_config.EquipmentSlots(mods...).All(ctx, r.db)
}

// Count 统计槽位数量（用于分页）
func (r *EquipmentSlotConfigRepositoryImpl) Count(ctx context.Context, params interfaces.ListSlotConfigParams) (int64, error) {
	// 构建查询条件
	mods := []qm.QueryMod{
		game_config.EquipmentSlotWhere.DeletedAt.IsNull(),
	}

	// 槽位类型筛选
	if params.SlotType != nil {
		mods = append(mods, game_config.EquipmentSlotWhere.SlotType.EQ(*params.SlotType))
	}

	// 激活状态筛选
	if params.IsActive != nil {
		mods = append(mods, game_config.EquipmentSlotWhere.IsActive.EQ(*params.IsActive))
	}

	// 关键词搜索
	if params.Keyword != nil && *params.Keyword != "" {
		keyword := "%" + *params.Keyword + "%"
		mods = append(mods, qm.Where("slot_code ILIKE ? OR slot_name ILIKE ?", keyword, keyword))
	}

	return game_config.EquipmentSlots(mods...).Count(ctx, r.db)
}

// Update 更新槽位配置
func (r *EquipmentSlotConfigRepositoryImpl) Update(ctx context.Context, slot *game_config.EquipmentSlot) error {
	_, err := slot.Update(ctx, r.db, boil.Infer())
	return err
}

// Delete 删除槽位配置（软删除）
func (r *EquipmentSlotConfigRepositoryImpl) Delete(ctx context.Context, id string) error {
	slot, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()
	slot.DeletedAt.SetValid(now)
	_, err = slot.Update(ctx, r.db, boil.Whitelist(game_config.EquipmentSlotColumns.DeletedAt))
	return err
}

