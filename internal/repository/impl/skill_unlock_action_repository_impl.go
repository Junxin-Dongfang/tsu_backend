package impl

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/google/uuid"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/interfaces"
)

type skillUnlockActionRepositoryImpl struct {
	db *sql.DB
}

// NewSkillUnlockActionRepository 创建技能解锁动作仓储实例
func NewSkillUnlockActionRepository(db *sql.DB) interfaces.SkillUnlockActionRepository {
	return &skillUnlockActionRepositoryImpl{db: db}
}

func (r *skillUnlockActionRepositoryImpl) GetByID(ctx context.Context, id string) (*game_config.SkillUnlockAction, error) {
	unlockAction, err := game_config.SkillUnlockActions(
		qm.Where("id = ? AND deleted_at IS NULL", id),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("技能解锁动作不存在: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("查询技能解锁动作失败: %w", err)
	}

	return unlockAction, nil
}

func (r *skillUnlockActionRepositoryImpl) GetBySkillID(ctx context.Context, skillID string) ([]*game_config.SkillUnlockAction, error) {
	unlockActions, err := game_config.SkillUnlockActions(
		qm.Where("skill_id = ? AND deleted_at IS NULL", skillID),
		qm.OrderBy("unlock_level ASC, is_default DESC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询技能解锁动作失败: %w", err)
	}

	return unlockActions, nil
}

func (r *skillUnlockActionRepositoryImpl) Create(ctx context.Context, unlockAction *game_config.SkillUnlockAction) error {
	if unlockAction.ID == "" {
		unlockAction.ID = uuid.New().String()
	}

	unlockAction.CreatedAt.SetValid(time.Now())

	if err := unlockAction.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("创建技能解锁动作失败: %w", err)
	}

	return nil
}

func (r *skillUnlockActionRepositoryImpl) Update(ctx context.Context, unlockAction *game_config.SkillUnlockAction) error {
	// 只更新允许更新的字段
	cols := []string{"unlock_level", "is_default", "level_scaling_config"}

	_, err := unlockAction.Update(ctx, r.db, boil.Whitelist(cols...))
	if err != nil {
		return fmt.Errorf("更新技能解锁动作失败: %w", err)
	}

	return nil
}

func (r *skillUnlockActionRepositoryImpl) Delete(ctx context.Context, unlockActionID string) error {
	unlockAction, err := game_config.SkillUnlockActions(
		qm.Where("id = ? AND deleted_at IS NULL", unlockActionID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return fmt.Errorf("技能解锁动作不存在: %s", unlockActionID)
	}
	if err != nil {
		return fmt.Errorf("查询技能解锁动作失败: %w", err)
	}

	unlockAction.DeletedAt.SetValid(time.Now())

	if _, err := unlockAction.Update(ctx, r.db, boil.Whitelist("deleted_at")); err != nil {
		return fmt.Errorf("删除技能解锁动作失败: %w", err)
	}

	return nil
}

func (r *skillUnlockActionRepositoryImpl) DeleteBySkillAndAction(ctx context.Context, skillID, actionID string) error {
	_, err := game_config.SkillUnlockActions(
		qm.Where("skill_id = ? AND action_id = ? AND deleted_at IS NULL", skillID, actionID),
	).UpdateAll(ctx, r.db, game_config.M{
		"deleted_at": time.Now(),
	})

	if err != nil {
		return fmt.Errorf("删除技能解锁动作失败: %w", err)
	}

	return nil
}

func (r *skillUnlockActionRepositoryImpl) DeleteAllBySkillID(ctx context.Context, skillID string) error {
	_, err := game_config.SkillUnlockActions(
		qm.Where("skill_id = ? AND deleted_at IS NULL", skillID),
	).UpdateAll(ctx, r.db, game_config.M{
		"deleted_at": time.Now(),
	})

	if err != nil {
		return fmt.Errorf("删除技能所有解锁动作失败: %w", err)
	}

	return nil
}

func (r *skillUnlockActionRepositoryImpl) BatchCreate(ctx context.Context, unlockActions []*game_config.SkillUnlockAction) error {
	for _, unlockAction := range unlockActions {
		if unlockAction.ID == "" {
			unlockAction.ID = uuid.New().String()
		}
		unlockAction.CreatedAt.SetValid(time.Now())

		if err := unlockAction.Insert(ctx, r.db, boil.Infer()); err != nil {
			return fmt.Errorf("批量创建技能解锁动作失败: %w", err)
		}
	}

	return nil
}
