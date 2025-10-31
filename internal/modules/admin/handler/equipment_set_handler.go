// Package handler 提供管理端HTTP请求处理器
package handler

import (
	"database/sql"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/modules/admin/dto"
	"tsu-self/internal/modules/admin/service"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/xerrors"
)

// EquipmentSetHandler 装备套装管理Handler
type EquipmentSetHandler struct {
	service    *service.EquipmentSetService
	respWriter response.Writer
}

// NewEquipmentSetHandler 创建装备套装管理Handler
func NewEquipmentSetHandler(db *sql.DB, respWriter response.Writer) *EquipmentSetHandler {
	return &EquipmentSetHandler{
		service:    service.NewEquipmentSetService(db),
		respWriter: respWriter,
	}
}

// CreateSet 创建套装配置
// @Summary 创建套装配置
// @Description 创建新的装备套装配置,定义套装效果和激活条件。
// @Description
// @Description **套装效果配置**:
// @Description - piece_count: 激活所需件数(如2件套、4件套)
// @Description - effect_description: 效果描述文本
// @Description - out_of_combat_effects: 局外效果(属性加成)
// @Description - in_combat_effects: 局内效果(战斗触发)
// @Description
// @Description **多档位支持**:
// @Description - 支持配置多个档位(如2件、4件、6件)
// @Description - 穿戴4件时会同时激活2件和4件效果
// @Description - 效果会累加计算
// @Description
// @Description **请求示例**:
// @Description ```json
// @Description {
// @Description   "set_code": "flame_set",
// @Description   "set_name": "烈焰套装",
// @Description   "description": "蕴含火焰之力的强大套装",
// @Description   "set_effects": [
// @Description     {
// @Description       "piece_count": 2,
// @Description       "effect_description": "2件套:攻击力+50",
// @Description       "out_of_combat_effects": [{"Data_type":"Status","Data_ID":"ATK","Bouns_type":"bonus","Bouns_Number":"50"}]
// @Description     },
// @Description     {
// @Description       "piece_count": 4,
// @Description       "effect_description": "4件套:攻击力+100,暴击率+10%",
// @Description       "out_of_combat_effects": [{"Data_type":"Status","Data_ID":"ATK","Bouns_type":"bonus","Bouns_Number":"100"}]
// @Description     }
// @Description   ]
// @Description }
// @Description ```
// @Tags 装备套装管理
// @Accept json
// @Produce json
// @Param request body dto.CreateEquipmentSetRequest true "创建套装请求"
// @Success 200 {object} response.Response{data=dto.EquipmentSetResponse} "创建成功,返回套装详情"
// @Failure 400 {object} response.Response "参数错误(100400): set_code重复、效果配置无效等"
// @Failure 500 {object} response.Response "服务器错误(100500)"
// @Router /admin/equipment-sets [post]
func (h *EquipmentSetHandler) CreateSet(c echo.Context) error {
	// 1. 解析请求参数
	var req dto.CreateEquipmentSetRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoError(c, h.respWriter, xerrors.Wrap(err, xerrors.CodeInvalidParams, "请求参数解析失败"))
	}

	// 2. 验证请求参数
	if err := c.Validate(&req); err != nil {
		return response.EchoValidationError(c, h.respWriter, err)
	}

	// 3. 调用Service创建
	result, err := h.service.CreateSet(c.Request().Context(), &req)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 4. 返回响应
	return response.EchoOK(c, h.respWriter, result)
}

// GetSetList 查询套装列表
// @Summary 查询套装列表
// @Description 分页查询装备套装配置列表，支持搜索和筛选
// @Tags 装备套装管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param keyword query string false "搜索关键词"
// @Param is_active query bool false "是否激活"
// @Param sort_by query string false "排序字段" default(created_at)
// @Param sort_order query string false "排序方向" default(desc) Enums(asc, desc)
// @Success 200 {object} response.Response{data=dto.EquipmentSetListResponse} "查询成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/equipment-sets [get]
func (h *EquipmentSetHandler) GetSetList(c echo.Context) error {
	// 1. 解析查询参数
	var req dto.ListEquipmentSetsRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoError(c, h.respWriter, xerrors.Wrap(err, xerrors.CodeInvalidParams, "请求参数解析失败"))
	}

	// 2. 调用Service查询
	result, err := h.service.GetSetList(c.Request().Context(), &req)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 3. 返回响应
	return response.EchoOK(c, h.respWriter, result)
}

// GetSet 查询套装详情
// @Summary 查询套装详情
// @Description 根据ID查询装备套装配置详情
// @Tags 装备套装管理
// @Accept json
// @Produce json
// @Param id path string true "套装ID"
// @Success 200 {object} response.Response{data=dto.EquipmentSetResponse} "查询成功"
// @Failure 404 {object} response.Response "套装不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/equipment-sets/{id} [get]
func (h *EquipmentSetHandler) GetSet(c echo.Context) error {
	// 1. 获取套装ID
	id := c.Param("id")
	if id == "" {
		return response.EchoError(c, h.respWriter, xerrors.New(xerrors.CodeInvalidParams, "套装ID不能为空"))
	}

	// 2. 调用Service查询
	result, err := h.service.GetSetByID(c.Request().Context(), id)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 3. 返回响应
	return response.EchoOK(c, h.respWriter, result)
}

// UpdateSet 更新套装配置
// @Summary 更新套装配置
// @Description 更新装备套装配置信息
// @Tags 装备套装管理
// @Accept json
// @Produce json
// @Param id path string true "套装ID"
// @Param request body dto.UpdateEquipmentSetRequest true "更新套装请求"
// @Success 200 {object} response.Response{data=dto.EquipmentSetResponse} "更新成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "套装不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/equipment-sets/{id} [put]
func (h *EquipmentSetHandler) UpdateSet(c echo.Context) error {
	// 1. 获取套装ID
	id := c.Param("id")
	if id == "" {
		return response.EchoError(c, h.respWriter, xerrors.New(xerrors.CodeInvalidParams, "套装ID不能为空"))
	}

	// 2. 解析请求参数
	var req dto.UpdateEquipmentSetRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoError(c, h.respWriter, xerrors.Wrap(err, xerrors.CodeInvalidParams, "请求参数解析失败"))
	}

	// 3. 验证请求参数
	if err := c.Validate(&req); err != nil {
		return response.EchoValidationError(c, h.respWriter, err)
	}

	// 4. 调用Service更新
	result, err := h.service.UpdateSet(c.Request().Context(), id, &req)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 5. 返回响应
	return response.EchoOK(c, h.respWriter, result)
}

// DeleteSet 删除套装配置
// @Summary 删除套装配置
// @Description 软删除装备套装配置
// @Tags 装备套装管理
// @Accept json
// @Produce json
// @Param id path string true "套装ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 404 {object} response.Response "套装不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/equipment-sets/{id} [delete]
func (h *EquipmentSetHandler) DeleteSet(c echo.Context) error {
	// 1. 获取套装ID
	id := c.Param("id")
	if id == "" {
		return response.EchoError(c, h.respWriter, xerrors.New(xerrors.CodeInvalidParams, "套装ID不能为空"))
	}

	// 2. 调用Service删除
	if err := h.service.DeleteSet(c.Request().Context(), id); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 3. 返回响应
	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": "套装配置删除成功",
	})
}

// GetSetItems 查询套装包含的装备列表
// @Summary 查询套装包含的装备列表
// @Description 查询指定套装包含的所有装备
// @Tags 装备套装管理
// @Accept json
// @Produce json
// @Param id path string true "套装ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=dto.SetItemListResponse} "查询成功"
// @Failure 404 {object} response.Response "套装不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/equipment-sets/{id}/items [get]
func (h *EquipmentSetHandler) GetSetItems(c echo.Context) error {
	// 1. 获取套装ID
	id := c.Param("id")
	if id == "" {
		return response.EchoError(c, h.respWriter, xerrors.New(xerrors.CodeInvalidParams, "套装ID不能为空"))
	}

	// 2. 解析分页参数
	var req dto.ListEquipmentSetsRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoError(c, h.respWriter, xerrors.Wrap(err, xerrors.CodeInvalidParams, "请求参数解析失败"))
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	// 3. 调用Service查询
	result, err := h.service.GetSetItems(c.Request().Context(), id, page, pageSize)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 4. 返回响应
	return response.EchoOK(c, h.respWriter, result)
}

// GetUnassignedItems 查询未关联套装的装备列表
// @Summary 查询未关联套装的装备列表
// @Description 查询所有未关联套装的装备，用于关联装备到套装
// @Tags 装备套装管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param keyword query string false "搜索关键词"
// @Success 200 {object} response.Response{data=dto.SetItemListResponse} "查询成功"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/equipment-sets/unassigned-items [get]
func (h *EquipmentSetHandler) GetUnassignedItems(c echo.Context) error {
	// 1. 解析查询参数
	var req dto.ListEquipmentSetsRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoError(c, h.respWriter, xerrors.Wrap(err, xerrors.CodeInvalidParams, "请求参数解析失败"))
	}

	// 2. 调用Service查询
	result, err := h.service.GetUnassignedItems(c.Request().Context(), &req)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 3. 返回响应
	return response.EchoOK(c, h.respWriter, result)
}

// BatchAssignItems godoc
// @Summary 批量分配物品到套装
// @Description 批量将装备物品分配到指定套装
// @Tags 装备套装管理
// @Accept json
// @Produce json
// @Param set_id path string true "套装ID"
// @Param request body dto.BatchAssignItemsToSetRequest true "批量分配请求"
// @Success 200 {object} response.Response{data=dto.BatchAssignItemsToSetResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/equipment-sets/{set_id}/items/batch-assign [post]
func (h *EquipmentSetHandler) BatchAssignItems(c echo.Context) error {
	// 1. 获取套装ID
	setID := c.Param("set_id")
	if setID == "" {
		return response.EchoError(c, h.respWriter, xerrors.New(xerrors.CodeInvalidParams, "套装ID不能为空"))
	}

	// 2. 解析请求参数
	var req dto.BatchAssignItemsToSetRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoError(c, h.respWriter, xerrors.Wrap(err, xerrors.CodeInvalidParams, "请求参数解析失败"))
	}

	// 3. 验证请求参数
	if err := c.Validate(&req); err != nil {
		return response.EchoValidationError(c, h.respWriter, err)
	}

	// 4. 调用Service批量分配
	result, err := h.service.BatchAssignItems(c.Request().Context(), setID, &req)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 5. 返回响应
	return response.EchoOK(c, h.respWriter, result)
}

// BatchRemoveItems godoc
// @Summary 批量移除物品从套装
// @Description 批量将物品从指定套装中移除
// @Tags 装备套装管理
// @Accept json
// @Produce json
// @Param set_id path string true "套装ID"
// @Param request body dto.BatchRemoveItemsFromSetRequest true "批量移除请求"
// @Success 200 {object} response.Response{data=dto.BatchRemoveItemsFromSetResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/equipment-sets/{set_id}/items/batch-remove [post]
func (h *EquipmentSetHandler) BatchRemoveItems(c echo.Context) error {
	// 1. 获取套装ID
	setID := c.Param("set_id")
	if setID == "" {
		return response.EchoError(c, h.respWriter, xerrors.New(xerrors.CodeInvalidParams, "套装ID不能为空"))
	}

	// 2. 解析请求参数
	var req dto.BatchRemoveItemsFromSetRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoError(c, h.respWriter, xerrors.Wrap(err, xerrors.CodeInvalidParams, "请求参数解析失败"))
	}

	// 3. 验证请求参数
	if err := c.Validate(&req); err != nil {
		return response.EchoValidationError(c, h.respWriter, err)
	}

	// 4. 调用Service批量移除
	result, err := h.service.BatchRemoveItems(c.Request().Context(), setID, &req)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 5. 返回响应
	return response.EchoOK(c, h.respWriter, result)
}

// RemoveItem godoc
// @Summary 移除单个物品从套装
// @Description 将单个物品从指定套装中移除
// @Tags 装备套装管理
// @Accept json
// @Produce json
// @Param set_id path string true "套装ID"
// @Param item_id path string true "物品ID"
// @Success 200 {object} response.Response{data=map[string]string}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/equipment-sets/{set_id}/items/{item_id} [delete]
func (h *EquipmentSetHandler) RemoveItem(c echo.Context) error {
	// 1. 获取套装ID和物品ID
	setID := c.Param("set_id")
	if setID == "" {
		return response.EchoError(c, h.respWriter, xerrors.New(xerrors.CodeInvalidParams, "套装ID不能为空"))
	}

	itemID := c.Param("item_id")
	if itemID == "" {
		return response.EchoError(c, h.respWriter, xerrors.New(xerrors.CodeInvalidParams, "物品ID不能为空"))
	}

	// 2. 调用Service移除
	if err := h.service.RemoveItem(c.Request().Context(), setID, itemID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 3. 返回响应
	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": "物品已从套装中移除",
	})
}
