package service

import (
	"context"
	"database/sql"
	"fmt"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// ActionFlagService 动作Flag服务
type ActionFlagService struct {
	repo interfaces.ActionFlagRepository
}

// NewActionFlagService 创建动作Flag服务
func NewActionFlagService(db *sql.DB) *ActionFlagService {
	return &ActionFlagService{
		repo: impl.NewActionFlagRepository(db),
	}
}

// GetActionFlags 获取动作Flag列表
func (s *ActionFlagService) GetActionFlags(ctx context.Context, params interfaces.ActionFlagQueryParams) ([]*game_config.ActionFlag, int64, error) {
	return s.repo.List(ctx, params)
}

// GetActionFlagByID 根据ID获取动作Flag
func (s *ActionFlagService) GetActionFlagByID(ctx context.Context, flagID string) (*game_config.ActionFlag, error) {
	return s.repo.GetByID(ctx, flagID)
}

// CreateActionFlag 创建动作Flag
func (s *ActionFlagService) CreateActionFlag(ctx context.Context, flag *game_config.ActionFlag) error {
	// 业务验证：检查Flag代码是否已存在
	exists, err := s.repo.Exists(ctx, flag.FlagCode)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("动作Flag代码已存在: %s", flag.FlagCode)
	}

	return s.repo.Create(ctx, flag)
}

// UpdateActionFlag 更新动作Flag信息
func (s *ActionFlagService) UpdateActionFlag(ctx context.Context, flagID string, updates map[string]interface{}) error {
	flag, err := s.repo.GetByID(ctx, flagID)
	if err != nil {
		return err
	}

	// 更新字段
	if flagCode, ok := updates["flag_code"].(string); ok && flagCode != "" {
		existing, err := s.repo.GetByCode(ctx, flagCode)
		if err == nil && existing.ID != flagID {
			return fmt.Errorf("动作Flag代码已被使用: %s", flagCode)
		}
		flag.FlagCode = flagCode
	}

	if flagName, ok := updates["flag_name"].(string); ok && flagName != "" {
		flag.FlagName = flagName
	}

	if category, ok := updates["category"].(string); ok {
		if category != "" {
			flag.Category.SetValid(category)
		} else {
			flag.Category.Valid = false
		}
	}

	if description, ok := updates["description"].(string); ok {
		if description != "" {
			flag.Description.SetValid(description)
		} else {
			flag.Description.Valid = false
		}
	}

	if isActive, ok := updates["is_active"].(bool); ok {
		flag.IsActive.SetValid(isActive)
	}

	return s.repo.Update(ctx, flag)
}

// DeleteActionFlag 删除动作Flag（软删除）
func (s *ActionFlagService) DeleteActionFlag(ctx context.Context, flagID string) error {
	return s.repo.Delete(ctx, flagID)
}
