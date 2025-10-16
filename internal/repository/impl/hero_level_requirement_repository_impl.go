package impl

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/queries/qm"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/interfaces"
)

type heroLevelRequirementRepositoryImpl struct {
	db *sql.DB
}

// NewHeroLevelRequirementRepository 创建英雄等级需求配置仓储实例
func NewHeroLevelRequirementRepository(db *sql.DB) interfaces.HeroLevelRequirementRepository {
	return &heroLevelRequirementRepositoryImpl{db: db}
}

// GetByLevel 根据等级获取需求配置
func (r *heroLevelRequirementRepositoryImpl) GetByLevel(ctx context.Context, level int) (*game_config.HeroLevelRequirement, error) {
	req, err := game_config.HeroLevelRequirements(
		qm.Where("level = ?", level),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("等级需求配置不存在: level=%d", level)
	}
	if err != nil {
		return nil, fmt.Errorf("查询等级需求配置失败: %w", err)
	}

	return req, nil
}

// GetNextLevelRequirement 获取下一级需求
func (r *heroLevelRequirementRepositoryImpl) GetNextLevelRequirement(ctx context.Context, currentLevel int) (*game_config.HeroLevelRequirement, error) {
	nextLevel := currentLevel + 1
	if nextLevel > 40 {
		return nil, fmt.Errorf("已达到最高等级")
	}

	return r.GetByLevel(ctx, nextLevel)
}

// GetAll 获取所有等级需求
func (r *heroLevelRequirementRepositoryImpl) GetAll(ctx context.Context) ([]*game_config.HeroLevelRequirement, error) {
	reqs, err := game_config.HeroLevelRequirements(
		qm.OrderBy("level ASC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询所有等级需求配置失败: %w", err)
	}

	return reqs, nil
}

// CheckCanLevelUp 检查是否可以升级（返回可升到的最高等级）
func (r *heroLevelRequirementRepositoryImpl) CheckCanLevelUp(ctx context.Context, experienceTotal int, currentLevel int) (canLevelUp bool, targetLevel int, error error) {
	if currentLevel >= 40 {
		return false, currentLevel, nil
	}

	// 获取所有等级需求配置
	reqs, err := game_config.HeroLevelRequirements(
		qm.Where("level > ? AND cumulative_xp <= ?", currentLevel, experienceTotal),
		qm.OrderBy("level DESC"),
		qm.Limit(1),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		// 经验不足以升级
		return false, currentLevel, nil
	}
	if err != nil {
		return false, currentLevel, fmt.Errorf("检查升级条件失败: %w", err)
	}

	return true, reqs.Level, nil
}

