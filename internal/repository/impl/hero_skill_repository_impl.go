package impl

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/repository/interfaces"
)

type heroSkillRepositoryImpl struct {
	db *sql.DB
}

// NewHeroSkillRepository 创建英雄技能仓储实例
func NewHeroSkillRepository(db *sql.DB) interfaces.HeroSkillRepository {
	return &heroSkillRepositoryImpl{db: db}
}

// Create 创建英雄技能记录
func (r *heroSkillRepositoryImpl) Create(ctx context.Context, execer boil.ContextExecutor, heroSkill *game_runtime.HeroSkill) error {
	if err := heroSkill.Insert(ctx, execer, boil.Infer()); err != nil {
		return fmt.Errorf("创建英雄技能记录失败: %w", err)
	}
	return nil
}

// GetByID 根据ID获取技能
func (r *heroSkillRepositoryImpl) GetByID(ctx context.Context, heroSkillID string) (*game_runtime.HeroSkill, error) {
	heroSkill, err := game_runtime.HeroSkills(
		qm.Where("id = ?", heroSkillID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("英雄技能不存在: %s", heroSkillID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询英雄技能失败: %w", err)
	}

	return heroSkill, nil
}

// GetByIDForUpdate 根据ID获取技能（带行锁）
func (r *heroSkillRepositoryImpl) GetByIDForUpdate(ctx context.Context, execer boil.ContextExecutor, heroSkillID string) (*game_runtime.HeroSkill, error) {
	heroSkill, err := game_runtime.HeroSkills(
		qm.Where("id = ?", heroSkillID),
		qm.For("UPDATE"),
	).One(ctx, execer)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("英雄技能不存在: %s", heroSkillID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询英雄技能失败（带锁）: %w", err)
	}

	return heroSkill, nil
}

// GetByHeroID 获取英雄的所有技能
func (r *heroSkillRepositoryImpl) GetByHeroID(ctx context.Context, heroID string) ([]*game_runtime.HeroSkill, error) {
	heroSkills, err := game_runtime.HeroSkills(
		qm.Where("hero_id = ?", heroID),
		qm.OrderBy("created_at ASC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询英雄技能列表失败: %w", err)
	}

	return heroSkills, nil
}

// GetByHeroAndSkillID 获取英雄的特定技能
func (r *heroSkillRepositoryImpl) GetByHeroAndSkillID(ctx context.Context, heroID, skillID string) (*game_runtime.HeroSkill, error) {
	heroSkill, err := game_runtime.HeroSkills(
		qm.Where("hero_id = ? AND skill_id = ?", heroID, skillID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, nil // 未学习该技能
	}
	if err != nil {
		return nil, fmt.Errorf("查询英雄技能失败: %w", err)
	}

	return heroSkill, nil
}

// GetByHeroAndSkillIDForUpdate 获取英雄的特定技能（带行锁）
func (r *heroSkillRepositoryImpl) GetByHeroAndSkillIDForUpdate(ctx context.Context, execer boil.ContextExecutor, heroID, skillID string) (*game_runtime.HeroSkill, error) {
	heroSkill, err := game_runtime.HeroSkills(
		qm.Where("hero_id = ? AND skill_id = ?", heroID, skillID),
		qm.For("UPDATE"),
	).One(ctx, execer)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("英雄未学习该技能")
	}
	if err != nil {
		return nil, fmt.Errorf("查询英雄技能失败（带锁）: %w", err)
	}

	return heroSkill, nil
}

// Update 更新技能信息
func (r *heroSkillRepositoryImpl) Update(ctx context.Context, execer boil.ContextExecutor, heroSkill *game_runtime.HeroSkill) error {
	if _, err := heroSkill.Update(ctx, execer, boil.Infer()); err != nil {
		return fmt.Errorf("更新英雄技能失败: %w", err)
	}
	return nil
}

// Delete 删除技能记录
func (r *heroSkillRepositoryImpl) Delete(ctx context.Context, execer boil.ContextExecutor, heroSkillID string) error {
	_, err := game_runtime.HeroSkills(
		qm.Where("id = ?", heroSkillID),
	).DeleteAll(ctx, execer, false)

	if err != nil {
		return fmt.Errorf("删除英雄技能失败: %w", err)
	}

	return nil
}

// DeleteAllByHeroID 删除英雄的所有技能（重生时使用）
func (r *heroSkillRepositoryImpl) DeleteAllByHeroID(ctx context.Context, execer boil.ContextExecutor, heroID string) error {
	_, err := game_runtime.HeroSkills(
		qm.Where("hero_id = ?", heroID),
	).DeleteAll(ctx, execer, false)

	if err != nil {
		return fmt.Errorf("删除英雄所有技能失败: %w", err)
	}

	return nil
}

// nullTimeNow 辅助函数：返回当前时间指针
func nullTimeNow() *time.Time {
	now := time.Now()
	return &now
}

