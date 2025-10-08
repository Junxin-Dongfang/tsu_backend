package service

import (
	"context"
	"database/sql"
	"fmt"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// BuffEffectService Buff效果关联服务
type BuffEffectService struct {
	buffEffectRepo interfaces.BuffEffectRepository
	buffRepo       interfaces.BuffRepository
	effectRepo     interfaces.EffectRepository
}

// NewBuffEffectService 创建Buff效果关联服务
func NewBuffEffectService(db *sql.DB) *BuffEffectService {
	return &BuffEffectService{
		buffEffectRepo: impl.NewBuffEffectRepository(db),
		buffRepo:       impl.NewBuffRepository(db),
		effectRepo:     impl.NewEffectRepository(db),
	}
}

// GetBuffEffects 获取Buff的所有效果
func (s *BuffEffectService) GetBuffEffects(ctx context.Context, buffID string) ([]*game_config.BuffEffect, error) {
	// 验证Buff存在
	if _, err := s.buffRepo.GetByID(ctx, buffID); err != nil {
		return nil, err
	}

	return s.buffEffectRepo.GetByBuffID(ctx, buffID)
}

// AddEffectToBuff 为Buff添加效果
func (s *BuffEffectService) AddEffectToBuff(ctx context.Context, buffEffect *game_config.BuffEffect) error {
	// 验证Buff存在
	if _, err := s.buffRepo.GetByID(ctx, buffEffect.BuffID); err != nil {
		return err
	}

	// 验证Effect存在
	if _, err := s.effectRepo.GetByID(ctx, buffEffect.EffectID); err != nil {
		return err
	}

	return s.buffEffectRepo.Create(ctx, buffEffect)
}

// RemoveEffectFromBuff 从Buff移除效果
func (s *BuffEffectService) RemoveEffectFromBuff(ctx context.Context, buffEffectID string) error {
	return s.buffEffectRepo.Delete(ctx, buffEffectID)
}

// BatchSetBuffEffects 批量设置Buff效果（先删后建）
func (s *BuffEffectService) BatchSetBuffEffects(ctx context.Context, buffID string, buffEffects []*game_config.BuffEffect) error {
	// 验证Buff存在
	if _, err := s.buffRepo.GetByID(ctx, buffID); err != nil {
		return err
	}

	// 验证所有Effect存在
	for _, buffEffect := range buffEffects {
		if _, err := s.effectRepo.GetByID(ctx, buffEffect.EffectID); err != nil {
			return fmt.Errorf("效果不存在: %s", buffEffect.EffectID)
		}
	}

	// 先删除旧关联
	if err := s.buffEffectRepo.DeleteAllByBuffID(ctx, buffID); err != nil {
		return err
	}

	// 批量创建新关联
	if len(buffEffects) > 0 {
		return s.buffEffectRepo.BatchCreate(ctx, buffEffects)
	}

	return nil
}
