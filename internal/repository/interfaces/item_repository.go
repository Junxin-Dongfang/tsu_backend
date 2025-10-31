package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// ItemRepository 装备/物品配置仓储接口
type ItemRepository interface {
	// GetByID 根据ID获取装备配置
	GetByID(ctx context.Context, itemID string) (*game_config.Item, error)

	// GetByCode 根据代码获取装备配置
	GetByCode(ctx context.Context, itemCode string) (*game_config.Item, error)

	// List 查询装备配置列表
	List(ctx context.Context, params ListItemParams) ([]*game_config.Item, int64, error)

	// GetByType 根据类型获取装备配置列表
	GetByType(ctx context.Context, itemType string) ([]*game_config.Item, error)

	// GetByIDs 根据ID列表批量获取装备配置
	GetByIDs(ctx context.Context, itemIDs []string) ([]*game_config.Item, error)

	// Create 创建装备配置
	Create(ctx context.Context, item *game_config.Item) error

	// Update 更新装备配置
	Update(ctx context.Context, item *game_config.Item) error

	// Delete 删除装备配置(软删除)
	Delete(ctx context.Context, itemID string) error

	// CheckCodeExists 检查物品代码是否已存在
	CheckCodeExists(ctx context.Context, itemCode string, excludeID *string) (bool, error)

	// GetUnassignedItems 查询未关联套装的装备
	GetUnassignedItems(ctx context.Context, params ListItemParams) ([]*game_config.Item, int64, error)
}

// ListItemParams 查询装备配置列表参数
type ListItemParams struct {
	ItemType    *string  // 物品类型
	ItemQuality *string  // 物品品质
	EquipSlot   *string  // 装备槽位
	MinLevel    *int     // 最低等级
	MaxLevel    *int     // 最高等级
	IsActive    *bool    // 是否启用
	TagIDs      []string // 标签ID列表（多标签筛选）
	Keyword     *string  // 关键词搜索（item_code或item_name）
	Page        int      // 页码(从1开始)
	PageSize    int      // 每页数量
}

