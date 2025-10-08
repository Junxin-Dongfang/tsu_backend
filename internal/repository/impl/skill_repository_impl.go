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

type skillRepositoryImpl struct {
	db *sql.DB
}

// NewSkillRepository 创建技能仓储实例
func NewSkillRepository(db *sql.DB) interfaces.SkillRepository {
	return &skillRepositoryImpl{db: db}
}

// GetByID 根据ID获取技能
func (r *skillRepositoryImpl) GetByID(ctx context.Context, skillID string) (*game_config.Skill, error) {
	skill, err := game_config.Skills(
		qm.Where("id = ? AND deleted_at IS NULL", skillID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("技能不存在: %s", skillID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询技能失败: %w", err)
	}

	return skill, nil
}

// GetByCode 根据代码获取技能
func (r *skillRepositoryImpl) GetByCode(ctx context.Context, code string) (*game_config.Skill, error) {
	skill, err := game_config.Skills(
		qm.Where("skill_code = ? AND deleted_at IS NULL", code),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("技能不存在: %s", code)
	}
	if err != nil {
		return nil, fmt.Errorf("查询技能失败: %w", err)
	}

	return skill, nil
}

// List 获取技能列表
func (r *skillRepositoryImpl) List(ctx context.Context, params interfaces.SkillQueryParams) ([]*game_config.Skill, int64, error) {
	// 构建基础查询条件
	var baseQueryMods []qm.QueryMod
	baseQueryMods = append(baseQueryMods, qm.Where("deleted_at IS NULL"))

	// 筛选条件
	if params.SkillType != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("skill_type = ?", *params.SkillType))
	}
	if params.CategoryID != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("category_id = ?", *params.CategoryID))
	}
	if params.IsActive != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("is_active = ?", *params.IsActive))
	}

	// 获取总数
	count, err := game_config.Skills(baseQueryMods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询技能总数失败: %w", err)
	}

	// 构建完整查询条件
	var queryMods []qm.QueryMod
	queryMods = append(queryMods, baseQueryMods...)

	// 排序
	queryMods = append(queryMods, qm.OrderBy("skill_code ASC"))

	// 分页
	if params.Limit > 0 {
		queryMods = append(queryMods, qm.Limit(params.Limit))
	}
	if params.Offset > 0 {
		queryMods = append(queryMods, qm.Offset(params.Offset))
	}

	// 查询列表
	skills, err := game_config.Skills(queryMods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询技能列表失败: %w", err)
	}

	return skills, count, nil
}

// Create 创建技能
func (r *skillRepositoryImpl) Create(ctx context.Context, skill *game_config.Skill) error {
	// 生成 UUID
	if skill.ID == "" {
		skill.ID = uuid.New().String()
	}

	// 设置时间戳
	now := time.Now()
	skill.CreatedAt.SetValid(now)
	skill.UpdatedAt.SetValid(now)

	// 插入数据库
	if err := skill.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("创建技能失败: %w", err)
	}

	return nil
}

// Update 更新技能
func (r *skillRepositoryImpl) Update(ctx context.Context, skill *game_config.Skill) error {
	// 更新时间戳
	skill.UpdatedAt.SetValid(time.Now())

	// 更新数据库
	if _, err := skill.Update(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("更新技能失败: %w", err)
	}

	return nil
}

// Delete 软删除技能
func (r *skillRepositoryImpl) Delete(ctx context.Context, skillID string) error {
	skill, err := r.GetByID(ctx, skillID)
	if err != nil {
		return err
	}

	// 软删除
	now := time.Now()
	skill.DeletedAt.SetValid(now)
	skill.UpdatedAt.SetValid(now)

	if _, err := skill.Update(ctx, r.db, boil.Whitelist("deleted_at", "updated_at")); err != nil {
		return fmt.Errorf("删除技能失败: %w", err)
	}

	return nil
}

// Exists 检查代码是否存在
func (r *skillRepositoryImpl) Exists(ctx context.Context, code string) (bool, error) {
	count, err := game_config.Skills(
		qm.Where("skill_code = ? AND deleted_at IS NULL", code),
	).Count(ctx, r.db)

	if err != nil {
		return false, fmt.Errorf("检查技能代码是否存在失败: %w", err)
	}

	return count > 0, nil
}
