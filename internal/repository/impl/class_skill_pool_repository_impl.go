package impl

import (
	"context"
	"database/sql"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/interfaces"
)

// ClassSkillPoolRepositoryImpl 职业技能池仓储实现
type ClassSkillPoolRepositoryImpl struct {
	db *sql.DB
}

// NewClassSkillPoolRepository 创建职业技能池仓储
func NewClassSkillPoolRepository(db *sql.DB) interfaces.ClassSkillPoolRepository {
	return &ClassSkillPoolRepositoryImpl{db: db}
}

// GetClassSkillPools 获取职业技能池列表
func (r *ClassSkillPoolRepositoryImpl) GetClassSkillPools(ctx context.Context, params interfaces.ClassSkillPoolQueryParams) ([]*game_config.ClassSkillPool, int64, error) {
	// 构建查询条件
	mods := []qm.QueryMod{
		qm.Where("deleted_at IS NULL"),
	}

	if params.ClassID != nil {
		mods = append(mods, qm.Where("class_id = ?", *params.ClassID))
	}

	if params.SkillID != nil {
		mods = append(mods, qm.Where("skill_id = ?", *params.SkillID))
	}

	if params.SkillTier != nil {
		mods = append(mods, qm.Where("skill_tier = ?", *params.SkillTier))
	}

	if params.IsCore != nil {
		mods = append(mods, qm.Where("is_core = ?", *params.IsCore))
	}

	if params.IsExclusive != nil {
		mods = append(mods, qm.Where("is_exclusive = ?", *params.IsExclusive))
	}

	if params.IsVisible != nil {
		mods = append(mods, qm.Where("is_visible = ?", *params.IsVisible))
	}

	// 获取总数
	countMods := append([]qm.QueryMod{}, mods...)
	total, err := game_config.ClassSkillPools(countMods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, err
	}

	// 分页
	if params.Limit > 0 {
		mods = append(mods, qm.Limit(params.Limit))
	}
	if params.Offset > 0 {
		mods = append(mods, qm.Offset(params.Offset))
	}

	// 排序
	mods = append(mods, qm.OrderBy("display_order ASC, created_at DESC"))

	// 查询
	pools, err := game_config.ClassSkillPools(mods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, err
	}

	return pools, total, nil
}

// GetClassSkillPoolByID 根据ID获取职业技能池
func (r *ClassSkillPoolRepositoryImpl) GetClassSkillPoolByID(ctx context.Context, id string) (*game_config.ClassSkillPool, error) {
	pool, err := game_config.ClassSkillPools(
		qm.Where("id = ?", id),
		qm.Where("deleted_at IS NULL"),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return pool, err
}

// GetClassSkillPoolsByClassID 获取指定职业的所有技能
func (r *ClassSkillPoolRepositoryImpl) GetClassSkillPoolsByClassID(ctx context.Context, classID string) ([]*game_config.ClassSkillPool, error) {
	pools, err := game_config.ClassSkillPools(
		qm.Where("class_id = ?", classID),
		qm.Where("deleted_at IS NULL"),
		qm.OrderBy("display_order ASC, skill_tier ASC"),
	).All(ctx, r.db)

	return pools, err
}

// CreateClassSkillPool 创建职业技能池配置
func (r *ClassSkillPoolRepositoryImpl) CreateClassSkillPool(ctx context.Context, pool *game_config.ClassSkillPool) error {
	return pool.Insert(ctx, r.db, boil.Infer())
}

// UpdateClassSkillPool 更新职业技能池配置
func (r *ClassSkillPoolRepositoryImpl) UpdateClassSkillPool(ctx context.Context, id string, updates map[string]interface{}) error {
	// 添加 updated_at
	updates["updated_at"] = time.Now()

	// 更新
	_, err := game_config.ClassSkillPools(
		qm.Where("id = ?", id),
		qm.Where("deleted_at IS NULL"),
	).UpdateAll(ctx, r.db, updates)

	return err
}

// DeleteClassSkillPool 删除职业技能池配置（软删除）
func (r *ClassSkillPoolRepositoryImpl) DeleteClassSkillPool(ctx context.Context, id string) error {
	updates := map[string]interface{}{
		"deleted_at": time.Now(),
	}

	_, err := game_config.ClassSkillPools(
		qm.Where("id = ?", id),
		qm.Where("deleted_at IS NULL"),
	).UpdateAll(ctx, r.db, updates)

	return err
}
