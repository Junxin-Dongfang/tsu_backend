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

type buffEffectRepositoryImpl struct {
	db *sql.DB
}

// NewBuffEffectRepository 创建Buff效果关联仓储实例
func NewBuffEffectRepository(db *sql.DB) interfaces.BuffEffectRepository {
	return &buffEffectRepositoryImpl{db: db}
}

func (r *buffEffectRepositoryImpl) GetByBuffID(ctx context.Context, buffID string) ([]*game_config.BuffEffect, error) {
	buffEffects, err := game_config.BuffEffects(
		qm.Where("buff_id = ? AND deleted_at IS NULL", buffID),
		qm.OrderBy("trigger_timing ASC, execution_order ASC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询Buff效果关联失败: %w", err)
	}

	return buffEffects, nil
}

func (r *buffEffectRepositoryImpl) Create(ctx context.Context, buffEffect *game_config.BuffEffect) error {
	if buffEffect.ID == "" {
		buffEffect.ID = uuid.New().String()
	}

	buffEffect.CreatedAt.SetValid(time.Now())

	if err := buffEffect.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("创建Buff效果关联失败: %w", err)
	}

	return nil
}

func (r *buffEffectRepositoryImpl) Delete(ctx context.Context, buffEffectID string) error {
	buffEffect, err := game_config.BuffEffects(
		qm.Where("id = ? AND deleted_at IS NULL", buffEffectID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return fmt.Errorf("Buff效果关联不存在: %s", buffEffectID)
	}
	if err != nil {
		return fmt.Errorf("查询Buff效果关联失败: %w", err)
	}

	buffEffect.DeletedAt.SetValid(time.Now())

	if _, err := buffEffect.Update(ctx, r.db, boil.Whitelist("deleted_at")); err != nil {
		return fmt.Errorf("删除Buff效果关联失败: %w", err)
	}

	return nil
}

func (r *buffEffectRepositoryImpl) DeleteByBuffAndEffect(ctx context.Context, buffID, effectID, triggerTiming string) error {
	_, err := game_config.BuffEffects(
		qm.Where("buff_id = ? AND effect_id = ? AND trigger_timing = ? AND deleted_at IS NULL", buffID, effectID, triggerTiming),
	).UpdateAll(ctx, r.db, game_config.M{
		"deleted_at": time.Now(),
	})

	if err != nil {
		return fmt.Errorf("删除Buff效果关联失败: %w", err)
	}

	return nil
}

func (r *buffEffectRepositoryImpl) DeleteAllByBuffID(ctx context.Context, buffID string) error {
	_, err := game_config.BuffEffects(
		qm.Where("buff_id = ? AND deleted_at IS NULL", buffID),
	).UpdateAll(ctx, r.db, game_config.M{
		"deleted_at": time.Now(),
	})

	if err != nil {
		return fmt.Errorf("删除Buff所有效果关联失败: %w", err)
	}

	return nil
}

func (r *buffEffectRepositoryImpl) BatchCreate(ctx context.Context, buffEffects []*game_config.BuffEffect) error {
	for _, buffEffect := range buffEffects {
		if buffEffect.ID == "" {
			buffEffect.ID = uuid.New().String()
		}
		buffEffect.CreatedAt.SetValid(time.Now())

		if err := buffEffect.Insert(ctx, r.db, boil.Infer()); err != nil {
			return fmt.Errorf("批量创建Buff效果关联失败: %w", err)
		}
	}

	return nil
}
