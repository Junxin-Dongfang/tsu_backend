package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// DungeonQueryParams 地城查询参数
type DungeonQueryParams struct {
	DungeonCode *string
	DungeonName *string
	MinLevel    *int16
	MaxLevel    *int16
	IsActive    *bool
	Limit       int
	Offset      int
	OrderBy     string
	OrderDesc   bool
}

// DungeonRepository 地城配置仓储接口
type DungeonRepository interface {
	// List 分页查询地城
	List(ctx context.Context, params DungeonQueryParams) ([]*game_config.Dungeon, int64, error)

	// GetByID 根据 ID 获取地城（未删除）
	GetByID(ctx context.Context, dungeonID string) (*game_config.Dungeon, error)

	// GetByCode 根据代码获取地城
	GetByCode(ctx context.Context, code string) (*game_config.Dungeon, error)

	// Exists 检查地城代码是否存在
	Exists(ctx context.Context, code string) (bool, error)

	// Create 创建地城
	Create(ctx context.Context, dungeon *game_config.Dungeon) error

	// Update 更新地城
	Update(ctx context.Context, dungeon *game_config.Dungeon) error

	// Delete 软删除地城
	Delete(ctx context.Context, dungeonID string) error
}
