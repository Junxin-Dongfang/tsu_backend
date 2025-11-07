package handler

import (
	"database/sql"
	"encoding/json"
	"strconv"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/modules/admin/dto"
	"tsu-self/internal/modules/admin/service"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/repository/interfaces"
)

// DungeonRoomHandler 房间 HTTP 处理器
type DungeonRoomHandler struct {
	service    *service.DungeonRoomService
	respWriter response.Writer
}

// NewDungeonRoomHandler 创建房间处理器
func NewDungeonRoomHandler(db *sql.DB, respWriter response.Writer) *DungeonRoomHandler {
	return &DungeonRoomHandler{
		service:    service.NewDungeonRoomService(db),
		respWriter: respWriter,
	}
}

// CreateRoom 创建房间
// @Summary 创建房间
// @Description 创建新的房间配置
// @Tags 地城房间管理
// @Accept json
// @Produce json
// @Param request body dto.CreateRoomRequest true "创建房间请求"
// @Success 201 {object} response.Response{data=dto.RoomResponse} "创建成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 409 {object} response.Response "房间代码已存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/dungeon-rooms [post]
func (h *DungeonRoomHandler) CreateRoom(c echo.Context) error {
	var req dto.CreateRoomRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求参数格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	room, err := h.service.CreateRoom(c.Request().Context(), &req)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	resp := h.toRoomResponse(room)
	return response.EchoOK(c, h.respWriter, resp)
}

// GetRooms 获取房间列表
// @Summary 获取房间列表
// @Description 分页查询房间列表,支持筛选
// @Tags 地城房间管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param room_code query string false "房间代码（模糊搜索）"
// @Param room_name query string false "房间名称（模糊搜索）"
// @Param room_type query string false "房间类型(battle/event/treasure/rest)"
// @Param is_active query bool false "是否启用"
// @Success 200 {object} response.Response{data=dto.RoomListResponse} "查询成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/dungeon-rooms [get]
func (h *DungeonRoomHandler) GetRooms(c echo.Context) error {
	// 解析分页参数
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 构建查询参数
	params := interfaces.DungeonRoomQueryParams{
		Limit:  pageSize,
		Offset: (page - 1) * pageSize,
	}

	if roomCode := c.QueryParam("room_code"); roomCode != "" {
		params.RoomCode = &roomCode
	}
	if roomName := c.QueryParam("room_name"); roomName != "" {
		params.RoomName = &roomName
	}
	if roomType := c.QueryParam("room_type"); roomType != "" {
		params.RoomType = &roomType
	}
	if isActive := c.QueryParam("is_active"); isActive != "" {
		if active, err := strconv.ParseBool(isActive); err == nil {
			params.IsActive = &active
		}
	}

	rooms, total, err := h.service.GetRooms(c.Request().Context(), params)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为响应格式
	list := make([]dto.RoomResponse, 0, len(rooms))
	for _, room := range rooms {
		list = append(list, h.toRoomResponse(room))
	}

	resp := dto.RoomListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// GetRoom 获取房间详情
// @Summary 获取房间详情
// @Description 根据ID获取房间详细信息
// @Tags 地城房间管理
// @Accept json
// @Produce json
// @Param id path string true "房间ID"
// @Success 200 {object} response.Response{data=dto.RoomResponse} "查询成功"
// @Failure 404 {object} response.Response "房间不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/dungeon-rooms/{id} [get]
func (h *DungeonRoomHandler) GetRoom(c echo.Context) error {
	roomID := c.Param("id")

	room, err := h.service.GetRoomByID(c.Request().Context(), roomID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	resp := h.toRoomResponse(room)
	return response.EchoOK(c, h.respWriter, resp)
}

// UpdateRoom 更新房间
// @Summary 更新房间
// @Description 更新房间配置信息
// @Tags 地城房间管理
// @Accept json
// @Produce json
// @Param id path string true "房间ID"
// @Param request body dto.UpdateRoomRequest true "更新房间请求"
// @Success 200 {object} response.Response{data=dto.RoomResponse} "更新成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "房间不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/dungeon-rooms/{id} [put]
func (h *DungeonRoomHandler) UpdateRoom(c echo.Context) error {
	roomID := c.Param("id")

	var req dto.UpdateRoomRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求参数格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	room, err := h.service.UpdateRoom(c.Request().Context(), roomID, &req)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	resp := h.toRoomResponse(room)
	return response.EchoOK(c, h.respWriter, resp)
}

// DeleteRoom 删除房间
// @Summary 删除房间
// @Description 软删除房间配置
// @Tags 地城房间管理
// @Accept json
// @Produce json
// @Param id path string true "房间ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 404 {object} response.Response "房间不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/dungeon-rooms/{id} [delete]
func (h *DungeonRoomHandler) DeleteRoom(c echo.Context) error {
	roomID := c.Param("id")

	if err := h.service.DeleteRoom(c.Request().Context(), roomID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]string{"message": "删除成功"})
}

// toRoomResponse 转换为响应格式
func (h *DungeonRoomHandler) toRoomResponse(room *game_config.DungeonRoom) dto.RoomResponse {
	resp := dto.RoomResponse{
		ID:       room.ID,
		RoomCode: room.RoomCode,
		RoomType: room.RoomType,
		IsActive: room.IsActive,
		CreatedAt: room.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: room.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if room.RoomName.Valid {
		name := room.RoomName.String
		resp.RoomName = &name
	}

	if room.TriggerID.Valid {
		triggerID := room.TriggerID.String
		resp.TriggerID = &triggerID
	}

	// 解析开启条件
	var openConditions map[string]interface{}
	if err := json.Unmarshal(room.OpenConditions, &openConditions); err == nil {
		resp.OpenConditions = openConditions
	}

	return resp
}

