package service

import (
	"context"
	"database/sql"
	"fmt"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// SkillLevelConfigService 技能等级配置服务
type SkillLevelConfigService struct {
	repo      interfaces.SkillLevelConfigRepository
	skillRepo interfaces.SkillRepository
	db        *sql.DB
}

// NewSkillLevelConfigService 创建技能等级配置服务
func NewSkillLevelConfigService(db *sql.DB) *SkillLevelConfigService {
	return &SkillLevelConfigService{
		repo:      impl.NewSkillLevelConfigRepository(db),
		skillRepo: impl.NewSkillRepository(db),
		db:        db,
	}
}

// GetSkillLevelConfigs 获取技能等级配置列表
func (s *SkillLevelConfigService) GetSkillLevelConfigs(ctx context.Context, params interfaces.SkillLevelConfigQueryParams) ([]*game_config.SkillLevelConfig, int64, error) {
	return s.repo.List(ctx, params)
}

// GetSkillLevelConfigsBySkillID 根据技能ID获取所有等级配置
func (s *SkillLevelConfigService) GetSkillLevelConfigsBySkillID(ctx context.Context, skillID string) ([]*game_config.SkillLevelConfig, error) {
	// 验证技能是否存在
	_, err := s.skillRepo.GetByID(ctx, skillID)
	if err != nil {
		return nil, err
	}

	return s.repo.ListBySkillID(ctx, skillID)
}

// GetSkillLevelConfigByID 根据ID获取配置
func (s *SkillLevelConfigService) GetSkillLevelConfigByID(ctx context.Context, id string) (*game_config.SkillLevelConfig, error) {
	return s.repo.GetByID(ctx, id)
}

// CreateSkillLevelConfig 创建技能等级配置
func (s *SkillLevelConfigService) CreateSkillLevelConfig(ctx context.Context, config *game_config.SkillLevelConfig) error {
	// 验证技能是否存在
	_, err := s.skillRepo.GetByID(ctx, config.SkillID)
	if err != nil {
		return fmt.Errorf("技能不存在: %w", err)
	}

	// 检查同一技能的同一等级是否已存在
	existing, _ := s.repo.GetBySkillIDAndLevel(ctx, config.SkillID, config.LevelNumber)
	if existing != nil {
		return fmt.Errorf("技能等级配置已存在: skillID=%s, level=%d", config.SkillID, config.LevelNumber)
	}

	return s.repo.Create(ctx, config)
}

// UpdateSkillLevelConfig 更新技能等级配置
func (s *SkillLevelConfigService) UpdateSkillLevelConfig(ctx context.Context, id string, updates map[string]interface{}) error {
	config, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 更新字段（这里简化处理，实际应用中需要更细致的字段映射）
	// 注意：skill_id 和 level_number 通常不应该被更新

	if damageMultiplier, ok := updates["damage_multiplier"].(string); ok {
		if damageMultiplier != "" {
			if err := config.DamageMultiplier.UnmarshalText([]byte(damageMultiplier)); err != nil {
				return fmt.Errorf("damage_multiplier 格式错误: %w", err)
			}
		}
	}

	if healingMultiplier, ok := updates["healing_multiplier"].(string); ok {
		if healingMultiplier != "" {
			if err := config.HealingMultiplier.UnmarshalText([]byte(healingMultiplier)); err != nil {
				return fmt.Errorf("healing_multiplier 格式错误: %w", err)
			}
		}
	}

	if durationModifier, ok := updates["duration_modifier"].(int); ok {
		config.DurationModifier.SetValid(durationModifier)
	}

	if rangeModifier, ok := updates["range_modifier"].(int); ok {
		config.RangeModifier.SetValid(rangeModifier)
	}

	if cooldownModifier, ok := updates["cooldown_modifier"].(int); ok {
		config.CooldownModifier.SetValid(cooldownModifier)
	}

	if manaCostModifier, ok := updates["mana_cost_modifier"].(int); ok {
		config.ManaCostModifier.SetValid(manaCostModifier)
	}

	if upgradeCostXP, ok := updates["upgrade_cost_xp"].(int); ok {
		config.UpgradeCostXP.SetValid(upgradeCostXP)
	}

	if upgradeCostGold, ok := updates["upgrade_cost_gold"].(int); ok {
		config.UpgradeCostGold.SetValid(upgradeCostGold)
	}

	return s.repo.Update(ctx, config)
}

// DeleteSkillLevelConfig 删除技能等级配置
func (s *SkillLevelConfigService) DeleteSkillLevelConfig(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// BatchSetSkillLevelConfigs 批量设置技能等级配置（先删后建）
func (s *SkillLevelConfigService) BatchSetSkillLevelConfigs(ctx context.Context, skillID string, configs []*game_config.SkillLevelConfig) error {
	// 验证技能是否存在
	_, err := s.skillRepo.GetByID(ctx, skillID)
	if err != nil {
		return fmt.Errorf("技能不存在: %w", err)
	}

	// 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开启事务失败: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// 删除该技能所有现有等级配置
	if err := s.repo.DeleteBySkillID(ctx, skillID); err != nil {
		return err
	}

	// 批量创建新配置
	for _, config := range configs {
		config.SkillID = skillID
		if err := s.repo.Create(ctx, config); err != nil {
			return err
		}
	}

	return nil
}
