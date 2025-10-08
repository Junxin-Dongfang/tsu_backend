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

// EffectTypeDefinitionHandler 元效果类型定义 HTTP 处理器
type EffectTypeDefinitionHandler struct {
	service    *service.EffectTypeDefinitionService
	respWriter response.Writer
}

// NewEffectTypeDefinitionHandler 创建元效果类型定义处理器
func NewEffectTypeDefinitionHandler(db *sql.DB, respWriter response.Writer) *EffectTypeDefinitionHandler {
	return &EffectTypeDefinitionHandler{
		service:    service.NewEffectTypeDefinitionService(db),
		respWriter: respWriter,
	}
}

// ==================== HTTP Models ====================

// EffectTypeDefinitionInfo 元效果类型定义信息响应
type EffectTypeDefinitionInfo struct {
	ID                    string   `json:"id"`
	EffectTypeCode        string   `json:"effect_type_code"`
	EffectTypeName        string   `json:"effect_type_name"`
	Description           string   `json:"description"`
	ParameterList         []string `json:"parameter_list"`
	ParameterDescriptions string   `json:"parameter_descriptions"`
	FailureHandling       string   `json:"failure_handling"`
	Example               string   `json:"example"`
	Notes                 string   `json:"notes"`
	IsActive              bool     `json:"is_active"`
	CreatedAt             int64    `json:"created_at"`
	UpdatedAt             int64    `json:"updated_at"`
}

// ==================== HTTP Handlers ====================

// GetEffectTypeDefinitions 获取元效果类型定义列表
// @Summary 获取元效果类型定义列表
// @Description 获取元效果类型定义的分页列表,支持按是否启用筛选。元数据配置表,定义游戏中效果类型的参数规范和验证规则。
// @Tags 元数据配置
// @Accept json
// @Produce json
// @Param is_active query bool false "是否启用"
// @Param limit query int false "每页数量"
// @Param offset query int false "偏移量"
// @Success 200 {object} response.Response{data=map[string]interface{}} "返回 list 和 total 字段"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/metadata/effect-type-definitions [get]
// @Security BearerAuth
func (h *EffectTypeDefinitionHandler) GetEffectTypeDefinitions(c echo.Context) error {
	ctx := c.Request().Context()

	// 解析查询参数
	params := interfaces.EffectTypeDefinitionQueryParams{}

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

	// 查询列表
	defs, total, err := h.service.GetList(ctx, params)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为响应格式
	result := make([]EffectTypeDefinitionInfo, len(defs))
	for i, def := range defs {
		result[i] = h.convertToInfo(def)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"list":  result,
		"total": total,
	})
}

// GetEffectTypeDefinition 获取元效果类型定义详情
// @Summary 获取元效果类型定义详情
// @Description 根据ID获取元效果类型定义的完整信息,包括参数列表、失败处理规则等配置。
// @Tags 元数据配置
// @Accept json
// @Produce json
// @Param id path string true "定义ID(UUID)"
// @Success 200 {object} response.Response{data=EffectTypeDefinitionInfo} "元效果类型定义详情"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "元效果类型定义不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/metadata/effect-type-definitions/{id} [get]
// @Security BearerAuth
func (h *EffectTypeDefinitionHandler) GetEffectTypeDefinition(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	def, err := h.service.GetByID(ctx, id)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToInfo(def))
}

// GetAllEffectTypeDefinitions 获取所有启用的元效果类型定义
// @Summary 获取所有启用的元效果类型定义
// @Description 获取所有启用状态的元效果类型定义列表,不分页,用于下拉选择器和配置表单。
// @Tags 元数据配置
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]EffectTypeDefinitionInfo} "所有启用的元效果类型定义"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/metadata/effect-type-definitions/all [get]
// @Security BearerAuth
func (h *EffectTypeDefinitionHandler) GetAllEffectTypeDefinitions(c echo.Context) error {
	ctx := c.Request().Context()

	defs, err := h.service.GetAll(ctx)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为响应格式
	result := make([]EffectTypeDefinitionInfo, len(defs))
	for i, def := range defs {
		result[i] = h.convertToInfo(def)
	}

	return response.EchoOK(c, h.respWriter, result)
}

// ==================== Helper Functions ====================

func (h *EffectTypeDefinitionHandler) convertToInfo(def *game_config.EffectTypeDefinition) EffectTypeDefinitionInfo {
	info := EffectTypeDefinitionInfo{
		ID:             def.ID,
		EffectTypeCode: def.EffectTypeCode,
		EffectTypeName: def.EffectTypeName,
	}

	if def.Description.Valid {
		info.Description = def.Description.String
	}

	// ParameterList 是 types.StringArray 类型
	if len(def.ParameterList) > 0 {
		info.ParameterList = def.ParameterList
	} else {
		info.ParameterList = []string{}
	}

	if def.ParameterDescriptions.Valid {
		info.ParameterDescriptions = def.ParameterDescriptions.String
	}

	if def.FailureHandling.Valid {
		info.FailureHandling = def.FailureHandling.String
	}

	if def.Example.Valid {
		info.Example = def.Example.String
	}

	if def.Notes.Valid {
		info.Notes = def.Notes.String
	}

	if def.IsActive.Valid {
		info.IsActive = def.IsActive.Bool
	}

	if def.CreatedAt.Valid {
		info.CreatedAt = def.CreatedAt.Time.Unix()
	}

	if def.UpdatedAt.Valid {
		info.UpdatedAt = def.UpdatedAt.Time.Unix()
	}

	return info
}
