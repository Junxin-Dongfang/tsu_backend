package handler

import (
	"tsu-self/internal/modules/game/service"
	"tsu-self/internal/pkg/response"

	"github.com/labstack/echo/v4"
)

// EquipmentSetHandler 装备套装处理器
type EquipmentSetHandler struct {
	equipmentSetSvc *service.EquipmentSetService
	respWriter      response.Writer
}

// NewEquipmentSetHandler 创建装备套装处理器
func NewEquipmentSetHandler(equipmentSetSvc *service.EquipmentSetService, respWriter response.Writer) *EquipmentSetHandler {
	return &EquipmentSetHandler{
		equipmentSetSvc: equipmentSetSvc,
		respWriter:      respWriter,
	}
}

// ListSetsRequest 查询套装列表请求
type ListSetsRequest struct {
	Page     int `query:"page"`
	PageSize int `query:"page_size"`
}

// ListSetsResponse 查询套装列表响应
type ListSetsResponse struct {
	Sets  []*service.SetInfo `json:"sets"`
	Total int64              `json:"total"`
	Page  int                `json:"page"`
}

// SetInfoResponse 套装详细信息响应
type SetInfoResponse struct {
	*service.SetInfo
}

// ActiveSetsResponse 激活的套装响应
type ActiveSetsResponse struct {
	ActiveSets []*service.ActiveSetInfo `json:"active_sets"`
}

// ListSets 查询可用套装列表
// @Summary 查询可用套装列表
// @Description 分页查询所有可用的装备套装
// @Tags Equipment
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=ListSetsResponse} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /game/equipment/sets [get]
func (h *EquipmentSetHandler) ListSets(c echo.Context) error {
	// 解析请求参数
	var req ListSetsRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数解析失败")
	}

	// 设置默认值
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}

	// 调用服务层
	sets, total, err := h.equipmentSetSvc.ListAvailableSets(c.Request().Context(), &service.ListSetsRequest{
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 返回响应
	return response.EchoOK(c, h.respWriter, ListSetsResponse{
		Sets:  sets,
		Total: total,
		Page:  req.Page,
	})
}

// GetSetInfo 查询套装详细信息
// @Summary 查询套装详细信息
// @Description 根据套装ID查询套装的详细信息，包括套装效果和包含的装备
// @Tags Equipment
// @Accept json
// @Produce json
// @Param set_id path string true "套装ID"
// @Success 200 {object} response.Response{data=SetInfoResponse} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "套装不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /game/equipment/sets/{set_id} [get]
func (h *EquipmentSetHandler) GetSetInfo(c echo.Context) error {
	// 获取套装ID
	setID := c.Param("set_id")
	if setID == "" {
		return response.EchoBadRequest(c, h.respWriter, "套装ID不能为空")
	}

	// 调用服务层
	setInfo, err := h.equipmentSetSvc.GetSetInfo(c.Request().Context(), setID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 返回响应
	return response.EchoOK(c, h.respWriter, SetInfoResponse{SetInfo: setInfo})
}

// GetActiveSets 查询英雄激活的套装
// @Summary 查询英雄激活的套装
// @Description 查询指定英雄当前激活的所有套装及其效果
// @Tags Equipment
// @Accept json
// @Produce json
// @Param hero_id path string true "英雄ID"
// @Success 200 {object} response.Response{data=ActiveSetsResponse} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "英雄不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /game/equipment/sets/active/{hero_id} [get]
func (h *EquipmentSetHandler) GetActiveSets(c echo.Context) error {
	// 获取英雄ID
	heroID := c.Param("hero_id")
	if heroID == "" {
		return response.EchoBadRequest(c, h.respWriter, "英雄ID不能为空")
	}

	// 调用服务层
	activeSets, err := h.equipmentSetSvc.GetActiveSets(c.Request().Context(), heroID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 返回响应
	return response.EchoOK(c, h.respWriter, ActiveSetsResponse{
		ActiveSets: activeSets,
	})
}

// RegisterRoutes 注册路由
func (h *EquipmentSetHandler) RegisterRoutes(g *echo.Group) {
	sets := g.Group("/equipment/sets")
	{
		sets.GET("", h.ListSets)                      // GET /game/equipment/sets
		sets.GET("/:set_id", h.GetSetInfo)            // GET /game/equipment/sets/:set_id
		sets.GET("/active/:hero_id", h.GetActiveSets) // GET /game/equipment/sets/active/:hero_id
	}
}

