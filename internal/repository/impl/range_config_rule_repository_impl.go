package impl

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/queries/qm"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/interfaces"
)

type rangeConfigRuleRepositoryImpl struct {
	db *sql.DB
}

func NewRangeConfigRuleRepository(db *sql.DB) interfaces.RangeConfigRuleRepository {
	return &rangeConfigRuleRepositoryImpl{db: db}
}

func (r *rangeConfigRuleRepositoryImpl) GetByID(ctx context.Context, id string) (*game_config.RangeConfigRule, error) {
	rule, err := game_config.RangeConfigRules(
		qm.Where("id = ?", id),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("范围配置规则不存在: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("查询范围配置规则失败: %w", err)
	}

	return rule, nil
}

func (r *rangeConfigRuleRepositoryImpl) List(ctx context.Context, params interfaces.RangeConfigRuleQueryParams) ([]*game_config.RangeConfigRule, int64, error) {
	var baseQueryMods []qm.QueryMod

	if params.IsActive != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("is_active = ?", *params.IsActive))
	}

	count, err := game_config.RangeConfigRules(baseQueryMods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询范围配置规则总数失败: %w", err)
	}

	var queryMods []qm.QueryMod
	queryMods = append(queryMods, baseQueryMods...)
	queryMods = append(queryMods, qm.OrderBy("parameter_type ASC"))

	if params.Limit > 0 {
		queryMods = append(queryMods, qm.Limit(params.Limit))
	}
	if params.Offset > 0 {
		queryMods = append(queryMods, qm.Offset(params.Offset))
	}

	rules, err := game_config.RangeConfigRules(queryMods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询范围配置规则列表失败: %w", err)
	}

	return rules, count, nil
}

func (r *rangeConfigRuleRepositoryImpl) GetAll(ctx context.Context) ([]*game_config.RangeConfigRule, error) {
	rules, err := game_config.RangeConfigRules(
		qm.Where("is_active = ?", true),
		qm.OrderBy("parameter_type ASC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询所有范围配置规则失败: %w", err)
	}

	return rules, nil
}
