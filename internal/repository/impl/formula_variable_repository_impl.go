package impl

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/queries/qm"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/interfaces"
)

type formulaVariableRepositoryImpl struct {
	db *sql.DB
}

// NewFormulaVariableRepository 创建公式变量仓储实例
func NewFormulaVariableRepository(db *sql.DB) interfaces.FormulaVariableRepository {
	return &formulaVariableRepositoryImpl{db: db}
}

func (r *formulaVariableRepositoryImpl) GetByID(ctx context.Context, id string) (*game_config.FormulaVariable, error) {
	v, err := game_config.FormulaVariables(
		qm.Where("id = ? AND deleted_at IS NULL", id),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("公式变量不存在: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("查询公式变量失败: %w", err)
	}

	return v, nil
}

func (r *formulaVariableRepositoryImpl) GetByCode(ctx context.Context, code string) (*game_config.FormulaVariable, error) {
	v, err := game_config.FormulaVariables(
		qm.Where("variable_code = ? AND deleted_at IS NULL", code),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("公式变量不存在: %s", code)
	}
	if err != nil {
		return nil, fmt.Errorf("查询公式变量失败: %w", err)
	}

	return v, nil
}

func (r *formulaVariableRepositoryImpl) List(ctx context.Context, params interfaces.FormulaVariableQueryParams) ([]*game_config.FormulaVariable, int64, error) {
	var baseQueryMods []qm.QueryMod
	baseQueryMods = append(baseQueryMods, qm.Where("deleted_at IS NULL"))

	if params.VariableType != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("variable_type = ?", *params.VariableType))
	}
	if params.Scope != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("scope = ?", *params.Scope))
	}
	if params.IsActive != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("is_active = ?", *params.IsActive))
	}

	count, err := game_config.FormulaVariables(baseQueryMods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询公式变量总数失败: %w", err)
	}

	var queryMods []qm.QueryMod
	queryMods = append(queryMods, baseQueryMods...)
	queryMods = append(queryMods, qm.OrderBy("variable_code ASC"))

	if params.Limit > 0 {
		queryMods = append(queryMods, qm.Limit(params.Limit))
	}
	if params.Offset > 0 {
		queryMods = append(queryMods, qm.Offset(params.Offset))
	}

	vars, err := game_config.FormulaVariables(queryMods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询公式变量列表失败: %w", err)
	}

	return vars, count, nil
}

func (r *formulaVariableRepositoryImpl) GetAll(ctx context.Context) ([]*game_config.FormulaVariable, error) {
	vars, err := game_config.FormulaVariables(
		qm.Where("deleted_at IS NULL"),
		qm.Where("is_active = ?", true),
		qm.OrderBy("variable_code ASC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询所有公式变量失败: %w", err)
	}

	return vars, nil
}
