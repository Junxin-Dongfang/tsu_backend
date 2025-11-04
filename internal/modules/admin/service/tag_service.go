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

// TagService 标签服务
type TagService struct {
	repo interfaces.TagRepository
}

// NewTagService 创建标签服务
func NewTagService(db *sql.DB) *TagService {
	return &TagService{
		repo: impl.NewTagRepository(db),
	}
}

// GetTags 获取标签列表
func (s *TagService) GetTags(ctx context.Context, params interfaces.TagQueryParams) ([]*game_config.Tag, int64, error) {
	return s.repo.List(ctx, params)
}

// GetTagByID 根据ID获取标签
func (s *TagService) GetTagByID(ctx context.Context, tagID string) (*game_config.Tag, error) {
	return s.repo.GetByID(ctx, tagID)
}

// CreateTag 创建标签
func (s *TagService) CreateTag(ctx context.Context, tag *game_config.Tag) error {
	// 业务验证：检查标签代码是否已存在
	exists, err := s.repo.Exists(ctx, tag.TagCode)
	if err != nil {
		return err
	}
	if exists {
		return xerrors.New(xerrors.CodeDuplicateResource, fmt.Sprintf("标签代码已存在: %s", tag.TagCode))
	}

	return s.repo.Create(ctx, tag)
}

// UpdateTag 更新标签信息
func (s *TagService) UpdateTag(ctx context.Context, tagID string, updates map[string]interface{}) error {
	// 获取标签
	tag, err := s.repo.GetByID(ctx, tagID)
	if err != nil {
		return err
	}

	// 更新字段
	if tagCode, ok := updates["tag_code"].(string); ok && tagCode != "" {
		// 检查标签代码是否已被使用
		existing, err := s.repo.GetByCode(ctx, tagCode)
		if err == nil && existing.ID != tagID {
			return xerrors.New(xerrors.CodeDuplicateResource, fmt.Sprintf("标签代码已被使用: %s", tagCode))
		}
		tag.TagCode = tagCode
	}

	if tagName, ok := updates["tag_name"].(string); ok && tagName != "" {
		tag.TagName = tagName
	}

	if category, ok := updates["category"].(string); ok && category != "" {
		tag.Category = category
	}

	if description, ok := updates["description"].(string); ok {
		if description != "" {
			tag.Description.SetValid(description)
		} else {
			tag.Description.Valid = false
		}
	}

	if icon, ok := updates["icon"].(string); ok {
		if icon != "" {
			tag.Icon.SetValid(icon)
		} else {
			tag.Icon.Valid = false
		}
	}

	if color, ok := updates["color"].(string); ok {
		if color != "" {
			tag.Color.SetValid(color)
		} else {
			tag.Color.Valid = false
		}
	}

	if displayOrder, ok := updates["display_order"].(int); ok {
		tag.DisplayOrder = displayOrder
	}

	if isActive, ok := updates["is_active"].(bool); ok {
		tag.IsActive = isActive
	}

	// 保存更新
	return s.repo.Update(ctx, tag)
}

// DeleteTag 删除标签
func (s *TagService) DeleteTag(ctx context.Context, tagID string) error {
	return s.repo.Delete(ctx, tagID)
}
