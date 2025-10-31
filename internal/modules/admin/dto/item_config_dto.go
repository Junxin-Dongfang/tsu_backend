package dto

import (
	"encoding/json"
	"time"
)

// CreateItemRequest 创建物品配置请求
type CreateItemRequest struct {
	// 基础信息
	ItemCode    string `json:"item_code" validate:"required,min=1,max=100"`
	ItemName    string `json:"item_name" validate:"required,min=1,max=200"`
	ItemType    string `json:"item_type" validate:"required,oneof=equipment consumable gem repair_material enhancement_material quest_item material other"`
	ItemQuality string `json:"item_quality" validate:"required,oneof=poor normal fine excellent superb master epic legendary mythic"`
	ItemLevel   int16  `json:"item_level" validate:"required,min=1,max=100"`

	// 描述信息
	Description string `json:"description" validate:"max=1000"`
	IconURL     string `json:"icon_url" validate:"omitempty,url"`

	// 装备相关
	EquipSlot        *string  `json:"equip_slot,omitempty" validate:"omitempty,oneof=head eyes ears neck cloak chest belt shoulder wrist gloves legs feet ring badge coat pocket summon_mount mainhand offhand twohand special"`
	RequiredClassIDs []string `json:"required_class_ids,omitempty" validate:"omitempty,dive,uuid4"` // 职业限制（空数组=通用装备）
	RequiredLevel    *int16   `json:"required_level,omitempty" validate:"omitempty,min=1,max=100"`
	MaterialType      *string `json:"material_type,omitempty"`
	MaxDurability     *int    `json:"max_durability,omitempty" validate:"omitempty,min=1"`
	UniquenessType    *string `json:"uniqueness_type,omitempty" validate:"omitempty,oneof=none unique unique_equipped"`

	// 效果相关
	// 局外效果 - 直接影响英雄属性的效果
	// 格式: [{"Data_type":"Status","Data_ID":"MAX_HP","Bouns_type":"bonus","Bouns_Number":"5"}]
	// Data_type: 固定为"Status"
	// Data_ID: 属性ID,如"STR"(力量),"ATK"(攻击力),"MAX_HP"(最大生命值),"CRIT_RATE"(暴击率)等
	// Bouns_type: 加成类型,"bonus"(固定值加成)或"percent"(百分比加成)
	// Bouns_Number: 加成数值,字符串格式,如"10"表示+10或+10%
	// 示例: [{"Data_type":"Status","Data_ID":"STR","Bouns_type":"bonus","Bouns_Number":"20"},{"Data_type":"Status","Data_ID":"ATK","Bouns_type":"bonus","Bouns_Number":"50"}]
	OutOfCombatEffects json.RawMessage `json:"out_of_combat_effects,omitempty" swaggertype:"string"`

	// 局内效果 - 战斗时触发的效果
	// 格式: [{"Data_type":"Buff","Data_ID":"buff_id","Trigger_type":"on_hit","Trigger_chance":"0.3"}]
	// Data_type: 效果类型,如"Buff","Damage","Heal"等
	// Data_ID: 效果ID或Buff ID
	// Trigger_type: 触发类型,如"on_hit"(命中时),"on_attack"(攻击时),"on_damaged"(受伤时)等
	// Trigger_chance: 触发概率,字符串格式,如"0.3"表示30%概率
	// 示例: [{"Data_type":"Buff","Data_ID":"fire_buff_001","Trigger_type":"on_hit","Trigger_chance":"0.3"}]
	InCombatEffects json.RawMessage `json:"in_combat_effects,omitempty" swaggertype:"string"`

	// 使用效果 - 消耗品使用时的效果
	// 格式: [{"Effect":"RESTORE_HP","params":{"amount":"100"}}]
	// Effect: 效果类型,如"RESTORE_HP"(恢复生命),"RESTORE_MP"(恢复魔法),"APPLY_BUFF"(施加Buff)等
	// params: 效果参数,根据不同的Effect类型有不同的参数
	// 示例: [{"Effect":"RESTORE_HP","params":{"amount":"100","percent":"false"}}]
	UseEffects json.RawMessage `json:"use_effects,omitempty" swaggertype:"string"`

	// 提供的技能 - 装备后获得的技能
	// 格式: [{"skill_id":"skill_uuid","skill_level":1}]
	// skill_id: 技能ID(UUID格式)
	// skill_level: 技能等级
	// 示例: [{"skill_id":"12345678-1234-4234-8234-123456789012","skill_level":1}]
	ProvidedSkills json.RawMessage `json:"provided_skills,omitempty" swaggertype:"string"`

	// 强化相关
	SocketType            *string `json:"socket_type,omitempty" validate:"omitempty,oneof=red blue yellow green prismatic"`
	SocketCount           *int16  `json:"socket_count,omitempty" validate:"omitempty,min=0,max=5"`
	EnhancementMaterialID *string `json:"enhancement_material_id,omitempty" validate:"omitempty,uuid4"`
	EnhancementCostGold   *int    `json:"enhancement_cost_gold,omitempty" validate:"omitempty,min=0"`

	// 宝石相关
	GemColor *string `json:"gem_color,omitempty" validate:"omitempty,oneof=red blue yellow green prismatic"`
	GemSize  *string `json:"gem_size,omitempty" validate:"omitempty,oneof=small medium large"`

	// 修理相关
	RepairDurabilityAmount  *int    `json:"repair_durability_amount,omitempty" validate:"omitempty,min=1"`
	RepairApplicableQuality *string `json:"repair_applicable_quality,omitempty"`
	RepairMaterialType      *string `json:"repair_material_type,omitempty"`

	// 堆叠和交易
	MaxStackSize *int16 `json:"max_stack_size,omitempty" validate:"omitempty,min=1,max=9999"`
	BaseValue    *int   `json:"base_value,omitempty" validate:"omitempty,min=0"`
	IsTradable   *bool  `json:"is_tradable,omitempty"`
	IsDroppable  *bool  `json:"is_droppable,omitempty"`

	// 标签
	TagIDs []string `json:"tag_ids,omitempty"`
}

// UpdateItemRequest 更新物品配置请求
type UpdateItemRequest struct {
	// 基础信息
	ItemCode    *string `json:"item_code,omitempty" validate:"omitempty,min=1,max=100"`
	ItemName    *string `json:"item_name,omitempty" validate:"omitempty,min=1,max=200"`
	ItemType    *string `json:"item_type,omitempty" validate:"omitempty,oneof=equipment consumable gem repair_material enhancement_material quest_item material other"`
	ItemQuality *string `json:"item_quality,omitempty" validate:"omitempty,oneof=poor normal fine excellent superb master epic legendary mythic"`
	ItemLevel   *int16  `json:"item_level,omitempty" validate:"omitempty,min=1,max=100"`

	// 描述信息
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	IconURL     *string `json:"icon_url,omitempty" validate:"omitempty,url"`

	// 装备相关
	EquipSlot        *string   `json:"equip_slot,omitempty" validate:"omitempty,oneof=head eyes ears neck cloak chest belt shoulder wrist gloves legs feet ring badge coat pocket summon_mount mainhand offhand twohand special"`
	RequiredClassIDs *[]string `json:"required_class_ids,omitempty" validate:"omitempty,dive,uuid4"` // nil=不更新，空数组=清除限制，非空=替换
	RequiredLevel    *int16    `json:"required_level,omitempty" validate:"omitempty,min=1,max=100"`
	MaterialType      *string `json:"material_type,omitempty"`
	MaxDurability     *int    `json:"max_durability,omitempty" validate:"omitempty,min=1"`
	UniquenessType    *string `json:"uniqueness_type,omitempty" validate:"omitempty,oneof=none unique unique_equipped"`

	// 效果相关
	// 局外效果 - 直接影响英雄属性的效果
	// 格式: [{"Data_type":"Status","Data_ID":"MAX_HP","Bouns_type":"bonus","Bouns_Number":"5"}]
	// 详细说明见CreateItemRequest
	OutOfCombatEffects json.RawMessage `json:"out_of_combat_effects,omitempty" swaggertype:"string"`

	// 局内效果 - 战斗时触发的效果
	// 格式: [{"Data_type":"Buff","Data_ID":"buff_id","Trigger_type":"on_hit","Trigger_chance":"0.3"}]
	// 详细说明见CreateItemRequest
	InCombatEffects json.RawMessage `json:"in_combat_effects,omitempty" swaggertype:"string"`

	// 使用效果 - 消耗品使用时的效果
	// 格式: [{"Effect":"RESTORE_HP","params":{"amount":"100"}}]
	// 详细说明见CreateItemRequest
	UseEffects json.RawMessage `json:"use_effects,omitempty" swaggertype:"string"`

	// 提供的技能 - 装备后获得的技能
	// 格式: [{"skill_id":"skill_uuid","skill_level":1}]
	// 详细说明见CreateItemRequest
	ProvidedSkills json.RawMessage `json:"provided_skills,omitempty" swaggertype:"string"`

	// 强化相关
	SocketType            *string `json:"socket_type,omitempty" validate:"omitempty,oneof=red blue yellow green prismatic"`
	SocketCount           *int16  `json:"socket_count,omitempty" validate:"omitempty,min=0,max=5"`
	EnhancementMaterialID *string `json:"enhancement_material_id,omitempty" validate:"omitempty,uuid4"`
	EnhancementCostGold   *int    `json:"enhancement_cost_gold,omitempty" validate:"omitempty,min=0"`

	// 宝石相关
	GemColor *string `json:"gem_color,omitempty" validate:"omitempty,oneof=red blue yellow green prismatic"`
	GemSize  *string `json:"gem_size,omitempty" validate:"omitempty,oneof=small medium large"`

	// 修理相关
	RepairDurabilityAmount  *int    `json:"repair_durability_amount,omitempty" validate:"omitempty,min=1"`
	RepairApplicableQuality *string `json:"repair_applicable_quality,omitempty"`
	RepairMaterialType      *string `json:"repair_material_type,omitempty"`

	// 堆叠和交易
	MaxStackSize *int16 `json:"max_stack_size,omitempty" validate:"omitempty,min=1,max=9999"`
	BaseValue    *int   `json:"base_value,omitempty" validate:"omitempty,min=0"`
	IsTradable   *bool  `json:"is_tradable,omitempty"`
	IsDroppable  *bool  `json:"is_droppable,omitempty"`

	// 套装关联
	SetID *string `json:"set_id,omitempty" validate:"omitempty,uuid4"` // 装备套装ID，null表示不属于任何套装
}

// ItemConfigResponse 物品配置响应
type ItemConfigResponse struct {
	// 基础信息
	ID          string `json:"id"`
	ItemCode    string `json:"item_code"`
	ItemName    string `json:"item_name"`
	ItemType    string `json:"item_type"`
	ItemQuality string `json:"item_quality"`
	ItemLevel   int16  `json:"item_level"`

	// 描述信息
	Description string `json:"description,omitempty"`
	IconURL     string `json:"icon_url,omitempty"`

	// 装备相关
	EquipSlot        *string  `json:"equip_slot,omitempty"`
	RequiredClassIDs []string `json:"required_class_ids"` // 职业限制列表（空数组=通用装备）
	RequiredLevel    *int16   `json:"required_level,omitempty"`
	MaterialType      *string `json:"material_type,omitempty"`
	MaxDurability     *int    `json:"max_durability,omitempty"`
	UniquenessType    *string `json:"uniqueness_type,omitempty"`

	// 效果相关
	// 局外效果 - 直接影响英雄属性的效果
	// 格式: [{"Data_type":"Status","Data_ID":"MAX_HP","Bouns_type":"bonus","Bouns_Number":"5"}]
	// 详细说明见CreateItemRequest
	OutOfCombatEffects json.RawMessage `json:"out_of_combat_effects,omitempty" swaggertype:"string"`

	// 局内效果 - 战斗时触发的效果
	// 格式: [{"Data_type":"Buff","Data_ID":"buff_id","Trigger_type":"on_hit","Trigger_chance":"0.3"}]
	// 详细说明见CreateItemRequest
	InCombatEffects json.RawMessage `json:"in_combat_effects,omitempty" swaggertype:"string"`

	// 使用效果 - 消耗品使用时的效果
	// 格式: [{"Effect":"RESTORE_HP","params":{"amount":"100"}}]
	// 详细说明见CreateItemRequest
	UseEffects json.RawMessage `json:"use_effects,omitempty" swaggertype:"string"`

	// 提供的技能 - 装备后获得的技能
	// 格式: [{"skill_id":"skill_uuid","skill_level":1}]
	// 详细说明见CreateItemRequest
	ProvidedSkills json.RawMessage `json:"provided_skills,omitempty" swaggertype:"string"`

	// 强化相关
	SocketType            *string `json:"socket_type,omitempty"`
	SocketCount           *int16  `json:"socket_count,omitempty"`
	EnhancementMaterialID *string `json:"enhancement_material_id,omitempty"`
	EnhancementCostGold   *int    `json:"enhancement_cost_gold,omitempty"`

	// 宝石相关
	GemColor *string `json:"gem_color,omitempty"`
	GemSize  *string `json:"gem_size,omitempty"`

	// 修理相关
	RepairDurabilityAmount  *int    `json:"repair_durability_amount,omitempty"`
	RepairApplicableQuality *string `json:"repair_applicable_quality,omitempty"`
	RepairMaterialType      *string `json:"repair_material_type,omitempty"`

	// 堆叠和交易
	MaxStackSize *int16 `json:"max_stack_size,omitempty"`
	BaseValue    *int   `json:"base_value,omitempty"`
	IsTradable   bool   `json:"is_tradable"`
	IsDroppable  bool   `json:"is_droppable"`
	IsActive     bool   `json:"is_active"`

	// 标签
	Tags []TagResponse `json:"tags,omitempty"`

	// 套装信息
	SetID   *string `json:"set_id,omitempty"`   // 装备套装ID
	SetName *string `json:"set_name,omitempty"` // 装备套装名称
	SetCode *string `json:"set_code,omitempty"` // 装备套装代码

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AddItemTagsRequest 添加物品标签请求
type AddItemTagsRequest struct {
	TagIDs []string `json:"tag_ids" validate:"required,min=1"`
}

// UpdateItemTagsRequest 批量更新物品标签请求
type UpdateItemTagsRequest struct {
	TagIDs []string `json:"tag_ids" validate:"required"`
}

// ItemListResponse 物品列表响应
type ItemListResponse struct {
	Items    []ItemConfigResponse `json:"items"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

// AddItemClassesRequest 添加物品职业限制请求
type AddItemClassesRequest struct {
	ClassIDs []string `json:"class_ids" validate:"required,min=1,dive,uuid4"`
}

// UpdateItemClassesRequest 批量更新物品职业限制请求
type UpdateItemClassesRequest struct {
	ClassIDs []string `json:"class_ids" validate:"required,dive,uuid4"`
}

// BatchAssignItemsToSetRequest 批量分配物品到套装请求
type BatchAssignItemsToSetRequest struct {
	ItemIDs []string `json:"item_ids" validate:"required,min=1,max=100,dive,uuid4"`
}

// BatchAssignItemsToSetResponse 批量分配物品到套装响应
type BatchAssignItemsToSetResponse struct {
	AssignedCount int          `json:"assigned_count"`
	FailedItems   []FailedItem `json:"failed_items"`
}

// FailedItem 失败的物品信息
type FailedItem struct {
	ItemID string `json:"item_id"`
	Reason string `json:"reason"`
}

// BatchRemoveItemsFromSetRequest 批量移除物品从套装请求
type BatchRemoveItemsFromSetRequest struct {
	ItemIDs []string `json:"item_ids" validate:"required,min=1,max=100,dive,uuid4"`
}

// BatchRemoveItemsFromSetResponse 批量移除物品从套装响应
type BatchRemoveItemsFromSetResponse struct {
	RemovedCount int `json:"removed_count"`
}
