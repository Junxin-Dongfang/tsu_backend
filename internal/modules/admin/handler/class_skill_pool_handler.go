package handler

import (
	"database/sql"
	"encoding/json"
	"strconv"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/modules/admin/service"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/repository/interfaces"
)

// ClassSkillPoolHandler 职业技能池 HTTP 处理器
type ClassSkillPoolHandler struct {
	service    *service.ClassSkillPoolService
	respWriter response.Writer
}

// NewClassSkillPoolHandler 创建职业技能池处理器
func NewClassSkillPoolHandler(db *sql.DB, respWriter response.Writer) *ClassSkillPoolHandler {
	return &ClassSkillPoolHandler{
		service:    service.NewClassSkillPoolService(db),
		respWriter: respWriter,
	}
}

// ==================== HTTP Models ====================

// CreateClassSkillPoolRequest 创建职业技能池请求
type CreateClassSkillPoolRequest struct {
	ClassID              string   `json:"class_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"` // 职业ID
	SkillID              string   `json:"skill_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"` // 技能ID
	RequiredLevel        int      `json:"required_level" validate:"required,min=1" example:"5"`                        // 需要的角色等级
	RequiredAttributes   string   `json:"required_attributes" example:"{\"STR\":15,\"INT\":10}"`                       // 所需属性要求(JSONB)
	PrerequisiteSkillIds []string `json:"prerequisite_skill_ids" example:"550e8400-e29b-41d4-a716-446655440000"`       // 前置技能ID数组
	LearnCostXP          int      `json:"learn_cost_xp" example:"1000"`                                                // 学习经验消耗
	SkillTier            int      `json:"skill_tier" validate:"min=1,max=5" example:"2"`                               // 技能等级(1-5)
	IsCore               bool     `json:"is_core" example:"true"`                                                      // 是否核心技能
	IsExclusive          bool     `json:"is_exclusive" example:"false"`                                                // 是否职业专属
	MaxLearnableLevel    int      `json:"max_learnable_level" example:"10"`                                            // 可学习的最大等级
	DisplayOrder         int      `json:"display_order" example:"1"`                                                   // 显示顺序
	IsVisible            bool     `json:"is_visible" example:"true"`                                                   // 是否可见
	CustomIcon           string   `json:"custom_icon" example:"warrior_fireball.png"`                                  // 自定义图标
	CustomDescription    string   `json:"custom_description" example:"战士专用的强化火球术"`                                     // 自定义描述
}

// UpdateClassSkillPoolRequest 更新职业技能池请求
type UpdateClassSkillPoolRequest struct {
	RequiredLevel        int      `json:"required_level" example:"5"`                            // 需要的角色等级
	RequiredAttributes   string   `json:"required_attributes" example:"{\"STR\":15,\"INT\":10}"` // 所需属性要求(JSONB)
	PrerequisiteSkillIds []string `json:"prerequisite_skill_ids" example:"550e8400-e29b-41d4"`   // 前置技能ID数组
	LearnCostXP          int      `json:"learn_cost_xp" example:"1000"`                          // 学习经验消耗
	SkillTier            int      `json:"skill_tier" example:"2"`                                // 技能等级(1-5)
	IsCore               bool     `json:"is_core" example:"true"`                                // 是否核心技能
	IsExclusive          bool     `json:"is_exclusive" example:"false"`                          // 是否职业专属
	MaxLearnableLevel    int      `json:"max_learnable_level" example:"10"`                      // 可学习的最大等级
	DisplayOrder         int      `json:"display_order" example:"1"`                             // 显示顺序
	IsVisible            bool     `json:"is_visible" example:"true"`                             // 是否可见
	CustomIcon           string   `json:"custom_icon" example:"warrior_fireball.png"`            // 自定义图标
	CustomDescription    string   `json:"custom_description" example:"战士专用的强化火球术"`               // 自定义描述
}

// ClassSkillPoolInfo 职业技能池信息响应
type ClassSkillPoolInfo struct {
	ID                   string   `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`       // ID
	ClassID              string   `json:"class_id" example:"550e8400-e29b-41d4-a716-446655440000"` // 职业ID
	SkillID              string   `json:"skill_id" example:"550e8400-e29b-41d4-a716-446655440000"` // 技能ID
	RequiredLevel        int      `json:"required_level" example:"5"`                              // 需要的角色等级
	RequiredAttributes   string   `json:"required_attributes" example:"{\"STR\":15,\"INT\":10}"`   // 所需属性要求(JSONB)
	PrerequisiteSkillIds []string `json:"prerequisite_skill_ids" example:"550e8400-e29b-41d4"`     // 前置技能ID数组
	LearnCostXP          int      `json:"learn_cost_xp" example:"1000"`                            // 学习经验消耗
	SkillTier            int      `json:"skill_tier" example:"2"`                                  // 技能等级(1-5)
	IsCore               bool     `json:"is_core" example:"true"`                                  // 是否核心技能
	IsExclusive          bool     `json:"is_exclusive" example:"false"`                            // 是否职业专属
	MaxLearnableLevel    int      `json:"max_learnable_level" example:"10"`                        // 可学习的最大等级
	DisplayOrder         int      `json:"display_order" example:"1"`                               // 显示顺序
	IsVisible            bool     `json:"is_visible" example:"true"`                               // 是否可见
	CustomIcon           string   `json:"custom_icon" example:"warrior_fireball.png"`              // 自定义图标
	CustomDescription    string   `json:"custom_description" example:"战士专用的强化火球术"`                 // 自定义描述
	CreatedAt            int64    `json:"created_at" example:"1633024800"`                         // 创建时间戳
	UpdatedAt            int64    `json:"updated_at" example:"1633024800"`                         // 更新时间戳
}

// ==================== HTTP Handlers ====================

// GetClassSkillPools 获取职业技能池列表
// @Summary 获取职业技能池列表
// @Description 分页查询职业技能池配置,支持按职业ID、技能ID、技能等级等条件筛选
// @Tags 职业技能池
// @Accept json
// @Produce json
// @Param class_id query string false "职业ID筛选"
// @Param skill_id query string false "技能ID筛选"
// @Param skill_tier query int false "技能等级筛选(1-5)"
// @Param is_core query bool false "是否核心技能筛选"
// @Param is_exclusive query bool false "是否专属技能筛选"
// @Param is_visible query bool false "是否可见筛选"
// @Param limit query int false "每页数量(默认10)"
// @Param offset query int false "偏移量(默认0)"
// @Success 200 {object} response.Response{data=object{list=[]ClassSkillPoolInfo,total=int}} "成功返回列表,包含list和total字段"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/class-skill-pools [get]
// @Security BearerAuth
func (h *ClassSkillPoolHandler) GetClassSkillPools(c echo.Context) error {
	ctx := c.Request().Context()

	// 解析查询参数
	params := interfaces.ClassSkillPoolQueryParams{}

	if classID := c.QueryParam("class_id"); classID != "" {
		params.ClassID = &classID
	}

	if skillID := c.QueryParam("skill_id"); skillID != "" {
		params.SkillID = &skillID
	}

	if skillTierStr := c.QueryParam("skill_tier"); skillTierStr != "" {
		skillTier, _ := strconv.Atoi(skillTierStr)
		params.SkillTier = &skillTier
	}

	if isCoreStr := c.QueryParam("is_core"); isCoreStr != "" {
		isCore, _ := strconv.ParseBool(isCoreStr)
		params.IsCore = &isCore
	}

	if isExclusiveStr := c.QueryParam("is_exclusive"); isExclusiveStr != "" {
		isExclusive, _ := strconv.ParseBool(isExclusiveStr)
		params.IsExclusive = &isExclusive
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

	// 查询列表
	pools, total, err := h.service.GetClassSkillPools(ctx, params)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为响应格式
	result := make([]ClassSkillPoolInfo, len(pools))
	for i, pool := range pools {
		result[i] = h.convertToClassSkillPoolInfo(pool)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"list":  result,
		"total": total,
	})
}

// GetClassSkillPoolsByClassID 获取指定职业的所有技能
// @Summary 获取指定职业的所有可学习技能
// @Description 根据职业ID获取该职业的所有可学习技能配置
// @Tags 职业技能池
// @Accept json
// @Produce json
// @Param class_id path string true "职业ID(UUID格式)"
// @Success 200 {object} response.Response{data=[]ClassSkillPoolInfo} "成功返回技能列表"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/classes/{class_id}/skill-pools [get]
// @Security BearerAuth
func (h *ClassSkillPoolHandler) GetClassSkillPoolsByClassID(c echo.Context) error {
	ctx := c.Request().Context()
	classID := c.Param("class_id")

	pools, err := h.service.GetClassSkillPoolsByClassID(ctx, classID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为响应格式
	result := make([]ClassSkillPoolInfo, len(pools))
	for i, pool := range pools {
		result[i] = h.convertToClassSkillPoolInfo(pool)
	}

	return response.EchoOK(c, h.respWriter, result)
}

// GetClassSkillPool 获取职业技能池详情
// @Summary 获取职业技能池详情
// @Description 根据ID获取单个职业技能池配置的详细信息
// @Tags 职业技能池
// @Accept json
// @Produce json
// @Param id path string true "职业技能池ID(UUID格式)"
// @Success 200 {object} response.Response{data=ClassSkillPoolInfo} "成功返回详情"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "配置不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/class-skill-pools/{id} [get]
// @Security BearerAuth
func (h *ClassSkillPoolHandler) GetClassSkillPool(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	pool, err := h.service.GetClassSkillPoolByID(ctx, id)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToClassSkillPoolInfo(pool))
}

// CreateClassSkillPool 创建职业技能池配置
// @Summary 添加技能到职业技能池
// @Description 为指定职业添加可学习的技能,配置学习要求和特性
// @Tags 职业技能池
// @Accept json
// @Produce json
// @Param request body CreateClassSkillPoolRequest true "创建职业技能池请求,class_id、skill_id、required_level为必填"
// @Success 200 {object} response.Response{data=ClassSkillPoolInfo} "成功创建并返回详情"
// @Failure 400 {object} response.Response "请求参数错误或验证失败"
// @Failure 409 {object} response.Response "该职业已配置此技能"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/class-skill-pools [post]
// @Security BearerAuth
func (h *ClassSkillPoolHandler) CreateClassSkillPool(c echo.Context) error {
	ctx := c.Request().Context()

	var req CreateClassSkillPoolRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 构造实体
	pool := &game_config.ClassSkillPool{
		ClassID:       req.ClassID,
		SkillID:       req.SkillID,
		RequiredLevel: req.RequiredLevel,
	}

	// 设置可选字段
	if req.RequiredAttributes != "" {
		// 验证 JSON 格式
		var jsonData interface{}
		if err := json.Unmarshal([]byte(req.RequiredAttributes), &jsonData); err != nil {
			return response.EchoBadRequest(c, h.respWriter, "required_attributes 必须是有效的 JSON")
		}
		pool.RequiredAttributes.UnmarshalJSON([]byte(req.RequiredAttributes))
	}

	if len(req.PrerequisiteSkillIds) > 0 {
		pool.PrerequisiteSkillIds = req.PrerequisiteSkillIds
	}

	if req.LearnCostXP > 0 {
		pool.LearnCostXP.SetValid(req.LearnCostXP)
	}

	if req.SkillTier > 0 {
		pool.SkillTier.SetValid(req.SkillTier)
	}

	pool.IsCore.SetValid(req.IsCore)
	pool.IsExclusive.SetValid(req.IsExclusive)

	if req.MaxLearnableLevel > 0 {
		pool.MaxLearnableLevel.SetValid(req.MaxLearnableLevel)
	}

	if req.DisplayOrder > 0 {
		pool.DisplayOrder.SetValid(req.DisplayOrder)
	}

	pool.IsVisible.SetValid(req.IsVisible)

	if req.CustomIcon != "" {
		pool.CustomIcon.SetValid(req.CustomIcon)
	}

	if req.CustomDescription != "" {
		pool.CustomDescription.SetValid(req.CustomDescription)
	}

	// 创建
	if err := h.service.CreateClassSkillPool(ctx, pool); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToClassSkillPoolInfo(pool))
}

// UpdateClassSkillPool 更新职业技能池配置
// @Summary 更新职业技能池配置
// @Description 更新指定职业技能池的配置信息,仅更新提供的字段
// @Tags 职业技能池
// @Accept json
// @Produce json
// @Param id path string true "职业技能池ID(UUID格式)"
// @Param request body UpdateClassSkillPoolRequest true "更新请求,仅提供需要更新的字段"
// @Success 200 {object} response.Response{data=object{message=string}} "成功更新"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "配置不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/class-skill-pools/{id} [put]
// @Security BearerAuth
func (h *ClassSkillPoolHandler) UpdateClassSkillPool(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	var req UpdateClassSkillPoolRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	// 构造更新字段
	updates := make(map[string]interface{})

	if req.RequiredLevel > 0 {
		updates[game_config.ClassSkillPoolColumns.RequiredLevel] = req.RequiredLevel
	}
	if req.RequiredAttributes != "" {
		var jsonData interface{}
		if err := json.Unmarshal([]byte(req.RequiredAttributes), &jsonData); err != nil {
			return response.EchoBadRequest(c, h.respWriter, "required_attributes 必须是有效的 JSON")
		}
		updates[game_config.ClassSkillPoolColumns.RequiredAttributes] = json.RawMessage(req.RequiredAttributes)
	}
	// 注意：不更新 prerequisite_skill_ids，因为它是数组类型，需要特殊处理
	// 如果需要更新，请使用 pq.Array 包装
	if len(req.PrerequisiteSkillIds) > 0 {
		updates[game_config.ClassSkillPoolColumns.PrerequisiteSkillIds] = req.PrerequisiteSkillIds
	}
	if req.LearnCostXP > 0 {
		updates[game_config.ClassSkillPoolColumns.LearnCostXP] = req.LearnCostXP
	}
	if req.SkillTier > 0 {
		updates[game_config.ClassSkillPoolColumns.SkillTier] = req.SkillTier
	}
	// Boolean 字段需要特殊处理，因为零值也可能是有效更新
	updates[game_config.ClassSkillPoolColumns.IsCore] = req.IsCore
	updates[game_config.ClassSkillPoolColumns.IsExclusive] = req.IsExclusive
	if req.MaxLearnableLevel > 0 {
		updates[game_config.ClassSkillPoolColumns.MaxLearnableLevel] = req.MaxLearnableLevel
	}
	if req.DisplayOrder >= 0 {
		updates[game_config.ClassSkillPoolColumns.DisplayOrder] = req.DisplayOrder
	}
	updates[game_config.ClassSkillPoolColumns.IsVisible] = req.IsVisible
	if req.CustomIcon != "" {
		updates[game_config.ClassSkillPoolColumns.CustomIcon] = req.CustomIcon
	}
	if req.CustomDescription != "" {
		updates[game_config.ClassSkillPoolColumns.CustomDescription] = req.CustomDescription
	}

	// 更新
	if err := h.service.UpdateClassSkillPool(ctx, id, updates); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": "职业技能池配置更新成功",
	})
}

// DeleteClassSkillPool 删除职业技能池配置
// @Summary 删除职业技能池配置(软删除)
// @Description 从职业技能池中移除指定技能,数据不会被物理删除
// @Tags 职业技能池
// @Accept json
// @Produce json
// @Param id path string true "职业技能池ID(UUID格式)"
// @Success 200 {object} response.Response{data=object{message=string}} "成功删除"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "配置不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/class-skill-pools/{id} [delete]
// @Security BearerAuth
func (h *ClassSkillPoolHandler) DeleteClassSkillPool(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	if err := h.service.DeleteClassSkillPool(ctx, id); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": "职业技能池配置删除成功",
	})
}

// ==================== Helper Functions ====================

func (h *ClassSkillPoolHandler) convertToClassSkillPoolInfo(pool *game_config.ClassSkillPool) ClassSkillPoolInfo {
	info := ClassSkillPoolInfo{
		ID:            pool.ID,
		ClassID:       pool.ClassID,
		SkillID:       pool.SkillID,
		RequiredLevel: pool.RequiredLevel,
	}

	// JSONB 字段处理
	if pool.RequiredAttributes.Valid {
		jsonBytes, _ := pool.RequiredAttributes.MarshalJSON()
		info.RequiredAttributes = string(jsonBytes)
	}

	// 数组字段处理
	if len(pool.PrerequisiteSkillIds) > 0 {
		info.PrerequisiteSkillIds = pool.PrerequisiteSkillIds
	} else {
		info.PrerequisiteSkillIds = []string{}
	}

	// 数值字段处理
	if pool.LearnCostXP.Valid {
		info.LearnCostXP = pool.LearnCostXP.Int
	}

	if pool.SkillTier.Valid {
		info.SkillTier = pool.SkillTier.Int
	}

	if pool.IsCore.Valid {
		info.IsCore = pool.IsCore.Bool
	}

	if pool.IsExclusive.Valid {
		info.IsExclusive = pool.IsExclusive.Bool
	}

	if pool.MaxLearnableLevel.Valid {
		info.MaxLearnableLevel = pool.MaxLearnableLevel.Int
	}

	if pool.DisplayOrder.Valid {
		info.DisplayOrder = pool.DisplayOrder.Int
	}

	if pool.IsVisible.Valid {
		info.IsVisible = pool.IsVisible.Bool
	}

	if pool.CustomIcon.Valid {
		info.CustomIcon = pool.CustomIcon.String
	}

	if pool.CustomDescription.Valid {
		info.CustomDescription = pool.CustomDescription.String
	}

	info.CreatedAt = pool.CreatedAt.Unix()
	info.UpdatedAt = pool.UpdatedAt.Unix()

	return info
}
