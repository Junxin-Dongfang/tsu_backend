package handler

import (
	"database/sql"
	"strconv"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/modules/admin/dto"
	"tsu-self/internal/modules/admin/service"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/repository/interfaces"
)

// EquipmentSlotHandler 装备槽位配置Handler
type EquipmentSlotHandler struct {
	service    *service.EquipmentSlotService
	respWriter response.Writer
}

// NewEquipmentSlotHandler 创建装备槽位配置Handler
func NewEquipmentSlotHandler(db *sql.DB, respWriter response.Writer) *EquipmentSlotHandler {
	return &EquipmentSlotHandler{
		service:    service.NewEquipmentSlotService(db),
		respWriter: respWriter,
	}
}

// CreateSlot 创建槽位配置
// @Summary 创建槽位配置
// @Description 创建新的装备槽位配置,定义角色可装备的槽位类型和规则。
// @Description
// @Description **槽位类型**:
// @Description - weapon: 武器槽位(mainhand主手, offhand副手)
// @Description - armor: 护甲槽位(head头部, chest胸部, legs腿部, feet脚部, hands手部, waist腰部)
// @Description - accessory: 饰品槽位(neck项链, ring戒指, trinket饰品)
// @Description - special: 特殊槽位(自定义)
// @Description
// @Description **配置说明**:
// @Description - slot_code: 槽位唯一代码(如mainhand, head等)
// @Description - slot_name: 槽位显示名称
// @Description - slot_type: 槽位类型分类(weapon/armor/accessory/special)
// @Description - display_order: 显示顺序(用于UI排序,数字越小越靠前)
// @Description - icon: 槽位图标URL(可选)
// @Description - description: 槽位描述(可选)
// @Description
// @Description **请求示例**:
// @Description ```json
// @Description {
// @Description   "slot_code": "mainhand",
// @Description   "slot_name": "主手",
// @Description   "slot_type": "weapon",
// @Description   "display_order": 1,
// @Description   "icon": "/icons/slots/mainhand.png",
// @Description   "description": "主手武器槽位"
// @Description }
// @Description ```
// @Tags 装备槽位配置
// @Accept json
// @Produce json
// @Param request body dto.CreateSlotRequest true "槽位配置请求"
// @Success 200 {object} response.Response{data=dto.SlotConfigResponse} "创建成功,返回槽位详情"
// @Failure 400 {object} response.Response "参数错误(100400): slot_code重复、类型无效等"
// @Failure 500 {object} response.Response "服务器错误(100500)"
// @Router /admin/equipment-slots [post]
func (h *EquipmentSlotHandler) CreateSlot(c echo.Context) error {
	// 1. 解析请求
	var req dto.CreateSlotRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 2. 验证请求
	if err := c.Validate(&req); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 3. 创建槽位配置
	resp, err := h.service.CreateSlot(c.Request().Context(), &req)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// GetSlotList 查询槽位列表
// @Summary 查询槽位列表
// @Description 查询装备槽位配置列表,支持分页、筛选、搜索和排序。
// @Description
// @Description **筛选条件**:
// @Description - slot_type: 按槽位类型筛选(weapon/armor/accessory/special)
// @Description - is_active: 按激活状态筛选
// @Description - keyword: 关键词搜索,匹配slot_code或slot_name
// @Description
// @Description **排序**:
// @Description - sort_by: 排序字段(display_order/created_at/slot_code)
// @Description - sort_order: 排序方向(asc升序/desc降序)
// @Description - 默认按display_order升序(按显示顺序排列)
// @Tags 装备槽位配置
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1) minimum(1)
// @Param page_size query int false "每页数量" default(20) minimum(1) maximum(100)
// @Param slot_type query string false "槽位类型筛选" Enums(weapon, armor, accessory, special)
// @Param is_active query bool false "激活状态筛选"
// @Param keyword query string false "关键词搜索(slot_code, slot_name)"
// @Param sort_by query string false "排序字段" Enums(display_order, created_at, slot_code) default(display_order)
// @Param sort_order query string false "排序方向" Enums(asc, desc) default(asc)
// @Success 200 {object} response.Response{data=dto.SlotListResponse} "查询成功"
// @Failure 400 {object} response.Response "参数错误(100400)"
// @Failure 500 {object} response.Response "服务器错误(100500)"
// @Router /admin/equipment-slots [get]
func (h *EquipmentSlotHandler) GetSlotList(c echo.Context) error {
	// 1. 解析查询参数
	params := interfaces.ListSlotConfigParams{
		Page:      1,
		PageSize:  20,
		SortBy:    "display_order",
		SortOrder: "asc",
	}

	if page := c.QueryParam("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			params.Page = p
		}
	}

	if pageSize := c.QueryParam("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil {
			params.PageSize = ps
		}
	}

	if slotType := c.QueryParam("slot_type"); slotType != "" {
		params.SlotType = &slotType
	}

	if isActive := c.QueryParam("is_active"); isActive != "" {
		if active, err := strconv.ParseBool(isActive); err == nil {
			params.IsActive = &active
		}
	}

	if keyword := c.QueryParam("keyword"); keyword != "" {
		params.Keyword = &keyword
	}

	if sortBy := c.QueryParam("sort_by"); sortBy != "" {
		params.SortBy = sortBy
	}

	if sortOrder := c.QueryParam("sort_order"); sortOrder != "" {
		params.SortOrder = sortOrder
	}

	// 2. 查询槽位列表
	resp, err := h.service.GetSlotList(c.Request().Context(), params)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// GetSlot 获取槽位详情
// @Summary 获取槽位详情
// @Description 根据ID获取装备槽位配置详情
// @Tags 装备槽位配置
// @Accept json
// @Produce json
// @Param id path string true "槽位ID"
// @Success 200 {object} response.Response{data=dto.SlotConfigResponse}
// @Router /admin/equipment-slots/{id} [get]
func (h *EquipmentSlotHandler) GetSlot(c echo.Context) error {
	// 1. 获取槽位ID
	slotID := c.Param("id")

	// 2. 查询槽位详情
	resp, err := h.service.GetSlotByID(c.Request().Context(), slotID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// UpdateSlot 更新槽位配置
// @Summary 更新槽位配置
// @Description 更新装备槽位配置（支持部分更新）
// @Tags 装备槽位配置
// @Accept json
// @Produce json
// @Param id path string true "槽位ID"
// @Param request body dto.UpdateSlotRequest true "更新请求"
// @Success 200 {object} response.Response{data=dto.SlotConfigResponse}
// @Router /admin/equipment-slots/{id} [put]
func (h *EquipmentSlotHandler) UpdateSlot(c echo.Context) error {
	// 1. 获取槽位ID
	slotID := c.Param("id")

	// 2. 解析请求
	var req dto.UpdateSlotRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 3. 验证请求
	if err := c.Validate(&req); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 4. 更新槽位配置
	resp, err := h.service.UpdateSlot(c.Request().Context(), slotID, &req)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// DeleteSlot 删除槽位配置
// @Summary 删除槽位配置
// @Description 删除装备槽位配置（软删除）
// @Tags 装备槽位配置
// @Accept json
// @Produce json
// @Param id path string true "槽位ID"
// @Success 200 {object} response.Response
// @Router /admin/equipment-slots/{id} [delete]
func (h *EquipmentSlotHandler) DeleteSlot(c echo.Context) error {
	// 1. 获取槽位ID
	slotID := c.Param("id")

	// 2. 删除槽位配置
	if err := h.service.DeleteSlot(c.Request().Context(), slotID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]string{"message": "槽位配置删除成功"})
}
