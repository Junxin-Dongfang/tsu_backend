package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// DungeonEventRepository 事件配置仓储接口
type DungeonEventRepository interface {
	// GetByID 根据ID获取事件配置
	GetByID(ctx context.Context, eventID string) (*game_config.DungeonEvent, error)

	// GetByCode 根据代码获取事件配置
	GetByCode(ctx context.Context, code string) (*game_config.DungeonEvent, error)

	// Create 创建事件配置
	Create(ctx context.Context, event *game_config.DungeonEvent) error

	// Update 更新事件配置
	Update(ctx context.Context, event *game_config.DungeonEvent) error

	// Delete 软删除事件配置
	Delete(ctx context.Context, eventID string) error

	// Exists 检查代码是否存在
	Exists(ctx context.Context, code string) (bool, error)

	// ExistsExcludingID 检查代码是否存在（排除指定ID）
	ExistsExcludingID(ctx context.Context, code string, excludeID string) (bool, error)
}

