package interfaces

import (
	"context"
	"database/sql"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"tsu-self/internal/entity/game_runtime"
)

// HeroClassHistoryRepository 英雄职业历史仓储接口
type HeroClassHistoryRepository interface {
	// Create 创建职业历史记录
	Create(ctx context.Context, execer boil.ContextExecutor, history *game_runtime.HeroClassHistory) error

	// GetByHeroID 获取英雄的职业历史
	GetByHeroID(ctx context.Context, heroID string) ([]*game_runtime.HeroClassHistory, error)

	// GetCurrentClass 获取当前职业
	GetCurrentClass(ctx context.Context, heroID string) (*game_runtime.HeroClassHistory, error)

	// GetAvailableClassesForSkills 获取可学习技能的职业列表（排除 transfer 之前的历史）
	GetAvailableClassesForSkills(ctx context.Context, heroID string) ([]*game_runtime.HeroClassHistory, error)

	// SetCurrentClass 设置当前职业（更新 is_current 标志）
	SetCurrentClass(ctx context.Context, tx *sql.Tx, heroID, classID, acquisitionType string) error

	// GetLastTransferTime 获取最后一次转职时间
	GetLastTransferTime(ctx context.Context, heroID string) (*time.Time, error)

	// SetNonCurrent 将英雄的当前职业设为非当前
	SetNonCurrent(ctx context.Context, tx *sql.Tx, heroID string) error
}

