// Package dto 定义Admin模块的数据传输对象
package dto

import (
	"encoding/json"
	"time"
)

// CreateDropPoolRequest 创建掉落池请求
// @Description 创建掉落池配置
type CreateDropPoolRequest struct {
	PoolCode        string  `json:"pool_code" validate:"required,min=1,max=50" example:"monster_goblin_elite"`                                                                         // 掉落池代码，唯一标识，1-50字符
	PoolName        string  `json:"pool_name" validate:"required,min=1,max=100" example:"精英哥布林掉落池"`                                                                                    // 掉落池名称，1-100字符
	PoolType        string  `json:"pool_type" validate:"required,oneof=monster dungeon quest activity boss other" example:"monster" enums:"monster,dungeon,quest,activity,boss,other"` // 掉落池类型：monster(怪物)、dungeon(副本)、quest(任务)、activity(活动)、boss(Boss)、other(其他)
	Description     *string `json:"description,omitempty" example:"精英哥布林的掉落池配置"`                                                                                                       // 掉落池描述（可选）
	MinDrops        int16   `json:"min_drops" validate:"min=0" example:"1"`                                                                                                            // 最小掉落数量，必须 >= 0
	MaxDrops        int16   `json:"max_drops" validate:"min=0" example:"3"`                                                                                                            // 最大掉落数量，必须 >= min_drops
	GuaranteedDrops int16   `json:"guaranteed_drops" validate:"min=0" example:"1"`                                                                                                     // 保底掉落数量，必须 <= min_drops
}

// UpdateDropPoolRequest 更新掉落池请求
type UpdateDropPoolRequest struct {
	PoolCode        *string `json:"pool_code,omitempty" validate:"omitempty,min=1,max=50"`
	PoolName        *string `json:"pool_name,omitempty" validate:"omitempty,min=1,max=100"`
	PoolType        *string `json:"pool_type,omitempty" validate:"omitempty,oneof=monster dungeon quest activity boss other"`
	Description     *string `json:"description,omitempty"`
	MinDrops        *int16  `json:"min_drops,omitempty" validate:"omitempty,min=0"`
	MaxDrops        *int16  `json:"max_drops,omitempty" validate:"omitempty,min=0"`
	GuaranteedDrops *int16  `json:"guaranteed_drops,omitempty" validate:"omitempty,min=0"`
	IsActive        *bool   `json:"is_active,omitempty"`
}

// DropPoolResponse 掉落池响应
type DropPoolResponse struct {
	ID              string     `json:"id"`
	PoolCode        string     `json:"pool_code"`
	PoolName        string     `json:"pool_name"`
	PoolType        string     `json:"pool_type"`
	Description     *string    `json:"description,omitempty"`
	MinDrops        int16      `json:"min_drops"`
	MaxDrops        int16      `json:"max_drops"`
	GuaranteedDrops int16      `json:"guaranteed_drops"`
	IsActive        bool       `json:"is_active"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}

// DropPoolListResponse 掉落池列表响应
type DropPoolListResponse struct {
	Items    []DropPoolResponse `json:"items"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

// AddDropPoolItemRequest 添加掉落物品请求
type AddDropPoolItemRequest struct {
	ItemID     string   `json:"item_id" validate:"required,uuid"`
	DropWeight *int     `json:"drop_weight,omitempty" validate:"omitempty,min=1"`
	DropRate   *float64 `json:"drop_rate,omitempty" validate:"omitempty,gt=0,lte=1"`

	// 品质权重 - 定义不同品质的掉落权重
	// 格式: {"poor":10,"normal":50,"fine":30,"excellent":8,"superb":1.5,"master":0.4,"epic":0.09,"legendary":0.01}
	// 键: 品质名称(poor/normal/fine/excellent/superb/master/epic/legendary/mythic)
	// 值: 权重值,数值越大掉落概率越高
	// 示例: {"normal":50,"fine":30,"excellent":15,"superb":5}
	QualityWeights json.RawMessage `json:"quality_weights,omitempty" swaggertype:"string"`

	MinQuantity int16  `json:"min_quantity" validate:"required,min=1"`
	MaxQuantity int16  `json:"max_quantity" validate:"required,min=1"`
	MinLevel    *int16 `json:"min_level,omitempty" validate:"omitempty,min=1,max=100"`
	MaxLevel    *int16 `json:"max_level,omitempty" validate:"omitempty,min=1,max=100"`
}

// UpdateDropPoolItemRequest 更新掉落物品请求
type UpdateDropPoolItemRequest struct {
	DropWeight *int     `json:"drop_weight,omitempty" validate:"omitempty,min=1"`
	DropRate   *float64 `json:"drop_rate,omitempty" validate:"omitempty,gt=0,lte=1"`

	// 品质权重 - 详细说明见AddDropPoolItemRequest
	QualityWeights json.RawMessage `json:"quality_weights,omitempty" swaggertype:"string"`

	MinQuantity *int16 `json:"min_quantity,omitempty" validate:"omitempty,min=1"`
	MaxQuantity *int16 `json:"max_quantity,omitempty" validate:"omitempty,min=1"`
	MinLevel    *int16 `json:"min_level,omitempty" validate:"omitempty,min=1,max=100"`
	MaxLevel    *int16 `json:"max_level,omitempty" validate:"omitempty,min=1,max=100"`
	IsActive    *bool  `json:"is_active,omitempty"`
}

// DropPoolItemResponse 掉落物品响应
type DropPoolItemResponse struct {
	ID         string   `json:"id"`
	DropPoolID string   `json:"drop_pool_id"`
	ItemID     string   `json:"item_id"`
	ItemCode   string   `json:"item_code"`
	ItemName   string   `json:"item_name"`
	DropWeight *int     `json:"drop_weight,omitempty"`
	DropRate   *float64 `json:"drop_rate,omitempty"`

	// 品质权重 - 详细说明见AddDropPoolItemRequest
	QualityWeights json.RawMessage `json:"quality_weights,omitempty" swaggertype:"string"`

	MinQuantity int16      `json:"min_quantity"`
	MaxQuantity int16      `json:"max_quantity"`
	MinLevel    *int16     `json:"min_level,omitempty"`
	MaxLevel    *int16     `json:"max_level,omitempty"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// DropPoolItemListResponse 掉落物品列表响应
type DropPoolItemListResponse struct {
	Items    []DropPoolItemResponse `json:"items"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}
