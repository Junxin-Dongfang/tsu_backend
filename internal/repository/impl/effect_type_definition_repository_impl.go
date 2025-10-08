package impl

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/queries/qm"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/interfaces"
)

type effectTypeDefinitionRepositoryImpl struct {
	db *sql.DB
}

// NewEffectTypeDefinitionRepository 创建元效果类型定义仓储实例
func NewEffectTypeDefinitionRepository(db *sql.DB) interfaces.EffectTypeDefinitionRepository {
	return &effectTypeDefinitionRepositoryImpl{db: db}
}

// GetByID 根据ID获取元效果类型定义
func (r *effectTypeDefinitionRepositoryImpl) GetByID(ctx context.Context, id string) (*game_config.EffectTypeDefinition, error) {
	def, err := game_config.EffectTypeDefinitions(
		qm.Where("id = ? AND deleted_at IS NULL", id),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("元效果类型定义不存在: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("查询元效果类型定义失败: %w", err)
	}

	return def, nil
}

// GetByCode 根据代码获取元效果类型定义
func (r *effectTypeDefinitionRepositoryImpl) GetByCode(ctx context.Context, code string) (*game_config.EffectTypeDefinition, error) {
	def, err := game_config.EffectTypeDefinitions(
		qm.Where("effect_type_code = ? AND deleted_at IS NULL", code),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("元效果类型定义不存在: %s", code)
	}
	if err != nil {
		return nil, fmt.Errorf("查询元效果类型定义失败: %w", err)
	}

	return def, nil
}

// List 获取元效果类型定义列表
func (r *effectTypeDefinitionRepositoryImpl) List(ctx context.Context, params interfaces.EffectTypeDefinitionQueryParams) ([]*game_config.EffectTypeDefinition, int64, error) {
	// 构建基础查询条件
	var baseQueryMods []qm.QueryMod
	baseQueryMods = append(baseQueryMods, qm.Where("deleted_at IS NULL"))

	// 筛选条件
	if params.IsActive != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("is_active = ?", *params.IsActive))
	}

	// 获取总数
	count, err := game_config.EffectTypeDefinitions(baseQueryMods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询元效果类型定义总数失败: %w", err)
	}

	// 构建完整查询条件
	var queryMods []qm.QueryMod
	queryMods = append(queryMods, baseQueryMods...)

	// 排序
	queryMods = append(queryMods, qm.OrderBy("effect_type_code ASC"))

	// 分页
	if params.Limit > 0 {
		queryMods = append(queryMods, qm.Limit(params.Limit))
	}
	if params.Offset > 0 {
		queryMods = append(queryMods, qm.Offset(params.Offset))
	}

	// 查询列表
	defs, err := game_config.EffectTypeDefinitions(queryMods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询元效果类型定义列表失败: %w", err)
	}

	return defs, count, nil
}

// GetAll 获取所有启用的元效果类型定义
func (r *effectTypeDefinitionRepositoryImpl) GetAll(ctx context.Context) ([]*game_config.EffectTypeDefinition, error) {
	defs, err := game_config.EffectTypeDefinitions(
		qm.Where("deleted_at IS NULL"),
		qm.Where("is_active = ?", true),
		qm.OrderBy("effect_type_code ASC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询所有元效果类型定义失败: %w", err)
	}

	return defs, nil
}
