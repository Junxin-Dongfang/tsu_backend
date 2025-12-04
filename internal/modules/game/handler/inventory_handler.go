package handler

import (
	"database/sql"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/modules/game/service"
	"tsu-self/internal/pkg/ctxkey"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/xerrors"
)

// InventoryHandler 背包管理处理器
type InventoryHandler struct {
	db           *sql.DB
	inventorySvc *service.InventoryService
	respWriter   response.Writer
}

// NewInventoryHandler 创建背包管理处理器
func NewInventoryHandler(db *sql.DB, respWriter response.Writer) *InventoryHandler {
	return &InventoryHandler{
		db:           db,
		inventorySvc: service.NewInventoryService(db),
		respWriter:   respWriter,
	}
}

// ==================== HTTP Request/Response Models ====================

// GetInventoryRequest 查询背包请求
type GetInventoryRequest struct {
	OwnerID      string  `query:"owner_id" validate:"required" example:"123e4567-e89b-12d3-a456-426614174000"`           // 所有者ID（必填）
	ItemLocation string  `query:"item_location" validate:"required,oneof=backpack warehouse storage" example:"backpack"` // 物品位置（必填：backpack/warehouse/storage）
	ItemType     *string `query:"item_type" example:"equipment"`                                                         // 物品类型（可选）
	ItemQuality  *string `query:"item_quality" example:"epic"`                                                           // 物品品质（可选）
	Page         int     `query:"page" example:"1"`                                                                      // 页码（默认1）
	PageSize     int     `query:"page_size" example:"20"`                                                                // 每页数量（默认20）
}

// GetInventoryResponse 查询背包响应
type GetInventoryResponse struct {
	Items      []*PlayerItemInfo `json:"items"`       // 物品列表
	TotalCount int64             `json:"total_count"` // 总数量
	Page       int               `json:"page"`        // 当前页码
	PageSize   int               `json:"page_size"`   // 每页数量
}

// MoveItemRequest 移动物品请求
type MoveItemRequest struct {
	ItemInstanceID string `json:"item_instance_id" validate:"required" example:"660e8400-e29b-41d4-a716-446655440001"` // 物品实例ID（必填）
	FromLocation   string `json:"from_location" validate:"required" example:"backpack"`                                // 源位置（必填）
	ToLocation     string `json:"to_location" validate:"required" example:"warehouse"`                                 // 目标位置（必填）
}

// MoveItemResponse 移动物品响应
type MoveItemResponse struct {
	Success bool   `json:"success" example:"true"`   // 是否成功
	Message string `json:"message" example:"物品移动成功"` // 消息
}

// DiscardItemRequest 丢弃物品请求
type DiscardItemRequest struct {
	ItemInstanceID string `json:"item_instance_id" validate:"required" example:"660e8400-e29b-41d4-a716-446655440001"` // 物品实例ID（必填）
	Quantity       int    `json:"quantity" validate:"required,min=1" example:"5"`                                      // 丢弃数量（必填，<= 当前堆叠数；等于堆叠数则删除实例）
}

// DiscardItemResponse 丢弃物品响应
type DiscardItemResponse struct {
	Success bool   `json:"success" example:"true"`   // 是否成功
	Message string `json:"message" example:"物品丢弃成功"` // 消息
}

// SortInventoryRequest 整理背包请求
type SortInventoryRequest struct {
	OwnerID      string `json:"owner_id" validate:"required" example:"123e4567-e89b-12d3-a456-426614174000"` // 所有者ID（必填）
	ItemLocation string `json:"item_location" validate:"required" example:"backpack"`                        // 物品位置（必填）
}

// SortInventoryResponse 整理背包响应
type SortInventoryResponse struct {
	Success bool              `json:"success" example:"true"`   // 是否成功
	Message string            `json:"message" example:"背包整理成功"` // 消息
	Items   []*PlayerItemInfo `json:"items"`                    // 整理后的物品列表
}

// ==================== HTTP Handlers ====================

// GetInventory 查询背包/仓库
// @Summary 查询背包/仓库
// @Description 查询玩家的背包或仓库物品列表,支持分页和筛选。`owner_id` 为用户ID（非英雄ID）。只能查询自己的背包，若 owner_id 与登录用户不一致返回 403。
// @Tags Inventory
// @Accept json
// @Produce json
// @Param owner_id query string true "所有者ID（用户ID）"
// @Param item_location query string true "物品位置(backpack/warehouse/storage)"
// @Param item_type query string false "物品类型"
// @Param item_quality query string false "物品品质"
// @Param page query int false "页码(默认1)"
// @Param page_size query int false "每页数量(默认20)"
// @Success 200 {object} response.Response{data=GetInventoryResponse} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 403 {object} response.Response "只能查询自己的背包"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /game/inventory [get]
func (h *InventoryHandler) GetInventory(c echo.Context) error {
	// 1. 解析请求
	var req GetInventoryRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoError(c, h.respWriter, xerrors.New(xerrors.CodeInvalidParams, "请求参数格式错误"))
	}

	// 2. 验证请求
	if err := c.Validate(&req); err != nil {
		return response.EchoError(c, h.respWriter, xerrors.New(xerrors.CodeInvalidParams, "请求参数验证失败"))
	}

	// 3. 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	// 4. 权限校验：只能查询自己的背包
	userID, _ := c.Get("user_id").(string)
	if userID == "" || userID != req.OwnerID {
		return response.EchoError(c, h.respWriter, xerrors.New(xerrors.CodePermissionDenied, "只能查询自己的背包"))
	}

	// 5. 调用服务
	svcReq := &service.GetInventoryRequest{
		OwnerID:      req.OwnerID,
		ItemLocation: req.ItemLocation,
		ItemType:     req.ItemType,
		ItemQuality:  req.ItemQuality,
		Page:         req.Page,
		PageSize:     req.PageSize,
	}

	svcResp, err := h.inventorySvc.GetInventory(c.Request().Context(), svcReq)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 6. 构建响应
	resp := &GetInventoryResponse{
		Items:      make([]*PlayerItemInfo, 0, len(svcResp.Items)),
		TotalCount: svcResp.TotalCount,
		Page:       svcResp.Page,
		PageSize:   svcResp.PageSize,
	}

	for _, item := range svcResp.Items {
		resp.Items = append(resp.Items, convertServiceItemToHandlerItem(item))
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// MoveItem 移动物品
// @Summary 移动物品
// @Description 在背包和仓库之间移动物品；只能操作当前登录英雄拥有的实例，越权请求返回 "只能操作自己的物品"。
// @Tags Inventory
// @Accept json
// @Produce json
// @Param request body MoveItemRequest true "移动物品请求"
// @Success 200 {object} response.Response{data=MoveItemResponse} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 403 {object} response.Response "只能操作自己的物品"
// @Failure 404 {object} response.Response "资源不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /game/inventory/move [post]
func (h *InventoryHandler) MoveItem(c echo.Context) error {
	// 1. 解析请求
	var req MoveItemRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoError(c, h.respWriter, xerrors.New(xerrors.CodeInvalidParams, "请求参数格式错误"))
	}

	// 2. 验证请求
	if err := c.Validate(&req); err != nil {
		return response.EchoError(c, h.respWriter, xerrors.New(xerrors.CodeInvalidParams, "请求参数验证失败"))
	}

	// 3. 调用服务
	ownerID := c.Get("hero_id")
	if ownerIDStr, ok := ownerID.(string); ok && ownerIDStr != "" {
		// 强制使用当前英雄作为 owner
	}

	svcReq := &service.MoveItemRequest{
		ItemInstanceID: req.ItemInstanceID,
		FromLocation:   req.FromLocation,
		ToLocation:     req.ToLocation,
	}

	ctx := c.Request().Context()
	if userID, ok := c.Get("user_id").(string); ok && userID != "" {
		ctx = ctxkey.WithValue(ctx, ctxkey.UserID, userID)
	} else {
		return response.EchoError(c, h.respWriter, xerrors.New(xerrors.CodePermissionDenied, "未登录或会话已过期"))
	}
	if heroID, ok := c.Get(string(ctxkey.HeroID)).(string); ok && heroID != "" {
		ctx = ctxkey.WithValue(ctx, ctxkey.HeroID, heroID)
	}

	svcResp, err := h.inventorySvc.MoveItem(ctx, svcReq)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 4. 构建响应
	resp := &MoveItemResponse{
		Success: svcResp.Success,
		Message: svcResp.Message,
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// DiscardItem 丢弃物品
// @Summary 丢弃物品
// @Description 丢弃指定的物品实例（软删除）。支持按数量丢弃：数量小于堆叠数则减少堆叠，等于堆叠数则删除该实例。
// @Tags Inventory
// @Accept json
// @Produce json
// @Param request body DiscardItemRequest true "丢弃物品请求"
// @Success 200 {object} response.Response{data=DiscardItemResponse} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "资源不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /game/inventory/discard [post]
func (h *InventoryHandler) DiscardItem(c echo.Context) error {
	// 1. 解析请求
	var req DiscardItemRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoError(c, h.respWriter, xerrors.New(xerrors.CodeInvalidParams, "请求参数格式错误"))
	}

	// 2. 验证请求
	if err := c.Validate(&req); err != nil {
		return response.EchoError(c, h.respWriter, xerrors.New(xerrors.CodeInvalidParams, "请求参数验证失败"))
	}

	// 3. 调用服务
	svcReq := &service.DiscardItemRequest{
		ItemInstanceID: req.ItemInstanceID,
		Quantity:       req.Quantity,
	}

	svcResp, err := h.inventorySvc.DiscardItem(c.Request().Context(), svcReq)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 4. 构建响应
	resp := &DiscardItemResponse{
		Success: svcResp.Success,
		Message: svcResp.Message,
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// SortInventory 整理背包
// @Summary 整理背包
// @Description 按类型、品质、等级排序整理背包
// @Tags Inventory
// @Accept json
// @Produce json
// @Param request body SortInventoryRequest true "整理背包请求"
// @Success 200 {object} response.Response{data=SortInventoryResponse} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /game/inventory/sort [post]
func (h *InventoryHandler) SortInventory(c echo.Context) error {
	// 1. 解析请求
	var req SortInventoryRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoError(c, h.respWriter, xerrors.New(xerrors.CodeInvalidParams, "请求参数格式错误"))
	}

	// 2. 验证请求
	if err := c.Validate(&req); err != nil {
		return response.EchoError(c, h.respWriter, xerrors.New(xerrors.CodeInvalidParams, "请求参数验证失败"))
	}

	// 3. 调用服务
	svcReq := &service.SortInventoryRequest{
		OwnerID:      req.OwnerID,
		ItemLocation: req.ItemLocation,
	}

	svcResp, err := h.inventorySvc.SortInventory(c.Request().Context(), svcReq)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 4. 构建响应
	resp := &SortInventoryResponse{
		Success: svcResp.Success,
		Message: svcResp.Message,
		Items:   make([]*PlayerItemInfo, 0, len(svcResp.Items)),
	}

	for _, item := range svcResp.Items {
		resp.Items = append(resp.Items, convertServiceItemToHandlerItem(item))
	}

	return response.EchoOK(c, h.respWriter, resp)
}
