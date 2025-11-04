// Package dto 定义Admin模块的数据传输对象
package dto

import (
	"encoding/json"
	"time"
)

// CreateWorldDropRequest 创建世界掉落请求
type CreateWorldDropRequest struct {
	ItemID          string `json:"item_id" validate:"required,uuid"`
	TotalDropLimit  *int   `json:"total_drop_limit,omitempty" validate:"omitempty,min=0"`
	DailyDropLimit  *int   `json:"daily_drop_limit,omitempty" validate:"omitempty,min=0"`
	HourlyDropLimit *int   `json:"hourly_drop_limit,omitempty" validate:"omitempty,min=0"`
	MinDropInterval *int   `json:"min_drop_interval,omitempty" validate:"omitempty,min=0"`
	MaxDropInterval *int   `json:"max_drop_interval,omitempty" validate:"omitempty,min=0"`

	// 触发条件 - 定义何时可以掉落该物品
	// 格式: {"min_player_level":10,"max_player_level":20,"required_quest":"quest_id","zone":"zone_name"}
	// min_player_level: 最低玩家等级
	// max_player_level: 最高玩家等级
	// required_quest: 需要完成的任务ID(可选)
	// zone: 限定区域(可选)
	// 示例: {"min_player_level":10,"max_player_level":20,"zone":"forest"}
	TriggerConditions json.RawMessage `json:"trigger_conditions,omitempty" swaggertype:"string"`

	BaseDropRate float64 `json:"base_drop_rate" validate:"required,gt=0,lte=1"`

	// 掉落率修正器 - 根据不同条件调整掉落率
	// 格式: {"time_of_day":{"morning":1.2,"night":0.8},"player_luck_bonus":0.1}
	// time_of_day: 时间段修正(可选)
	// player_luck_bonus: 玩家幸运加成(可选)
	// party_size_bonus: 队伍人数加成(可选)
	// 示例: {"time_of_day":{"morning":1.2,"night":0.8},"player_luck_bonus":0.1}
	DropRateModifiers json.RawMessage `json:"drop_rate_modifiers,omitempty" swaggertype:"string"`
}

// UpdateWorldDropRequest 更新世界掉落请求
type UpdateWorldDropRequest struct {
	TotalDropLimit  *int `json:"total_drop_limit,omitempty" validate:"omitempty,min=0"`
	DailyDropLimit  *int `json:"daily_drop_limit,omitempty" validate:"omitempty,min=0"`
	HourlyDropLimit *int `json:"hourly_drop_limit,omitempty" validate:"omitempty,min=0"`
	MinDropInterval *int `json:"min_drop_interval,omitempty" validate:"omitempty,min=0"`
	MaxDropInterval *int `json:"max_drop_interval,omitempty" validate:"omitempty,min=0"`

	// 触发条件 - 详细说明见CreateWorldDropRequest
	TriggerConditions json.RawMessage `json:"trigger_conditions,omitempty" swaggertype:"string"`

	BaseDropRate *float64 `json:"base_drop_rate,omitempty" validate:"omitempty,gt=0,lte=1"`

	// 掉落率修正器 - 详细说明见CreateWorldDropRequest
	DropRateModifiers json.RawMessage `json:"drop_rate_modifiers,omitempty" swaggertype:"string"`

	IsActive *bool `json:"is_active,omitempty"`
}

// WorldDropResponse 世界掉落响应
type WorldDropResponse struct {
	ID              string `json:"id"`
	ItemID          string `json:"item_id"`
	ItemCode        string `json:"item_code"`
	ItemName        string `json:"item_name"`
	TotalDropLimit  *int   `json:"total_drop_limit,omitempty"`
	DailyDropLimit  *int   `json:"daily_drop_limit,omitempty"`
	HourlyDropLimit *int   `json:"hourly_drop_limit,omitempty"`
	MinDropInterval *int   `json:"min_drop_interval,omitempty"`
	MaxDropInterval *int   `json:"max_drop_interval,omitempty"`

	// 触发条件 - 详细说明见CreateWorldDropRequest
	TriggerConditions json.RawMessage `json:"trigger_conditions,omitempty" swaggertype:"string"`

	BaseDropRate float64 `json:"base_drop_rate"`

	// 掉落率修正器 - 详细说明见CreateWorldDropRequest
	DropRateModifiers json.RawMessage `json:"drop_rate_modifiers,omitempty" swaggertype:"string"`

	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// WorldDropListResponse 世界掉落列表响应
type WorldDropListResponse struct {
	Items    []WorldDropResponse `json:"items"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}
