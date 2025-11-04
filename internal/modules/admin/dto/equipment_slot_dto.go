package dto

import "time"

// CreateSlotRequest 创建槽位配置请求
type CreateSlotRequest struct {
	SlotCode     string  `json:"slot_code" validate:"required,min=1,max=50"`
	SlotName     string  `json:"slot_name" validate:"required,min=1,max=100"`
	SlotType     string  `json:"slot_type" validate:"required,oneof=weapon armor accessory special"`
	DisplayOrder int     `json:"display_order" validate:"required,min=0"`
	Icon         *string `json:"icon,omitempty" validate:"omitempty,max=255"`
	Description  *string `json:"description,omitempty"`
}

// UpdateSlotRequest 更新槽位配置请求
type UpdateSlotRequest struct {
	SlotName     *string `json:"slot_name,omitempty" validate:"omitempty,min=1,max=100"`
	DisplayOrder *int    `json:"display_order,omitempty" validate:"omitempty,min=0"`
	Icon         *string `json:"icon,omitempty" validate:"omitempty,max=255"`
	Description  *string `json:"description,omitempty"`
	IsActive     *bool   `json:"is_active,omitempty"`
}

// SlotConfigResponse 槽位配置响应
type SlotConfigResponse struct {
	ID           string    `json:"id"`
	SlotCode     string    `json:"slot_code"`
	SlotName     string    `json:"slot_name"`
	SlotType     string    `json:"slot_type"`
	DisplayOrder int       `json:"display_order"`
	Icon         *string   `json:"icon,omitempty"`
	Description  *string   `json:"description,omitempty"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SlotListResponse 槽位列表响应
type SlotListResponse struct {
	Slots    []SlotConfigResponse `json:"slots"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}
