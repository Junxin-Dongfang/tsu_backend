package dto

import (
	"encoding/json"
	"time"
)

// CreateWorldDropItemRequest 创建世界掉落物品
type CreateWorldDropItemRequest struct {
	ItemID         string          `json:"item_id" validate:"required,uuid"`
	DropRate       *float64        `json:"drop_rate,omitempty" validate:"omitempty,gt=0,lte=1"`
	DropWeight     *int            `json:"drop_weight,omitempty" validate:"omitempty,min=1"`
	MinQuantity    int             `json:"min_quantity" validate:"required,min=1"`
	MaxQuantity    int             `json:"max_quantity" validate:"required,min=1"`
	MinLevel       *int            `json:"min_level,omitempty" validate:"omitempty,min=1"`
	MaxLevel       *int            `json:"max_level,omitempty" validate:"omitempty,min=1"`
	GuaranteedDrop bool            `json:"guaranteed_drop"`
	Metadata       json.RawMessage `json:"metadata,omitempty"`
}

// UpdateWorldDropItemRequest 更新世界掉落物品
type UpdateWorldDropItemRequest struct {
	ItemID         *string         `json:"item_id" validate:"omitempty,uuid"`
	DropRate       *float64        `json:"drop_rate,omitempty" validate:"omitempty,gt=0,lte=1"`
	DropWeight     *int            `json:"drop_weight,omitempty" validate:"omitempty,min=1"`
	MinQuantity    *int            `json:"min_quantity" validate:"omitempty,min=1"`
	MaxQuantity    *int            `json:"max_quantity" validate:"omitempty,min=1"`
	MinLevel       *int            `json:"min_level,omitempty" validate:"omitempty,min=1"`
	MaxLevel       *int            `json:"max_level,omitempty" validate:"omitempty,min=1"`
	GuaranteedDrop *bool           `json:"guaranteed_drop,omitempty"`
	Metadata       json.RawMessage `json:"metadata,omitempty"`
}

// WorldDropItemResponse 返回详情
type WorldDropItemResponse struct {
	ID                string          `json:"id"`
	WorldDropConfigID string          `json:"world_drop_config_id"`
	ItemID            string          `json:"item_id"`
	ItemCode          string          `json:"item_code"`
	ItemName          string          `json:"item_name"`
	ItemQuality       string          `json:"item_quality"`
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

type WorldDropItemListResponse struct {
	Items    []WorldDropItemResponse `json:"items"`
	Total    int64                   `json:"total"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"page_size"`
}
