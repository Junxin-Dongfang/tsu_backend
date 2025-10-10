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

// SkillCategoryHandler 技能类别 HTTP 处理器
type SkillCategoryHandler struct {
	service    *service.SkillCategoryService
	respWriter response.Writer
}

// NewSkillCategoryHandler 创建技能类别处理器
func NewSkillCategoryHandler(db *sql.DB, respWriter response.Writer) *SkillCategoryHandler {
	return &SkillCategoryHandler{
		service:    service.NewSkillCategoryService(db),
		respWriter: respWriter,
	}
}

// ==================== HTTP Models ====================

// CreateSkillCategoryRequest 创建技能类别请求
type CreateSkillCategoryRequest struct {
	CategoryCode string `json:"category_code" validate:"required,max=32"`
	CategoryName string `json:"category_name" validate:"required,max=64"`
	Description  string `json:"description"`
	Icon         string `json:"icon"`
	Color        string `json:"color"`
	DisplayOrder *int   `json:"display_order"`
	IsActive     *bool  `json:"is_active"`
}

// UpdateSkillCategoryRequest 更新技能类别请求
type UpdateSkillCategoryRequest struct {
	CategoryCode *string `json:"category_code" validate:"omitempty,max=32"`
	CategoryName *string `json:"category_name" validate:"omitempty,max=64"`
	Description  *string `json:"description"`
	Icon         *string `json:"icon"`
	Color        *string `json:"color"`
	DisplayOrder *int    `json:"display_order"`
	IsActive     *bool   `json:"is_active"`
}

// SkillCategoryInfo 技能类别信息响应
type SkillCategoryInfo struct {
	ID           string `json:"id"`
	CategoryCode string `json:"category_code"`
	CategoryName string `json:"category_name"`
	Description  string `json:"description"`
	Icon         string `json:"icon"`
	Color        string `json:"color"`
	DisplayOrder int    `json:"display_order"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}

// ==================== HTTP Handlers ====================

// GetSkillCategories 获取技能类别列表
// @Summary 获取技能类别列表
// @Description 获取技能类别列表，支持按启用状态筛选，支持分页
// @Tags 基础配置-技能类别
// @Accept json
// @Produce json
// @Param is_active query bool false "是否启用"
// @Param limit query int false "每页数量(默认20)"
// @Param offset query int false "偏移量(默认0)"
// @Success 200 {object} response.Response{data=object{list=[]SkillCategoryInfo,total=int}} "成功返回技能类别列表和总数"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/skill-categories [get]
func (h *SkillCategoryHandler) GetSkillCategories(c echo.Context) error {
	ctx := c.Request().Context()

	// 解析查询参数
	params := interfaces.SkillCategoryQueryParams{}

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

	// 查询技能类别列表
	categories, total, err := h.service.GetSkillCategories(ctx, params)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为响应格式
	result := make([]SkillCategoryInfo, len(categories))
	for i, category := range categories {
		result[i] = h.convertToSkillCategoryInfo(category)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"list":  result,
		"total": total,
	})
}

// GetSkillCategory 获取技能类别详情
// @Summary 获取技能类别详情
// @Description 根据技能类别ID获取单个技能类别的详细信息
// @Tags 基础配置-技能类别
// @Accept json
// @Produce json
// @Param id path string true "技能类别ID(UUID格式)"
// @Success 200 {object} response.Response{data=SkillCategoryInfo} "成功返回技能类别详情"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "技能类别不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/skill-categories/{id} [get]
func (h *SkillCategoryHandler) GetSkillCategory(c echo.Context) error {
	ctx := c.Request().Context()
	categoryID := c.Param("id")

	category, err := h.service.GetSkillCategoryByID(ctx, categoryID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToSkillCategoryInfo(category))
}

// CreateSkillCategory 创建技能类别
// @Summary 创建技能类别
// @Description 创建新的技能类别，类别代码必须唯一
// @Tags 基础配置-技能类别
// @Accept json
// @Produce json
// @Param request body CreateSkillCategoryRequest true "创建技能类别请求"
// @Success 200 {object} response.Response{data=SkillCategoryInfo} "成功返回创建的技能类别信息"
// @Failure 400 {object} response.Response "请求参数错误或类别代码已存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/skill-categories [post]
func (h *SkillCategoryHandler) CreateSkillCategory(c echo.Context) error {
	ctx := c.Request().Context()

	var req CreateSkillCategoryRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 创建技能类别实体
	category := &game_config.SkillCategory{
		CategoryCode: req.CategoryCode,
		CategoryName: req.CategoryName,
	}

	if req.Description != "" {
		category.Description.SetValid(req.Description)
	}
	if req.Icon != "" {
		category.Icon.SetValid(req.Icon)
	}
	if req.Color != "" {
		category.Color.SetValid(req.Color)
	}
	if req.DisplayOrder != nil {
		category.DisplayOrder.SetValid(*req.DisplayOrder)
	}
	if req.IsActive != nil {
		category.IsActive.SetValid(*req.IsActive)
	}

	// 创建技能类别
	if err := h.service.CreateSkillCategory(ctx, category); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToSkillCategoryInfo(category))
}

// UpdateSkillCategory 更新技能类别
// @Summary 更新技能类别
// @Description 更新指定ID的技能类别信息，支持部分字段更新
// @Tags 基础配置-技能类别
// @Accept json
// @Produce json
// @Param id path string true "技能类别ID(UUID格式)"
// @Param request body UpdateSkillCategoryRequest true "更新技能类别请求"
// @Success 200 {object} response.Response "成功返回更新结果"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "技能类别不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/skill-categories/{id} [put]
func (h *SkillCategoryHandler) UpdateSkillCategory(c echo.Context) error {
	ctx := c.Request().Context()
	categoryID := c.Param("id")

	var req UpdateSkillCategoryRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	if req.CategoryCode != nil {
		updates["category_code"] = *req.CategoryCode
	}
	if req.CategoryName != nil {
		updates["category_name"] = *req.CategoryName
	}
	if req.Description != nil {
		updates["description"] = *req.Description
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

	// 更新技能类别
	if err := h.service.UpdateSkillCategory(ctx, categoryID, updates); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"category_id": categoryID,
		"message":     "技能类别更新成功",
	})
}

// DeleteSkillCategory 删除技能类别
// @Summary 删除技能类别(软删除)
// @Description 软删除指定ID的技能类别，不会真正删除数据
// @Tags 基础配置-技能类别
// @Accept json
// @Produce json
// @Param id path string true "技能类别ID(UUID格式)"
// @Success 200 {object} response.Response "成功返回删除结果"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "技能类别不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/skill-categories/{id} [delete]
func (h *SkillCategoryHandler) DeleteSkillCategory(c echo.Context) error {
	ctx := c.Request().Context()
	categoryID := c.Param("id")

	if err := h.service.DeleteSkillCategory(ctx, categoryID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"category_id": categoryID,
		"message":     "技能类别删除成功",
	})
}

// ==================== Helper Functions ====================

func (h *SkillCategoryHandler) convertToSkillCategoryInfo(category *game_config.SkillCategory) SkillCategoryInfo {
	info := SkillCategoryInfo{
		ID:           category.ID,
		CategoryCode: category.CategoryCode,
		CategoryName: category.CategoryName,
	}

	if category.CreatedAt.Valid {
		info.CreatedAt = category.CreatedAt.Time.Unix()
	}
	if category.UpdatedAt.Valid {
		info.UpdatedAt = category.UpdatedAt.Time.Unix()
	}
	if category.Description.Valid {
		info.Description = category.Description.String
	}
	if category.Icon.Valid {
		info.Icon = category.Icon.String
	}
	if category.Color.Valid {
		info.Color = category.Color.String
	}
	if category.DisplayOrder.Valid {
		info.DisplayOrder = category.DisplayOrder.Int
	}
	if category.IsActive.Valid {
		info.IsActive = category.IsActive.Bool
	}

	return info
}
