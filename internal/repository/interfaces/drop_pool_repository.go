package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// DropPoolRepository 掉落池仓储接口
type DropPoolRepository interface {
	// GetByID 根据ID获取掉落池
	GetByID(ctx context.Context, poolID string) (*game_config.DropPool, error)

	// GetByCode 根据代码获取掉落池
	GetByCode(ctx context.Context, poolCode string) (*game_config.DropPool, error)

	// GetByType 根据类型获取掉落池列表
	GetByType(ctx context.Context, poolType string) ([]*game_config.DropPool, error)

	// GetPoolItems 获取掉落池中的所有物品
	GetPoolItems(ctx context.Context, poolID string) ([]*game_config.DropPoolItem, error)

	// GetPoolItemsByLevel 获取掉落池中符合等级要求的物品
	GetPoolItemsByLevel(ctx context.Context, poolID string, playerLevel int) ([]*game_config.DropPoolItem, error)

	// List 查询掉落池列表
	List(ctx context.Context, params ListDropPoolParams) ([]*game_config.DropPool, int64, error)

	// Create 创建掉落池
	Create(ctx context.Context, pool *game_config.DropPool) error

	// Update 更新掉落池
	Update(ctx context.Context, pool *game_config.DropPool) error

	// Delete 删除掉落池（软删除）
	Delete(ctx context.Context, poolID string) error

	// Count 统计掉落池数量
	Count(ctx context.Context, params ListDropPoolParams) (int64, error)

	// CreatePoolItem 添加掉落池物品
	CreatePoolItem(ctx context.Context, item *game_config.DropPoolItem) error

	// GetPoolItemByID 根据ID获取掉落池物品
	GetPoolItemByID(ctx context.Context, poolID, itemID string) (*game_config.DropPoolItem, error)

	// UpdatePoolItem 更新掉落池物品
	UpdatePoolItem(ctx context.Context, item *game_config.DropPoolItem) error

	// DeletePoolItem 删除掉落池物品（软删除）
	DeletePoolItem(ctx context.Context, poolID, itemID string) error

	// ListPoolItems 查询掉落池物品列表
	ListPoolItems(ctx context.Context, params ListDropPoolItemParams) ([]*game_config.DropPoolItem, int64, error)

	// CountPoolItems 统计掉落池物品数量
	CountPoolItems(ctx context.Context, params ListDropPoolItemParams) (int64, error)
}

// ListDropPoolParams 查询掉落池列表参数
type ListDropPoolParams struct {
	PoolType *string // 掉落池类型
	IsActive *bool   // 是否启用
	Keyword  *string // 关键词搜索（pool_code, pool_name）
	SortBy   string  // 排序字段
	SortOrder string // 排序方向
	Page     int     // 页码(从1开始)
	PageSize int     // 每页数量
}

// ListDropPoolItemParams 查询掉落池物品列表参数
type ListDropPoolItemParams struct {
	DropPoolID string  // 掉落池ID
	IsActive   *bool   // 是否启用
	MinLevel   *int16  // 最低等级
	MaxLevel   *int16  // 最高等级
	Page       int     // 页码(从1开始)
	PageSize   int     // 每页数量
}

