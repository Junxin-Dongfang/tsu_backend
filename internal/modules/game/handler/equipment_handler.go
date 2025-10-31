package handler

import (
	"database/sql"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/modules/game/service"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/xerrors"
)

// EquipmentHandler 装备穿戴处理器
type EquipmentHandler struct {
	db              *sql.DB
	equipmentSvc    *service.EquipmentService
	respWriter      response.Writer
}

// NewEquipmentHandler 创建装备穿戴处理器
func NewEquipmentHandler(db *sql.DB, respWriter response.Writer) *EquipmentHandler {
	return &EquipmentHandler{
		db:           db,
		equipmentSvc: service.NewEquipmentService(db),
		respWriter:   respWriter,
	}
}

// ==================== HTTP Request/Response Models ====================

// EquipItemRequest 穿戴装备请求
type EquipItemRequest struct {
	HeroID         string  `json:"hero_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`         // 英雄ID（必填）
	ItemInstanceID string  `json:"item_instance_id" validate:"required" example:"660e8400-e29b-41d4-a716-446655440001"` // 物品实例ID（必填）
	SlotType       *string `json:"slot_type,omitempty" example:"weapon"`                                               // 槽位类型（可选，不指定则自动选择）
}

// EquipItemResponse 穿戴装备响应
type EquipItemResponse struct {
	Success        bool                  `json:"success" example:"true"`                                  // 是否成功
	Message        string                `json:"message" example:"装备穿戴成功"`                                // 消息
	EquippedSlot   *EquipmentSlotInfo    `json:"equipped_slot,omitempty"`                                 // 装备的槽位信息
	UnequippedItem *PlayerItemInfo       `json:"unequipped_item,omitempty"`                               // 被替换的装备（如果有）
}

// UnequipItemRequest 卸下装备请求
type UnequipItemRequest struct {
	HeroID string `json:"hero_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"` // 英雄ID（必填）
	SlotID string `json:"slot_id" validate:"required" example:"770e8400-e29b-41d4-a716-446655440002"` // 槽位ID（必填）
}

// UnequipItemResponse 卸下装备响应
type UnequipItemResponse struct {
	Success        bool            `json:"success" example:"true"`     // 是否成功
	Message        string          `json:"message" example:"装备卸下成功"`   // 消息
	UnequippedItem *PlayerItemInfo `json:"unequipped_item,omitempty"`  // 卸下的装备信息
}

// EquipmentSlotInfo 装备槽位信息
type EquipmentSlotInfo struct {
	ID              string  `json:"id" example:"770e8400-e29b-41d4-a716-446655440002"`         // 槽位ID
	HeroID          string  `json:"hero_id" example:"550e8400-e29b-41d4-a716-446655440000"`    // 英雄ID
	SlotType        string  `json:"slot_type" example:"weapon"`                                // 槽位类型
	SlotIndex       int     `json:"slot_index" example:"0"`                                    // 槽位索引
	IsUnlocked      bool    `json:"is_unlocked" example:"true"`                                // 是否已解锁
	UnlockLevel     *int    `json:"unlock_level,omitempty" example:"1"`                        // 解锁等级
	EquippedItemID  *string `json:"equipped_item_id,omitempty" example:"660e8400-e29b-41d4-a716-446655440001"` // 已装备物品ID
	AddedByItemID   *string `json:"added_by_item_id,omitempty"`                                // 由哪个装备增加的槽位
}

// PlayerItemInfo 玩家物品信息
type PlayerItemInfo struct {
	ID                    string  `json:"id" example:"660e8400-e29b-41d4-a716-446655440001"`         // 物品实例ID
	ItemID                string  `json:"item_id" example:"item-sword-001"`                          // 物品配置ID
	ItemName              string  `json:"item_name" example:"烈焰之剑"`                                  // 物品名称
	ItemType              string  `json:"item_type" example:"equipment"`                             // 物品类型
	ItemQuality           string  `json:"item_quality" example:"epic"`                               // 物品品质
	OwnerID               string  `json:"owner_id" example:"123e4567-e89b-12d3-a456-426614174000"`   // 所有者ID
	ItemLocation          string  `json:"item_location" example:"backpack"`                          // 物品位置
	CurrentDurability     *int    `json:"current_durability,omitempty" example:"100"`                // 当前耐久度
	MaxDurability         *int    `json:"max_durability,omitempty" example:"100"`                    // 最大耐久度
	EnhancementLevel      *int    `json:"enhancement_level,omitempty" example:"5"`                   // 强化等级
	StackCount            *int    `json:"stack_count,omitempty" example:"1"`                         // 堆叠数量
}

// GetEquippedItemsResponse 查询已装备物品响应
type GetEquippedItemsResponse struct {
	Items []*PlayerItemInfo `json:"items"` // 已装备物品列表
}

// GetEquipmentSlotsResponse 查询装备槽位响应
type GetEquipmentSlotsResponse struct {
	Slots []*EquipmentSlotInfo `json:"slots"` // 装备槽位列表
}

// GetEquipmentBonusResponse 查询装备属性加成响应
type GetEquipmentBonusResponse struct {
	Bonuses map[string]*AttributeBonusInfo `json:"bonuses"` // 属性加成映射
}

// AttributeBonusInfo 属性加成信息
type AttributeBonusInfo struct {
	AttributeCode string  `json:"attribute_code" example:"STR"`  // 属性代码
	FlatBonus     float64 `json:"flat_bonus" example:"10.5"`     // 固定加值
	PercentBonus  float64 `json:"percent_bonus" example:"0.15"`  // 百分比加成(0.15表示15%)
}

// ==================== HTTP Handlers ====================

// EquipItem 穿戴装备
// @Summary 穿戴装备
// @Description 为英雄穿戴装备到指定槽位
// @Tags Equipment
// @Accept json
// @Produce json
// @Param request body EquipItemRequest true "穿戴装备请求"
// @Success 200 {object} response.Response{data=EquipItemResponse} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "资源不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /game/equipment/equip [post]
func (h *EquipmentHandler) EquipItem(c echo.Context) error {
	// 1. 解析请求
	var req EquipItemRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求参数格式错误")
	}

	// 2. 验证请求
	if err := c.Validate(&req); err != nil {
		return response.EchoValidationError(c, h.respWriter, err)
	}

	// 3. 调用服务
	svcReq := &service.EquipItemRequest{
		HeroID:         req.HeroID,
		ItemInstanceID: req.ItemInstanceID,
	}
	if req.SlotType != nil {
		svcReq.SlotType = *req.SlotType
	}

	svcResp, err := h.equipmentSvc.EquipItem(c.Request().Context(), svcReq)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 4. 构建响应
	resp := &EquipItemResponse{
		Success: svcResp.Success,
		Message: svcResp.Message,
	}

	if svcResp.EquippedSlot != nil {
		resp.EquippedSlot = convertServiceSlotToHandlerSlot(svcResp.EquippedSlot)
	}

	if svcResp.UnequippedItem != nil {
		resp.UnequippedItem = convertServiceItemToHandlerItem(svcResp.UnequippedItem)
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// UnequipItem 卸下装备
// @Summary 卸下装备
// @Description 从指定槽位卸下装备
// @Tags Equipment
// @Accept json
// @Produce json
// @Param request body UnequipItemRequest true "卸下装备请求"
// @Success 200 {object} response.Response{data=UnequipItemResponse} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "资源不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /game/equipment/unequip [post]
func (h *EquipmentHandler) UnequipItem(c echo.Context) error {
	// 1. 解析请求
	var req UnequipItemRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoError(c, h.respWriter, xerrors.New(xerrors.CodeInvalidParams, "请求参数格式错误"))
	}

	// 2. 验证请求
	if err := c.Validate(&req); err != nil {
		return response.EchoError(c, h.respWriter, xerrors.New(xerrors.CodeInvalidParams, "请求参数验证失败"))
	}

	// 3. 调用服务
	svcReq := &service.UnequipItemRequest{
		HeroID: req.HeroID,
		SlotID: req.SlotID,
	}

	svcResp, err := h.equipmentSvc.UnequipItem(c.Request().Context(), svcReq)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 4. 构建响应
	resp := &UnequipItemResponse{
		Success: svcResp.Success,
		Message: svcResp.Message,
	}

	if svcResp.UnequippedItem != nil {
		resp.UnequippedItem = convertServiceItemToHandlerItem(svcResp.UnequippedItem)
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// GetEquippedItems 查询已装备物品
// @Summary 查询已装备物品
// @Description 查询英雄当前已装备的所有物品
// @Tags Equipment
// @Accept json
// @Produce json
// @Param hero_id path string true "英雄ID"
// @Success 200 {object} response.Response{data=GetEquippedItemsResponse} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "资源不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /game/equipment/equipped/{hero_id} [get]
func (h *EquipmentHandler) GetEquippedItems(c echo.Context) error {
	// 1. 获取参数
	heroID := c.Param("hero_id")
	if heroID == "" {
		return response.EchoError(c, h.respWriter, xerrors.New(xerrors.CodeInvalidParams, "英雄ID不能为空"))
	}

	// 2. 调用服务
	items, err := h.equipmentSvc.GetEquippedItems(c.Request().Context(), heroID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 3. 构建响应
	resp := &GetEquippedItemsResponse{
		Items: make([]*PlayerItemInfo, 0, len(items)),
	}

	for _, item := range items {
		resp.Items = append(resp.Items, convertServiceItemToHandlerItem(item))
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// GetEquipmentSlots 查询装备槽位
// @Summary 查询装备槽位
// @Description 查询英雄的所有装备槽位
// @Tags Equipment
// @Accept json
// @Produce json
// @Param hero_id path string true "英雄ID"
// @Success 200 {object} response.Response{data=GetEquipmentSlotsResponse} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "资源不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /game/equipment/slots/{hero_id} [get]
func (h *EquipmentHandler) GetEquipmentSlots(c echo.Context) error {
	// 1. 获取参数
	heroID := c.Param("hero_id")
	if heroID == "" {
		return response.EchoError(c, h.respWriter, xerrors.New(xerrors.CodeInvalidParams, "英雄ID不能为空"))
	}

	// 2. 调用服务
	slots, err := h.equipmentSvc.GetEquipmentSlots(c.Request().Context(), heroID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 3. 构建响应
	resp := &GetEquipmentSlotsResponse{
		Slots: make([]*EquipmentSlotInfo, 0, len(slots)),
	}

	for _, slot := range slots {
		resp.Slots = append(resp.Slots, convertServiceSlotToHandlerSlot(slot))
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// GetEquipmentBonus 查询装备属性加成
// @Summary 查询装备属性加成
// @Description 查询英雄当前装备提供的属性加成
// @Tags Equipment
// @Accept json
// @Produce json
// @Param hero_id path string true "英雄ID"
// @Success 200 {object} response.Response{data=GetEquipmentBonusResponse} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "资源不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /game/equipment/bonus/{hero_id} [get]
func (h *EquipmentHandler) GetEquipmentBonus(c echo.Context) error {
	// 1. 获取参数
	heroID := c.Param("hero_id")
	if heroID == "" {
		return response.EchoError(c, h.respWriter, xerrors.New(xerrors.CodeInvalidParams, "英雄ID不能为空"))
	}

	// 2. 调用服务
	bonuses, err := h.equipmentSvc.GetEquipmentAttributeBonus(c.Request().Context(), heroID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 3. 构建响应
	resp := &GetEquipmentBonusResponse{
		Bonuses: make(map[string]*AttributeBonusInfo),
	}

	for attrCode, bonus := range bonuses {
		resp.Bonuses[attrCode] = &AttributeBonusInfo{
			AttributeCode: bonus.AttributeCode,
			FlatBonus:     bonus.FlatBonus,
			PercentBonus:  bonus.PercentBonus,
		}
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// ==================== Helper Functions ====================

// convertServiceSlotToHandlerSlot 转换Service层DTO为Handler层DTO
func convertServiceSlotToHandlerSlot(dto *service.EquipmentSlotDTO) *EquipmentSlotInfo {
	if dto == nil {
		return nil
	}

	return &EquipmentSlotInfo{
		ID:              dto.ID,
		HeroID:          dto.HeroID,
		SlotType:        dto.SlotType,
		SlotIndex:       dto.SlotIndex,
		IsUnlocked:      dto.IsUnlocked,
		UnlockLevel:     dto.UnlockLevel,
		EquippedItemID:  dto.EquippedItemID,
		AddedByItemID:   dto.AddedByItemID,
	}
}

// convertServiceItemToHandlerItem 转换Service层DTO为Handler层DTO
func convertServiceItemToHandlerItem(dto *service.PlayerItemDTO) *PlayerItemInfo {
	if dto == nil {
		return nil
	}

	return &PlayerItemInfo{
		ID:                dto.ID,
		ItemID:            dto.ItemID,
		OwnerID:           dto.OwnerID,
		ItemLocation:      dto.ItemLocation,
		CurrentDurability: dto.CurrentDurability,
		MaxDurability:     dto.MaxDurability,
		EnhancementLevel:  dto.EnhancementLevel,
		StackCount:        dto.StackCount,
		// TODO: ItemName, ItemType, ItemQuality 需要从 item config 获取
		// 可以在 Service 层扩展 PlayerItemDTO 包含这些信息
	}
}

