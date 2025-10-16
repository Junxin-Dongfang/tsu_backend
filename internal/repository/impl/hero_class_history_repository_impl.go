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

type heroClassHistoryRepositoryImpl struct {
	db *sql.DB
}

// NewHeroClassHistoryRepository 创建英雄职业历史仓储实例
func NewHeroClassHistoryRepository(db *sql.DB) interfaces.HeroClassHistoryRepository {
	return &heroClassHistoryRepositoryImpl{db: db}
}

// Create 创建职业历史记录
func (r *heroClassHistoryRepositoryImpl) Create(ctx context.Context, execer boil.ContextExecutor, history *game_runtime.HeroClassHistory) error {
	if err := history.Insert(ctx, execer, boil.Infer()); err != nil {
		return fmt.Errorf("创建职业历史记录失败: %w", err)
	}
	return nil
}

// GetByHeroID 获取英雄的职业历史
func (r *heroClassHistoryRepositoryImpl) GetByHeroID(ctx context.Context, heroID string) ([]*game_runtime.HeroClassHistory, error) {
	histories, err := game_runtime.HeroClassHistories(
		qm.Where("hero_id = ?", heroID),
		qm.OrderBy("acquired_at ASC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询职业历史失败: %w", err)
	}

	return histories, nil
}

// GetCurrentClass 获取当前职业
func (r *heroClassHistoryRepositoryImpl) GetCurrentClass(ctx context.Context, heroID string) (*game_runtime.HeroClassHistory, error) {
	history, err := game_runtime.HeroClassHistories(
		qm.Where("hero_id = ? AND is_current = ?", heroID, true),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("英雄当前职业不存在: %s", heroID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询当前职业失败: %w", err)
	}

	return history, nil
}

// GetAvailableClassesForSkills 获取可学习技能的职业列表（排除 transfer 之前的历史）
func (r *heroClassHistoryRepositoryImpl) GetAvailableClassesForSkills(ctx context.Context, heroID string) ([]*game_runtime.HeroClassHistory, error) {
	// 获取最后一次 transfer 的时间
	lastTransferTime, err := r.GetLastTransferTime(ctx, heroID)
	if err != nil {
		return nil, err
	}

	// 构建查询条件：hero_id = ? AND acquisition_type IN ('initial', 'advancement')
	whereClause := "hero_id = ? AND acquisition_type IN ('initial', 'advancement')"
	queries := []qm.QueryMod{
		qm.Where(whereClause, heroID),
		qm.OrderBy("acquired_at ASC"),
	}

	// 如果有转职记录，只保留转职之后的职业
	if lastTransferTime != nil {
		queries = append(queries, qm.Where("acquired_at > ?", *lastTransferTime))
	}

	histories, err := game_runtime.HeroClassHistories(queries...).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("查询可学习技能的职业列表失败: %w", err)
	}

	return histories, nil
}

// SetCurrentClass 设置当前职业（更新 is_current 标志）
func (r *heroClassHistoryRepositoryImpl) SetCurrentClass(ctx context.Context, tx *sql.Tx, heroID, classID, acquisitionType string) error {
	// 1. 将所有旧职业的 is_current 设为 false
	_, err := game_runtime.HeroClassHistories(
		qm.Where("hero_id = ?", heroID),
	).UpdateAll(ctx, tx, game_runtime.M{"is_current": false})
	if err != nil {
		return fmt.Errorf("更新旧职业状态失败: %w", err)
	}

	// 2. 创建新职业记录
	newHistory := &game_runtime.HeroClassHistory{
		HeroID:          heroID,
		ClassID:         classID,
		IsCurrent:       true,
		AcquiredAt:      time.Now(),
		AcquisitionType: acquisitionType,
	}

	if err := r.Create(ctx, tx, newHistory); err != nil {
		return err
	}

	return nil
}

// GetLastTransferTime 获取最后一次转职时间
func (r *heroClassHistoryRepositoryImpl) GetLastTransferTime(ctx context.Context, heroID string) (*time.Time, error) {
	history, err := game_runtime.HeroClassHistories(
		qm.Where("hero_id = ? AND acquisition_type = ?", heroID, "transfer"),
		qm.OrderBy("acquired_at DESC"),
		qm.Limit(1),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, nil // 没有转职记录
	}
	if err != nil {
		return nil, fmt.Errorf("查询最后转职时间失败: %w", err)
	}

	return &history.AcquiredAt, nil
}

