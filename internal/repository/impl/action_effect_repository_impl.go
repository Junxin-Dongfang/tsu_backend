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

type actionEffectRepositoryImpl struct {
	db *sql.DB
}

// NewActionEffectRepository 创建动作效果关联仓储实例
func NewActionEffectRepository(db *sql.DB) interfaces.ActionEffectRepository {
	return &actionEffectRepositoryImpl{db: db}
}

func (r *actionEffectRepositoryImpl) GetByActionID(ctx context.Context, actionID string) ([]*game_config.ActionEffect, error) {
	actionEffects, err := game_config.ActionEffects(
		qm.Where("action_id = ? AND deleted_at IS NULL", actionID),
		qm.OrderBy("execution_order ASC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询动作效果关联失败: %w", err)
	}

	return actionEffects, nil
}

func (r *actionEffectRepositoryImpl) Create(ctx context.Context, actionEffect *game_config.ActionEffect) error {
	if actionEffect.ID == "" {
		actionEffect.ID = uuid.New().String()
	}

	actionEffect.CreatedAt.SetValid(time.Now())

	if err := actionEffect.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("创建动作效果关联失败: %w", err)
	}

	return nil
}

func (r *actionEffectRepositoryImpl) Delete(ctx context.Context, actionEffectID string) error {
	actionEffect, err := game_config.ActionEffects(
		qm.Where("id = ? AND deleted_at IS NULL", actionEffectID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return fmt.Errorf("动作效果关联不存在: %s", actionEffectID)
	}
	if err != nil {
		return fmt.Errorf("查询动作效果关联失败: %w", err)
	}

	actionEffect.DeletedAt.SetValid(time.Now())

	if _, err := actionEffect.Update(ctx, r.db, boil.Whitelist("deleted_at")); err != nil {
		return fmt.Errorf("删除动作效果关联失败: %w", err)
	}

	return nil
}

func (r *actionEffectRepositoryImpl) DeleteByActionAndEffect(ctx context.Context, actionID, effectID string, executionOrder int) error {
	_, err := game_config.ActionEffects(
		qm.Where("action_id = ? AND effect_id = ? AND execution_order = ? AND deleted_at IS NULL", actionID, effectID, executionOrder),
	).UpdateAll(ctx, r.db, game_config.M{
		"deleted_at": time.Now(),
	})

	if err != nil {
		return fmt.Errorf("删除动作效果关联失败: %w", err)
	}

	return nil
}

func (r *actionEffectRepositoryImpl) DeleteAllByActionID(ctx context.Context, actionID string) error {
	_, err := game_config.ActionEffects(
		qm.Where("action_id = ? AND deleted_at IS NULL", actionID),
	).UpdateAll(ctx, r.db, game_config.M{
		"deleted_at": time.Now(),
	})

	if err != nil {
		return fmt.Errorf("删除动作所有效果关联失败: %w", err)
	}

	return nil
}

func (r *actionEffectRepositoryImpl) BatchCreate(ctx context.Context, actionEffects []*game_config.ActionEffect) error {
	for _, actionEffect := range actionEffects {
		if actionEffect.ID == "" {
			actionEffect.ID = uuid.New().String()
		}
		actionEffect.CreatedAt.SetValid(time.Now())

		if err := actionEffect.Insert(ctx, r.db, boil.Infer()); err != nil {
			return fmt.Errorf("批量创建动作效果关联失败: %w", err)
		}
	}

	return nil
}
