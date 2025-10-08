package service

import (
	"context"
	"database/sql"
	"fmt"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// ActionEffectService 动作效果关联服务
type ActionEffectService struct {
	actionEffectRepo interfaces.ActionEffectRepository
	actionRepo       interfaces.ActionRepository
	effectRepo       interfaces.EffectRepository
}

// NewActionEffectService 创建动作效果关联服务
func NewActionEffectService(db *sql.DB) *ActionEffectService {
	return &ActionEffectService{
		actionEffectRepo: impl.NewActionEffectRepository(db),
		actionRepo:       impl.NewActionRepository(db),
		effectRepo:       impl.NewEffectRepository(db),
	}
}

// GetActionEffects 获取动作的所有效果
func (s *ActionEffectService) GetActionEffects(ctx context.Context, actionID string) ([]*game_config.ActionEffect, error) {
	// 验证动作存在
	if _, err := s.actionRepo.GetByID(ctx, actionID); err != nil {
		return nil, err
	}

	return s.actionEffectRepo.GetByActionID(ctx, actionID)
}

// AddEffectToAction 为动作添加效果
func (s *ActionEffectService) AddEffectToAction(ctx context.Context, actionEffect *game_config.ActionEffect) error {
	// 验证动作存在
	if _, err := s.actionRepo.GetByID(ctx, actionEffect.ActionID); err != nil {
		return err
	}

	// 验证效果存在
	if _, err := s.effectRepo.GetByID(ctx, actionEffect.EffectID); err != nil {
		return err
	}

	return s.actionEffectRepo.Create(ctx, actionEffect)
}

// RemoveEffectFromAction 从动作移除效果
func (s *ActionEffectService) RemoveEffectFromAction(ctx context.Context, actionEffectID string) error {
	return s.actionEffectRepo.Delete(ctx, actionEffectID)
}

// BatchSetActionEffects 批量设置动作效果（先删后建）
func (s *ActionEffectService) BatchSetActionEffects(ctx context.Context, actionID string, actionEffects []*game_config.ActionEffect) error {
	// 验证动作存在
	if _, err := s.actionRepo.GetByID(ctx, actionID); err != nil {
		return err
	}

	// 验证所有效果存在
	for _, actionEffect := range actionEffects {
		if _, err := s.effectRepo.GetByID(ctx, actionEffect.EffectID); err != nil {
			return fmt.Errorf("效果不存在: %s", actionEffect.EffectID)
		}
	}

	// 先删除旧关联
	if err := s.actionEffectRepo.DeleteAllByActionID(ctx, actionID); err != nil {
		return err
	}

	// 批量创建新关联
	if len(actionEffects) > 0 {
		return s.actionEffectRepo.BatchCreate(ctx, actionEffects)
	}

	return nil
}
