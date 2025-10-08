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

type skillCategoryRepositoryImpl struct {
	db *sql.DB
}

// NewSkillCategoryRepository 创建技能类别仓储实例
func NewSkillCategoryRepository(db *sql.DB) interfaces.SkillCategoryRepository {
	return &skillCategoryRepositoryImpl{db: db}
}

// GetByID 根据ID获取技能类别
func (r *skillCategoryRepositoryImpl) GetByID(ctx context.Context, categoryID string) (*game_config.SkillCategory, error) {
	category, err := game_config.SkillCategories(
		qm.Where("id = ? AND deleted_at IS NULL", categoryID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("技能类别不存在: %s", categoryID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询技能类别失败: %w", err)
	}

	return category, nil
}

// GetByCode 根据代码获取技能类别
func (r *skillCategoryRepositoryImpl) GetByCode(ctx context.Context, code string) (*game_config.SkillCategory, error) {
	category, err := game_config.SkillCategories(
		qm.Where("category_code = ? AND deleted_at IS NULL", code),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("技能类别不存在: %s", code)
	}
	if err != nil {
		return nil, fmt.Errorf("查询技能类别失败: %w", err)
	}

	return category, nil
}

// List 获取技能类别列表
func (r *skillCategoryRepositoryImpl) List(ctx context.Context, params interfaces.SkillCategoryQueryParams) ([]*game_config.SkillCategory, int64, error) {
	// 构建基础查询条件
	var baseQueryMods []qm.QueryMod
	baseQueryMods = append(baseQueryMods, qm.Where("deleted_at IS NULL"))

	// 筛选条件
	if params.IsActive != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("is_active = ?", *params.IsActive))
	}

	// 获取总数 (只使用筛选条件)
	count, err := game_config.SkillCategories(baseQueryMods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询技能类别总数失败: %w", err)
	}

	// 构建完整查询条件 (包含排序和分页)
	var queryMods []qm.QueryMod
	queryMods = append(queryMods, baseQueryMods...)

	// 排序
	queryMods = append(queryMods, qm.OrderBy("display_order ASC, created_at DESC"))

	// 分页
	if params.Limit > 0 {
		queryMods = append(queryMods, qm.Limit(params.Limit))
	}
	if params.Offset > 0 {
		queryMods = append(queryMods, qm.Offset(params.Offset))
	}

	// 查询列表
	categories, err := game_config.SkillCategories(queryMods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询技能类别列表失败: %w", err)
	}

	return categories, count, nil
}

// Create 创建技能类别
func (r *skillCategoryRepositoryImpl) Create(ctx context.Context, category *game_config.SkillCategory) error {
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
		return fmt.Errorf("创建技能类别失败: %w", err)
	}

	return nil
}

// Update 更新技能类别
func (r *skillCategoryRepositoryImpl) Update(ctx context.Context, category *game_config.SkillCategory) error {
	// 更新时间戳
	category.UpdatedAt.SetValid(time.Now())

	// 更新数据库
	if _, err := category.Update(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("更新技能类别失败: %w", err)
	}

	return nil
}

// Delete 软删除技能类别
func (r *skillCategoryRepositoryImpl) Delete(ctx context.Context, categoryID string) error {
	category, err := r.GetByID(ctx, categoryID)
	if err != nil {
		return err
	}

	// 软删除
	category.DeletedAt.SetValid(time.Now())
	category.UpdatedAt.SetValid(time.Now())

	if _, err := category.Update(ctx, r.db, boil.Whitelist("deleted_at", "updated_at")); err != nil {
		return fmt.Errorf("删除技能类别失败: %w", err)
	}

	return nil
}

// Exists 检查代码是否存在
func (r *skillCategoryRepositoryImpl) Exists(ctx context.Context, code string) (bool, error) {
	count, err := game_config.SkillCategories(
		qm.Where("category_code = ? AND deleted_at IS NULL", code),
	).Count(ctx, r.db)

	if err != nil {
		return false, fmt.Errorf("检查技能类别代码是否存在失败: %w", err)
	}

	return count > 0, nil
}
