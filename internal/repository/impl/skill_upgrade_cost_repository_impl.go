package impl

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/interfaces"
)

// SkillUpgradeCostRepositoryImpl 技能升级消耗仓储实现
type SkillUpgradeCostRepositoryImpl struct {
	db *sql.DB
}

// NewSkillUpgradeCostRepository 创建技能升级消耗仓储
func NewSkillUpgradeCostRepository(db *sql.DB) interfaces.SkillUpgradeCostRepository {
	return &SkillUpgradeCostRepositoryImpl{
		db: db,
	}
}

// Create 创建升级消耗配置
func (r *SkillUpgradeCostRepositoryImpl) Create(ctx context.Context, cost *game_config.SkillUpgradeCost) error {
	if err := cost.Insert(ctx, r.db, boil.Infer()); err != nil {
		return xerrors.Wrap(err, 600001, "创建升级消耗配置失败")
	}
	return nil
}

// GetByID 根据ID获取
func (r *SkillUpgradeCostRepositoryImpl) GetByID(ctx context.Context, id string) (*game_config.SkillUpgradeCost, error) {
	cost, err := game_config.FindSkillUpgradeCost(ctx, r.db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, xerrors.NewNotFoundError("skill_upgrade_cost", id)
		}
		return nil, xerrors.Wrap(err, 600002, "查询升级消耗配置失败")
	}
	return cost, nil
}

// GetByLevel 根据等级获取
func (r *SkillUpgradeCostRepositoryImpl) GetByLevel(ctx context.Context, level int) (*game_config.SkillUpgradeCost, error) {
	cost, err := game_config.SkillUpgradeCosts(
		qm.Where("level_number = ?", level),
	).One(ctx, r.db)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, xerrors.NewNotFoundError("skill_upgrade_cost", fmt.Sprintf("level_%d", level))
		}
		return nil, xerrors.Wrap(err, 600002, "查询升级消耗配置失败")
	}
	return cost, nil
}

// List 获取所有升级消耗配置
func (r *SkillUpgradeCostRepositoryImpl) List(ctx context.Context) ([]*game_config.SkillUpgradeCost, error) {
	costs, err := game_config.SkillUpgradeCosts(
		qm.OrderBy("level_number ASC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, xerrors.Wrap(err, 600003, "查询升级消耗配置列表失败")
	}
	return costs, nil
}

// Update 更新升级消耗配置
func (r *SkillUpgradeCostRepositoryImpl) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	cost, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 应用更新
	cols := make([]string, 0)
	for key, value := range updates {
		cols = append(cols, key)
		switch key {
		case "level_number":
			if v, ok := value.(int); ok {
				cost.LevelNumber = v
			}
		case "cost_xp":
			if v, ok := value.(int); ok {
				cost.CostXP.SetValid(v)
			}
		case "cost_gold":
			if v, ok := value.(int); ok {
				cost.CostGold.SetValid(v)
			}
		case "cost_materials":
			// JSONB 字段处理
			if v, ok := value.([]byte); ok {
				cost.CostMaterials.UnmarshalJSON(v)
			}
		}
	}

	if _, err := cost.Update(ctx, r.db, boil.Whitelist(cols...)); err != nil {
		return xerrors.Wrap(err, 600004, "更新升级消耗配置失败")
	}

	return nil
}

// Delete 删除升级消耗配置
func (r *SkillUpgradeCostRepositoryImpl) Delete(ctx context.Context, id string) error {
	cost, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if _, err := cost.Delete(ctx, r.db); err != nil {
		return xerrors.Wrap(err, 600005, "删除升级消耗配置失败")
	}

	return nil
}
