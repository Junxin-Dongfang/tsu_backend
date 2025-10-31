package interfaces

import (
	"context"

	"github.com/aarondl/sqlboiler/v4/boil"
	"tsu-self/internal/entity/game_runtime"
)

// ItemDropRecordRepository 物品掉落记录仓储接口
type ItemDropRecordRepository interface {
	// Create 创建掉落记录
	Create(ctx context.Context, execer boil.ContextExecutor, record *game_runtime.ItemDropRecord) error

	// GetByID 根据ID获取掉落记录
	GetByID(ctx context.Context, recordID string) (*game_runtime.ItemDropRecord, error)

	// GetByReceiver 查询玩家的掉落记录
	GetByReceiver(ctx context.Context, receiverID string, limit int) ([]*game_runtime.ItemDropRecord, error)

	// GetByItemConfig 查询特定物品的掉落记录
	GetByItemConfig(ctx context.Context, itemConfigID string, limit int) ([]*game_runtime.ItemDropRecord, error)

	// GetBySource 查询特定来源的掉落记录
	GetBySource(ctx context.Context, dropSource string, sourceID string, limit int) ([]*game_runtime.ItemDropRecord, error)

	// GetRecentDrops 查询最近的掉落记录
	GetRecentDrops(ctx context.Context, limit int) ([]*game_runtime.ItemDropRecord, error)

	// CountByItemConfigAndTimeRange 统计特定物品在时间范围内的掉落次数
	CountByItemConfigAndTimeRange(ctx context.Context, itemConfigID string, startTime, endTime string) (int64, error)
}

