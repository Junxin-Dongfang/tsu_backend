package interfaces

import (
	"context"
	"encoding/json"
	"time"
)

// WorldDropItem 保存世界掉落物品原始字段
type WorldDropItem struct {
	ID                string          `json:"id"`
	WorldDropConfigID string          `json:"world_drop_config_id"`
	ItemID            string          `json:"item_id"`
	DropRate          *float64        `json:"drop_rate,omitempty"`
	DropWeight        *int            `json:"drop_weight,omitempty"`
	MinQuantity       int             `json:"min_quantity"`
	MaxQuantity       int             `json:"max_quantity"`
	MinLevel          *int            `json:"min_level,omitempty"`
	MaxLevel          *int            `json:"max_level,omitempty"`
	GuaranteedDrop    bool            `json:"guaranteed_drop"`
	Metadata          json.RawMessage `json:"metadata,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

// WorldDropItemWithItem 附带物品基础信息
type WorldDropItemWithItem struct {
	WorldDropItem
	ItemCode    string `json:"item_code"`
	ItemName    string `json:"item_name"`
	ItemQuality string `json:"item_quality"`
}

// ListWorldDropItemParams 查询参数
type ListWorldDropItemParams struct {
	WorldDropConfigID string
	Page              int
	PageSize          int
}

// WorldDropItemRepository 世界掉落物品仓储接口
type WorldDropItemRepository interface {
    ListByConfig(ctx context.Context, params ListWorldDropItemParams) ([]WorldDropItemWithItem, int64, error)
    GetByID(ctx context.Context, configID, itemEntryID string) (*WorldDropItemWithItem, error)
    HasItemInConfig(ctx context.Context, configID, itemID string) (bool, error)
    Create(ctx context.Context, item *WorldDropItem) error
    Update(ctx context.Context, item *WorldDropItem) error
    SoftDelete(ctx context.Context, configID, itemEntryID string) error
    ExistsActiveItem(ctx context.Context, itemID, excludeConfigID string) (bool, error)
    SumDropRates(ctx context.Context, configID string, excludeItemEntryID *string) (float64, error)
}
