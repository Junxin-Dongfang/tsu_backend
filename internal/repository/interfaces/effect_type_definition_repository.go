package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// EffectTypeDefinitionQueryParams 元效果类型定义查询参数
type EffectTypeDefinitionQueryParams struct {
	IsActive *bool // 是否启用
	Limit    int   // 每页数量
	Offset   int   // 偏移量
}

// EffectTypeDefinitionRepository 元效果类型定义仓储接口
type EffectTypeDefinitionRepository interface {
	// GetByID 根据ID获取元效果类型定义
	GetByID(ctx context.Context, id string) (*game_config.EffectTypeDefinition, error)

	// GetByCode 根据代码获取元效果类型定义
	GetByCode(ctx context.Context, code string) (*game_config.EffectTypeDefinition, error)

	// List 获取元效果类型定义列表
	List(ctx context.Context, params EffectTypeDefinitionQueryParams) ([]*game_config.EffectTypeDefinition, int64, error)

	// GetAll 获取所有启用的元效果类型定义（用于下拉选择）
	GetAll(ctx context.Context) ([]*game_config.EffectTypeDefinition, error)
}
