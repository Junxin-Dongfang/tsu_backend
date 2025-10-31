// Package dto 定义数据传输对象
package dto

import "time"

// CreateEquipmentSetRequest 创建套装请求
type CreateEquipmentSetRequest struct {
	SetCode     string         `json:"set_code" validate:"required,min=3,max=64"`
	SetName     string         `json:"set_name" validate:"required,min=1,max=128"`
	Description *string        `json:"description,omitempty"`
	SetEffects  []SetEffectDTO `json:"set_effects" validate:"required,min=1,dive"`
	IsActive    *bool          `json:"is_active,omitempty"`
}

// UpdateEquipmentSetRequest 更新套装请求
type UpdateEquipmentSetRequest struct {
	SetName     *string        `json:"set_name,omitempty" validate:"omitempty,min=1,max=128"`
	Description *string        `json:"description,omitempty"`
	SetEffects  []SetEffectDTO `json:"set_effects,omitempty" validate:"omitempty,min=1,dive"`
	IsActive    *bool          `json:"is_active,omitempty"`
}

// ListEquipmentSetsRequest 查询套装列表请求
type ListEquipmentSetsRequest struct {
	Page      int     `query:"page"`
	PageSize  int     `query:"page_size"`
	Keyword   *string `query:"keyword"`
	IsActive  *bool   `query:"is_active"`
	SortBy    *string `query:"sort_by"`
	SortOrder *string `query:"sort_order"`
}

// SetEffectDTO 套装效果DTO（用于验证）
type SetEffectDTO struct {
	PieceCount         int                 `json:"piece_count" validate:"required,min=1"`
	EffectDescription  string              `json:"effect_description" validate:"required"`
	OutOfCombatEffects []OutOfCombatEffect `json:"out_of_combat_effects,omitempty" validate:"omitempty,dive"`
	InCombatEffects    []InCombatEffect    `json:"in_combat_effects,omitempty" validate:"omitempty,dive"`
}

// OutOfCombatEffect 局外加成效果
type OutOfCombatEffect struct {
	DataType    string `json:"Data_type" validate:"required,eq=Status"`
	DataID      string `json:"Data_ID" validate:"required"`
	BonusType   string `json:"Bouns_type" validate:"required,oneof=bonus percent"`
	BonusNumber string `json:"Bouns_Number" validate:"required"`
}

// InCombatEffect 局内加成效果
type InCombatEffect struct {
	DataType      string  `json:"Data_type" validate:"required"`
	DataID        string  `json:"Data_ID" validate:"required"`
	TriggerType   string  `json:"Trigger_type" validate:"required"`
	TriggerChance *string `json:"Trigger_chance,omitempty"`
}

// EquipmentSetResponse 套装配置响应
type EquipmentSetResponse struct {
	ID          string              `json:"id"`
	SetCode     string              `json:"set_code"`
	SetName     string              `json:"set_name"`
	Description *string             `json:"description,omitempty"`
	SetEffects  []SetEffectResponse `json:"set_effects"`
	IsActive    bool                `json:"is_active"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// SetEffectResponse 套装效果响应
type SetEffectResponse struct {
	PieceCount         int                 `json:"piece_count"`
	EffectDescription  string              `json:"effect_description"`
	OutOfCombatEffects []OutOfCombatEffect `json:"out_of_combat_effects,omitempty"`
	InCombatEffects    []InCombatEffect    `json:"in_combat_effects,omitempty"`
}

// EquipmentSetListResponse 套装列表响应
type EquipmentSetListResponse struct {
	Sets  []EquipmentSetResponse `json:"sets"`
	Total int64                  `json:"total"`
	Page  int                    `json:"page"`
}

// SetItemResponse 套装装备响应
type SetItemResponse struct {
	ID          string  `json:"id"`
	ItemCode    string  `json:"item_code"`
	ItemName    string  `json:"item_name"`
	ItemType    string  `json:"item_type"`
	ItemQuality string  `json:"item_quality"`
	EquipSlot   *string `json:"equip_slot,omitempty"`
	SetID       *string `json:"set_id,omitempty"`
}

// SetItemListResponse 套装装备列表响应
type SetItemListResponse struct {
	Items []SetItemResponse `json:"items"`
	Total int64             `json:"total"`
	Page  int               `json:"page"`
}
