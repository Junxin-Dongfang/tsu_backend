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

// SkillCategoryService 技能类别服务
type SkillCategoryService struct {
	repo interfaces.SkillCategoryRepository
}

// NewSkillCategoryService 创建技能类别服务
func NewSkillCategoryService(db *sql.DB) *SkillCategoryService {
	return &SkillCategoryService{
		repo: impl.NewSkillCategoryRepository(db),
	}
}

// GetSkillCategories 获取技能类别列表
func (s *SkillCategoryService) GetSkillCategories(ctx context.Context, params interfaces.SkillCategoryQueryParams) ([]*game_config.SkillCategory, int64, error) {
	return s.repo.List(ctx, params)
}

// GetSkillCategoryByID 根据ID获取技能类别
func (s *SkillCategoryService) GetSkillCategoryByID(ctx context.Context, categoryID string) (*game_config.SkillCategory, error) {
	return s.repo.GetByID(ctx, categoryID)
}

// CreateSkillCategory 创建技能类别
func (s *SkillCategoryService) CreateSkillCategory(ctx context.Context, category *game_config.SkillCategory) error {
	// 业务验证：检查类别代码是否已存在
	exists, err := s.repo.Exists(ctx, category.CategoryCode)
	if err != nil {
		return err
	}
	if exists {
		return xerrors.New(xerrors.CodeDuplicateResource, fmt.Sprintf("技能类别代码已存在: %s", category.CategoryCode))
	}

	return s.repo.Create(ctx, category)
}

// UpdateSkillCategory 更新技能类别信息
func (s *SkillCategoryService) UpdateSkillCategory(ctx context.Context, categoryID string, updates map[string]interface{}) error {
	// 获取技能类别
	category, err := s.repo.GetByID(ctx, categoryID)
	if err != nil {
		return err
	}

	// 更新字段
	if categoryCode, ok := updates["category_code"].(string); ok && categoryCode != "" {
		// 检查类别代码是否已被使用
		existing, err := s.repo.GetByCode(ctx, categoryCode)
		if err == nil && existing.ID != categoryID {
			return xerrors.New(xerrors.CodeDuplicateResource, fmt.Sprintf("技能类别代码已被使用: %s", categoryCode))
		}
		category.CategoryCode = categoryCode
	}

	if categoryName, ok := updates["category_name"].(string); ok && categoryName != "" {
		category.CategoryName = categoryName
	}

	if description, ok := updates["description"].(string); ok {
		if description != "" {
			category.Description.SetValid(description)
		} else {
			category.Description.Valid = false
		}
	}

	if icon, ok := updates["icon"].(string); ok {
		if icon != "" {
			category.Icon.SetValid(icon)
		} else {
			category.Icon.Valid = false
		}
	}

	if color, ok := updates["color"].(string); ok {
		if color != "" {
			category.Color.SetValid(color)
		} else {
			category.Color.Valid = false
		}
	}

	if displayOrder, ok := updates["display_order"].(int); ok {
		category.DisplayOrder.SetValid(displayOrder)
	}

	if isActive, ok := updates["is_active"].(bool); ok {
		category.IsActive.SetValid(isActive)
	}

	// 保存更新
	return s.repo.Update(ctx, category)
}

// DeleteSkillCategory 删除技能类别
func (s *SkillCategoryService) DeleteSkillCategory(ctx context.Context, categoryID string) error {
	return s.repo.Delete(ctx, categoryID)
}
