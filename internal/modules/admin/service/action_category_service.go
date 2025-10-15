package service

import (
	"tsu-self/internal/pkg/xerrors"
	"context"
	"database/sql"
	"fmt"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// ActionCategoryService 动作类别服务
type ActionCategoryService struct {
	repo interfaces.ActionCategoryRepository
}

// NewActionCategoryService 创建动作类别服务
func NewActionCategoryService(db *sql.DB) *ActionCategoryService {
	return &ActionCategoryService{
		repo: impl.NewActionCategoryRepository(db),
	}
}

// GetActionCategories 获取动作类别列表
func (s *ActionCategoryService) GetActionCategories(ctx context.Context, params interfaces.ActionCategoryQueryParams) ([]*game_config.ActionCategory, int64, error) {
	return s.repo.List(ctx, params)
}

// GetActionCategoryByID 根据ID获取动作类别
func (s *ActionCategoryService) GetActionCategoryByID(ctx context.Context, categoryID string) (*game_config.ActionCategory, error) {
	return s.repo.GetByID(ctx, categoryID)
}

// CreateActionCategory 创建动作类别
func (s *ActionCategoryService) CreateActionCategory(ctx context.Context, category *game_config.ActionCategory) error {
	// 业务验证：检查类别代码是否已存在
	exists, err := s.repo.Exists(ctx, category.CategoryCode)
	if err != nil {
		return err
	}
	if exists {
		return xerrors.New(xerrors.CodeDuplicateResource, fmt.Sprintf("动作类别代码已存在: %s", category.CategoryCode))
	}

	return s.repo.Create(ctx, category)
}

// UpdateActionCategory 更新动作类别信息
func (s *ActionCategoryService) UpdateActionCategory(ctx context.Context, categoryID string, updates map[string]interface{}) error {
	// 获取动作类别
	category, err := s.repo.GetByID(ctx, categoryID)
	if err != nil {
		return err
	}

	// 更新字段
	if categoryCode, ok := updates["category_code"].(string); ok && categoryCode != "" {
		// 检查类别代码是否已被使用
		existing, err := s.repo.GetByCode(ctx, categoryCode)
		if err == nil && existing.ID != categoryID {
			return xerrors.New(xerrors.CodeDuplicateResource, fmt.Sprintf("动作类别代码已被使用: %s", categoryCode))
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

	if isActive, ok := updates["is_active"].(bool); ok {
		category.IsActive.SetValid(isActive)
	}

	// 保存更新
	return s.repo.Update(ctx, category)
}

// DeleteActionCategory 删除动作类别
func (s *ActionCategoryService) DeleteActionCategory(ctx context.Context, categoryID string) error {
	return s.repo.Delete(ctx, categoryID)
}
