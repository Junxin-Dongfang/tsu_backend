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

// HeroAttributeTypeService 属性类型服务
type HeroAttributeTypeService struct {
	repo interfaces.HeroAttributeTypeRepository
}

// NewHeroAttributeTypeService 创建属性类型服务
func NewHeroAttributeTypeService(db *sql.DB) *HeroAttributeTypeService {
	return &HeroAttributeTypeService{
		repo: impl.NewHeroAttributeTypeRepository(db),
	}
}

// GetHeroAttributeTypes 获取属性类型列表
func (s *HeroAttributeTypeService) GetHeroAttributeTypes(ctx context.Context, params interfaces.HeroAttributeTypeQueryParams) ([]*game_config.HeroAttributeType, int64, error) {
	return s.repo.List(ctx, params)
}

// GetHeroAttributeTypeByID 根据ID获取属性类型
func (s *HeroAttributeTypeService) GetHeroAttributeTypeByID(ctx context.Context, attributeTypeID string) (*game_config.HeroAttributeType, error) {
	return s.repo.GetByID(ctx, attributeTypeID)
}

// CreateHeroAttributeType 创建属性类型
func (s *HeroAttributeTypeService) CreateHeroAttributeType(ctx context.Context, attributeType *game_config.HeroAttributeType) error {
	// 业务验证：检查代码是否已存在
	exists, err := s.repo.Exists(ctx, attributeType.AttributeCode)
	if err != nil {
		return err
	}
	if exists {
		return xerrors.New(xerrors.CodeDuplicateResource, fmt.Sprintf("属性类型代码已存在: %s", attributeType.AttributeCode))
	}

	return s.repo.Create(ctx, attributeType)
}

// UpdateHeroAttributeType 更新属性类型信息
func (s *HeroAttributeTypeService) UpdateHeroAttributeType(ctx context.Context, attributeTypeID string, updates map[string]interface{}) error {
	// 获取属性类型
	attributeType, err := s.repo.GetByID(ctx, attributeTypeID)
	if err != nil {
		return err
	}

	// 更新字段
	if code, ok := updates["attribute_code"].(string); ok && code != "" {
		// 检查代码是否已被使用
		existing, err := s.repo.GetByCode(ctx, code)
		if err == nil && existing.ID != attributeTypeID {
			return xerrors.New(xerrors.CodeDuplicateResource, fmt.Sprintf("属性类型代码已被使用: %s", code))
		}
		attributeType.AttributeCode = code
	}

	if name, ok := updates["attribute_name"].(string); ok && name != "" {
		attributeType.AttributeName = name
	}

	if category, ok := updates["category"].(string); ok && category != "" {
		attributeType.Category = category
	}

	if description, ok := updates["description"].(string); ok {
		if description != "" {
			attributeType.Description.SetValid(description)
		} else {
			attributeType.Description.Valid = false
		}
	}

	if icon, ok := updates["icon"].(string); ok {
		if icon != "" {
			attributeType.Icon.SetValid(icon)
		} else {
			attributeType.Icon.Valid = false
		}
	}

	if color, ok := updates["color"].(string); ok {
		if color != "" {
			attributeType.Color.SetValid(color)
		} else {
			attributeType.Color.Valid = false
		}
	}

	if displayOrder, ok := updates["display_order"].(int); ok {
		attributeType.DisplayOrder = displayOrder
	}

	if isActive, ok := updates["is_active"].(bool); ok {
		attributeType.IsActive = isActive
	}

	if isVisible, ok := updates["is_visible"].(bool); ok {
		attributeType.IsVisible = isVisible
	}

	// 保存更新
	return s.repo.Update(ctx, attributeType)
}

// DeleteHeroAttributeType 删除属性类型
func (s *HeroAttributeTypeService) DeleteHeroAttributeType(ctx context.Context, attributeTypeID string) error {
	return s.repo.Delete(ctx, attributeTypeID)
}
