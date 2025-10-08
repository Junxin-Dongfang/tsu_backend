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

// TagHandler 标签 HTTP 处理器
type TagHandler struct {
	service    *service.TagService
	respWriter response.Writer
}

// NewTagHandler 创建标签处理器
func NewTagHandler(db *sql.DB, respWriter response.Writer) *TagHandler {
	return &TagHandler{
		service:    service.NewTagService(db),
		respWriter: respWriter,
	}
}

// ==================== HTTP Models ====================

// CreateTagRequest 创建标签请求
type CreateTagRequest struct {
	TagCode      string `json:"tag_code" validate:"required,max=32"`
	TagName      string `json:"tag_name" validate:"required,max=64"`
	Category     string `json:"category" validate:"required,oneof=class item skill monster"`
	Description  string `json:"description"`
	Icon         string `json:"icon"`
	Color        string `json:"color"`
	DisplayOrder *int   `json:"display_order"`
	IsActive     *bool  `json:"is_active"`
}

// UpdateTagRequest 更新标签请求
type UpdateTagRequest struct {
	TagCode      *string `json:"tag_code" validate:"omitempty,max=32"`
	TagName      *string `json:"tag_name" validate:"omitempty,max=64"`
	Category     *string `json:"category" validate:"omitempty,oneof=class item skill monster"`
	Description  *string `json:"description"`
	Icon         *string `json:"icon"`
	Color        *string `json:"color"`
	DisplayOrder *int    `json:"display_order"`
	IsActive     *bool   `json:"is_active"`
}

// TagInfo 标签信息响应
type TagInfo struct {
	ID           string `json:"id"`
	TagCode      string `json:"tag_code"`
	TagName      string `json:"tag_name"`
	Category     string `json:"category"`
	Description  string `json:"description"`
	Icon         string `json:"icon"`
	Color        string `json:"color"`
	DisplayOrder int    `json:"display_order"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}

// ==================== HTTP Handlers ====================

// GetTags 获取标签列表
// @Summary 获取标签列表
// @Description 获取标签列表，支持按类别和启用状态筛选，支持分页
// @Tags 基础配置-标签
// @Accept json
// @Produce json
// @Param category query string false "标签类别(class/item/skill/monster)"
// @Param is_active query bool false "是否启用"
// @Param limit query int false "每页数量(默认20)"
// @Param offset query int false "偏移量(默认0)"
// @Success 200 {object} response.Response{data=map[string]interface{}} "成功返回标签列表和总数"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/tags [get]
func (h *TagHandler) GetTags(c echo.Context) error {
	ctx := c.Request().Context()

	// 解析查询参数
	params := interfaces.TagQueryParams{}

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

	// 查询标签列表
	tags, total, err := h.service.GetTags(ctx, params)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为响应格式
	result := make([]TagInfo, len(tags))
	for i, tag := range tags {
		result[i] = h.convertToTagInfo(tag)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"list":  result,
		"total": total,
	})
}

// GetTag 获取标签详情
// @Summary 获取标签详情
// @Description 根据标签ID获取单个标签的详细信息
// @Tags 基础配置-标签
// @Accept json
// @Produce json
// @Param id path string true "标签ID"
// @Success 200 {object} response.Response{data=TagInfo} "成功返回标签详情"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "标签不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/tags/{id} [get]
func (h *TagHandler) GetTag(c echo.Context) error {
	ctx := c.Request().Context()
	tagID := c.Param("id")

	tag, err := h.service.GetTagByID(ctx, tagID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToTagInfo(tag))
}

// CreateTag 创建标签
// @Summary 创建标签
// @Description 创建新的标签，标签代码必须唯一
// @Tags 基础配置-标签
// @Accept json
// @Produce json
// @Param request body CreateTagRequest true "创建标签请求"
// @Success 200 {object} response.Response{data=TagInfo} "成功返回创建的标签信息"
// @Failure 400 {object} response.Response "请求参数错误或标签代码已存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/tags [post]
func (h *TagHandler) CreateTag(c echo.Context) error {
	ctx := c.Request().Context()

	var req CreateTagRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 创建标签实体
	tag := &game_config.Tag{
		TagCode:  req.TagCode,
		TagName:  req.TagName,
		Category: req.Category,
	}

	if req.Description != "" {
		tag.Description.SetValid(req.Description)
	}
	if req.Icon != "" {
		tag.Icon.SetValid(req.Icon)
	}
	if req.Color != "" {
		tag.Color.SetValid(req.Color)
	}
	if req.DisplayOrder != nil {
		tag.DisplayOrder = *req.DisplayOrder
	}
	if req.IsActive != nil {
		tag.IsActive = *req.IsActive
	}

	// 创建标签
	if err := h.service.CreateTag(ctx, tag); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToTagInfo(tag))
}

// UpdateTag 更新标签
// @Summary 更新标签
// @Description 更新指定ID的标签信息，支持部分字段更新
// @Tags 基础配置-标签
// @Accept json
// @Produce json
// @Param id path string true "标签ID"
// @Param request body UpdateTagRequest true "更新标签请求"
// @Success 200 {object} response.Response "成功返回更新结果"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "标签不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/tags/{id} [put]
func (h *TagHandler) UpdateTag(c echo.Context) error {
	ctx := c.Request().Context()
	tagID := c.Param("id")

	var req UpdateTagRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	if req.TagCode != nil {
		updates["tag_code"] = *req.TagCode
	}
	if req.TagName != nil {
		updates["tag_name"] = *req.TagName
	}
	if req.Category != nil {
		updates["category"] = *req.Category
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

	// 更新标签
	if err := h.service.UpdateTag(ctx, tagID, updates); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"tag_id":  tagID,
		"message": "标签更新成功",
	})
}

// DeleteTag 删除标签
// @Summary 删除标签(软删除)
// @Description 软删除指定ID的标签，不会真正删除数据
// @Tags 基础配置-标签
// @Accept json
// @Produce json
// @Param id path string true "标签ID"
// @Success 200 {object} response.Response "成功返回删除结果"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "标签不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/tags/{id} [delete]
func (h *TagHandler) DeleteTag(c echo.Context) error {
	ctx := c.Request().Context()
	tagID := c.Param("id")

	if err := h.service.DeleteTag(ctx, tagID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"tag_id":  tagID,
		"message": "标签删除成功",
	})
}

// ==================== Helper Functions ====================

func (h *TagHandler) convertToTagInfo(tag *game_config.Tag) TagInfo {
	info := TagInfo{
		ID:           tag.ID,
		TagCode:      tag.TagCode,
		TagName:      tag.TagName,
		Category:     tag.Category,
		DisplayOrder: tag.DisplayOrder,
		IsActive:     tag.IsActive,
	}

	// CreatedAt 和 UpdatedAt 是 time.Time 类型，不需要判断 Valid
	if !tag.CreatedAt.IsZero() {
		info.CreatedAt = tag.CreatedAt.Unix()
	}
	if !tag.UpdatedAt.IsZero() {
		info.UpdatedAt = tag.UpdatedAt.Unix()
	}

	// null.String 类型需要判断 Valid
	if tag.Description.Valid {
		info.Description = tag.Description.String
	}
	if tag.Icon.Valid {
		info.Icon = tag.Icon.String
	}
	if tag.Color.Valid {
		info.Color = tag.Color.String
	}

	return info
}
