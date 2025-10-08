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

// DamageTypeHandler 伤害类型 HTTP 处理器
type DamageTypeHandler struct {
	service    *service.DamageTypeService
	respWriter response.Writer
}

// NewDamageTypeHandler 创建伤害类型处理器
func NewDamageTypeHandler(db *sql.DB, respWriter response.Writer) *DamageTypeHandler {
	return &DamageTypeHandler{
		service:    service.NewDamageTypeService(db),
		respWriter: respWriter,
	}
}

// ==================== HTTP Models ====================

// CreateDamageTypeRequest 创建伤害类型请求
type CreateDamageTypeRequest struct {
	Code                         string  `json:"code" validate:"required,max=32" example:"fire"`           // 伤害类型代码(必需,最多32字符)
	Name                         string  `json:"name" validate:"required,max=64" example:"火焰"`             // 伤害类型名称(必需,最多64字符)
	Category                     *string `json:"category" example:"elemental"`                             // 伤害类别(可选)
	ResistanceAttributeCode      *string `json:"resistance_attribute_code" example:"fire_resistance"`      // 抗性属性代码(可选)
	DamageReductionAttributeCode *string `json:"damage_reduction_attribute_code" example:"fire_reduction"` // 伤害减免属性代码(可选)
	ResistanceCap                *int    `json:"resistance_cap" example:"75"`                              // 抗性上限(可选)
	Color                        *string `json:"color" example:"#FF4500"`                                  // 颜色(可选)
	Icon                         *string `json:"icon" example:"fire_icon.png"`                             // 图标(可选)
	Description                  *string `json:"description" example:"火焰伤害"`                               // 描述(可选)
	IsActive                     *bool   `json:"is_active" example:"true"`                                 // 是否启用(可选)
}

// UpdateDamageTypeRequest 更新伤害类型请求
type UpdateDamageTypeRequest struct {
	Code                         *string `json:"code" validate:"omitempty,max=32"`
	Name                         *string `json:"name" validate:"omitempty,max=64"`
	Category                     *string `json:"category"`
	ResistanceAttributeCode      *string `json:"resistance_attribute_code"`
	DamageReductionAttributeCode *string `json:"damage_reduction_attribute_code"`
	ResistanceCap                *int    `json:"resistance_cap"`
	Color                        *string `json:"color"`
	Icon                         *string `json:"icon"`
	Description                  *string `json:"description"`
	IsActive                     *bool   `json:"is_active"`
}

// DamageTypeInfo 伤害类型信息响应
type DamageTypeInfo struct {
	ID                           string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`        // 伤害类型ID
	Code                         string `json:"code" example:"fire"`                                      // 伤害类型代码
	Name                         string `json:"name" example:"火焰"`                                        // 伤害类型名称
	Category                     string `json:"category" example:"elemental"`                             // 伤害类别
	ResistanceAttributeCode      string `json:"resistance_attribute_code" example:"fire_resistance"`      // 抗性属性代码
	DamageReductionAttributeCode string `json:"damage_reduction_attribute_code" example:"fire_reduction"` // 伤害减免属性代码
	ResistanceCap                int    `json:"resistance_cap" example:"75"`                              // 抗性上限
	Color                        string `json:"color" example:"#FF4500"`                                  // 颜色
	Icon                         string `json:"icon" example:"fire_icon.png"`                             // 图标
	Description                  string `json:"description" example:"火焰伤害"`                               // 描述
	IsActive                     bool   `json:"is_active" example:"true"`                                 // 是否启用
	CreatedAt                    int64  `json:"created_at" example:"1633024800"`                          // 创建时间戳
	UpdatedAt                    int64  `json:"updated_at" example:"1633024800"`                          // 更新时间戳
}

// ==================== HTTP Handlers ====================

// GetDamageTypes 获取伤害类型列表
// @Summary 获取伤害类型列表
// @Description 获取伤害类型列表，支持按类别和启用状态筛选，支持分页
// @Tags 基础配置-伤害类型
// @Accept json
// @Produce json
// @Param category query string false "伤害类别"
// @Param is_active query bool false "是否启用"
// @Param limit query int false "每页数量(默认20)"
// @Param offset query int false "偏移量(默认0)"
// @Success 200 {object} response.Response{data=map[string]interface{}} "成功返回伤害类型列表和总数"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/damage-types [get]
func (h *DamageTypeHandler) GetDamageTypes(c echo.Context) error {
	ctx := c.Request().Context()

	// 解析查询参数
	params := interfaces.DamageTypeQueryParams{}

	if category := c.QueryParam("category"); category != "" {
		params.Category = &category
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

	// 查询伤害类型列表
	damageTypes, total, err := h.service.GetDamageTypes(ctx, params)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为响应格式
	result := make([]DamageTypeInfo, len(damageTypes))
	for i, damageType := range damageTypes {
		result[i] = h.convertToDamageTypeInfo(damageType)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"list":  result,
		"total": total,
	})
}

// GetDamageType 获取伤害类型详情
// @Summary 获取伤害类型详情
// @Description 根据伤害类型ID获取单个伤害类型的详细信息
// @Tags 基础配置-伤害类型
// @Accept json
// @Produce json
// @Param id path string true "伤害类型ID(UUID格式)"
// @Success 200 {object} response.Response{data=DamageTypeInfo} "成功返回伤害类型详情"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "伤害类型不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/damage-types/{id} [get]
func (h *DamageTypeHandler) GetDamageType(c echo.Context) error {
	ctx := c.Request().Context()
	damageTypeID := c.Param("id")

	damageType, err := h.service.GetDamageTypeByID(ctx, damageTypeID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToDamageTypeInfo(damageType))
}

// CreateDamageType 创建伤害类型
// @Summary 创建伤害类型
// @Description 创建新的伤害类型，伤害类型代码必须唯一
// @Tags 基础配置-伤害类型
// @Accept json
// @Produce json
// @Param request body CreateDamageTypeRequest true "创建伤害类型请求"
// @Success 200 {object} response.Response{data=DamageTypeInfo} "成功返回创建的伤害类型信息"
// @Failure 400 {object} response.Response "请求参数错误或伤害类型代码已存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/damage-types [post]
func (h *DamageTypeHandler) CreateDamageType(c echo.Context) error {
	ctx := c.Request().Context()

	var req CreateDamageTypeRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 创建伤害类型实体
	damageType := &game_config.DamageType{
		Code: req.Code,
		Name: req.Name,
	}

	if req.Category != nil && *req.Category != "" {
		damageType.Category.SetValid(*req.Category)
	}
	if req.ResistanceAttributeCode != nil && *req.ResistanceAttributeCode != "" {
		damageType.ResistanceAttributeCode.SetValid(*req.ResistanceAttributeCode)
	}
	if req.DamageReductionAttributeCode != nil && *req.DamageReductionAttributeCode != "" {
		damageType.DamageReductionAttributeCode.SetValid(*req.DamageReductionAttributeCode)
	}
	if req.ResistanceCap != nil {
		damageType.ResistanceCap.SetValid(*req.ResistanceCap)
	}
	if req.Color != nil && *req.Color != "" {
		damageType.Color.SetValid(*req.Color)
	}
	if req.Icon != nil && *req.Icon != "" {
		damageType.Icon.SetValid(*req.Icon)
	}
	if req.Description != nil && *req.Description != "" {
		damageType.Description.SetValid(*req.Description)
	}
	if req.IsActive != nil {
		damageType.IsActive.SetValid(*req.IsActive)
	}

	// 创建伤害类型
	if err := h.service.CreateDamageType(ctx, damageType); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToDamageTypeInfo(damageType))
}

// UpdateDamageType 更新伤害类型
// @Summary 更新伤害类型
// @Description 更新指定ID的伤害类型信息，支持部分字段更新
// @Tags 基础配置-伤害类型
// @Accept json
// @Produce json
// @Param id path string true "伤害类型ID(UUID格式)"
// @Param request body UpdateDamageTypeRequest true "更新伤害类型请求"
// @Success 200 {object} response.Response "成功返回更新结果"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "伤害类型不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/damage-types/{id} [put]
func (h *DamageTypeHandler) UpdateDamageType(c echo.Context) error {
	ctx := c.Request().Context()
	damageTypeID := c.Param("id")

	var req UpdateDamageTypeRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	if req.Code != nil {
		updates["code"] = *req.Code
	}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.ResistanceAttributeCode != nil {
		updates["resistance_attribute_code"] = *req.ResistanceAttributeCode
	}
	if req.DamageReductionAttributeCode != nil {
		updates["damage_reduction_attribute_code"] = *req.DamageReductionAttributeCode
	}
	if req.ResistanceCap != nil {
		updates["resistance_cap"] = *req.ResistanceCap
	}
	if req.Color != nil {
		updates["color"] = *req.Color
	}
	if req.Icon != nil {
		updates["icon"] = *req.Icon
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	// 更新伤害类型
	if err := h.service.UpdateDamageType(ctx, damageTypeID, updates); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"damage_type_id": damageTypeID,
		"message":        "伤害类型更新成功",
	})
}

// DeleteDamageType 删除伤害类型
// @Summary 删除伤害类型(软删除)
// @Description 软删除指定ID的伤害类型，不会真正删除数据
// @Tags 基础配置-伤害类型
// @Accept json
// @Produce json
// @Param id path string true "伤害类型ID(UUID格式)"
// @Success 200 {object} response.Response "成功返回删除结果"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "伤害类型不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/damage-types/{id} [delete]
func (h *DamageTypeHandler) DeleteDamageType(c echo.Context) error {
	ctx := c.Request().Context()
	damageTypeID := c.Param("id")

	if err := h.service.DeleteDamageType(ctx, damageTypeID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"damage_type_id": damageTypeID,
		"message":        "伤害类型删除成功",
	})
}

// ==================== Helper Functions ====================

func (h *DamageTypeHandler) convertToDamageTypeInfo(damageType *game_config.DamageType) DamageTypeInfo {
	info := DamageTypeInfo{
		ID:   damageType.ID,
		Code: damageType.Code,
		Name: damageType.Name,
	}

	if damageType.CreatedAt.Valid {
		info.CreatedAt = damageType.CreatedAt.Time.Unix()
	}
	if damageType.UpdatedAt.Valid {
		info.UpdatedAt = damageType.UpdatedAt.Time.Unix()
	}
	if damageType.Category.Valid {
		info.Category = damageType.Category.String
	}
	if damageType.ResistanceAttributeCode.Valid {
		info.ResistanceAttributeCode = damageType.ResistanceAttributeCode.String
	}
	if damageType.DamageReductionAttributeCode.Valid {
		info.DamageReductionAttributeCode = damageType.DamageReductionAttributeCode.String
	}
	if damageType.ResistanceCap.Valid {
		info.ResistanceCap = damageType.ResistanceCap.Int
	}
	if damageType.Color.Valid {
		info.Color = damageType.Color.String
	}
	if damageType.Icon.Valid {
		info.Icon = damageType.Icon.String
	}
	if damageType.Description.Valid {
		info.Description = damageType.Description.String
	}
	if damageType.IsActive.Valid {
		info.IsActive = damageType.IsActive.Bool
	}

	return info
}
