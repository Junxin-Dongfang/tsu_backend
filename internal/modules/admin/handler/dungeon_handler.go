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

// DungeonHandler 地城 HTTP 处理器
type DungeonHandler struct {
	service    *service.DungeonService
	respWriter response.Writer
}

// NewDungeonHandler 创建地城处理器
func NewDungeonHandler(db *sql.DB, respWriter response.Writer) *DungeonHandler {
	return &DungeonHandler{
		service:    service.NewDungeonService(db),
		respWriter: respWriter,
	}
}

// CreateDungeon 创建地城
// @Summary 创建地城
// @Description 创建新的地城配置
// @Tags 地城管理
// @Accept json
// @Produce json
// @Param request body dto.CreateDungeonRequest true "创建地城请求"
// @Success 201 {object} response.Response{data=dto.DungeonResponse} "创建成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 409 {object} response.Response "地城代码已存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/dungeons [post]
func (h *DungeonHandler) CreateDungeon(c echo.Context) error {
	var req dto.CreateDungeonRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求参数格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	dungeon, err := h.service.CreateDungeon(c.Request().Context(), &req)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	resp := h.toDungeonResponse(dungeon)
	return response.EchoOK(c, h.respWriter, resp)
}

// GetDungeons 获取地城列表
// @Summary 获取地城列表
// @Description 分页查询地城列表,支持筛选
// @Tags 地城管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param dungeon_code query string false "地城代码（模糊搜索）"
// @Param dungeon_name query string false "地城名称（模糊搜索）"
// @Param min_level query int false "最小等级"
// @Param max_level query int false "最大等级"
// @Param is_active query bool false "是否启用"
// @Success 200 {object} response.Response{data=dto.DungeonListResponse} "查询成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/dungeons [get]
func (h *DungeonHandler) GetDungeons(c echo.Context) error {
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
	params := interfaces.DungeonQueryParams{
		Limit:  pageSize,
		Offset: (page - 1) * pageSize,
	}

	if dungeonCode := c.QueryParam("dungeon_code"); dungeonCode != "" {
		params.DungeonCode = &dungeonCode
	}
	if dungeonName := c.QueryParam("dungeon_name"); dungeonName != "" {
		params.DungeonName = &dungeonName
	}
	if minLevel := c.QueryParam("min_level"); minLevel != "" {
		if level, err := strconv.ParseInt(minLevel, 10, 16); err == nil {
			l := int16(level)
			params.MinLevel = &l
		}
	}
	if maxLevel := c.QueryParam("max_level"); maxLevel != "" {
		if level, err := strconv.ParseInt(maxLevel, 10, 16); err == nil {
			l := int16(level)
			params.MaxLevel = &l
		}
	}
	if isActive := c.QueryParam("is_active"); isActive != "" {
		if active, err := strconv.ParseBool(isActive); err == nil {
			params.IsActive = &active
		}
	}

	dungeons, total, err := h.service.GetDungeons(c.Request().Context(), params)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为响应格式
	list := make([]dto.DungeonResponse, 0, len(dungeons))
	for _, dungeon := range dungeons {
		list = append(list, h.toDungeonResponse(dungeon))
	}

	resp := dto.DungeonListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// GetDungeon 获取地城详情
// @Summary 获取地城详情
// @Description 根据ID获取地城详细信息
// @Tags 地城管理
// @Accept json
// @Produce json
// @Param id path string true "地城ID"
// @Success 200 {object} response.Response{data=dto.DungeonResponse} "查询成功"
// @Failure 404 {object} response.Response "地城不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/dungeons/{id} [get]
func (h *DungeonHandler) GetDungeon(c echo.Context) error {
	dungeonID := c.Param("id")

	dungeon, err := h.service.GetDungeonByID(c.Request().Context(), dungeonID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	resp := h.toDungeonResponse(dungeon)
	return response.EchoOK(c, h.respWriter, resp)
}

// UpdateDungeon 更新地城
// @Summary 更新地城
// @Description 更新地城配置信息
// @Tags 地城管理
// @Accept json
// @Produce json
// @Param id path string true "地城ID"
// @Param request body dto.UpdateDungeonRequest true "更新地城请求"
// @Success 200 {object} response.Response{data=dto.DungeonResponse} "更新成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "地城不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/dungeons/{id} [put]
func (h *DungeonHandler) UpdateDungeon(c echo.Context) error {
	dungeonID := c.Param("id")

	var req dto.UpdateDungeonRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求参数格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	dungeon, err := h.service.UpdateDungeon(c.Request().Context(), dungeonID, &req)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	resp := h.toDungeonResponse(dungeon)
	return response.EchoOK(c, h.respWriter, resp)
}

// DeleteDungeon 删除地城
// @Summary 删除地城
// @Description 软删除地城配置
// @Tags 地城管理
// @Accept json
// @Produce json
// @Param id path string true "地城ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 404 {object} response.Response "地城不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/dungeons/{id} [delete]
func (h *DungeonHandler) DeleteDungeon(c echo.Context) error {
	dungeonID := c.Param("id")

	if err := h.service.DeleteDungeon(c.Request().Context(), dungeonID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]string{"message": "删除成功"})
}

// toDungeonResponse 转换为响应格式
func (h *DungeonHandler) toDungeonResponse(dungeon *game_config.Dungeon) dto.DungeonResponse {
	resp := dto.DungeonResponse{
		ID:               dungeon.ID,
		DungeonCode:      dungeon.DungeonCode,
		DungeonName:      dungeon.DungeonName,
		MinLevel:         dungeon.MinLevel,
		MaxLevel:         dungeon.MaxLevel,
		IsTimeLimited:    dungeon.IsTimeLimited,
		RequiresAttempts: dungeon.RequiresAttempts,
		IsActive:         dungeon.IsActive,
		CreatedAt:        dungeon.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:        dungeon.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if dungeon.Description.Valid {
		desc := dungeon.Description.String
		resp.Description = &desc
	}

	if dungeon.TimeLimitStart.Valid {
		start := dungeon.TimeLimitStart.Time.Format("2006-01-02T15:04:05Z07:00")
		resp.TimeLimitStart = &start
	}

	if dungeon.TimeLimitEnd.Valid {
		end := dungeon.TimeLimitEnd.Time.Format("2006-01-02T15:04:05Z07:00")
		resp.TimeLimitEnd = &end
	}

	if dungeon.MaxAttemptsPerDay.Valid {
		attempts := dungeon.MaxAttemptsPerDay.Int16
		resp.MaxAttemptsPerDay = &attempts
	}

	// 解析房间序列
	var roomSequence []dto.RoomSequenceItem
	if err := json.Unmarshal(dungeon.RoomSequence, &roomSequence); err == nil {
		resp.RoomSequence = roomSequence
	}

	return resp
}

