// Package impl 提供Repository接口的实现
package impl

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/interfaces"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

// EquipmentSetRepositoryImpl 装备套装Repository实现
type EquipmentSetRepositoryImpl struct {
	db *sql.DB
}

// NewEquipmentSetRepository 创建装备套装Repository
func NewEquipmentSetRepository(db *sql.DB) interfaces.EquipmentSetRepository {
	return &EquipmentSetRepositoryImpl{
		db: db,
	}
}

// GetByID 根据ID查询套装配置
func (r *EquipmentSetRepositoryImpl) GetByID(ctx context.Context, id string) (*game_config.EquipmentSetConfig, error) {
	setConfig, err := game_config.FindEquipmentSetConfig(ctx, r.db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, xerrors.NewNotFoundError("equipment_set", id)
		}
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "failed to get equipment set by ID")
	}
	return setConfig, nil
}

// GetByIDs 根据ID列表批量查询套装配置
func (r *EquipmentSetRepositoryImpl) GetByIDs(ctx context.Context, ids []string) ([]*game_config.EquipmentSetConfig, error) {
	if len(ids) == 0 {
		return []*game_config.EquipmentSetConfig{}, nil
	}

	// Convert string IDs to interface{} for IN clause
	idsInterface := make([]interface{}, len(ids))
	for i, id := range ids {
		idsInterface[i] = id
	}

	setConfigs, err := game_config.EquipmentSetConfigs(
		qm.WhereIn("id IN ?", idsInterface...),
		qm.Where("deleted_at IS NULL"),
	).All(ctx, r.db)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "failed to get equipment sets by IDs")
	}

	return setConfigs, nil
}

// GetByCode 根据代码查询套装配置
func (r *EquipmentSetRepositoryImpl) GetByCode(ctx context.Context, code string) (*game_config.EquipmentSetConfig, error) {
	setConfig, err := game_config.EquipmentSetConfigs(
		qm.Where("set_code = ?", code),
		qm.Where("deleted_at IS NULL"),
	).One(ctx, r.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 不存在返回nil，不是错误
		}
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "failed to get equipment set by code")
	}
	return setConfig, nil
}

// List 查询套装列表
func (r *EquipmentSetRepositoryImpl) List(ctx context.Context, req *interfaces.ListEquipmentSetsRequest) ([]*game_config.EquipmentSetConfig, int64, error) {
	// 设置默认分页参数
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 构建查询条件
	queryMods := r.buildQueryMods(req)

	// 查询总数
	total, err := r.Count(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	// 添加排序和分页
	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortOrder := req.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}
	orderClause := fmt.Sprintf("%s %s", sortBy, sortOrder)
	queryMods = append(queryMods,
		qm.OrderBy(orderClause),
		qm.Limit(pageSize),
		qm.Offset(offset),
	)

	// 查询列表
	setConfigs, err := game_config.EquipmentSetConfigs(queryMods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, xerrors.Wrap(err, xerrors.CodeInternalError, "failed to list equipment sets")
	}

	return setConfigs, total, nil
}

// Count 统计套装总数
func (r *EquipmentSetRepositoryImpl) Count(ctx context.Context, req *interfaces.ListEquipmentSetsRequest) (int64, error) {
	queryMods := r.buildQueryMods(req)
	total, err := game_config.EquipmentSetConfigs(queryMods...).Count(ctx, r.db)
	if err != nil {
		return 0, xerrors.Wrap(err, xerrors.CodeInternalError, "failed to count equipment sets")
	}
	return total, nil
}

// buildQueryMods 构建查询条件
func (r *EquipmentSetRepositoryImpl) buildQueryMods(req *interfaces.ListEquipmentSetsRequest) []qm.QueryMod {
	queryMods := []qm.QueryMod{
		qm.Where("deleted_at IS NULL"),
	}

	// 关键词搜索
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		queryMods = append(queryMods, qm.Where("(set_code ILIKE ? OR set_name ILIKE ? OR description ILIKE ?)", keyword, keyword, keyword))
	}

	// 激活状态筛选
	if req.IsActive != nil {
		queryMods = append(queryMods, qm.Where("is_active = ?", *req.IsActive))
	}

	return queryMods
}

// Create 创建套装配置
func (r *EquipmentSetRepositoryImpl) Create(ctx context.Context, config *game_config.EquipmentSetConfig) error {
	err := config.Insert(ctx, r.db, boil.Infer())
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "failed to create equipment set")
	}
	return nil
}

// Update 更新套装配置
func (r *EquipmentSetRepositoryImpl) Update(ctx context.Context, config *game_config.EquipmentSetConfig) error {
	_, err := config.Update(ctx, r.db, boil.Infer())
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "failed to update equipment set")
	}
	return nil
}

// Delete 软删除套装配置
func (r *EquipmentSetRepositoryImpl) Delete(ctx context.Context, id string) error {
	// 查询套装配置
	setConfig, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 设置删除时间
	setConfig.DeletedAt.SetValid(time.Now())

	// 更新记录
	_, err = setConfig.Update(ctx, r.db, boil.Infer())
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "failed to delete equipment set")
	}

	return nil
}

// GetItemsBySetID 查询套装包含的装备列表
func (r *EquipmentSetRepositoryImpl) GetItemsBySetID(ctx context.Context, setID string) ([]*game_config.Item, error) {
	items, err := game_config.Items(
		qm.Where("set_id = ?", setID),
		qm.Where("deleted_at IS NULL"),
		qm.OrderBy("item_code"),
	).All(ctx, r.db)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, fmt.Sprintf("failed to get items by set ID: %s", setID))
	}

	return items, nil
}
