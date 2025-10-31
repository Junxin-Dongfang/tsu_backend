// Package handler 提供Admin模块的HTTP请求处理器
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

// WorldDropHandler 世界掉落配置Handler
type WorldDropHandler struct {
	service    *service.WorldDropService
	respWriter response.Writer
}

// NewWorldDropHandler 创建世界掉落配置Handler
func NewWorldDropHandler(db *sql.DB, respWriter response.Writer) *WorldDropHandler {
	return &WorldDropHandler{
		service:    service.NewWorldDropService(db),
		respWriter: respWriter,
	}
}

// CreateWorldDrop 创建世界掉落配置
// @Summary 创建世界掉落配置
// @Description 创建新的世界掉落配置,用于全局掉落系统(如世界BOSS、野外精英等)。
// @Description
// @Description **世界掉落特点**:
// @Description - 全局生效,不限于特定掉落池
// @Description - 支持基础掉落率和掉落率修正器
// @Description - 支持触发条件(等级、任务、区域等)
// @Description - 支持掉落限制(总量、每日、每小时)
// @Description
// @Description **掉落率配置**:
// @Description - base_drop_rate: 基础掉落概率(0-1之间)
// @Description - drop_rate_modifiers: 掉落率修正器(JSON对象),根据不同条件调整掉落率
// @Description
// @Description **触发条件**(trigger_conditions):
// @Description - min_player_level: 最低玩家等级
// @Description - max_player_level: 最高玩家等级
// @Description - required_quest: 需要完成的任务ID(可选)
// @Description - zone: 限定区域(可选)
// @Description
// @Description **掉落限制**:
// @Description - total_drop_limit: 总掉落次数限制
// @Description - daily_drop_limit: 每日掉落次数限制
// @Description - hourly_drop_limit: 每小时掉落次数限制
// @Description - min_drop_interval: 最小掉落间隔(秒)
// @Description - max_drop_interval: 最大掉落间隔(秒)
// @Description
// @Description **请求示例**:
// @Description ```json
// @Description {
// @Description   "item_id": "550e8400-e29b-41d4-a716-446655440000",
// @Description   "base_drop_rate": 0.01,
// @Description   "trigger_conditions": "{\"min_player_level\":30,\"max_player_level\":60,\"zone\":\"dark_forest\"}",
// @Description   "drop_rate_modifiers": "{\"time_of_day\":{\"morning\":1.2,\"night\":0.8},\"player_luck_bonus\":0.1}",
// @Description   "daily_drop_limit": 10
// @Description }
// @Description ```
// @Tags 世界掉落配置
// @Accept json
// @Produce json
// @Param request body dto.CreateWorldDropRequest true "世界掉落配置请求"
// @Success 200 {object} response.Response{data=dto.WorldDropResponse} "创建成功,返回世界掉落详情"
// @Failure 400 {object} response.Response "参数错误(100400): item_id不存在、掉落率无效等"
// @Failure 500 {object} response.Response "服务器错误(100500)"
// @Router /admin/world-drops [post]
func (h *WorldDropHandler) CreateWorldDrop(c echo.Context) error {
	// 1. 解析请求
	var req dto.CreateWorldDropRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 2. 验证请求
	if err := c.Validate(&req); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 3. 创建世界掉落配置
	resp, err := h.service.CreateWorldDrop(c.Request().Context(), &req)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// GetWorldDropList 查询世界掉落列表
// @Summary 查询世界掉落列表
// @Description 查询世界掉落配置列表,支持分页、筛选和排序。
// @Description
// @Description **筛选条件**:
// @Description - item_id: 按物品ID筛选
// @Description - is_active: 按激活状态筛选
// @Description
// @Description **排序**:
// @Description - sort_by: 排序字段(base_drop_rate/created_at/updated_at)
// @Description - sort_order: 排序方向(asc升序/desc降序)
// @Description - 默认按base_drop_rate降序(掉落率从高到低)
// @Tags 世界掉落配置
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1) minimum(1)
// @Param page_size query int false "每页数量" default(20) minimum(1) maximum(100)
// @Param item_id query string false "物品ID筛选(UUID格式)"
// @Param is_active query bool false "激活状态筛选"
// @Param sort_by query string false "排序字段" Enums(base_drop_rate, created_at, updated_at) default(base_drop_rate)
// @Param sort_order query string false "排序方向" Enums(asc, desc) default(desc)
// @Success 200 {object} response.Response{data=dto.WorldDropListResponse} "查询成功"
// @Failure 400 {object} response.Response "参数错误(100400)"
// @Failure 500 {object} response.Response "服务器错误(100500)"
// @Router /admin/world-drops [get]
func (h *WorldDropHandler) GetWorldDropList(c echo.Context) error {
	// 1. 解析查询参数
	params := interfaces.ListWorldDropConfigParams{
		Page:     1,
		PageSize: 20,
	}

	if page := c.QueryParam("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			params.Page = p
		}
	}

	if pageSize := c.QueryParam("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 {
			params.PageSize = ps
		}
	}

	if itemID := c.QueryParam("item_id"); itemID != "" {
		params.ItemID = &itemID
	}

	if isActiveStr := c.QueryParam("is_active"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			params.IsActive = &isActive
		}
	}

	params.SortBy = c.QueryParam("sort_by")
	params.SortOrder = c.QueryParam("sort_order")

	// 2. 查询世界掉落列表
	resp, err := h.service.GetWorldDropList(c.Request().Context(), params)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// GetWorldDrop 获取世界掉落详情
// @Summary 获取世界掉落详情
// @Description 根据ID获取世界掉落配置详情
// @Tags 世界掉落配置
// @Accept json
// @Produce json
// @Param id path string true "世界掉落配置ID"
// @Success 200 {object} response.Response{data=dto.WorldDropResponse}
// @Router /admin/world-drops/{id} [get]
func (h *WorldDropHandler) GetWorldDrop(c echo.Context) error {
	// 1. 获取配置ID
	configID := c.Param("id")

	// 2. 查询世界掉落详情
	resp, err := h.service.GetWorldDropByID(c.Request().Context(), configID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// UpdateWorldDrop 更新世界掉落配置
// @Summary 更新世界掉落配置
// @Description 更新世界掉落配置信息
// @Tags 世界掉落配置
// @Accept json
// @Produce json
// @Param id path string true "世界掉落配置ID"
// @Param request body dto.UpdateWorldDropRequest true "更新请求"
// @Success 200 {object} response.Response{data=dto.WorldDropResponse}
// @Router /admin/world-drops/{id} [put]
func (h *WorldDropHandler) UpdateWorldDrop(c echo.Context) error {
	// 1. 获取配置ID
	configID := c.Param("id")

	// 2. 解析请求
	var req dto.UpdateWorldDropRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 3. 验证请求
	if err := c.Validate(&req); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 4. 更新世界掉落配置
	resp, err := h.service.UpdateWorldDrop(c.Request().Context(), configID, &req)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// DeleteWorldDrop 删除世界掉落配置
// @Summary 删除世界掉落配置
// @Description 删除世界掉落配置（软删除）
// @Tags 世界掉落配置
// @Accept json
// @Produce json
// @Param id path string true "世界掉落配置ID"
// @Success 200 {object} response.Response
// @Router /admin/world-drops/{id} [delete]
func (h *WorldDropHandler) DeleteWorldDrop(c echo.Context) error {
	// 1. 获取配置ID
	configID := c.Param("id")

	// 2. 删除世界掉落配置
	if err := h.service.DeleteWorldDrop(c.Request().Context(), configID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]string{"message": "删除成功"})
}
