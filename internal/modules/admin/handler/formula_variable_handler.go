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

type FormulaVariableHandler struct {
	service    *service.FormulaVariableService
	respWriter response.Writer
}

func NewFormulaVariableHandler(db *sql.DB, respWriter response.Writer) *FormulaVariableHandler {
	return &FormulaVariableHandler{
		service:    service.NewFormulaVariableService(db),
		respWriter: respWriter,
	}
}

type FormulaVariableInfo struct {
	ID           string `json:"id"`
	VariableCode string `json:"variable_code"`
	VariableName string `json:"variable_name"`
	VariableType string `json:"variable_type"`
	Scope        string `json:"scope"`
	DataType     string `json:"data_type"`
	Description  string `json:"description"`
	Example      string `json:"example"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}

// GetFormulaVariables 获取公式变量列表
// @Summary 获取公式变量列表
// @Description 获取公式变量的分页列表,支持按变量类型、作用域、启用状态筛选。元数据配置表,定义游戏公式中可用的变量及其类型。
// @Tags 元数据配置
// @Accept json
// @Produce json
// @Param variable_type query string false "变量类型(attribute/runtime/config/calculated)"
// @Param scope query string false "作用域(hero/skill/buff/global)"
// @Param is_active query bool false "是否启用"
// @Param limit query int false "每页数量"
// @Param offset query int false "偏移量"
// @Success 200 {object} response.Response{data=map[string]interface{}} "返回 list 和 total 字段"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/metadata/formula-variables [get]
// @Security BearerAuth
func (h *FormulaVariableHandler) GetFormulaVariables(c echo.Context) error {
	ctx := c.Request().Context()

	params := interfaces.FormulaVariableQueryParams{}

	if vType := c.QueryParam("variable_type"); vType != "" {
		params.VariableType = &vType
	}
	if scope := c.QueryParam("scope"); scope != "" {
		params.Scope = &scope
	}
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

	vars, total, err := h.service.GetList(ctx, params)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	result := make([]FormulaVariableInfo, len(vars))
	for i, v := range vars {
		result[i] = h.convertToInfo(v)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"list":  result,
		"total": total,
	})
}

// GetFormulaVariable 获取公式变量详情
// @Summary 获取公式变量详情
// @Description 根据ID获取公式变量的完整信息,包括变量代码、名称、类型、作用域、数据类型等配置。
// @Tags 元数据配置
// @Accept json
// @Produce json
// @Param id path string true "变量ID(UUID)"
// @Success 200 {object} response.Response{data=FormulaVariableInfo} "公式变量详情"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "公式变量不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/metadata/formula-variables/{id} [get]
// @Security BearerAuth
func (h *FormulaVariableHandler) GetFormulaVariable(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	v, err := h.service.GetByID(ctx, id)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToInfo(v))
}

// GetAllFormulaVariables 获取所有启用的公式变量
// @Summary 获取所有启用的公式变量
// @Description 获取所有启用状态的公式变量列表,不分页,用于公式编辑器的变量选择和自动补全功能。
// @Tags 元数据配置
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]FormulaVariableInfo} "所有启用的公式变量"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/metadata/formula-variables/all [get]
// @Security BearerAuth
func (h *FormulaVariableHandler) GetAllFormulaVariables(c echo.Context) error {
	ctx := c.Request().Context()

	vars, err := h.service.GetAll(ctx)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	result := make([]FormulaVariableInfo, len(vars))
	for i, v := range vars {
		result[i] = h.convertToInfo(v)
	}

	return response.EchoOK(c, h.respWriter, result)
}

func (h *FormulaVariableHandler) convertToInfo(v *game_config.FormulaVariable) FormulaVariableInfo {
	info := FormulaVariableInfo{
		ID:           v.ID,
		VariableCode: v.VariableCode,
		VariableName: v.VariableName,
		VariableType: v.VariableType,
		Scope:        v.Scope,
		DataType:     v.DataType,
	}

	if v.Description.Valid {
		info.Description = v.Description.String
	}
	if v.Example.Valid {
		info.Example = v.Example.String
	}
	if v.IsActive.Valid {
		info.IsActive = v.IsActive.Bool
	}
	if v.CreatedAt.Valid {
		info.CreatedAt = v.CreatedAt.Time.Unix()
	}
	if v.UpdatedAt.Valid {
		info.UpdatedAt = v.UpdatedAt.Time.Unix()
	}

	return info
}
