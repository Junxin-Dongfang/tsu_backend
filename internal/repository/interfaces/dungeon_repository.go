package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// DungeonQueryParams 地城查询参数
type DungeonQueryParams struct {
	DungeonCode *string // 地城代码（模糊搜索）
	DungeonName *string // 地城名称（模糊搜索）
	MinLevel    *int16  // 最小等级
	MaxLevel    *int16  // 最大等级
	IsActive    *bool   // 是否启用
	Limit       int     // 每页数量
	Offset      int     // 偏移量
	OrderBy     string  // 排序字段（created_at, updated_at, min_level）
	OrderDesc   bool    // 是否降序
}

// DungeonRepository 地城仓储接口
type DungeonRepository interface {
	// GetByID 根据ID获取地城
	GetByID(ctx context.Context, dungeonID string) (*game_config.Dungeon, error)

	// GetByCode 根据代码获取地城
	GetByCode(ctx context.Context, code string) (*game_config.Dungeon, error)

	// List 获取地城列表
	List(ctx context.Context, params DungeonQueryParams) ([]*game_config.Dungeon, int64, error)

	// Create 创建地城
	Create(ctx context.Context, dungeon *game_config.Dungeon) error

	// Update 更新地城
	Update(ctx context.Context, dungeon *game_config.Dungeon) error

	// Delete 软删除地城
	Delete(ctx context.Context, dungeonID string) error

	// Exists 检查代码是否存在
	Exists(ctx context.Context, code string) (bool, error)

	// ExistsExcludingID 检查代码是否存在（排除指定ID）
	ExistsExcludingID(ctx context.Context, code string, excludeID string) (bool, error)
}

