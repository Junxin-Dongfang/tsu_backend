package impl

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/google/uuid"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/interfaces"
)

type damageTypeRepositoryImpl struct {
	db *sql.DB
}

// NewDamageTypeRepository 创建伤害类型仓储实例
func NewDamageTypeRepository(db *sql.DB) interfaces.DamageTypeRepository {
	return &damageTypeRepositoryImpl{db: db}
}

// GetByID 根据ID获取伤害类型
func (r *damageTypeRepositoryImpl) GetByID(ctx context.Context, damageTypeID string) (*game_config.DamageType, error) {
	damageType, err := game_config.DamageTypes(
		qm.Where("id = ? AND deleted_at IS NULL", damageTypeID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("伤害类型不存在: %s", damageTypeID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询伤害类型失败: %w", err)
	}

	return damageType, nil
}

// GetByCode 根据代码获取伤害类型
func (r *damageTypeRepositoryImpl) GetByCode(ctx context.Context, code string) (*game_config.DamageType, error) {
	damageType, err := game_config.DamageTypes(
		qm.Where("code = ? AND deleted_at IS NULL", code),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("伤害类型不存在: %s", code)
	}
	if err != nil {
		return nil, fmt.Errorf("查询伤害类型失败: %w", err)
	}

	return damageType, nil
}

// List 获取伤害类型列表
func (r *damageTypeRepositoryImpl) List(ctx context.Context, params interfaces.DamageTypeQueryParams) ([]*game_config.DamageType, int64, error) {
	// 构建基础查询条件
	var baseQueryMods []qm.QueryMod
	baseQueryMods = append(baseQueryMods, qm.Where("deleted_at IS NULL"))

	// 筛选条件
	if params.Category != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("category = ?", *params.Category))
	}
	if params.IsActive != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("is_active = ?", *params.IsActive))
	}

	// 获取总数 (只使用筛选条件)
	count, err := game_config.DamageTypes(baseQueryMods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询伤害类型总数失败: %w", err)
	}

	// 构建完整查询条件 (包含排序和分页)
	var queryMods []qm.QueryMod
	queryMods = append(queryMods, baseQueryMods...)

	// 排序 (按类别和名称)
	queryMods = append(queryMods, qm.OrderBy("category ASC, name ASC"))

	// 分页
	if params.Limit > 0 {
		queryMods = append(queryMods, qm.Limit(params.Limit))
	}
	if params.Offset > 0 {
		queryMods = append(queryMods, qm.Offset(params.Offset))
	}

	// 查询列表
	damageTypes, err := game_config.DamageTypes(queryMods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询伤害类型列表失败: %w", err)
	}

	return damageTypes, count, nil
}

// Create 创建伤害类型
func (r *damageTypeRepositoryImpl) Create(ctx context.Context, damageType *game_config.DamageType) error {
	// 生成 UUID
	if damageType.ID == "" {
		damageType.ID = uuid.New().String()
	}

	// 设置时间戳
	now := time.Now()
	damageType.CreatedAt.SetValid(now)
	damageType.UpdatedAt.SetValid(now)

	// 插入数据库
	if err := damageType.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("创建伤害类型失败: %w", err)
	}

	return nil
}

// Update 更新伤害类型
func (r *damageTypeRepositoryImpl) Update(ctx context.Context, damageType *game_config.DamageType) error {
	// 更新时间戳
	damageType.UpdatedAt.SetValid(time.Now())

	// 更新数据库
	if _, err := damageType.Update(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("更新伤害类型失败: %w", err)
	}

	return nil
}

// Delete 软删除伤害类型
func (r *damageTypeRepositoryImpl) Delete(ctx context.Context, damageTypeID string) error {
	damageType, err := r.GetByID(ctx, damageTypeID)
	if err != nil {
		return err
	}

	// 软删除
	damageType.DeletedAt.SetValid(time.Now())
	damageType.UpdatedAt.SetValid(time.Now())

	if _, err := damageType.Update(ctx, r.db, boil.Whitelist("deleted_at", "updated_at")); err != nil {
		return fmt.Errorf("删除伤害类型失败: %w", err)
	}

	return nil
}

// Exists 检查代码是否存在
func (r *damageTypeRepositoryImpl) Exists(ctx context.Context, code string) (bool, error) {
	count, err := game_config.DamageTypes(
		qm.Where("code = ? AND deleted_at IS NULL", code),
	).Count(ctx, r.db)

	if err != nil {
		return false, fmt.Errorf("检查伤害类型代码是否存在失败: %w", err)
	}

	return count > 0, nil
}
