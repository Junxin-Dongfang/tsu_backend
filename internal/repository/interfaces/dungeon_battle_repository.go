package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// DungeonBattleRepository 战斗配置仓储接口
type DungeonBattleRepository interface {
	// GetByID 根据ID获取战斗配置
	GetByID(ctx context.Context, battleID string) (*game_config.DungeonBattle, error)

	// GetByCode 根据代码获取战斗配置
	GetByCode(ctx context.Context, code string) (*game_config.DungeonBattle, error)

	// Create 创建战斗配置
	Create(ctx context.Context, battle *game_config.DungeonBattle) error

	// Update 更新战斗配置
	Update(ctx context.Context, battle *game_config.DungeonBattle) error

	// Delete 软删除战斗配置
	Delete(ctx context.Context, battleID string) error

	// Exists 检查代码是否存在
	Exists(ctx context.Context, code string) (bool, error)

	// ExistsExcludingID 检查代码是否存在（排除指定ID）
	ExistsExcludingID(ctx context.Context, code string, excludeID string) (bool, error)
}

