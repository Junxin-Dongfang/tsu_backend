package interfaces

import (
	"context"
	"database/sql"

	"github.com/aarondl/sqlboiler/v4/boil"
	"tsu-self/internal/entity/game_runtime"
)

// WorldDropStatsRepository 世界掉落统计仓储接口
type WorldDropStatsRepository interface {
	// GetByConfigID 根据配置ID获取统计信息
	GetByConfigID(ctx context.Context, configID string) (*game_runtime.WorldDropStat, error)

	// GetByConfigIDForUpdate 根据配置ID获取统计信息（带行锁）
	GetByConfigIDForUpdate(ctx context.Context, tx *sql.Tx, configID string) (*game_runtime.WorldDropStat, error)

	// Create 创建统计记录
	Create(ctx context.Context, execer boil.ContextExecutor, stats *game_runtime.WorldDropStat) error

	// Update 更新统计记录
	Update(ctx context.Context, execer boil.ContextExecutor, stats *game_runtime.WorldDropStat) error

	// IncrementDropCount 增加掉落计数
	IncrementDropCount(ctx context.Context, execer boil.ContextExecutor, configID string) error

	// ResetDailyStats 重置每日统计
	ResetDailyStats(ctx context.Context, execer boil.ContextExecutor, configID string) error

	// ResetHourlyStats 重置每小时统计
	ResetHourlyStats(ctx context.Context, execer boil.ContextExecutor, configID string) error

	// GetStatsNeedingDailyReset 获取需要每日重置的统计记录
	GetStatsNeedingDailyReset(ctx context.Context) ([]*game_runtime.WorldDropStat, error)

	// GetStatsNeedingHourlyReset 获取需要每小时重置的统计记录
	GetStatsNeedingHourlyReset(ctx context.Context) ([]*game_runtime.WorldDropStat, error)
}

