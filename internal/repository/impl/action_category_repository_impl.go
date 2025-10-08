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

type actionCategoryRepositoryImpl struct {
	db *sql.DB
}

// NewActionCategoryRepository 创建动作类别仓储实例
func NewActionCategoryRepository(db *sql.DB) interfaces.ActionCategoryRepository {
	return &actionCategoryRepositoryImpl{db: db}
}

// GetByID 根据ID获取动作类别
func (r *actionCategoryRepositoryImpl) GetByID(ctx context.Context, categoryID string) (*game_config.ActionCategory, error) {
	category, err := game_config.ActionCategories(
		qm.Where("id = ? AND deleted_at IS NULL", categoryID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("动作类别不存在: %s", categoryID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询动作类别失败: %w", err)
	}

	return category, nil
}

// GetByCode 根据代码获取动作类别
func (r *actionCategoryRepositoryImpl) GetByCode(ctx context.Context, code string) (*game_config.ActionCategory, error) {
	category, err := game_config.ActionCategories(
		qm.Where("category_code = ? AND deleted_at IS NULL", code),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("动作类别不存在: %s", code)
	}
	if err != nil {
		return nil, fmt.Errorf("查询动作类别失败: %w", err)
	}

	return category, nil
}

// List 获取动作类别列表
func (r *actionCategoryRepositoryImpl) List(ctx context.Context, params interfaces.ActionCategoryQueryParams) ([]*game_config.ActionCategory, int64, error) {
	// 构建基础查询条件
	var baseQueryMods []qm.QueryMod
	baseQueryMods = append(baseQueryMods, qm.Where("deleted_at IS NULL"))

	// 筛选条件
	if params.IsActive != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("is_active = ?", *params.IsActive))
	}

	// 获取总数 (只使用筛选条件)
	count, err := game_config.ActionCategories(baseQueryMods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询动作类别总数失败: %w", err)
	}

	// 构建完整查询条件 (包含排序和分页)
	var queryMods []qm.QueryMod
	queryMods = append(queryMods, baseQueryMods...)

	// 排序 (按创建时间倒序)
	queryMods = append(queryMods, qm.OrderBy("created_at DESC"))

	// 分页
	if params.Limit > 0 {
		queryMods = append(queryMods, qm.Limit(params.Limit))
	}
	if params.Offset > 0 {
		queryMods = append(queryMods, qm.Offset(params.Offset))
	}

	// 查询列表
	categories, err := game_config.ActionCategories(queryMods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询动作类别列表失败: %w", err)
	}

	return categories, count, nil
}

// Create 创建动作类别
func (r *actionCategoryRepositoryImpl) Create(ctx context.Context, category *game_config.ActionCategory) error {
	// 生成 UUID
	if category.ID == "" {
		category.ID = uuid.New().String()
	}

	// 设置时间戳
	now := time.Now()
	category.CreatedAt.SetValid(now)
	category.UpdatedAt.SetValid(now)

	// 插入数据库
	if err := category.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("创建动作类别失败: %w", err)
	}

	return nil
}

// Update 更新动作类别
func (r *actionCategoryRepositoryImpl) Update(ctx context.Context, category *game_config.ActionCategory) error {
	// 更新时间戳
	category.UpdatedAt.SetValid(time.Now())

	// 更新数据库
	if _, err := category.Update(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("更新动作类别失败: %w", err)
	}

	return nil
}

// Delete 软删除动作类别
func (r *actionCategoryRepositoryImpl) Delete(ctx context.Context, categoryID string) error {
	category, err := r.GetByID(ctx, categoryID)
	if err != nil {
		return err
	}

	// 软删除
	category.DeletedAt.SetValid(time.Now())
	category.UpdatedAt.SetValid(time.Now())

	if _, err := category.Update(ctx, r.db, boil.Whitelist("deleted_at", "updated_at")); err != nil {
		return fmt.Errorf("删除动作类别失败: %w", err)
	}

	return nil
}

// Exists 检查代码是否存在
func (r *actionCategoryRepositoryImpl) Exists(ctx context.Context, code string) (bool, error) {
	count, err := game_config.ActionCategories(
		qm.Where("category_code = ? AND deleted_at IS NULL", code),
	).Count(ctx, r.db)

	if err != nil {
		return false, fmt.Errorf("检查动作类别代码是否存在失败: %w", err)
	}

	return count > 0, nil
}
