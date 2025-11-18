package handler

import (
	"database/sql"
	"encoding/json"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/modules/admin/dto"
	"tsu-self/internal/modules/admin/service"
	"tsu-self/internal/pkg/response"
)

// DungeonEventHandler 事件配置 HTTP 处理器
type DungeonEventHandler struct {
	service    *service.DungeonEventService
	respWriter response.Writer
}

// NewDungeonEventHandler 创建事件配置处理器
func NewDungeonEventHandler(db *sql.DB, respWriter response.Writer) *DungeonEventHandler {
	return &DungeonEventHandler{
		service:    service.NewDungeonEventService(db),
		respWriter: respWriter,
	}
}

// CreateEvent 创建事件配置
// @Summary 创建事件配置
// @Description 创建新的事件配置,用于剧情/互动节点,可设置奖励效果、掉落与描述文案。
// @Description
// @Description **字段说明**:
// @Description - `event_code`: 唯一代码
// @Description - `event_description`: 前置描述,支持富文本
// @Description - `apply_effects`: 效果列表,每项包含 `buff_code`, `caster_level`, `target`, `buff_params`(JSON)
// @Description - `drop_config`: 事件掉落,可关联掉落池 + 配置保底奖励
// @Description - `reward_exp`: 固定经验奖励
// @Description - `event_end_desc`: 结束描述
// @Description - `is_active`: 是否在房间中可用
// @Tags 地城事件管理
// @Accept json
// @Produce json
// @Param request body dto.CreateEventRequest true "创建事件配置请求"
// @Success 201 {object} response.Response{data=dto.EventResponse} "创建成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 409 {object} response.Response "事件配置代码已存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/dungeon-events [post]
func (h *DungeonEventHandler) CreateEvent(c echo.Context) error {
	var req dto.CreateEventRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求参数格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	event, err := h.service.CreateEvent(c.Request().Context(), &req)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	resp := h.toEventResponse(event)
	return response.EchoOK(c, h.respWriter, resp)
}

// GetEvent 获取事件配置详情
// @Summary 获取事件配置详情
// @Description 根据ID获取事件配置详细信息,包含效果列表、掉落、经验值和描述文案。
// @Tags 地城事件管理
// @Accept json
// @Produce json
// @Param id path string true "事件配置ID"
// @Success 200 {object} response.Response{data=dto.EventResponse} "查询成功"
// @Failure 404 {object} response.Response "事件配置不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/dungeon-events/{id} [get]
func (h *DungeonEventHandler) GetEvent(c echo.Context) error {
	eventID := c.Param("id")

	event, err := h.service.GetEventByID(c.Request().Context(), eventID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	resp := h.toEventResponse(event)
	return response.EchoOK(c, h.respWriter, resp)
}

// UpdateEvent 更新事件配置
// @Summary 更新事件配置
// @Description 更新事件配置信息,支持部分字段更新（例如替换掉落、增删效果、调整描述或启用状态）。
// @Tags 地城事件管理
// @Accept json
// @Produce json
// @Param id path string true "事件配置ID"
// @Param request body dto.UpdateEventRequest true "更新事件配置请求"
// @Success 200 {object} response.Response{data=dto.EventResponse} "更新成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "事件配置不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/dungeon-events/{id} [put]
func (h *DungeonEventHandler) UpdateEvent(c echo.Context) error {
	eventID := c.Param("id")

	var req dto.UpdateEventRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求参数格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	event, err := h.service.UpdateEvent(c.Request().Context(), eventID, &req)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	resp := h.toEventResponse(event)
	return response.EchoOK(c, h.respWriter, resp)
}

// DeleteEvent 删除事件配置
// @Summary 删除事件配置
// @Description 软删除事件配置,删除后无法再被房间引用,但历史记录保留。
// @Tags 地城事件管理
// @Accept json
// @Produce json
// @Param id path string true "事件配置ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 404 {object} response.Response "事件配置不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/dungeon-events/{id} [delete]
func (h *DungeonEventHandler) DeleteEvent(c echo.Context) error {
	eventID := c.Param("id")

	if err := h.service.DeleteEvent(c.Request().Context(), eventID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]string{"message": "删除成功"})
}

// toEventResponse 转换为响应格式
func (h *DungeonEventHandler) toEventResponse(event *game_config.DungeonEvent) dto.EventResponse {
	resp := dto.EventResponse{
		ID:        event.ID,
		EventCode: event.EventCode,
		RewardExp: event.RewardExp.Int,
		IsActive:  event.IsActive,
		CreatedAt: event.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: event.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if event.EventDescription.Valid {
		desc := event.EventDescription.String
		resp.EventDescription = &desc
	}

	// 解析施加效果
	var applyEffects []dto.ApplyEffectItem
	if err := json.Unmarshal(event.ApplyEffects, &applyEffects); err == nil {
		resp.ApplyEffects = applyEffects
	}

	// 解析掉落配置
	var dropConfig dto.DropConfig
	if err := json.Unmarshal(event.DropConfig, &dropConfig); err == nil {
		resp.DropConfig = dropConfig
	}

	if event.EventEndDesc.Valid {
		desc := event.EventEndDesc.String
		resp.EventEndDesc = &desc
	}

	return resp
}
