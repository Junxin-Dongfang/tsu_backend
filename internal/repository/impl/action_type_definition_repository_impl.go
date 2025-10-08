package impl

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/queries/qm"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/interfaces"
)

type actionTypeDefinitionRepositoryImpl struct {
	db *sql.DB
}

func NewActionTypeDefinitionRepository(db *sql.DB) interfaces.ActionTypeDefinitionRepository {
	return &actionTypeDefinitionRepositoryImpl{db: db}
}

func (r *actionTypeDefinitionRepositoryImpl) GetByID(ctx context.Context, id string) (*game_config.ActionTypeDefinition, error) {
	def, err := game_config.ActionTypeDefinitions(
		qm.Where("id = ?", id),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("动作类型定义不存在: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("查询动作类型定义失败: %w", err)
	}

	return def, nil
}

func (r *actionTypeDefinitionRepositoryImpl) GetByActionType(ctx context.Context, actionType string) (*game_config.ActionTypeDefinition, error) {
	def, err := game_config.ActionTypeDefinitions(
		qm.Where("action_type = ?", actionType),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("动作类型定义不存在: %s", actionType)
	}
	if err != nil {
		return nil, fmt.Errorf("查询动作类型定义失败: %w", err)
	}

	return def, nil
}

func (r *actionTypeDefinitionRepositoryImpl) List(ctx context.Context, params interfaces.ActionTypeDefinitionQueryParams) ([]*game_config.ActionTypeDefinition, int64, error) {
	var baseQueryMods []qm.QueryMod

	if params.IsActive != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("is_active = ?", *params.IsActive))
	}

	count, err := game_config.ActionTypeDefinitions(baseQueryMods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询动作类型定义总数失败: %w", err)
	}

	var queryMods []qm.QueryMod
	queryMods = append(queryMods, baseQueryMods...)
	queryMods = append(queryMods, qm.OrderBy("action_type ASC"))

	if params.Limit > 0 {
		queryMods = append(queryMods, qm.Limit(params.Limit))
	}
	if params.Offset > 0 {
		queryMods = append(queryMods, qm.Offset(params.Offset))
	}

	defs, err := game_config.ActionTypeDefinitions(queryMods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询动作类型定义列表失败: %w", err)
	}

	return defs, count, nil
}

func (r *actionTypeDefinitionRepositoryImpl) GetAll(ctx context.Context) ([]*game_config.ActionTypeDefinition, error) {
	defs, err := game_config.ActionTypeDefinitions(
		qm.Where("is_active = ?", true),
		qm.OrderBy("action_type ASC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询所有动作类型定义失败: %w", err)
	}

	return defs, nil
}
