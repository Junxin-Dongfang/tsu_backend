package handler

import (
	"strconv"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/modules/game/service"
	"tsu-self/internal/pkg/response"
)

// TeamWarehouseHandler 团队仓库管理 Handler
type TeamWarehouseHandler struct {
	warehouseService *service.TeamWarehouseService
	respWriter       response.Writer
}

// NewTeamWarehouseHandler 创建团队仓库管理 Handler
func NewTeamWarehouseHandler(serviceContainer *service.ServiceContainer, respWriter response.Writer) *TeamWarehouseHandler {
	return &TeamWarehouseHandler{
		warehouseService: serviceContainer.GetTeamWarehouseService(),
		respWriter:       respWriter,
	}
}

// ==================== HTTP Request/Response Models ====================

// WarehouseResponse HTTP 仓库响应
type WarehouseResponse struct {
	ID         string `json:"id" example:"warehouse-uuid-001"`          // 仓库ID
	TeamID     string `json:"team_id" example:"team-uuid-001"`          // 团队ID
	GoldAmount int64  `json:"gold_amount" example:"10000"`              // 金币数量
	CreatedAt  string `json:"created_at" example:"2025-01-01 12:00:00"` // 创建时间
	UpdatedAt  string `json:"updated_at" example:"2025-01-01 12:00:00"` // 更新时间
}

// DistributeGoldRequest HTTP 分配金币请求
type DistributeGoldRequest struct {
	DistributorID string           `json:"distributor_id" validate:"required" example:"hero-uuid-001"` // 分配者英雄ID（必填）
	Distributions map[string]int64 `json:"distributions" validate:"required"`                          // 接收者英雄ID -> 金币数量（必填）
}

// DistributeItemsRequest HTTP 分配物品请求
type DistributeItemsRequest struct {
	DistributorID string                    `json:"distributor_id" validate:"required" example:"hero-uuid-001"` // 分配者英雄ID
	Distributions map[string]map[string]int `json:"distributions" validate:"required"`                          // 接收者英雄ID -> (物品ID -> 数量)
}

// WarehouseItemsResponse 仓库物品列表响应
type WarehouseItemsResponse struct {
	Items  []*game_runtime.TeamWarehouseItem `json:"items"`  // 仓库物品列表
	Total  int64                             `json:"total"`  // 总数
	Limit  int                               `json:"limit"`  // 每页数量
	Offset int                               `json:"offset"` // 偏移量
}

// DistributionHistoryRecord 分配记录（响应体）
type DistributionHistoryRecord struct {
	ID                string  `json:"id" example:"dist-uuid-001"`
	TeamID            string  `json:"team_id" example:"team-uuid"`
	WarehouseID       string  `json:"warehouse_id" example:"warehouse-uuid"`
	DistributorHeroID string  `json:"distributor_hero_id" example:"hero-distributor"`
	RecipientHeroID   string  `json:"recipient_hero_id" example:"hero-recipient"`
	ItemType          string  `json:"item_type" example:"item|gold"`
	ItemID            *string `json:"item_id,omitempty" example:"item-uuid"`
	Quantity          int64   `json:"quantity" example:"100"`
	DistributedAt     string  `json:"distributed_at" example:"2025-12-03T08:45:21Z"`
}

// DistributionHistoryResponse 分配历史响应
type DistributionHistoryResponse struct {
	Records []*DistributionHistoryRecord `json:"records"` // 分配记录
	Total   int64                        `json:"total"`   // 总数
	Limit   int                          `json:"limit"`   // 每页数量
	Offset  int                          `json:"offset"`  // 偏移量
}

// ==================== HTTP Handlers ====================

// GetWarehouse 查看团队仓库
// @Summary 查看团队仓库
// @Description 查看团队仓库（任何团队成员均可查看）
// @Tags 团队仓库
// @Accept json
// @Produce json
// @Param team_id path string true "团队ID"
// @Param hero_id query string true "操作者英雄ID"
// @Success 200 {object} response.Response{data=WarehouseResponse} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "您不是该团队成员"
// @Failure 403 {object} response.Response "需要管理员或队长权限"
// @Failure 404 {object} response.Response "您不是该团队成员"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/teams/{team_id}/warehouse [get]
func (h *TeamWarehouseHandler) GetWarehouse(c echo.Context) error {
	// 1. 获取路径参数
	teamID := c.Param("team_id")
	if teamID == "" {
		return response.EchoBadRequest(c, h.respWriter, "团队ID不能为空")
	}

	heroID := c.QueryParam("hero_id")
	if heroID == "" {
		return response.EchoBadRequest(c, h.respWriter, "英雄ID不能为空")
	}

	// 2. 调用 Service
	warehouse, err := h.warehouseService.GetWarehouse(c.Request().Context(), teamID, heroID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 3. 转换为 HTTP 响应
	resp := &WarehouseResponse{
		ID:         warehouse.ID,
		TeamID:     warehouse.TeamID,
		GoldAmount: warehouse.GoldAmount,
		CreatedAt:  warehouse.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  warehouse.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// DistributeGold 分配金币
// @Summary 分配金币
// @Description 分配金币给团队成员（队长和管理员可操作）
// @Tags 团队仓库
// @Accept json
// @Produce json
// @Param team_id path string true "团队ID"
// @Param request body DistributeGoldRequest true "分配金币请求"
// @Success 200 {object} response.Response "分配成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 403 {object} response.Response "您不是该团队成员"
// @Failure 403 {object} response.Response "需要管理员或队长权限"
// @Failure 404 {object} response.Response "您不是该团队成员"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/teams/{team_id}/warehouse/distribute-gold [post]
func (h *TeamWarehouseHandler) DistributeGold(c echo.Context) error {
	// 1. 获取路径参数
	teamID := c.Param("team_id")
	if teamID == "" {
		return response.EchoBadRequest(c, h.respWriter, "团队ID不能为空")
	}

	// 2. 绑定和验证 HTTP 请求
	var req DistributeGoldRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 3. 调用 Service
	distributeReq := &service.DistributeGoldRequest{
		TeamID:        teamID,
		DistributorID: req.DistributorID,
		Distributions: req.Distributions,
	}

	if err := h.warehouseService.DistributeGold(c.Request().Context(), distributeReq); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{})
}

// DistributeItems 分配物品
// @Summary 分配物品
// @Description 分配仓库物品给成员（队长和管理员）。注意：单个物品的分配数量不得超过该物品的 max_stack_size（装备类通常为1）。
// @Tags 团队仓库
// @Accept json
// @Produce json
// @Param team_id path string true "团队ID"
// @Param request body DistributeItemsRequest true "分配物品请求"
// @Success 200 {object} response.Response "分配成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "您不是该团队成员"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/teams/{team_id}/warehouse/distribute-items [post]
func (h *TeamWarehouseHandler) DistributeItems(c echo.Context) error {
	teamID := c.Param("team_id")
	if teamID == "" {
		return response.EchoBadRequest(c, h.respWriter, "团队ID不能为空")
	}

	var req DistributeItemsRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}
	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	distributeReq := &service.DistributeItemsRequest{
		TeamID:        teamID,
		DistributorID: req.DistributorID,
		Distributions: req.Distributions,
	}

	if err := h.warehouseService.DistributeItems(c.Request().Context(), distributeReq); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{})
}

// GetWarehouseItems 获取仓库物品列表
// @Summary 获取仓库物品列表
// @Description 获取仓库物品列表（任何团队成员均可查看）
// @Tags 团队仓库
// @Accept json
// @Produce json
// @Param team_id path string true "团队ID"
// @Param hero_id query string true "操作者英雄ID"
// @Param limit query int false "每页数量" default(20)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {object} response.Response{data=WarehouseItemsResponse} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 403 {object} response.Response "需要管理员或队长权限"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/teams/{team_id}/warehouse/items [get]
func (h *TeamWarehouseHandler) GetWarehouseItems(c echo.Context) error {
	// 1. 获取路径参数
	teamID := c.Param("team_id")
	if teamID == "" {
		return response.EchoBadRequest(c, h.respWriter, "团队ID不能为空")
	}

	heroID := c.QueryParam("hero_id")
	if heroID == "" {
		return response.EchoBadRequest(c, h.respWriter, "英雄ID不能为空")
	}

	// 2. 获取分页参数
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 20
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	// 3. 调用 Service
	items, total, err := h.warehouseService.GetWarehouseItems(c.Request().Context(), teamID, heroID, limit, offset)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 4. 返回响应
	resp := &WarehouseItemsResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// GetDistributionHistory 获取分配历史
// @Summary 获取分配历史
// @Tags 团队仓库
// @Accept json
// @Produce json
// @Param team_id path string true "团队ID"
// @Param hero_id query string true "操作者英雄ID"
// @Param start_at query string false "开始时间 RFC3339"
// @Param end_at query string false "结束时间 RFC3339"
// @Param limit query int false "每页数量" default(20)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {object} response.Response{data=DistributionHistoryResponse} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/teams/{team_id}/warehouse/distributions [get]
func (h *TeamWarehouseHandler) GetDistributionHistory(c echo.Context) error {
	teamID := c.Param("team_id")
	heroID := c.QueryParam("hero_id")
	if teamID == "" || heroID == "" {
		return response.EchoBadRequest(c, h.respWriter, "参数不能为空")
	}
	startAt := c.QueryParam("start_at")
	endAt := c.QueryParam("end_at")
	var startPtr, endPtr *string
	if startAt != "" {
		startPtr = &startAt
	}
	if endAt != "" {
		endPtr = &endAt
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	rows, total, err := h.warehouseService.GetDistributionHistory(c.Request().Context(), teamID, heroID, startPtr, endPtr, limit, offset)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为响应 DTO（字段小写、带 json tag）
	records := make([]*DistributionHistoryRecord, 0, len(rows))
	for _, r := range rows {
		records = append(records, &DistributionHistoryRecord{
			ID:                r.ID,
			TeamID:            r.TeamID,
			WarehouseID:       r.WarehouseID,
			DistributorHeroID: r.DistributorHeroID,
			RecipientHeroID:   r.RecipientHeroID,
			ItemType:          r.ItemType,
			ItemID:            r.ItemID,
			Quantity:          r.Quantity,
			DistributedAt:     r.DistributedAt,
		})
	}

	resp := &DistributionHistoryResponse{
		Records: records,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	}

	return response.EchoOK(c, h.respWriter, resp)
}
