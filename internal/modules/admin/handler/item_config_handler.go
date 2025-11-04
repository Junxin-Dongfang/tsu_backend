package handler

import (
	"database/sql"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/modules/admin/dto"
	"tsu-self/internal/modules/admin/service"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/repository/interfaces"
)

// ItemConfigHandler 物品配置管理Handler
type ItemConfigHandler struct {
	service    *service.ItemConfigService
	respWriter response.Writer
}

// NewItemConfigHandler 创建物品配置管理Handler
func NewItemConfigHandler(db *sql.DB, respWriter response.Writer) *ItemConfigHandler {
	return &ItemConfigHandler{
		service:    service.NewItemConfigService(db),
		respWriter: respWriter,
	}
}

// CreateItem 创建物品配置
// @Summary 创建物品配置
// @Description 创建新的物品配置,支持装备、消耗品、材料等多种物品类型。
// @Description
// @Description **物品类型**:
// @Description - equipment: 装备(需设置equip_slot)
// @Description - consumable: 消耗品
// @Description - material: 材料
// @Description - quest: 任务物品
// @Description - currency: 货币
// @Description
// @Description **品质等级**:
// @Description - poor: 劣质(灰色)
// @Description - normal: 普通(白色)
// @Description - fine: 精良(绿色)
// @Description - excellent: 卓越(蓝色)
// @Description - superb: 极品(紫色)
// @Description - master: 大师(橙色)
// @Description - epic: 史诗(红色)
// @Description - legendary: 传说(金色)
// @Description - mythic: 神话(彩色)
// @Description
// @Description **装备槽位**(仅装备类型需要):
// @Description - mainhand: 主手, offhand: 副手, head: 头部, chest: 胸部
// @Description - legs: 腿部, feet: 脚部, hands: 手部, waist: 腰部
// @Description - neck: 项链, ring: 戒指, trinket: 饰品
// @Description
// @Description **效果配置**:
// @Description - out_of_combat_effects: 局外效果(JSON数组),直接属性加成
// @Description - in_combat_effects: 局内效果(JSON数组),战斗触发效果
// @Tags 物品配置管理
// @Accept json
// @Produce json
// @Param request body dto.CreateItemRequest true "创建物品配置请求"
// @Success 200 {object} response.Response{data=dto.ItemConfigResponse} "创建成功,返回物品详情"
// @Failure 400 {object} response.Response "参数错误(100400): item_code重复、类型无效、装备缺少槽位等"
// @Failure 500 {object} response.Response "服务器错误(100500)"
// @Router /admin/items [post]
func (h *ItemConfigHandler) CreateItem(c echo.Context) error {
	ctx := c.Request().Context()

	var req dto.CreateItemRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求参数无效")
	}

	// 创建物品配置
	item, err := h.service.CreateItem(ctx, &req)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, item)
}

// GetItem 获取物品配置详情
// @Summary 获取物品配置详情
// @Description 根据ID获取物品配置详情
// @Tags 物品配置管理
// @Accept json
// @Produce json
// @Param id path string true "物品ID"
// @Success 200 {object} response.Response{data=dto.ItemConfigResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/items/{id} [get]
func (h *ItemConfigHandler) GetItem(c echo.Context) error {
	ctx := c.Request().Context()
	itemID := c.Param("id")

	item, err := h.service.GetItemByID(ctx, itemID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, item)
}

// ListItems 查询物品配置列表
// @Summary 查询物品配置列表
// @Description 查询物品配置列表,支持多维度筛选、分页和排序。
// @Description
// @Description **筛选条件**:
// @Description - item_type: 按物品类型筛选(equipment/consumable/material/quest/currency)
// @Description - item_quality: 按品质筛选(poor/normal/fine/excellent/superb/master/epic/legendary/mythic)
// @Description - equip_slot: 按装备槽位筛选(仅装备类型)
// @Description - min_level/max_level: 按等级范围筛选
// @Description - is_active: 按启用状态筛选
// @Description - keyword: 关键词搜索(匹配item_code或item_name)
// @Description - tag_ids: 按标签筛选(逗号分隔的标签ID列表)
// @Description
// @Description **排序**:
// @Description - 默认按created_at降序排列
// @Description - 支持按item_level、item_quality等字段排序
// @Tags 物品配置管理
// @Accept json
// @Produce json
// @Param item_type query string false "物品类型" Enums(equipment, consumable, material, quest, currency)
// @Param item_quality query string false "物品品质" Enums(poor, normal, fine, excellent, superb, master, epic, legendary, mythic)
// @Param equip_slot query string false "装备槽位" Enums(mainhand, offhand, head, chest, legs, feet, hands, waist, neck, ring, trinket)
// @Param min_level query int false "最低等级" minimum(1) maximum(100)
// @Param max_level query int false "最高等级" minimum(1) maximum(100)
// @Param is_active query bool false "是否启用"
// @Param keyword query string false "关键词搜索(item_code, item_name)"
// @Param tag_ids query string false "标签ID列表(逗号分隔)"
// @Param page query int false "页码" default(1) minimum(1)
// @Param page_size query int false "每页数量" default(20) minimum(1) maximum(100)
// @Success 200 {object} response.Response{data=object{items=[]dto.ItemConfigResponse,total=int64,page=int,page_size=int}} "查询成功"
// @Failure 400 {object} response.Response "参数错误(100400)"
// @Failure 500 {object} response.Response "服务器错误(100500)"
// @Router /admin/items [get]
func (h *ItemConfigHandler) ListItems(c echo.Context) error {
	ctx := c.Request().Context()

	// 解析查询参数
	params := interfaces.ListItemParams{
		Page:     1,
		PageSize: 20,
	}

	if page, err := strconv.Atoi(c.QueryParam("page")); err == nil && page > 0 {
		params.Page = page
	}
	if pageSize, err := strconv.Atoi(c.QueryParam("page_size")); err == nil && pageSize > 0 {
		params.PageSize = pageSize
	}

	if itemType := c.QueryParam("item_type"); itemType != "" {
		params.ItemType = &itemType
	}
	if itemQuality := c.QueryParam("item_quality"); itemQuality != "" {
		params.ItemQuality = &itemQuality
	}
	if equipSlot := c.QueryParam("equip_slot"); equipSlot != "" {
		params.EquipSlot = &equipSlot
	}
	if minLevel, err := strconv.Atoi(c.QueryParam("min_level")); err == nil {
		params.MinLevel = &minLevel
	}
	if maxLevel, err := strconv.Atoi(c.QueryParam("max_level")); err == nil {
		params.MaxLevel = &maxLevel
	}
	if isActive := c.QueryParam("is_active"); isActive != "" {
		active := isActive == "true"
		params.IsActive = &active
	}
	if keyword := c.QueryParam("keyword"); keyword != "" {
		params.Keyword = &keyword
	}
	if tagIDs := c.QueryParam("tag_ids"); tagIDs != "" {
		params.TagIDs = parseCommaSeparated(tagIDs)
	}

	// 查询列表
	items, total, err := h.service.GetItems(ctx, params)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 返回分页响应
	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"items":     items,
		"total":     total,
		"page":      params.Page,
		"page_size": params.PageSize,
	})
}

// UpdateItem 更新物品配置
// @Summary 更新物品配置
// @Description 更新物品配置信息
// @Tags 物品配置管理
// @Accept json
// @Produce json
// @Param id path string true "物品ID"
// @Param request body dto.UpdateItemRequest true "更新物品配置请求"
// @Success 200 {object} response.Response{data=dto.ItemConfigResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/items/{id} [put]
func (h *ItemConfigHandler) UpdateItem(c echo.Context) error {
	ctx := c.Request().Context()
	itemID := c.Param("id")

	var req dto.UpdateItemRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求参数无效")
	}

	// 更新物品配置
	item, err := h.service.UpdateItem(ctx, itemID, &req)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, item)
}

// DeleteItem 删除物品配置
// @Summary 删除物品配置
// @Description 删除物品配置(软删除)
// @Tags 物品配置管理
// @Accept json
// @Produce json
// @Param id path string true "物品ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/items/{id} [delete]
func (h *ItemConfigHandler) DeleteItem(c echo.Context) error {
	ctx := c.Request().Context()
	itemID := c.Param("id")

	if err := h.service.DeleteItem(ctx, itemID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]string{"message": "删除成功"})
}

// AddItemTags 为物品添加标签
// @Summary 为物品添加标签
// @Description 为物品添加一个或多个标签,用于分类、筛选和搜索。
// @Description
// @Description **标签用途**:
// @Description - 分类标签: 武器、防具、消耗品等
// @Description - 属性标签: 火属性、冰属性等
// @Description - 稀有度标签: 普通、稀有、史诗等
// @Description - 功能标签: 可交易、可强化等
// @Description
// @Description **请求示例**:
// @Description ```json
// @Description {
// @Description   "tag_ids": ["tag-weapon", "tag-fire", "tag-legendary"]
// @Description }
// @Description ```
// @Tags 物品配置管理
// @Accept json
// @Produce json
// @Param id path string true "物品ID(UUID格式)"
// @Param request body map[string][]string true "标签ID列表"
// @Success 200 {object} response.Response "添加成功"
// @Failure 400 {object} response.Response "参数错误(100400): tag_id不存在、标签已存在等"
// @Failure 404 {object} response.Response "物品不存在(100404)"
// @Failure 500 {object} response.Response "服务器错误(100500)"
// @Router /admin/items/{id}/tags [post]
func (h *ItemConfigHandler) AddItemTags(c echo.Context) error {
	ctx := c.Request().Context()
	itemID := c.Param("id")

	var req struct {
		TagIDs []string `json:"tag_ids"`
	}
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求参数无效")
	}

	if err := h.service.AddItemTags(ctx, itemID, req.TagIDs); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]string{"message": "标签添加成功"})
}

// GetItemTags 查询物品的所有标签
// @Summary 查询物品的所有标签
// @Description 查询物品的所有标签
// @Tags 物品配置管理
// @Accept json
// @Produce json
// @Param id path string true "物品ID"
// @Success 200 {object} response.Response{data=[]dto.TagResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/items/{id}/tags [get]
func (h *ItemConfigHandler) GetItemTags(c echo.Context) error {
	ctx := c.Request().Context()
	itemID := c.Param("id")

	tags, err := h.service.GetItemTags(ctx, itemID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, tags)
}

// UpdateItemTags 批量更新物品标签
// @Summary 批量更新物品标签
// @Description 批量更新物品的所有标签，会替换现有的所有标签
// @Tags 物品配置管理
// @Accept json
// @Produce json
// @Param id path string true "物品ID"
// @Param request body dto.UpdateItemTagsRequest true "标签ID列表"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/items/{id}/tags [put]
func (h *ItemConfigHandler) UpdateItemTags(c echo.Context) error {
	ctx := c.Request().Context()
	itemID := c.Param("id")

	var req dto.UpdateItemTagsRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求参数无效")
	}

	if err := h.service.UpdateItemTags(ctx, itemID, req.TagIDs); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]string{"message": "标签更新成功"})
}

// RemoveItemTag 移除物品标签
// @Summary 移除物品标签
// @Description 移除物品的单个标签
// @Tags 物品配置管理
// @Accept json
// @Produce json
// @Param id path string true "物品ID"
// @Param tag_id path string true "标签ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/items/{id}/tags/{tag_id} [delete]
func (h *ItemConfigHandler) RemoveItemTag(c echo.Context) error {
	ctx := c.Request().Context()
	itemID := c.Param("id")
	tagID := c.Param("tag_id")

	if err := h.service.RemoveItemTag(ctx, itemID, tagID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]string{"message": "标签移除成功"})
}

// parseCommaSeparated 解析逗号分隔的字符串
func parseCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}
	var result []string
	for _, item := range strings.Split(s, ",") {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// AddItemClasses 为物品添加职业限制
// @Summary 添加物品职业限制
// @Description 为物品添加一个或多个职业限制,限制只有特定职业可以使用该物品。
// @Description
// @Description **职业限制说明**:
// @Description - 添加职业限制后,只有指定职业的角色可以装备/使用该物品
// @Description - 未添加职业限制的物品,所有职业都可以使用
// @Description - 支持同时限制多个职业(如战士和骑士都可以使用)
// @Description
// @Description **使用场景**:
// @Description - 职业专属装备: 只有战士可以装备的重甲
// @Description - 职业专属道具: 只有法师可以使用的魔法卷轴
// @Description - 多职业共享: 战士和骑士都可以使用的盾牌
// @Description
// @Description **请求示例**:
// @Description ```json
// @Description {
// @Description   "class_ids": ["class-warrior", "class-paladin"]
// @Description }
// @Description ```
// @Tags 物品配置管理
// @Accept json
// @Produce json
// @Param id path string true "物品ID(UUID格式)"
// @Param request body dto.AddItemClassesRequest true "职业ID列表"
// @Success 200 {object} response.Response "添加成功"
// @Failure 400 {object} response.Response "参数错误(100400): class_id不存在、职业限制已存在等"
// @Failure 404 {object} response.Response "物品不存在(100404)"
// @Failure 500 {object} response.Response "服务器错误(100500)"
// @Router /admin/items/{id}/classes [post]
func (h *ItemConfigHandler) AddItemClasses(c echo.Context) error {
	// 1. 获取物品ID
	itemID := c.Param("id")

	// 2. 解析请求
	var req dto.AddItemClassesRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 3. 验证请求
	if err := c.Validate(&req); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 4. 添加职业关联
	if err := h.service.AddItemClasses(c.Request().Context(), itemID, req.ClassIDs); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]string{"message": "职业限制添加成功"})
}

// GetItemClasses 查询物品的职业限制
// @Summary 查询物品职业限制
// @Description 查询物品关联的所有职业
// @Tags 物品配置管理
// @Accept json
// @Produce json
// @Param id path string true "物品ID"
// @Success 200 {object} response.Response{data=[]string}
// @Router /admin/items/{id}/classes [get]
func (h *ItemConfigHandler) GetItemClasses(c echo.Context) error {
	// 1. 获取物品ID
	itemID := c.Param("id")

	// 2. 查询职业关联
	classIDs, err := h.service.GetItemClasses(c.Request().Context(), itemID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, classIDs)
}

// UpdateItemClasses 批量更新物品职业限制
// @Summary 批量更新物品职业限制
// @Description 替换物品的所有职业限制
// @Tags 物品配置管理
// @Accept json
// @Produce json
// @Param id path string true "物品ID"
// @Param request body dto.UpdateItemClassesRequest true "职业ID列表"
// @Success 200 {object} response.Response
// @Router /admin/items/{id}/classes [put]
func (h *ItemConfigHandler) UpdateItemClasses(c echo.Context) error {
	// 1. 获取物品ID
	itemID := c.Param("id")

	// 2. 解析请求
	var req dto.UpdateItemClassesRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 3. 验证请求
	if err := c.Validate(&req); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 4. 批量更新职业关联
	if err := h.service.UpdateItemClasses(c.Request().Context(), itemID, req.ClassIDs); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]string{"message": "职业限制更新成功"})
}

// RemoveItemClass 移除物品职业限制
// @Summary 移除物品职业限制
// @Description 移除物品的单个职业限制
// @Tags 物品配置管理
// @Accept json
// @Produce json
// @Param id path string true "物品ID"
// @Param class_id path string true "职业ID"
// @Success 200 {object} response.Response
// @Router /admin/items/{id}/classes/{class_id} [delete]
func (h *ItemConfigHandler) RemoveItemClass(c echo.Context) error {
	// 1. 获取物品ID和职业ID
	itemID := c.Param("id")
	classID := c.Param("class_id")

	// 2. 移除职业关联
	if err := h.service.RemoveItemClass(c.Request().Context(), itemID, classID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]string{"message": "职业限制移除成功"})
}
