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

type RangeConfigRuleHandler struct {
	service    *service.RangeConfigRuleService
	respWriter response.Writer
}

func NewRangeConfigRuleHandler(db *sql.DB, respWriter response.Writer) *RangeConfigRuleHandler {
	return &RangeConfigRuleHandler{
		service:    service.NewRangeConfigRuleService(db),
		respWriter: respWriter,
	}
}

type RangeConfigRuleInfo struct {
	ID              string `json:"id"`
	ParameterType   string `json:"parameter_type"`
	ParameterFormat string `json:"parameter_format"`
	Description     string `json:"description"`
	Example         string `json:"example"`
	Notes           string `json:"notes"`
	IsActive        bool   `json:"is_active"`
	CreatedAt       int64  `json:"created_at"`
}

// GetRangeConfigRules 获取范围配置规则列表
// @Summary 获取范围配置规则列表
// @Description 获取范围配置规则的分页列表,支持按启用状态筛选。元数据配置表,定义技能/动作射程配置的参数格式和验证规则。
// @Tags 元数据
// @Accept json
// @Produce json
// @Param is_active query bool false "是否启用"
// @Param limit query int false "每页数量"
// @Param offset query int false "偏移量"
// @Success 200 {object} response.Response{data=object{list=[]RangeConfigRuleInfo,total=int}} "返回 list 和 total 字段"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/metadata/range-config-rules [get]
// @Security BearerAuth
func (h *RangeConfigRuleHandler) GetRangeConfigRules(c echo.Context) error {
	ctx := c.Request().Context()

	params := interfaces.RangeConfigRuleQueryParams{}

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

	rules, total, err := h.service.GetList(ctx, params)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	result := make([]RangeConfigRuleInfo, len(rules))
	for i, rule := range rules {
		result[i] = h.convertToInfo(rule)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"list":  result,
		"total": total,
	})
}

// GetRangeConfigRule 获取范围配置规则详情
// @Summary 获取范围配置规则详情
// @Description 根据ID获取范围配置规则的完整信息,包括参数类型、参数格式、示例等配置。
// @Tags 元数据
// @Accept json
// @Produce json
// @Param id path string true "规则ID(UUID)"
// @Success 200 {object} response.Response{data=RangeConfigRuleInfo} "范围配置规则详情"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "范围配置规则不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/metadata/range-config-rules/{id} [get]
// @Security BearerAuth
func (h *RangeConfigRuleHandler) GetRangeConfigRule(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	rule, err := h.service.GetByID(ctx, id)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToInfo(rule))
}

// GetAllRangeConfigRules 获取所有启用的范围配置规则
// @Summary 获取所有启用的范围配置规则
// @Description 获取所有启用状态的范围配置规则列表,不分页,用于射程配置表单的参数验证和格式提示。
// @Tags 元数据
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]RangeConfigRuleInfo} "所有启用的范围配置规则"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/metadata/range-config-rules/all [get]
// @Security BearerAuth
func (h *RangeConfigRuleHandler) GetAllRangeConfigRules(c echo.Context) error {
	ctx := c.Request().Context()

	rules, err := h.service.GetAll(ctx)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	result := make([]RangeConfigRuleInfo, len(rules))
	for i, rule := range rules {
		result[i] = h.convertToInfo(rule)
	}

	return response.EchoOK(c, h.respWriter, result)
}

func (h *RangeConfigRuleHandler) convertToInfo(rule *game_config.RangeConfigRule) RangeConfigRuleInfo {
	info := RangeConfigRuleInfo{
		ID:              rule.ID,
		ParameterType:   rule.ParameterType,
		ParameterFormat: rule.ParameterFormat,
	}

	if rule.Description.Valid {
		info.Description = rule.Description.String
	}
	if rule.Example.Valid {
		info.Example = rule.Example.String
	}
	if rule.Notes.Valid {
		info.Notes = rule.Notes.String
	}
	if rule.IsActive.Valid {
		info.IsActive = rule.IsActive.Bool
	}
	if rule.CreatedAt.Valid {
		info.CreatedAt = rule.CreatedAt.Time.Unix()
	}

	return info
}
