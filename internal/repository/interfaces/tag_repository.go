package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// TagQueryParams 标签查询参数
type TagQueryParams struct {
	Category *string // tag_type_enum: class/item/skill/monster
	IsActive *bool
	Limit    int
	Offset   int
}

// TagRepository 标签仓储接口
type TagRepository interface {
	// GetByID 根据ID获取标签
	GetByID(ctx context.Context, tagID string) (*game_config.Tag, error)

	// GetByCode 根据代码获取标签
	GetByCode(ctx context.Context, code string) (*game_config.Tag, error)

	// List 获取标签列表
	List(ctx context.Context, params TagQueryParams) ([]*game_config.Tag, int64, error)

	// Create 创建标签
	Create(ctx context.Context, tag *game_config.Tag) error

	// Update 更新标签
	Update(ctx context.Context, tag *game_config.Tag) error

	// Delete 软删除标签
	Delete(ctx context.Context, tagID string) error

	// Exists 检查代码是否存在
	Exists(ctx context.Context, code string) (bool, error)
}
