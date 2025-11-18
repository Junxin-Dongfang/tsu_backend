// Package handler 提供Admin模块的HTTP请求处理器
package handler

import (
	"database/sql"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"tsu-self/internal/modules/admin/dto"
	"tsu-self/internal/modules/admin/service"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/interfaces"
)

// DropPoolHandler 掉落池配置Handler
type DropPoolHandler struct {
	service    *service.DropPoolService
	respWriter response.Writer
}

// NewDropPoolHandler 创建掉落池配置Handler
func NewDropPoolHandler(db *sql.DB, respWriter response.Writer) *DropPoolHandler {
	return &DropPoolHandler{
		service:    service.NewDropPoolService(db),
		respWriter: respWriter,
	}
}

// CreateDropPool 创建掉落池配置
// @Summary 创建掉落池配置
// @Description 创建新的掉落池配置。掉落池用于定义怪物、副本、任务等场景的物品掉落规则。
// @Description
// @Description **业务规则**:
// @Description - pool_code必须唯一,建议使用场景_名称格式(如: monster_goblin_elite)
// @Description - min_drops必须小于等于max_drops
// @Description - guaranteed_drops必须小于等于max_drops
// @Description - pool_type支持: monster(怪物)、dungeon(副本)、quest(任务)、activity(活动)、boss(Boss)、other(其他)
// @Description
// @Description **请求示例**:
// @Description ```json
// @Description {
// @Description   "pool_code": "monster_goblin_elite",
// @Description   "pool_name": "精英哥布林掉落池",
// @Description   "pool_type": "monster",
// @Description   "description": "10-20级精英哥布林的掉落配置",
// @Description   "min_drops": 1,
// @Description   "max_drops": 3,
// @Description   "guaranteed_drops": 1
// @Description }
// @Description ```
// @Tags 掉落池配置
// @Accept json
// @Produce json
// @Param request body dto.CreateDropPoolRequest true "掉落池配置请求"
// @Success 200 {object} response.Response{data=dto.DropPoolResponse} "创建成功,返回掉落池详情"
// @Failure 400 {object} response.Response "参数错误(100400): pool_code重复、数量范围无效等"
// @Failure 500 {object} response.Response "服务器错误(100500)"
// @Router /admin/drop-pools [post]
func (h *DropPoolHandler) CreateDropPool(c echo.Context) error {
	// 1. 解析请求
	var req dto.CreateDropPoolRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 2. 验证请求
	if err := c.Validate(&req); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 3. 创建掉落池配置
	resp, err := h.service.CreateDropPool(c.Request().Context(), &req)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// GetDropPoolList 查询掉落池列表
// @Summary 查询掉落池列表
// @Description 查询掉落池配置列表,支持分页、筛选、搜索和排序。
// @Description
// @Description **筛选条件**:
// @Description - pool_type: 按类型筛选(monster/dungeon/quest/activity/boss/other)
// @Description - is_active: 按激活状态筛选(true/false)
// @Description - keyword: 关键词搜索,匹配pool_code或pool_name
// @Description
// @Description **排序**:
// @Description - sort_by: 排序字段(created_at/updated_at/pool_code/pool_name)
// @Description - sort_order: 排序方向(asc升序/desc降序)
// @Description
// @Description **响应示例**:
// @Description ```json
// @Description {
// @Description   "code": 100000,
// @Description   "message": "success",
// @Description   "data": {
// @Description     "items": [...],
// @Description     "total": 25,
// @Description     "page": 1,
// @Description     "page_size": 20
// @Description   }
// @Description }
// @Description ```
// @Tags 掉落池配置
// @Accept json
// @Produce json
// @Param page query int false "页码(默认1)" default(1) minimum(1)
// @Param page_size query int false "每页数量(默认20,最大100)" default(20) minimum(1) maximum(100)
// @Param pool_type query string false "掉落池类型筛选" Enums(monster, dungeon, quest, activity, boss, other)
// @Param is_active query bool false "激活状态筛选"
// @Param keyword query string false "关键词搜索(pool_code, pool_name)"
// @Param sort_by query string false "排序字段" Enums(created_at, updated_at, pool_code, pool_name) default(created_at)
// @Param sort_order query string false "排序方向" Enums(asc, desc) default(desc)
// @Success 200 {object} response.Response{data=dto.DropPoolListResponse} "查询成功"
// @Failure 400 {object} response.Response "参数错误(100400)"
// @Failure 500 {object} response.Response "服务器错误(100500)"
// @Router /admin/drop-pools [get]
func (h *DropPoolHandler) GetDropPoolList(c echo.Context) error {
	// 1. 解析查询参数
	params := interfaces.ListDropPoolParams{
		Page:     1,
		PageSize: 20,
	}

	if page := c.QueryParam("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			params.Page = p
		}
	}

	if pageSize := c.QueryParam("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 {
			params.PageSize = ps
		}
	}

	if poolType := c.QueryParam("pool_type"); poolType != "" {
		params.PoolType = &poolType
	}

	if isActiveStr := c.QueryParam("is_active"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			params.IsActive = &isActive
		}
	}

	if keyword := c.QueryParam("keyword"); keyword != "" {
		params.Keyword = &keyword
	}

	params.SortBy = c.QueryParam("sort_by")
	params.SortOrder = c.QueryParam("sort_order")

	// 2. 查询掉落池列表
	resp, err := h.service.GetDropPoolList(c.Request().Context(), params)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// GetDropPool 获取掉落池详情
// @Summary 获取掉落池详情
// @Description 根据ID获取掉落池配置详情
// @Tags 掉落池配置
// @Accept json
// @Produce json
// @Param id path string true "掉落池ID"
// @Success 200 {object} response.Response{data=dto.DropPoolResponse}
// @Router /admin/drop-pools/{id} [get]
func (h *DropPoolHandler) GetDropPool(c echo.Context) error {
	// 1. 获取掉落池ID
	poolID := c.Param("id")

	// 2. 查询掉落池详情
	resp, err := h.service.GetDropPoolByID(c.Request().Context(), poolID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// UpdateDropPool 更新掉落池配置
// @Summary 更新掉落池配置
// @Description 更新掉落池配置信息,支持部分字段更新。
// @Description
// @Description **可更新字段**:
// @Description - pool_code: 掉落池代码(需保证唯一性)
// @Description - pool_name: 掉落池名称
// @Description - pool_type: 掉落池类型
// @Description - description: 描述
// @Description - min_drops, max_drops, guaranteed_drops: 掉落数量配置
// @Description - is_active: 激活状态
// @Description
// @Description **业务规则**:
// @Description - 更新后仍需满足: min_drops <= max_drops, guaranteed_drops <= max_drops
// @Description - pool_code更新后需保证唯一性
// @Description
// @Description **请求示例**:
// @Description ```json
// @Description {
// @Description   "pool_name": "精英哥布林掉落池(已更新)",
// @Description   "min_drops": 2,
// @Description   "max_drops": 5,
// @Description   "is_active": true
// @Description }
// @Description ```
// @Tags 掉落池配置
// @Accept json
// @Produce json
// @Param id path string true "掉落池ID(UUID格式)"
// @Param request body dto.UpdateDropPoolRequest true "更新请求(所有字段可选)"
// @Success 200 {object} response.Response{data=dto.DropPoolResponse} "更新成功,返回更新后的掉落池详情"
// @Failure 400 {object} response.Response "参数错误(100400): pool_code重复、数量范围无效等"
// @Failure 404 {object} response.Response "掉落池不存在(100404)"
// @Failure 500 {object} response.Response "服务器错误(100500)"
// @Router /admin/drop-pools/{id} [put]
func (h *DropPoolHandler) UpdateDropPool(c echo.Context) error {
	// 1. 获取掉落池ID
	poolID := c.Param("id")

	// 2. 解析请求
	var req dto.UpdateDropPoolRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 3. 验证请求
	if err := c.Validate(&req); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 4. 更新掉落池配置
	resp, err := h.service.UpdateDropPool(c.Request().Context(), poolID, &req)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// DeleteDropPool 删除掉落池配置
// @Summary 删除掉落池配置
// @Description 删除掉落池配置(软删除),可选择是否级联删除掉落池中的物品。
// @Description
// @Description **删除模式**:
// @Description - cascade=false(默认): 如果掉落池中还有物品,删除失败,需先删除所有物品
// @Description - cascade=true: 级联删除,同时删除掉落池中的所有物品
// @Description
// @Description **注意事项**:
// @Description - 删除为软删除,数据不会真正删除,只是标记为已删除
// @Description - 级联删除会同时软删除掉落池中的所有物品
// @Description - 删除后的掉落池不会出现在列表查询中
// @Tags 掉落池配置
// @Accept json
// @Produce json
// @Param id path string true "掉落池ID(UUID格式)"
// @Param cascade query bool false "是否级联删除物品" default(false)
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "参数错误(100400): 掉落池中还有物品且未设置级联删除"
// @Failure 404 {object} response.Response "掉落池不存在(100404)"
// @Failure 500 {object} response.Response "服务器错误(100500)"
// @Router /admin/drop-pools/{id} [delete]
func (h *DropPoolHandler) DeleteDropPool(c echo.Context) error {
	// 1. 获取掉落池ID
	poolID := c.Param("id")

	// 2. 获取级联删除参数
	cascade := false
	if cascadeStr := c.QueryParam("cascade"); cascadeStr != "" {
		if c, err := strconv.ParseBool(cascadeStr); err == nil {
			cascade = c
		}
	}

	// 3. 删除掉落池配置
	if err := h.service.DeleteDropPool(c.Request().Context(), poolID, cascade); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]string{"message": "删除成功"})
}

// AddDropPoolItem 添加掉落物品
// @Summary 添加掉落物品
// @Description 向掉落池添加物品,配置掉落权重、概率、数量等参数。
// @Description
// @Description **掉落模式**:
// @Description - **权重模式**: 设置drop_weight,系统根据权重计算掉落概率。适用于多个物品竞争掉落的场景。
// @Description - **固定概率模式**: 设置drop_rate(0-1之间),每个物品独立判定。适用于固定概率掉落的场景。
// @Description - 两种模式二选一,不能同时设置
// @Description
// @Description **品质权重**:
// @Description - quality_weights定义不同品质的掉落权重,格式为JSON对象
// @Description - 键: 品质名称(poor/normal/fine/excellent/superb/master/epic/legendary/mythic)
// @Description - 值: 权重值,数值越大掉落概率越高
// @Description - 示例: {"normal":50,"fine":30,"excellent":15,"superb":5}
// @Description
// @Description **数量和等级**:
// @Description - min_quantity和max_quantity定义掉落数量范围
// @Description - min_level和max_level定义等级限制(可选)
// @Description
// @Description **请求示例**:
// @Description ```json
// @Description {
// @Description   "item_id": "550e8400-e29b-41d4-a716-446655440000",
// @Description   "drop_weight": 100,
// @Description   "quality_weights": "{\"normal\":50,\"fine\":30,\"excellent\":15,\"superb\":5}",
// @Description   "min_quantity": 1,
// @Description   "max_quantity": 3,
// @Description   "min_level": 10,
// @Description   "max_level": 20
// @Description }
// @Description ```
// @Tags 掉落池配置
// @Accept json
// @Produce json
// @Param pool_id path string true "掉落池ID(UUID格式)"
// @Param request body dto.AddDropPoolItemRequest true "添加物品请求"
// @Success 200 {object} response.Response{data=dto.DropPoolItemResponse} "添加成功,返回掉落物品详情"
// @Failure 400 {object} response.Response "参数错误(100400): item_id不存在、数量范围无效、同时设置drop_weight和drop_rate等"
// @Failure 404 {object} response.Response "掉落池不存在(100404)"
// @Failure 500 {object} response.Response "服务器错误(100500)"
// @Router /admin/drop-pools/{pool_id}/items [post]
func (h *DropPoolHandler) AddDropPoolItem(c echo.Context) error {
	// 1. 获取掉落池ID
	poolID := c.Param("pool_id")
	if _, err := uuid.Parse(poolID); err != nil {
		return response.EchoError(c, h.respWriter, xerrors.New(xerrors.CodeResourceNotFound, "掉落池不存在"))
	}

	// 2. 解析请求
	var req dto.AddDropPoolItemRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 3. 验证请求
	if err := c.Validate(&req); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 4. 添加掉落物品
	resp, err := h.service.AddDropPoolItem(c.Request().Context(), poolID, &req)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// GetDropPoolItems 查询掉落池物品列表
// @Summary 查询掉落池物品列表
// @Description 查询掉落池的物品列表（支持分页、筛选）
// @Tags 掉落池配置
// @Accept json
// @Produce json
// @Param pool_id path string true "掉落池ID"
// @Param page query int false "页码（默认1）"
// @Param page_size query int false "每页数量（默认20，最大100）"
// @Param is_active query bool false "激活状态筛选"
// @Param min_level query int false "最低等级筛选"
// @Param max_level query int false "最高等级筛选"
// @Success 200 {object} response.Response{data=dto.DropPoolItemListResponse}
// @Router /admin/drop-pools/{pool_id}/items [get]
func (h *DropPoolHandler) GetDropPoolItems(c echo.Context) error {
	// 1. 获取掉落池ID
	poolID := c.Param("pool_id")
	if _, err := uuid.Parse(poolID); err != nil {
		return response.EchoError(c, h.respWriter, xerrors.New(xerrors.CodeResourceNotFound, "掉落池不存在"))
	}

	// 2. 解析查询参数
	params := interfaces.ListDropPoolItemParams{
		DropPoolID: poolID,
		Page:       1,
		PageSize:   20,
	}

	if page := c.QueryParam("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			params.Page = p
		}
	}

	if pageSize := c.QueryParam("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 {
			params.PageSize = ps
		}
	}

	if isActiveStr := c.QueryParam("is_active"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			params.IsActive = &isActive
		}
	}

	if minLevelStr := c.QueryParam("min_level"); minLevelStr != "" {
		if minLevel, err := strconv.ParseInt(minLevelStr, 10, 16); err == nil {
			level := int16(minLevel)
			params.MinLevel = &level
		}
	}

	if maxLevelStr := c.QueryParam("max_level"); maxLevelStr != "" {
		if maxLevel, err := strconv.ParseInt(maxLevelStr, 10, 16); err == nil {
			level := int16(maxLevel)
			params.MaxLevel = &level
		}
	}

	// 3. 查询掉落池物品列表
	resp, err := h.service.GetDropPoolItems(c.Request().Context(), params)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// GetDropPoolItem 获取掉落物品详情
// @Summary 获取掉落物品详情
// @Description 根据ID获取掉落池物品详情
// @Tags 掉落池配置
// @Accept json
// @Produce json
// @Param pool_id path string true "掉落池ID"
// @Param item_id path string true "物品ID"
// @Success 200 {object} response.Response{data=dto.DropPoolItemResponse}
// @Router /admin/drop-pools/{pool_id}/items/{item_id} [get]
func (h *DropPoolHandler) GetDropPoolItem(c echo.Context) error {
	// 1. 获取掉落池ID和物品ID
	poolID := c.Param("pool_id")
	itemID := c.Param("item_id")

	// 2. 查询掉落物品详情
	resp, err := h.service.GetDropPoolItem(c.Request().Context(), poolID, itemID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// UpdateDropPoolItem 更新掉落物品配置
// @Summary 更新掉落物品配置
// @Description 更新掉落池物品配置信息
// @Tags 掉落池配置
// @Accept json
// @Produce json
// @Param pool_id path string true "掉落池ID"
// @Param item_id path string true "物品ID"
// @Param request body dto.UpdateDropPoolItemRequest true "更新请求"
// @Success 200 {object} response.Response{data=dto.DropPoolItemResponse}
// @Router /admin/drop-pools/{pool_id}/items/{item_id} [put]
func (h *DropPoolHandler) UpdateDropPoolItem(c echo.Context) error {
	// 1. 获取掉落池ID和物品ID
	poolID := c.Param("pool_id")
	itemID := c.Param("item_id")

	// 2. 解析请求
	var req dto.UpdateDropPoolItemRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 3. 验证请求
	if err := c.Validate(&req); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 4. 更新掉落物品配置
	resp, err := h.service.UpdateDropPoolItem(c.Request().Context(), poolID, itemID, &req)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// RemoveDropPoolItem 移除掉落物品
// @Summary 移除掉落物品
// @Description 从掉落池移除物品（软删除）
// @Tags 掉落池配置
// @Accept json
// @Produce json
// @Param pool_id path string true "掉落池ID"
// @Param item_id path string true "物品ID"
// @Success 200 {object} response.Response
// @Router /admin/drop-pools/{pool_id}/items/{item_id} [delete]
func (h *DropPoolHandler) RemoveDropPoolItem(c echo.Context) error {
	// 1. 获取掉落池ID和物品ID
	poolID := c.Param("pool_id")
	itemID := c.Param("item_id")

	// 2. 移除掉落物品
	if err := h.service.RemoveDropPoolItem(c.Request().Context(), poolID, itemID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]string{"message": "移除成功"})
}
