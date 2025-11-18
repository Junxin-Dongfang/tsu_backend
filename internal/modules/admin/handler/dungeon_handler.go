package handler

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"

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
// @Description 创建新的地城配置,一次性定义基础信息、等级限制、挑战次数以及房间流程。
// @Description
// @Description **填写说明**:
// @Description - `dungeon_code`/`dungeon_name`: 运营唯一标识 + 展示名称,代码全局唯一
// @Description - `min_level`/`max_level`: 限定英雄等级区间,超出范围无法进入
// @Description - `is_time_limited` + `time_limit_start/time_limit_end`: 可配置限时开放,时间使用 ISO8601
// @Description - `requires_attempts` + `max_attempts_per_day`: 是否扣挑战次数以及每日上限
// @Description - `room_sequence`: 定义房间流程,数组顺序即关卡顺序
// @Description   - `room_id`: 房间ID(UUID)
// @Description   - `sort`: 房间顺序(从1开始)
// @Description   - `conditional_skip`: 满足条件时跳过本房间 `{ "condition": "has_key", "jump_to": 4 }`
// @Description   - `conditional_return`: 失败后回退 `{ "condition": "fail_boss", "return_to": 2 }`
// @Description - `is_active`: 是否立即上线(用于灰度发布)
// @Description
// @Description **请求示例**:
// @Description ```json
// @Description {
// @Description   "dungeon_code": "dungeon_forest_beginner",
// @Description   "dungeon_name": "初心者森林",
// @Description   "min_level": 5,
// @Description   "max_level": 20,
// @Description   "requires_attempts": true,
// @Description   "max_attempts_per_day": 3,
// @Description   "room_sequence": [
// @Description     {"room_id": "c5e6fb80-0d44-4e41-85eb-3e52fa45209d", "sort": 1},
// @Description     {"room_id": "f82b1a2c-e2b4-4a66-8a37-2e3a682a5c88", "sort": 2}
// @Description   ],
// @Description   "is_active": true
// @Description }
// @Description ```
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
// @Description 分页查询地城列表,支持按代码/名称/等级/启用状态筛选,并提供排序能力。常用于后台列表页与选择器。
// @Description
// @Description **筛选参数**:
// @Description - `dungeon_code`: 模糊匹配地城代码,便于快速定位(`"dungeon_forest"` -> 匹配`dungeon_forest_beginner`)
// @Description - `dungeon_name`: 模糊匹配地城名称
// @Description - `min_level`/`max_level`: 仅返回等级区间重叠的地城(输入10-30会列出与该区间相交的配置)
// @Description - `is_active`: true=仅展示上线地城, false=仅展示下线地城, 未传=全部
// @Description
// @Description **分页&排序**:
// @Description - `page` + `page_size`: 默认第1页,每页20条,最大100
// @Description - `order_by`: 支持 `created_at` / `updated_at` / `dungeon_code`
// @Description - `order_desc`: true=降序(默认), false=升序
// @Tags 地城管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1) minimum(1)
// @Param page_size query int false "每页数量" default(20) minimum(1) maximum(100)
// @Param dungeon_code query string false "按地城代码模糊搜索(大小写不敏感)"
// @Param dungeon_name query string false "按地城名称模糊搜索"
// @Param min_level query int false "最小等级" minimum(1) maximum(100)
// @Param max_level query int false "最大等级" minimum(1) maximum(100)
// @Param is_active query bool false "是否启用 true=仅启用 false=仅停用"
// @Param order_by query string false "排序字段" Enums(created_at,updated_at,dungeon_code) default(created_at)
// @Param order_desc query bool false "是否降序" default(true)
// @Success 200 {object} response.Response{data=dto.DungeonListResponse} "查询成功,返回列表/总数/排序信息"
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

	orderBy := c.QueryParam("order_by")
	if orderBy == "" {
		orderBy = "created_at"
	}
	switch strings.ToLower(orderBy) {
	case "created_at", "updated_at", "dungeon_code":
		params.OrderBy = orderBy
	default:
		return response.EchoBadRequest(c, h.respWriter, "order_by 不受支持")
	}

	params.OrderDesc = true
	if orderDesc := c.QueryParam("order_desc"); orderDesc != "" {
		if desc, err := strconv.ParseBool(orderDesc); err == nil {
			params.OrderDesc = desc
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
		List:      list,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
		OrderBy:   params.OrderBy,
		OrderDesc: params.OrderDesc,
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// GetDungeon 获取地城详情
// @Summary 获取地城详情
// @Description 根据ID获取完整地城配置(基础信息、房间流程、挑战限制等),用于详情页或编辑表单回显。
// @Description
// @Description **返回字段**:
// @Description - `dungeon_code/dungeon_name`: 基础信息
// @Description - `min_level/max_level`: 等级限制
// @Description - `is_time_limited` + `time_limit_start/end`: 限时配置
// @Description - `requires_attempts` + `max_attempts_per_day`: 挑战次数限制
// @Description - `room_sequence`: 房间流程列表,包含顺序与条件跳转
// @Description - `is_active`: 启用状态
// @Description - `created_at/updated_at`: 操作审计
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
// @Description 更新地城配置信息,支持部分字段更新。只需传需要调整的字段即可,未传字段保持原值。
// @Description
// @Description **常见场景**:
// @Description - 调整开放等级/挑战次数: 仅传 `min_level/max_level/max_attempts_per_day`
// @Description - 替换房间流程: 仅传 `room_sequence`, 其余字段不受影响
// @Description - 暂停/恢复地城: 仅传 `is_active`
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
// @Description 软删除地城配置（记录保留,但不会再出现在列表/选择器中）,常用于下线废弃地城。
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
