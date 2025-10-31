package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// MetadataDictionaryRepository 元数据字典仓储接口
type MetadataDictionaryRepository interface {
	// GetByID 根据ID获取字典项
	GetByID(ctx context.Context, id string) (*game_config.MetadataDictionary, error)

	// GetByCodeAndCategory 根据代码和分类获取字典项
	GetByCodeAndCategory(ctx context.Context, code, category string) (*game_config.MetadataDictionary, error)

	// GetByCategory 根据分类获取所有字典项
	GetByCategory(ctx context.Context, category string) ([]*game_config.MetadataDictionary, error)

	// GetActionAttributes 获取所有动作属性字典项
	GetActionAttributes(ctx context.Context) ([]*game_config.MetadataDictionary, error)

	// GetFormulaVariables 获取所有公式变量字典项
	GetFormulaVariables(ctx context.Context) ([]*game_config.MetadataDictionary, error)
}
