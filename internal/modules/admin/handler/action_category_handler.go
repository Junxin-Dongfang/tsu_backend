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

// ActionCategoryHandler 动作类别 HTTP 处理器
type ActionCategoryHandler struct {
	service    *service.ActionCategoryService
	respWriter response.Writer
}

// NewActionCategoryHandler 创建动作类别处理器
func NewActionCategoryHandler(db *sql.DB, respWriter response.Writer) *ActionCategoryHandler {
	return &ActionCategoryHandler{
		service:    service.NewActionCategoryService(db),
		respWriter: respWriter,
	}
}

// ==================== HTTP Models ====================

// CreateActionCategoryRequest 创建动作类别请求
type CreateActionCategoryRequest struct {
	CategoryCode string `json:"category_code" validate:"required,max=32" example:"melee_attack"` // 类别代码,唯一标识,最大32字符
	CategoryName string `json:"category_name" validate:"required,max=64" example:"近战攻击"`         // 类别名称,最大64字符
	Description  string `json:"description" example:"近距离物理攻击动作"`                                 // 类别描述
	IsActive     *bool  `json:"is_active" example:"true"`                                        // 是否启用
}

// UpdateActionCategoryRequest 更新动作类别请求
type UpdateActionCategoryRequest struct {
	CategoryCode *string `json:"category_code" validate:"omitempty,max=32" example:"melee_attack"` // 类别代码,唯一标识,最大32字符
	CategoryName *string `json:"category_name" validate:"omitempty,max=64" example:"近战攻击"`         // 类别名称,最大64字符
	Description  *string `json:"description" example:"近距离物理攻击动作"`                                  // 类别描述
	IsActive     *bool   `json:"is_active" example:"true"`                                         // 是否启用
}

// ActionCategoryInfo 动作类别信息响应
type ActionCategoryInfo struct {
	ID           string `json:"id" example:"01d132ed-6378-4e0b-bc16-a5b224e5b04a"` // 动作类别ID,UUID格式
	CategoryCode string `json:"category_code" example:"melee_attack"`              // 类别代码,唯一标识
	CategoryName string `json:"category_name" example:"近战攻击"`                      // 类别名称
	Description  string `json:"description" example:"近距离物理攻击动作"`                   // 类别描述
	IsActive     bool   `json:"is_active" example:"true"`                          // 是否启用
	CreatedAt    int64  `json:"created_at" example:"1759501201"`                   // 创建时间戳(Unix时间戳)
	UpdatedAt    int64  `json:"updated_at" example:"1759501201"`                   // 更新时间戳(Unix时间戳)
}

// ==================== HTTP Handlers ====================

// GetActionCategories 获取动作类别列表
// @Summary 获取动作类别列表
// @Description 获取动作类别列表，支持按启用状态筛选，支持分页
// @Tags 基础配置-动作类别
// @Accept json
// @Produce json
// @Param is_active query bool false "是否启用"
// @Param limit query int false "每页数量(默认20)"
// @Param offset query int false "偏移量(默认0)"
// @Success 200 {object} response.Response{data=object{list=[]ActionCategoryInfo,total=int}} "成功返回动作类别列表和总数"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth
// @Router /admin/action-categories [get]
func (h *ActionCategoryHandler) GetActionCategories(c echo.Context) error {
	ctx := c.Request().Context()

	// 解析查询参数
	params := interfaces.ActionCategoryQueryParams{}

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

	// 查询动作类别列表
	categories, total, err := h.service.GetActionCategories(ctx, params)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为响应格式
	result := make([]ActionCategoryInfo, len(categories))
	for i, category := range categories {
		result[i] = h.convertToActionCategoryInfo(category)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"list":  result,
		"total": total,
	})
}

// GetActionCategory 获取动作类别详情
// @Summary 获取动作类别详情
// @Description 根据动作类别ID获取单个动作类别的详细信息
// @Tags 基础配置-动作类别
// @Accept json
// @Produce json
// @Param id path string true "动作类别ID(UUID格式)"
// @Success 200 {object} response.Response{data=ActionCategoryInfo} "成功返回动作类别详情"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "动作类别不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth
// @Router /admin/action-categories/{id} [get]
func (h *ActionCategoryHandler) GetActionCategory(c echo.Context) error {
	ctx := c.Request().Context()
	categoryID := c.Param("id")

	category, err := h.service.GetActionCategoryByID(ctx, categoryID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToActionCategoryInfo(category))
}

// CreateActionCategory 创建动作类别
// @Summary 创建动作类别
// @Description 创建新的动作类别，类别代码必须唯一
// @Tags 基础配置-动作类别
// @Accept json
// @Produce json
// @Param request body CreateActionCategoryRequest true "创建动作类别请求"
// @Success 200 {object} response.Response{data=ActionCategoryInfo} "成功返回创建的动作类别信息"
// @Failure 400 {object} response.Response "请求参数错误或类别代码已存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth
// @Router /admin/action-categories [post]
func (h *ActionCategoryHandler) CreateActionCategory(c echo.Context) error {
	ctx := c.Request().Context()

	var req CreateActionCategoryRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 创建动作类别实体
	category := &game_config.ActionCategory{
		CategoryCode: req.CategoryCode,
		CategoryName: req.CategoryName,
	}

	if req.Description != "" {
		category.Description.SetValid(req.Description)
	}
	if req.IsActive != nil {
		category.IsActive.SetValid(*req.IsActive)
	}

	// 创建动作类别
	if err := h.service.CreateActionCategory(ctx, category); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToActionCategoryInfo(category))
}

// UpdateActionCategory 更新动作类别
// @Summary 更新动作类别
// @Description 更新指定ID的动作类别信息，支持部分字段更新
// @Tags 基础配置-动作类别
// @Accept json
// @Produce json
// @Param id path string true "动作类别ID(UUID格式)"
// @Param request body UpdateActionCategoryRequest true "更新动作类别请求"
// @Success 200 {object} response.Response "成功返回更新结果"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "动作类别不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth
// @Router /admin/action-categories/{id} [put]
func (h *ActionCategoryHandler) UpdateActionCategory(c echo.Context) error {
	ctx := c.Request().Context()
	categoryID := c.Param("id")

	var req UpdateActionCategoryRequest
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
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	// 更新动作类别
	if err := h.service.UpdateActionCategory(ctx, categoryID, updates); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"category_id": categoryID,
		"message":     "动作类别更新成功",
	})
}

// DeleteActionCategory 删除动作类别
// @Summary 删除动作类别(软删除)
// @Description 软删除指定ID的动作类别，不会真正删除数据
// @Tags 基础配置-动作类别
// @Accept json
// @Produce json
// @Param id path string true "动作类别ID(UUID格式)"
// @Success 200 {object} response.Response "成功返回删除结果"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "动作类别不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth
// @Router /admin/action-categories/{id} [delete]
func (h *ActionCategoryHandler) DeleteActionCategory(c echo.Context) error {
	ctx := c.Request().Context()
	categoryID := c.Param("id")

	if err := h.service.DeleteActionCategory(ctx, categoryID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"category_id": categoryID,
		"message":     "动作类别删除成功",
	})
}

// ==================== Helper Functions ====================

func (h *ActionCategoryHandler) convertToActionCategoryInfo(category *game_config.ActionCategory) ActionCategoryInfo {
	info := ActionCategoryInfo{
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
	if category.IsActive.Valid {
		info.IsActive = category.IsActive.Bool
	}

	return info
}
