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

// HeroAttributeTypeHandler 属性类型 HTTP 处理器
type HeroAttributeTypeHandler struct {
	service    *service.HeroAttributeTypeService
	respWriter response.Writer
}

// NewHeroAttributeTypeHandler 创建属性类型处理器
func NewHeroAttributeTypeHandler(db *sql.DB, respWriter response.Writer) *HeroAttributeTypeHandler {
	return &HeroAttributeTypeHandler{
		service:    service.NewHeroAttributeTypeService(db),
		respWriter: respWriter,
	}
}

// ==================== HTTP Models ====================

// CreateHeroAttributeTypeRequest 创建属性类型请求
type CreateHeroAttributeTypeRequest struct {
	AttributeCode string  `json:"attribute_code" validate:"required,max=32" example:"strength"`                             // 属性代码(必需,最多32字符)
	AttributeName string  `json:"attribute_name" validate:"required,max=64" example:"力量"`                                   // 属性名称(必需,最多64字符)
	Category      string  `json:"category" validate:"required,oneof=basic derived resistance" example:"basic"`              // 属性类别(必需,basic=基础属性/derived=派生属性/resistance=抗性属性)
	DataType      string  `json:"data_type" validate:"required,oneof=integer decimal percentage boolean" example:"integer"` // 数据类型(必需,integer=整数/decimal=小数/percentage=百分比/boolean=布尔)
	Icon          *string `json:"icon" example:"strength_icon.png"`                                                         // 图标(可选)
	Color         *string `json:"color" example:"#FF0000"`                                                                  // 颜色(可选)
	Unit          *string `json:"unit" example:"点"`                                                                         // 单位(可选)
	DisplayOrder  *int    `json:"display_order" example:"1"`                                                                // 显示顺序(可选)
	IsActive      *bool   `json:"is_active" example:"true"`                                                                 // 是否启用(可选)
	IsVisible     *bool   `json:"is_visible" example:"true"`                                                                // 是否可见(可选)
	Description   *string `json:"description" example:"角色的力量属性"`                                                            // 描述(可选)
}

// UpdateHeroAttributeTypeRequest 更新属性类型请求
type UpdateHeroAttributeTypeRequest struct {
	AttributeCode *string `json:"attribute_code" validate:"omitempty,max=32"`
	AttributeName *string `json:"attribute_name" validate:"omitempty,max=64"`
	Category      *string `json:"category" validate:"omitempty,oneof=basic derived resistance"` // 属性类别(可选,basic=基础属性/derived=派生属性/resistance=抗性属性)
	Icon          *string `json:"icon"`
	Color         *string `json:"color"`
	DisplayOrder  *int    `json:"display_order"`
	IsActive      *bool   `json:"is_active"`
	IsVisible     *bool   `json:"is_visible"`
	Description   *string `json:"description"`
}

// HeroAttributeTypeInfo 属性类型信息响应
type HeroAttributeTypeInfo struct {
	ID            string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"` // 属性类型ID
	AttributeCode string `json:"attribute_code" example:"strength"`                 // 属性代码
	AttributeName string `json:"attribute_name" example:"力量"`                       // 属性名称
	Category      string `json:"category" example:"primary"`                        // 属性类别
	DataType      string `json:"data_type" example:"integer"`                       // 数据类型
	Icon          string `json:"icon" example:"strength_icon.png"`                  // 图标
	Color         string `json:"color" example:"#FF0000"`                           // 颜色
	Unit          string `json:"unit" example:"点"`                                  // 单位
	DisplayOrder  int    `json:"display_order" example:"1"`                         // 显示顺序
	IsActive      bool   `json:"is_active" example:"true"`                          // 是否启用
	IsVisible     bool   `json:"is_visible" example:"true"`                         // 是否可见
	Description   string `json:"description" example:"角色的力量属性"`                     // 描述
	CreatedAt     int64  `json:"created_at" example:"1633024800"`                   // 创建时间戳
	UpdatedAt     int64  `json:"updated_at" example:"1633024800"`                   // 更新时间戳
}

// ==================== HTTP Handlers ====================

// GetHeroAttributeTypes 获取属性类型列表
// @Summary 获取属性类型列表
// @Description 获取英雄属性类型列表，支持按类别、启用状态和可见性筛选，支持分页
// @Tags 属性类型
// @Accept json
// @Produce json
// @Param category query string false "属性类别" Enums(basic, derived, resistance)
// @Param is_active query bool false "是否启用"
// @Param is_visible query bool false "是否可见"
// @Param limit query int false "每页数量(默认20)"
// @Param offset query int false "偏移量(默认0)"
// @Success 200 {object} response.Response{data=object{list=[]HeroAttributeTypeInfo,total=int}} "成功返回属性类型列表和总数"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/hero-attribute-types [get]
func (h *HeroAttributeTypeHandler) GetHeroAttributeTypes(c echo.Context) error {
	ctx := c.Request().Context()

	// 解析查询参数
	params := interfaces.HeroAttributeTypeQueryParams{}

	if category := c.QueryParam("category"); category != "" {
		params.Category = &category
	}

	if isActiveStr := c.QueryParam("is_active"); isActiveStr != "" {
		isActive, _ := strconv.ParseBool(isActiveStr)
		params.IsActive = &isActive
	}

	if isVisibleStr := c.QueryParam("is_visible"); isVisibleStr != "" {
		isVisible, _ := strconv.ParseBool(isVisibleStr)
		params.IsVisible = &isVisible
	}

	if limitStr := c.QueryParam("limit"); limitStr != "" {
		params.Limit, _ = strconv.Atoi(limitStr)
	}

	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		params.Offset, _ = strconv.Atoi(offsetStr)
	}

	// 查询属性类型列表
	attributeTypes, total, err := h.service.GetHeroAttributeTypes(ctx, params)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为响应格式
	result := make([]HeroAttributeTypeInfo, len(attributeTypes))
	for i, attributeType := range attributeTypes {
		result[i] = h.convertToHeroAttributeTypeInfo(attributeType)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"list":  result,
		"total": total,
	})
}

// GetHeroAttributeType 获取属性类型详情
// @Summary 获取属性类型详情
// @Description 根据属性类型ID获取单个英雄属性类型的详细信息
// @Tags 属性类型
// @Accept json
// @Produce json
// @Param id path string true "属性类型ID(UUID格式)"
// @Success 200 {object} response.Response{data=HeroAttributeTypeInfo} "成功返回属性类型详情"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "属性类型不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/hero-attribute-types/{id} [get]
func (h *HeroAttributeTypeHandler) GetHeroAttributeType(c echo.Context) error {
	ctx := c.Request().Context()
	attributeTypeID := c.Param("id")

	attributeType, err := h.service.GetHeroAttributeTypeByID(ctx, attributeTypeID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToHeroAttributeTypeInfo(attributeType))
}

// CreateHeroAttributeType 创建属性类型
// @Summary 创建属性类型
// @Description 创建新的英雄属性类型，属性代码必须唯一
// @Tags 属性类型
// @Accept json
// @Produce json
// @Param request body CreateHeroAttributeTypeRequest true "创建属性类型请求"
// @Success 200 {object} response.Response{data=HeroAttributeTypeInfo} "成功返回创建的属性类型信息"
// @Failure 400 {object} response.Response "请求参数错误或属性代码已存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/hero-attribute-types [post]
func (h *HeroAttributeTypeHandler) CreateHeroAttributeType(c echo.Context) error {
	ctx := c.Request().Context()

	var req CreateHeroAttributeTypeRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 创建属性类型实体
	attributeType := &game_config.HeroAttributeType{
		AttributeCode: req.AttributeCode,
		AttributeName: req.AttributeName,
		Category:      req.Category,
		DataType:      req.DataType,
		IsActive:      true, // 默认启用
		IsVisible:     true, // 默认可见
		DisplayOrder:  0,    // 默认排序
	}

	if req.Icon != nil && *req.Icon != "" {
		attributeType.Icon.SetValid(*req.Icon)
	}
	if req.Color != nil && *req.Color != "" {
		attributeType.Color.SetValid(*req.Color)
	}
	if req.Unit != nil && *req.Unit != "" {
		attributeType.Unit.SetValid(*req.Unit)
	}
	if req.DisplayOrder != nil {
		attributeType.DisplayOrder = *req.DisplayOrder
	}
	if req.IsActive != nil {
		attributeType.IsActive = *req.IsActive
	}
	if req.IsVisible != nil {
		attributeType.IsVisible = *req.IsVisible
	}
	if req.Description != nil && *req.Description != "" {
		attributeType.Description.SetValid(*req.Description)
	}

	// 创建属性类型
	if err := h.service.CreateHeroAttributeType(ctx, attributeType); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToHeroAttributeTypeInfo(attributeType))
}

// UpdateHeroAttributeType 更新属性类型
// @Summary 更新属性类型
// @Description 更新指定ID的英雄属性类型信息，支持部分字段更新
// @Tags 属性类型
// @Accept json
// @Produce json
// @Param id path string true "属性类型ID(UUID格式)"
// @Param request body UpdateHeroAttributeTypeRequest true "更新属性类型请求"
// @Success 200 {object} response.Response "成功返回更新结果"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "属性类型不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/hero-attribute-types/{id} [put]
func (h *HeroAttributeTypeHandler) UpdateHeroAttributeType(c echo.Context) error {
	ctx := c.Request().Context()
	attributeTypeID := c.Param("id")

	var req UpdateHeroAttributeTypeRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	if req.AttributeCode != nil {
		updates["attribute_code"] = *req.AttributeCode
	}
	if req.AttributeName != nil {
		updates["attribute_name"] = *req.AttributeName
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.Icon != nil {
		updates["icon"] = *req.Icon
	}
	if req.Color != nil {
		updates["color"] = *req.Color
	}
	if req.DisplayOrder != nil {
		updates["display_order"] = *req.DisplayOrder
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.IsVisible != nil {
		updates["is_visible"] = *req.IsVisible
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}

	// 更新属性类型
	if err := h.service.UpdateHeroAttributeType(ctx, attributeTypeID, updates); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"attribute_type_id": attributeTypeID,
		"message":           "属性类型更新成功",
	})
}

// DeleteHeroAttributeType 删除属性类型
// @Summary 删除属性类型(软删除)
// @Description 软删除指定ID的英雄属性类型，不会真正删除数据
// @Tags 属性类型
// @Accept json
// @Produce json
// @Param id path string true "属性类型ID(UUID格式)"
// @Success 200 {object} response.Response "成功返回删除结果"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "属性类型不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/hero-attribute-types/{id} [delete]
func (h *HeroAttributeTypeHandler) DeleteHeroAttributeType(c echo.Context) error {
	ctx := c.Request().Context()
	attributeTypeID := c.Param("id")

	if err := h.service.DeleteHeroAttributeType(ctx, attributeTypeID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"attribute_type_id": attributeTypeID,
		"message":           "属性类型删除成功",
	})
}

// ==================== Helper Functions ====================

func (h *HeroAttributeTypeHandler) convertToHeroAttributeTypeInfo(attributeType *game_config.HeroAttributeType) HeroAttributeTypeInfo {
	info := HeroAttributeTypeInfo{
		ID:            attributeType.ID,
		AttributeCode: attributeType.AttributeCode,
		AttributeName: attributeType.AttributeName,
		Category:      attributeType.Category,
		DataType:      attributeType.DataType,
		DisplayOrder:  attributeType.DisplayOrder,
		IsActive:      attributeType.IsActive,
		IsVisible:     attributeType.IsVisible,
		CreatedAt:     attributeType.CreatedAt.Unix(),
		UpdatedAt:     attributeType.UpdatedAt.Unix(),
	}

	if attributeType.Icon.Valid {
		info.Icon = attributeType.Icon.String
	}
	if attributeType.Color.Valid {
		info.Color = attributeType.Color.String
	}
	if attributeType.Unit.Valid {
		info.Unit = attributeType.Unit.String
	}
	if attributeType.Description.Valid {
		info.Description = attributeType.Description.String
	}

	return info
}
