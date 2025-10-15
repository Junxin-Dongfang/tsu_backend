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

// BuffService Buff服务
type BuffService struct {
	repo interfaces.BuffRepository
}

// NewBuffService 创建Buff服务
func NewBuffService(db *sql.DB) *BuffService {
	return &BuffService{
		repo: impl.NewBuffRepository(db),
	}
}

// GetBuffs 获取Buff列表
func (s *BuffService) GetBuffs(ctx context.Context, params interfaces.BuffQueryParams) ([]*game_config.Buff, int64, error) {
	return s.repo.List(ctx, params)
}

// GetBuffByID 根据ID获取Buff
func (s *BuffService) GetBuffByID(ctx context.Context, buffID string) (*game_config.Buff, error) {
	return s.repo.GetByID(ctx, buffID)
}

// CreateBuff 创建Buff
func (s *BuffService) CreateBuff(ctx context.Context, buff *game_config.Buff) error {
	// 业务验证：检查Buff代码是否已存在
	exists, err := s.repo.Exists(ctx, buff.BuffCode)
	if err != nil {
		return err
	}
	if exists {
		return xerrors.New(xerrors.CodeDuplicateResource, fmt.Sprintf("Buff代码已存在: %s", buff.BuffCode))
	}

	return s.repo.Create(ctx, buff)
}

// UpdateBuff 更新Buff信息
func (s *BuffService) UpdateBuff(ctx context.Context, buffID string, updates map[string]interface{}) error {
	buff, err := s.repo.GetByID(ctx, buffID)
	if err != nil {
		return err
	}

	// 更新字段
	if buffCode, ok := updates["buff_code"].(string); ok && buffCode != "" {
		existing, err := s.repo.GetByCode(ctx, buffCode)
		if err == nil && existing.ID != buffID {
			return xerrors.New(xerrors.CodeDuplicateResource, fmt.Sprintf("Buff代码已被使用: %s", buffCode))
		}
		buff.BuffCode = buffCode
	}

	if buffName, ok := updates["buff_name"].(string); ok && buffName != "" {
		buff.BuffName = buffName
	}

	if buffType, ok := updates["buff_type"].(string); ok && buffType != "" {
		buff.BuffType = buffType
	}

	if category, ok := updates["category"].(string); ok {
		if category != "" {
			buff.Category.SetValid(category)
		} else {
			buff.Category.Valid = false
		}
	}

	if description, ok := updates["description"].(string); ok {
		if description != "" {
			buff.Description.SetValid(description)
		} else {
			buff.Description.Valid = false
		}
	}

	if isActive, ok := updates["is_active"].(bool); ok {
		buff.IsActive.SetValid(isActive)
	}

	return s.repo.Update(ctx, buff)
}

// DeleteBuff 删除Buff（软删除）
func (s *BuffService) DeleteBuff(ctx context.Context, buffID string) error {
	return s.repo.Delete(ctx, buffID)
}
