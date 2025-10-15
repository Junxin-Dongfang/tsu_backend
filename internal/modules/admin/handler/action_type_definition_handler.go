package handler

import (
	"database/sql"
	"strconv"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/modules/admin/service"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/repository/interfaces"
)

type ActionTypeDefinitionHandler struct {
	service    *service.ActionTypeDefinitionService
	respWriter response.Writer
}

func NewActionTypeDefinitionHandler(db *sql.DB, respWriter response.Writer) *ActionTypeDefinitionHandler {
	return &ActionTypeDefinitionHandler{
		service:    service.NewActionTypeDefinitionService(db),
		respWriter: respWriter,
	}
}

type ActionTypeDefinitionInfo struct {
	ID           string `json:"id"`
	ActionType   string `json:"action_type"`
	Description  string `json:"description"`
	PerTurnLimit int    `json:"per_turn_limit"`
	UsageTiming  string `json:"usage_timing"`
	Example      string `json:"example"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    int64  `json:"created_at"`
}

// GetActionTypeDefinitions 获取动作类型定义列表
// @Summary 获取动作类型定义列表
// @Description 获取动作类型定义的分页列表,支持按启用状态筛选。元数据配置表,定义DnD 5e动作类型(主要动作/附赠动作/反应)的规则和限制。
// @Tags 元数据
// @Accept json
// @Produce json
// @Param is_active query bool false "是否启用"
// @Param limit query int false "每页数量"
// @Param offset query int false "偏移量"
// @Success 200 {object} response.Response{data=object{list=[]ActionTypeDefinitionInfo,total=int}} "返回 list 和 total 字段"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/metadata/action-type-definitions [get]
// @Security BearerAuth
func (h *ActionTypeDefinitionHandler) GetActionTypeDefinitions(c echo.Context) error {
	ctx := c.Request().Context()

	params := interfaces.ActionTypeDefinitionQueryParams{}

	if isActiveStr := c.QueryParam("is_active"); isActiveStr != "" {
		isActive, _ := strconv.ParseBool(isActiveStr)
		params.IsActive = &isActive
	}
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		params.Limit, _ = strconv.Atoi(limitStr)
	}
	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		params.Offset, _ = strconv.Atoi(offsetStr)
	}

	defs, total, err := h.service.GetList(ctx, params)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	result := make([]ActionTypeDefinitionInfo, len(defs))
	for i, def := range defs {
		result[i] = h.convertToInfo(def)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"list":  result,
		"total": total,
	})
}

// GetActionTypeDefinition 获取动作类型定义详情
// @Summary 获取动作类型定义详情
// @Description 根据ID获取动作类型定义的完整信息,包括动作类型、每回合限制、使用时机等配置。
// @Tags 元数据
// @Accept json
// @Produce json
// @Param id path string true "定义ID(UUID)"
// @Success 200 {object} response.Response{data=ActionTypeDefinitionInfo} "动作类型定义详情"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "动作类型定义不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/metadata/action-type-definitions/{id} [get]
// @Security BearerAuth
func (h *ActionTypeDefinitionHandler) GetActionTypeDefinition(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	def, err := h.service.GetByID(ctx, id)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToInfo(def))
}

// GetAllActionTypeDefinitions 获取所有启用的动作类型定义
// @Summary 获取所有启用的动作类型定义
// @Description 获取所有启用状态的动作类型定义列表,不分页,用于动作配置表单的类型选择和规则验证。
// @Tags 元数据
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]ActionTypeDefinitionInfo} "所有启用的动作类型定义"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/metadata/action-type-definitions/all [get]
// @Security BearerAuth
func (h *ActionTypeDefinitionHandler) GetAllActionTypeDefinitions(c echo.Context) error {
	ctx := c.Request().Context()

	defs, err := h.service.GetAll(ctx)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	result := make([]ActionTypeDefinitionInfo, len(defs))
	for i, def := range defs {
		result[i] = h.convertToInfo(def)
	}

	return response.EchoOK(c, h.respWriter, result)
}

func (h *ActionTypeDefinitionHandler) convertToInfo(def *game_config.ActionTypeDefinition) ActionTypeDefinitionInfo {
	info := ActionTypeDefinitionInfo{
		ID:         def.ID,
		ActionType: def.ActionType,
	}

	if def.Description.Valid {
		info.Description = def.Description.String
	}
	if def.PerTurnLimit.Valid {
		info.PerTurnLimit = def.PerTurnLimit.Int
	}
	if def.UsageTiming.Valid {
		info.UsageTiming = def.UsageTiming.String
	}
	if def.Example.Valid {
		info.Example = def.Example.String
	}
	if def.IsActive.Valid {
		info.IsActive = def.IsActive.Bool
	}
	if def.CreatedAt.Valid {
		info.CreatedAt = def.CreatedAt.Time.Unix()
	}

	return info
}
