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

// EffectService 效果服务
type EffectService struct {
	repo interfaces.EffectRepository
}

// NewEffectService 创建效果服务
func NewEffectService(db *sql.DB) *EffectService {
	return &EffectService{
		repo: impl.NewEffectRepository(db),
	}
}

// GetEffects 获取效果列表
func (s *EffectService) GetEffects(ctx context.Context, params interfaces.EffectQueryParams) ([]*game_config.Effect, int64, error) {
	return s.repo.List(ctx, params)
}

// GetEffectByID 根据ID获取效果
func (s *EffectService) GetEffectByID(ctx context.Context, effectID string) (*game_config.Effect, error) {
	return s.repo.GetByID(ctx, effectID)
}

// CreateEffect 创建效果
func (s *EffectService) CreateEffect(ctx context.Context, effect *game_config.Effect) error {
	// 业务验证：检查效果代码是否已存在
	exists, err := s.repo.Exists(ctx, effect.EffectCode)
	if err != nil {
		return err
	}
	if exists {
		return xerrors.New(xerrors.CodeDuplicateResource, fmt.Sprintf("效果代码已存在: %s", effect.EffectCode))
	}

	return s.repo.Create(ctx, effect)
}

// UpdateEffect 更新效果信息
func (s *EffectService) UpdateEffect(ctx context.Context, effectID string, updates map[string]interface{}) error {
	effect, err := s.repo.GetByID(ctx, effectID)
	if err != nil {
		return err
	}

	// 更新字段
	if effectCode, ok := updates["effect_code"].(string); ok && effectCode != "" {
		existing, err := s.repo.GetByCode(ctx, effectCode)
		if err == nil && existing.ID != effectID {
			return xerrors.New(xerrors.CodeDuplicateResource, fmt.Sprintf("效果代码已被使用: %s", effectCode))
		}
		effect.EffectCode = effectCode
	}

	if effectName, ok := updates["effect_name"].(string); ok && effectName != "" {
		effect.EffectName = effectName
	}

	if effectType, ok := updates["effect_type"].(string); ok && effectType != "" {
		effect.EffectType = effectType
	}

	if description, ok := updates["description"].(string); ok {
		if description != "" {
			effect.Description.SetValid(description)
		} else {
			effect.Description.Valid = false
		}
	}

	if isActive, ok := updates["is_active"].(bool); ok {
		effect.IsActive.SetValid(isActive)
	}

	return s.repo.Update(ctx, effect)
}

// DeleteEffect 删除效果
func (s *EffectService) DeleteEffect(ctx context.Context, effectID string) error {
	return s.repo.Delete(ctx, effectID)
}
