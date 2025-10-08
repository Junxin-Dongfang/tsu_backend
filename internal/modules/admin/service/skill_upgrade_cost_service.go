package service

import (
	"context"
	"database/sql"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// SkillUpgradeCostService 技能升级消耗服务
type SkillUpgradeCostService struct {
	repo interfaces.SkillUpgradeCostRepository
}

// NewSkillUpgradeCostService 创建技能升级消耗服务
func NewSkillUpgradeCostService(db *sql.DB) *SkillUpgradeCostService {
	return &SkillUpgradeCostService{
		repo: impl.NewSkillUpgradeCostRepository(db),
	}
}

// CreateSkillUpgradeCost 创建升级消耗配置
func (s *SkillUpgradeCostService) CreateSkillUpgradeCost(ctx context.Context, cost *game_config.SkillUpgradeCost) error {
	return s.repo.Create(ctx, cost)
}

// GetSkillUpgradeCostByID 根据ID获取
func (s *SkillUpgradeCostService) GetSkillUpgradeCostByID(ctx context.Context, id string) (*game_config.SkillUpgradeCost, error) {
	return s.repo.GetByID(ctx, id)
}

// GetSkillUpgradeCostByLevel 根据等级获取
func (s *SkillUpgradeCostService) GetSkillUpgradeCostByLevel(ctx context.Context, level int) (*game_config.SkillUpgradeCost, error) {
	return s.repo.GetByLevel(ctx, level)
}

// GetAllSkillUpgradeCosts 获取所有升级消耗配置
func (s *SkillUpgradeCostService) GetAllSkillUpgradeCosts(ctx context.Context) ([]*game_config.SkillUpgradeCost, error) {
	return s.repo.List(ctx)
}

// UpdateSkillUpgradeCost 更新升级消耗配置
func (s *SkillUpgradeCostService) UpdateSkillUpgradeCost(ctx context.Context, id string, updates map[string]interface{}) error {
	return s.repo.Update(ctx, id, updates)
}

// DeleteSkillUpgradeCost 删除升级消耗配置
func (s *SkillUpgradeCostService) DeleteSkillUpgradeCost(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
