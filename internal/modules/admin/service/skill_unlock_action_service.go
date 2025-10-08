package service

import (
	"context"
	"database/sql"
	"fmt"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// SkillUnlockActionService 技能解锁动作服务
type SkillUnlockActionService struct {
	unlockActionRepo interfaces.SkillUnlockActionRepository
	skillRepo        interfaces.SkillRepository
	actionRepo       interfaces.ActionRepository
}

// NewSkillUnlockActionService 创建技能解锁动作服务
func NewSkillUnlockActionService(db *sql.DB) *SkillUnlockActionService {
	return &SkillUnlockActionService{
		unlockActionRepo: impl.NewSkillUnlockActionRepository(db),
		skillRepo:        impl.NewSkillRepository(db),
		actionRepo:       impl.NewActionRepository(db),
	}
}

// GetSkillUnlockActions 获取技能的所有解锁动作
func (s *SkillUnlockActionService) GetSkillUnlockActions(ctx context.Context, skillID string) ([]*game_config.SkillUnlockAction, error) {
	// 验证技能存在
	if _, err := s.skillRepo.GetByID(ctx, skillID); err != nil {
		return nil, err
	}

	return s.unlockActionRepo.GetBySkillID(ctx, skillID)
}

// AddUnlockActionToSkill 为技能添加解锁动作
func (s *SkillUnlockActionService) AddUnlockActionToSkill(ctx context.Context, unlockAction *game_config.SkillUnlockAction) error {
	// 验证技能存在
	if _, err := s.skillRepo.GetByID(ctx, unlockAction.SkillID); err != nil {
		return err
	}

	// 验证动作存在
	if _, err := s.actionRepo.GetByID(ctx, unlockAction.ActionID); err != nil {
		return err
	}

	return s.unlockActionRepo.Create(ctx, unlockAction)
}

// RemoveUnlockActionFromSkill 从技能移除解锁动作
func (s *SkillUnlockActionService) RemoveUnlockActionFromSkill(ctx context.Context, unlockActionID string) error {
	return s.unlockActionRepo.Delete(ctx, unlockActionID)
}

// BatchSetSkillUnlockActions 批量设置技能解锁动作（先删后建）
func (s *SkillUnlockActionService) BatchSetSkillUnlockActions(ctx context.Context, skillID string, unlockActions []*game_config.SkillUnlockAction) error {
	// 验证技能存在
	if _, err := s.skillRepo.GetByID(ctx, skillID); err != nil {
		return err
	}

	// 验证所有动作存在
	for _, unlockAction := range unlockActions {
		if _, err := s.actionRepo.GetByID(ctx, unlockAction.ActionID); err != nil {
			return fmt.Errorf("动作不存在: %s", unlockAction.ActionID)
		}
	}

	// 先删除旧关联
	if err := s.unlockActionRepo.DeleteAllBySkillID(ctx, skillID); err != nil {
		return err
	}

	// 批量创建新关联
	if len(unlockActions) > 0 {
		return s.unlockActionRepo.BatchCreate(ctx, unlockActions)
	}

	return nil
}
