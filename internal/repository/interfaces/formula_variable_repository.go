package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// FormulaVariableQueryParams 公式变量查询参数
type FormulaVariableQueryParams struct {
	VariableType *string // 变量类型
	Scope        *string // 作用域
	IsActive     *bool   // 是否启用
	Limit        int     // 每页数量
	Offset       int     // 偏移量
}

// FormulaVariableRepository 公式变量仓储接口
type FormulaVariableRepository interface {
	// GetByID 根据ID获取公式变量
	GetByID(ctx context.Context, id string) (*game_config.FormulaVariable, error)

	// GetByCode 根据代码获取公式变量
	GetByCode(ctx context.Context, code string) (*game_config.FormulaVariable, error)

	// List 获取公式变量列表
	List(ctx context.Context, params FormulaVariableQueryParams) ([]*game_config.FormulaVariable, int64, error)

	// GetAll 获取所有启用的公式变量
	GetAll(ctx context.Context) ([]*game_config.FormulaVariable, error)
}
