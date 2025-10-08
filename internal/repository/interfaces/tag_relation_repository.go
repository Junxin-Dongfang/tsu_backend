package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// TagRelationQueryParams 标签关联查询参数
type TagRelationQueryParams struct {
	TagID      *string // 标签ID
	EntityType *string // 实体类型
	EntityID   *string // 实体ID
	Limit      int     // 每页数量
	Offset     int     // 偏移量
}

// TagRelationRepository 标签关联仓储接口
type TagRelationRepository interface {
	// GetByID 根据ID获取标签关联
	GetByID(ctx context.Context, relationID string) (*game_config.TagsRelation, error)

	// List 获取标签关联列表
	List(ctx context.Context, params TagRelationQueryParams) ([]*game_config.TagsRelation, int64, error)

	// GetEntityTags 获取实体的所有标签
	GetEntityTags(ctx context.Context, entityType string, entityID string) ([]*game_config.Tag, error)

	// GetTagEntities 获取使用某个标签的所有实体
	GetTagEntities(ctx context.Context, tagID string) ([]*game_config.TagsRelation, error)

	// Create 创建标签关联
	Create(ctx context.Context, relation *game_config.TagsRelation) error

	// Delete 软删除标签关联
	Delete(ctx context.Context, relationID string) error

	// DeleteByTagAndEntity 删除指定标签和实体的关联
	DeleteByTagAndEntity(ctx context.Context, tagID string, entityType string, entityID string) error

	// Exists 检查标签和实体的关联是否存在
	Exists(ctx context.Context, tagID string, entityType string, entityID string) (bool, error)

	// BatchCreate 批量创建标签关联
	BatchCreate(ctx context.Context, relations []*game_config.TagsRelation) error

	// DeleteByEntity 删除实体的所有标签关联
	DeleteByEntity(ctx context.Context, entityType string, entityID string) error
}
