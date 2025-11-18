package impl

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/google/uuid"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/interfaces"
)

type monsterSkillRepositoryImpl struct {
	exec     boil.ContextExecutor
	beginner boil.ContextBeginner
}

// NewMonsterSkillRepository 创建怪物技能仓储实例
func NewMonsterSkillRepository(db *sql.DB) interfaces.MonsterSkillRepository {
	return &monsterSkillRepositoryImpl{
		exec:     db,
		beginner: db,
	}
}

// NewMonsterSkillRepositoryWithExecutor 使用自定义执行器创建仓储实例
func NewMonsterSkillRepositoryWithExecutor(exec boil.ContextExecutor) interfaces.MonsterSkillRepository {
	var beginner boil.ContextBeginner
	if b, ok := exec.(boil.ContextBeginner); ok {
		beginner = b
	}
	return &monsterSkillRepositoryImpl{
		exec:     exec,
		beginner: beginner,
	}
}

// Create 创建怪物技能关联
func (r *monsterSkillRepositoryImpl) Create(ctx context.Context, monsterSkill *game_config.MonsterSkill) error {
	// 生成UUID
	if monsterSkill.ID == "" {
		monsterSkill.ID = uuid.New().String()
	}

	// 设置时间戳
	now := time.Now()
	monsterSkill.CreatedAt = now
	monsterSkill.UpdatedAt = now

	// 插入数据库
	if err := monsterSkill.Insert(ctx, r.exec, boil.Infer()); err != nil {
		return fmt.Errorf("创建怪物技能关联失败: %w", err)
	}

	return nil
}

// BatchCreate 批量创建怪物技能关联
func (r *monsterSkillRepositoryImpl) BatchCreate(ctx context.Context, monsterSkills []*game_config.MonsterSkill) error {
	if len(monsterSkills) == 0 {
		return nil
	}

	if tx, ok := r.exec.(*sql.Tx); ok {
		return r.batchCreateWithExecutor(ctx, tx, monsterSkills)
	}

	if r.beginner != nil {
		tx, err := r.beginner.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("开启事务失败: %w", err)
		}
		defer tx.Rollback()

		if err := r.batchCreateWithExecutor(ctx, tx, monsterSkills); err != nil {
			return err
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("提交事务失败: %w", err)
		}
		return nil
	}

	return r.batchCreateWithExecutor(ctx, r.exec, monsterSkills)
}

func (r *monsterSkillRepositoryImpl) batchCreateWithExecutor(ctx context.Context, exec boil.ContextExecutor, monsterSkills []*game_config.MonsterSkill) error {
	now := time.Now()
	for _, ms := range monsterSkills {
		if ms.ID == "" {
			ms.ID = uuid.New().String()
		}
		ms.CreatedAt = now
		ms.UpdatedAt = now

		if err := ms.Insert(ctx, exec, boil.Infer()); err != nil {
			return fmt.Errorf("批量创建怪物技能关联失败: %w", err)
		}
	}
	return nil
}

// GetByMonsterID 获取怪物的所有技能
func (r *monsterSkillRepositoryImpl) GetByMonsterID(ctx context.Context, monsterID string) ([]*game_config.MonsterSkill, error) {
	monsterSkills, err := game_config.MonsterSkills(
		qm.Where("monster_id = ? AND deleted_at IS NULL", monsterID),
		qm.OrderBy("display_order ASC, created_at ASC"),
	).All(ctx, r.exec)

	if err != nil {
		return nil, fmt.Errorf("查询怪物技能列表失败: %w", err)
	}

	return monsterSkills, nil
}

// GetByMonsterAndSkill 获取怪物的特定技能
func (r *monsterSkillRepositoryImpl) GetByMonsterAndSkill(ctx context.Context, monsterID, skillID string) (*game_config.MonsterSkill, error) {
	monsterSkill, err := game_config.MonsterSkills(
		qm.Where("monster_id = ? AND skill_id = ? AND deleted_at IS NULL", monsterID, skillID),
	).One(ctx, r.exec)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("怪物技能关联不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询怪物技能关联失败: %w", err)
	}

	return monsterSkill, nil
}

// Update 更新怪物技能配置
func (r *monsterSkillRepositoryImpl) Update(ctx context.Context, monsterSkill *game_config.MonsterSkill) error {
	// 更新时间戳
	monsterSkill.UpdatedAt = time.Now()

	// 更新数据库
	if _, err := monsterSkill.Update(ctx, r.exec, boil.Infer()); err != nil {
		return fmt.Errorf("更新怪物技能配置失败: %w", err)
	}

	return nil
}

// Delete 软删除怪物技能关联
func (r *monsterSkillRepositoryImpl) Delete(ctx context.Context, monsterID, skillID string) error {
	// 查询怪物技能关联
	monsterSkill, err := r.GetByMonsterAndSkill(ctx, monsterID, skillID)
	if err != nil {
		return err
	}

	// 设置删除时间
	now := time.Now()
	monsterSkill.DeletedAt = null.TimeFrom(now)
	monsterSkill.UpdatedAt = now

	// 更新数据库
	if _, err := monsterSkill.Update(ctx, r.exec, boil.Whitelist("deleted_at", "updated_at")); err != nil {
		return fmt.Errorf("删除怪物技能关联失败: %w", err)
	}

	return nil
}

// DeleteByMonsterID 删除怪物的所有技能（软删除）
func (r *monsterSkillRepositoryImpl) DeleteByMonsterID(ctx context.Context, monsterID string) error {
	// 查询怪物的所有技能
	monsterSkills, err := r.GetByMonsterID(ctx, monsterID)
	if err != nil {
		return err
	}

	if len(monsterSkills) == 0 {
		return nil
	}

	if tx, ok := r.exec.(*sql.Tx); ok {
		return r.softDeleteSkills(ctx, tx, monsterSkills)
	}

	if r.beginner != nil {
		tx, err := r.beginner.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("开启事务失败: %w", err)
		}
		defer tx.Rollback()

		if err := r.softDeleteSkills(ctx, tx, monsterSkills); err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("提交事务失败: %w", err)
		}
		return nil
	}

	return r.softDeleteSkills(ctx, r.exec, monsterSkills)
}

func (r *monsterSkillRepositoryImpl) softDeleteSkills(ctx context.Context, exec boil.ContextExecutor, monsterSkills []*game_config.MonsterSkill) error {
	now := time.Now()
	for _, ms := range monsterSkills {
		ms.DeletedAt = null.TimeFrom(now)
		ms.UpdatedAt = now

		if _, err := ms.Update(ctx, exec, boil.Whitelist("deleted_at", "updated_at")); err != nil {
			return fmt.Errorf("删除怪物技能关联失败: %w", err)
		}
	}
	return nil
}

// Exists 检查怪物技能关联是否存在
func (r *monsterSkillRepositoryImpl) Exists(ctx context.Context, monsterID, skillID string) (bool, error) {
	count, err := game_config.MonsterSkills(
		qm.Where("monster_id = ? AND skill_id = ? AND deleted_at IS NULL", monsterID, skillID),
	).Count(ctx, r.exec)

	if err != nil {
		return false, fmt.Errorf("检查怪物技能关联是否存在失败: %w", err)
	}

	return count > 0, nil
}
