package service

import (
	"context"
	"database/sql"
	"fmt"
	"tsu-self/internal/pkg/xerrors"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// TagRelationService 标签关联服务
type TagRelationService struct {
	relationRepo interfaces.TagRelationRepository
	tagRepo      interfaces.TagRepository
}

// NewTagRelationService 创建标签关联服务
func NewTagRelationService(db *sql.DB) *TagRelationService {
	return &TagRelationService{
		relationRepo: impl.NewTagRelationRepository(db),
		tagRepo:      impl.NewTagRepository(db),
	}
}

// GetEntityTags 获取实体的所有标签
func (s *TagRelationService) GetEntityTags(ctx context.Context, entityType string, entityID string) ([]*game_config.Tag, error) {
	return s.relationRepo.GetEntityTags(ctx, entityType, entityID)
}

// GetTagEntities 获取使用某个标签的所有实体
func (s *TagRelationService) GetTagEntities(ctx context.Context, tagID string) ([]*game_config.TagsRelation, error) {
	// 验证标签是否存在
	_, err := s.tagRepo.GetByID(ctx, tagID)
	if err != nil {
		return nil, err
	}

	return s.relationRepo.GetTagEntities(ctx, tagID)
}

// AddTagToEntity 为实体添加标签
func (s *TagRelationService) AddTagToEntity(ctx context.Context, tagID string, entityType string, entityID string) error {
	// 验证标签是否存在
	_, err := s.tagRepo.GetByID(ctx, tagID)
	if err != nil {
		return err
	}

	// 检查关联是否已存在
	exists, err := s.relationRepo.Exists(ctx, tagID, entityType, entityID)
	if err != nil {
		return err
	}
	if exists {
		return xerrors.New(xerrors.CodeDuplicateResource, "标签关联已存在")
	}

	// 创建关联
	relation := &game_config.TagsRelation{
		TagID:      tagID,
		EntityType: entityType,
		EntityID:   entityID,
	}

	return s.relationRepo.Create(ctx, relation)
}

// RemoveTagFromEntity 从实体移除标签
func (s *TagRelationService) RemoveTagFromEntity(ctx context.Context, tagID string, entityType string, entityID string) error {
	return s.relationRepo.DeleteByTagAndEntity(ctx, tagID, entityType, entityID)
}

// BatchSetEntityTags 批量设置实体的标签
func (s *TagRelationService) BatchSetEntityTags(ctx context.Context, entityType string, entityID string, tagIDs []string) error {
	// 验证所有标签是否存在
	for _, tagID := range tagIDs {
		_, err := s.tagRepo.GetByID(ctx, tagID)
		if err != nil {
			return fmt.Errorf("标签不存在: %s", tagID)
		}
	}

	// 删除实体的所有旧标签关联
	if err := s.relationRepo.DeleteByEntity(ctx, entityType, entityID); err != nil {
		return err
	}

	// 如果没有新标签，直接返回
	if len(tagIDs) == 0 {
		return nil
	}

	// 创建新的标签关联
	var relations []*game_config.TagsRelation
	for _, tagID := range tagIDs {
		relations = append(relations, &game_config.TagsRelation{
			TagID:      tagID,
			EntityType: entityType,
			EntityID:   entityID,
		})
	}

	return s.relationRepo.BatchCreate(ctx, relations)
}
