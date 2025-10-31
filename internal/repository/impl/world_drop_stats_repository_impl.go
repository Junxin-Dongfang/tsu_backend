package impl

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/repository/interfaces"
)

type worldDropStatsRepositoryImpl struct {
	db *sql.DB
}

// NewWorldDropStatsRepository 创建世界掉落统计仓储实例
func NewWorldDropStatsRepository(db *sql.DB) interfaces.WorldDropStatsRepository {
	return &worldDropStatsRepositoryImpl{db: db}
}

// GetByConfigID 根据配置ID获取统计信息
func (r *worldDropStatsRepositoryImpl) GetByConfigID(ctx context.Context, configID string) (*game_runtime.WorldDropStat, error) {
	stats, err := game_runtime.WorldDropStats(
		qm.Where("world_drop_config_id = ?", configID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("世界掉落统计不存在: %s", configID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询世界掉落统计失败: %w", err)
	}

	return stats, nil
}

// GetByConfigIDForUpdate 根据配置ID获取统计信息（带行锁）
func (r *worldDropStatsRepositoryImpl) GetByConfigIDForUpdate(ctx context.Context, tx *sql.Tx, configID string) (*game_runtime.WorldDropStat, error) {
	stats, err := game_runtime.WorldDropStats(
		qm.Where("world_drop_config_id = ?", configID),
		qm.For("UPDATE"),
	).One(ctx, tx)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("世界掉落统计不存在: %s", configID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询世界掉落统计失败（带锁）: %w", err)
	}

	return stats, nil
}

// Create 创建统计记录
func (r *worldDropStatsRepositoryImpl) Create(ctx context.Context, execer boil.ContextExecutor, stats *game_runtime.WorldDropStat) error {
	if err := stats.Insert(ctx, execer, boil.Infer()); err != nil {
		return fmt.Errorf("创建世界掉落统计失败: %w", err)
	}
	return nil
}

// Update 更新统计记录
func (r *worldDropStatsRepositoryImpl) Update(ctx context.Context, execer boil.ContextExecutor, stats *game_runtime.WorldDropStat) error {
	if _, err := stats.Update(ctx, execer, boil.Infer()); err != nil {
		return fmt.Errorf("更新世界掉落统计失败: %w", err)
	}
	return nil
}

// IncrementDropCount 增加掉落计数
func (r *worldDropStatsRepositoryImpl) IncrementDropCount(ctx context.Context, execer boil.ContextExecutor, configID string) error {
	// 获取统计记录（带锁）
	tx, ok := execer.(*sql.Tx)
	if !ok {
		return fmt.Errorf("execer必须是*sql.Tx类型")
	}

	stats, err := r.GetByConfigIDForUpdate(ctx, tx, configID)
	if err != nil {
		return err
	}

	now := time.Now()

	// 增加总掉落数量
	stats.TotalDropped++

	// 增加每日掉落数量
	stats.DailyDropped++

	// 增加每小时掉落数量
	stats.HourlyDropped++

	// 更新上次掉落时间
	stats.LastDropAt = null.TimeFrom(now)

	// 更新统计记录
	if err := r.Update(ctx, execer, stats); err != nil {
		return err
	}

	return nil
}

// ResetDailyStats 重置每日统计
func (r *worldDropStatsRepositoryImpl) ResetDailyStats(ctx context.Context, execer boil.ContextExecutor, configID string) error {
	stats, err := r.GetByConfigID(ctx, configID)
	if err != nil {
		return err
	}

	now := time.Now()

	// 重置每日掉落数量
	stats.DailyDropped = 0

	// 更新每日重置时间
	stats.DailyResetAt = null.TimeFrom(now)

	// 更新统计记录
	if err := r.Update(ctx, execer, stats); err != nil {
		return err
	}

	return nil
}

// ResetHourlyStats 重置每小时统计
func (r *worldDropStatsRepositoryImpl) ResetHourlyStats(ctx context.Context, execer boil.ContextExecutor, configID string) error {
	stats, err := r.GetByConfigID(ctx, configID)
	if err != nil {
		return err
	}

	now := time.Now()

	// 重置每小时掉落数量
	stats.HourlyDropped = 0

	// 更新每小时重置时间
	stats.HourlyResetAt = null.TimeFrom(now)

	// 更新统计记录
	if err := r.Update(ctx, execer, stats); err != nil {
		return err
	}

	return nil
}

// GetStatsNeedingDailyReset 获取需要每日重置的统计记录
func (r *worldDropStatsRepositoryImpl) GetStatsNeedingDailyReset(ctx context.Context) ([]*game_runtime.WorldDropStat, error) {
	// 查询上次重置时间超过24小时的记录
	oneDayAgo := time.Now().Add(-24 * time.Hour)

	stats, err := game_runtime.WorldDropStats(
		qm.Where("daily_reset_at IS NULL OR daily_reset_at < ?", oneDayAgo),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询需要每日重置的统计记录失败: %w", err)
	}

	return stats, nil
}

// GetStatsNeedingHourlyReset 获取需要每小时重置的统计记录
func (r *worldDropStatsRepositoryImpl) GetStatsNeedingHourlyReset(ctx context.Context) ([]*game_runtime.WorldDropStat, error) {
	// 查询上次重置时间超过1小时的记录
	oneHourAgo := time.Now().Add(-1 * time.Hour)

	stats, err := game_runtime.WorldDropStats(
		qm.Where("hourly_reset_at IS NULL OR hourly_reset_at < ?", oneHourAgo),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询需要每小时重置的统计记录失败: %w", err)
	}

	return stats, nil
}

