package handler

import (
	"database/sql"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/modules/admin/service"
	"tsu-self/internal/pkg/response"
)

// TagRelationHandler 标签关联 HTTP 处理器
type TagRelationHandler struct {
	service    *service.TagRelationService
	respWriter response.Writer
}

// NewTagRelationHandler 创建标签关联处理器
func NewTagRelationHandler(db *sql.DB, respWriter response.Writer) *TagRelationHandler {
	return &TagRelationHandler{
		service:    service.NewTagRelationService(db),
		respWriter: respWriter,
	}
}

// ==================== HTTP Models ====================

// AddTagToEntityRequest 为实体添加标签请求
type AddTagToEntityRequest struct {
	TagID string `json:"tag_id" validate:"required,uuid"`
}

// BatchSetEntityTagsRequest 批量设置实体标签请求
type BatchSetEntityTagsRequest struct {
	TagIDs []string `json:"tag_ids" validate:"required,dive,uuid"`
}

// EntityTagInfo 实体标签信息
type EntityTagInfo struct {
	ID           string `json:"id"`
	TagCode      string `json:"tag_code"`
	TagName      string `json:"tag_name"`
	Category     string `json:"category"`
	Description  string `json:"description"`
	Icon         string `json:"icon"`
	Color        string `json:"color"`
	DisplayOrder int    `json:"display_order"`
}

// TagEntityInfo 标签实体信息
type TagEntityInfo struct {
	ID         string `json:"id"`
	EntityType string `json:"entity_type"`
	EntityID   string `json:"entity_id"`
	CreatedAt  int64  `json:"created_at"`
}

// ==================== HTTP Handlers ====================

// GetEntityTags 获取实体的所有标签
// @Summary 获取实体的所有标签
// @Description 获取指定实体的所有关联标签，返回完整的标签信息
// @Tags 基础配置-标签关联
// @Accept json
// @Produce json
// @Param entity_type path string true "实体类型(skill/class/item/monster)"
// @Param entity_id path string true "实体ID(UUID格式)"
// @Success 200 {object} response.Response{data=[]EntityTagInfo} "成功返回标签列表"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "实体不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/entities/{entity_type}/{entity_id}/tags [get]
func (h *TagRelationHandler) GetEntityTags(c echo.Context) error {
	ctx := c.Request().Context()
	entityType := c.Param("entity_type")
	entityID := c.Param("entity_id")

	tags, err := h.service.GetEntityTags(ctx, entityType, entityID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为响应格式
	result := make([]EntityTagInfo, len(tags))
	for i, tag := range tags {
		info := EntityTagInfo{
			ID:           tag.ID,
			TagCode:      tag.TagCode,
			TagName:      tag.TagName,
			Category:     tag.Category,
			DisplayOrder: tag.DisplayOrder,
		}

		if tag.Description.Valid {
			info.Description = tag.Description.String
		}
		if tag.Icon.Valid {
			info.Icon = tag.Icon.String
		}
		if tag.Color.Valid {
			info.Color = tag.Color.String
		}

		result[i] = info
	}

	return response.EchoOK(c, h.respWriter, result)
}

// GetTagEntities 获取使用某个标签的所有实体
// @Summary 获取使用某个标签的所有实体
// @Description 查询使用指定标签的所有实体，返回实体类型和ID列表
// @Tags 基础配置-标签关联
// @Accept json
// @Produce json
// @Param tag_id path string true "标签ID(UUID格式)"
// @Success 200 {object} response.Response{data=[]TagEntityInfo} "成功返回实体列表"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "标签不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/tags/{tag_id}/entities [get]
func (h *TagRelationHandler) GetTagEntities(c echo.Context) error {
	ctx := c.Request().Context()
	tagID := c.Param("tag_id")

	relations, err := h.service.GetTagEntities(ctx, tagID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为响应格式
	result := make([]TagEntityInfo, len(relations))
	for i, relation := range relations {
		result[i] = TagEntityInfo{
			ID:         relation.ID,
			EntityType: relation.EntityType,
			EntityID:   relation.EntityID,
		}

		if !relation.CreatedAt.IsZero() {
			result[i].CreatedAt = relation.CreatedAt.Unix()
		}
	}

	return response.EchoOK(c, h.respWriter, result)
}

// AddTagToEntity 为实体添加标签
// @Summary 为实体添加标签
// @Description 为指定实体添加单个标签，如果已存在则返回错误
// @Tags 基础配置-标签关联
// @Accept json
// @Produce json
// @Param entity_type path string true "实体类型(skill/class/item/monster)"
// @Param entity_id path string true "实体ID(UUID格式)"
// @Param request body AddTagToEntityRequest true "添加标签请求"
// @Success 200 {object} response.Response "成功返回添加结果"
// @Failure 400 {object} response.Response "请求参数错误或标签已存在"
// @Failure 404 {object} response.Response "实体或标签不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/entities/{entity_type}/{entity_id}/tags [post]
func (h *TagRelationHandler) AddTagToEntity(c echo.Context) error {
	ctx := c.Request().Context()
	entityType := c.Param("entity_type")
	entityID := c.Param("entity_id")

	var req AddTagToEntityRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 添加标签
	if err := h.service.AddTagToEntity(ctx, req.TagID, entityType, entityID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message":     "标签添加成功",
		"entity_type": entityType,
		"entity_id":   entityID,
		"tag_id":      req.TagID,
	})
}

// RemoveTagFromEntity 从实体移除标签
// @Summary 从实体移除标签
// @Description 从指定实体移除单个标签关联
// @Tags 基础配置-标签关联
// @Accept json
// @Produce json
// @Param entity_type path string true "实体类型(skill/class/item/monster)"
// @Param entity_id path string true "实体ID(UUID格式)"
// @Param tag_id path string true "标签ID(UUID格式)"
// @Success 200 {object} response.Response "成功返回移除结果"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "实体或标签不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/entities/{entity_type}/{entity_id}/tags/{tag_id} [delete]
func (h *TagRelationHandler) RemoveTagFromEntity(c echo.Context) error {
	ctx := c.Request().Context()
	entityType := c.Param("entity_type")
	entityID := c.Param("entity_id")
	tagID := c.Param("tag_id")

	// 移除标签
	if err := h.service.RemoveTagFromEntity(ctx, tagID, entityType, entityID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message":     "标签移除成功",
		"entity_type": entityType,
		"entity_id":   entityID,
		"tag_id":      tagID,
	})
}

// BatchSetEntityTags 批量设置实体的标签
// @Summary 批量设置实体的标签
// @Description 批量设置实体的所有标签，会先删除原有标签再添加新标签(事务保证)
// @Tags 基础配置-标签关联
// @Accept json
// @Produce json
// @Param entity_type path string true "实体类型(skill/class/item/monster)"
// @Param entity_id path string true "实体ID(UUID格式)"
// @Param request body BatchSetEntityTagsRequest true "批量设置标签请求"
// @Success 200 {object} response.Response "成功返回设置结果"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "实体或标签不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/entities/{entity_type}/{entity_id}/tags/batch [post]
func (h *TagRelationHandler) BatchSetEntityTags(c echo.Context) error {
	ctx := c.Request().Context()
	entityType := c.Param("entity_type")
	entityID := c.Param("entity_id")

	var req BatchSetEntityTagsRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 批量设置标签
	if err := h.service.BatchSetEntityTags(ctx, entityType, entityID, req.TagIDs); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message":     "标签批量设置成功",
		"entity_type": entityType,
		"entity_id":   entityID,
		"tag_count":   len(req.TagIDs),
	})
}
