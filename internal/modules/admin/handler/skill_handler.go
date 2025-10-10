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

// SkillHandler 技能 HTTP 处理器
type SkillHandler struct {
	service    *service.SkillService
	respWriter response.Writer
}

// NewSkillHandler 创建技能处理器
func NewSkillHandler(db *sql.DB, respWriter response.Writer) *SkillHandler {
	return &SkillHandler{
		service:    service.NewSkillService(db),
		respWriter: respWriter,
	}
}

// ==================== HTTP Models ====================

// CreateSkillRequest 创建技能请求
// 注意：技能的学习限制（等级要求、职业限制、前置技能）现在由 class_skill_pools 表管理
type CreateSkillRequest struct {
	SkillCode           string   `json:"skill_code" validate:"required,max=50" example:"fireball"`   // 技能唯一代码
	SkillName           string   `json:"skill_name" validate:"required,max=100" example:"火球术"`       // 技能名称
	SkillType           string   `json:"skill_type" validate:"required" example:"magic"`             // 技能类型: weapon(武器技能), magic(魔法技能), physical(物理技能), usage(使用物品技能), reaction(反应技能), guard(防御技能), movement(移动技能), command(指挥技能)
	CategoryID          string   `json:"category_id" example:"550e8400-e29b-41d4-a716-446655440000"` // 技能类别ID
	MaxLevel            int      `json:"max_level" example:"5"`                                      // 最大等级
	FeatureTags         []string `json:"feature_tags" example:"fire,damage"`                         // 特性标签数组
	PassiveEffects      string   `json:"passive_effects" example:"{\"effects\":[]}"`                 // 被动效果JSON配置
	Description         string   `json:"description" example:"释放一个火球攻击敌人"`                           // 简短描述
	DetailedDescription string   `json:"detailed_description" example:"向目标发射一个强大的火球,造成火焰伤害"`         // 详细描述
	Icon                string   `json:"icon" example:"fireball_icon.png"`                           // 技能图标路径
	IsActive            bool     `json:"is_active" example:"true"`                                   // 是否启用
}

// UpdateSkillRequest 更新技能请求
type UpdateSkillRequest struct {
	SkillCode           string `json:"skill_code" example:"fireball"`                              // 技能唯一代码
	SkillName           string `json:"skill_name" example:"火球术"`                                   // 技能名称
	SkillType           string `json:"skill_type" example:"magic"`                                 // 技能类型: weapon(武器技能), magic(魔法技能), physical(物理技能), usage(使用物品技能), reaction(反应技能), guard(防御技能), movement(移动技能), command(指挥技能)
	CategoryID          string `json:"category_id" example:"550e8400-e29b-41d4-a716-446655440000"` // 技能类别ID
	MaxLevel            int    `json:"max_level" example:"5"`                                      // 最大等级
	Description         string `json:"description" example:"释放一个火球攻击敌人"`                           // 简短描述
	DetailedDescription string `json:"detailed_description" example:"向目标发射一个强大的火球"`                // 详细描述
	Icon                string `json:"icon" example:"fireball_icon.png"`                           // 技能图标路径
	IsActive            bool   `json:"is_active" example:"true"`                                   // 是否启用
}

// SkillInfo 技能信息响应
// 注意：技能的学习限制（等级要求、职业限制、前置技能）现在由 class_skill_pools 表管理
type SkillInfo struct {
	ID                  string   `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`          // 技能ID
	SkillCode           string   `json:"skill_code" example:"fireball"`                              // 技能唯一代码
	SkillName           string   `json:"skill_name" example:"火球术"`                                   // 技能名称
	SkillType           string   `json:"skill_type" example:"magic"`                                 // 技能类型: weapon(武器技能), magic(魔法技能), physical(物理技能), usage(使用物品技能), reaction(反应技能), guard(防御技能), movement(移动技能), command(指挥技能)
	CategoryID          string   `json:"category_id" example:"550e8400-e29b-41d4-a716-446655440000"` // 技能类别ID
	MaxLevel            int      `json:"max_level" example:"5"`                                      // 最大等级
	FeatureTags         []string `json:"feature_tags" example:"fire,damage"`                         // 特性标签数组
	PassiveEffects      string   `json:"passive_effects" example:"{\"effects\":[]}"`                 // 被动效果JSON配置
	Description         string   `json:"description" example:"释放一个火球攻击敌人"`                           // 简短描述
	DetailedDescription string   `json:"detailed_description" example:"向目标发射一个强大的火球"`                // 详细描述
	Icon                string   `json:"icon" example:"fireball_icon.png"`                           // 技能图标路径
	IsActive            bool     `json:"is_active" example:"true"`                                   // 是否启用
	CreatedAt           int64    `json:"created_at" example:"1633024800"`                            // 创建时间戳
	UpdatedAt           int64    `json:"updated_at" example:"1633024800"`                            // 更新时间戳
}

// ==================== HTTP Handlers ====================

// GetSkills 获取技能列表
// @Summary 获取技能列表
// @Description 分页查询技能列表,支持按技能类型、类别、启用状态筛选
// @Tags 技能系统
// @Accept json
// @Produce json
// @Param skill_type query string false "技能类型: weapon(武器技能), magic(魔法技能), physical(物理技能), usage(使用物品技能), reaction(反应技能), guard(防御技能), movement(移动技能), command(指挥技能)"
// @Param category_id query string false "类别ID (UUID格式)"
// @Param is_active query bool false "是否启用 (true/false)"
// @Param limit query int false "每页数量 (默认10)" default(10)
// @Param offset query int false "偏移量 (默认0)" default(0)
// @Success 200 {object} response.Response{data=object{list=[]SkillInfo,total=int}} "成功返回技能列表,包含 list 和 total 字段"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/skills [get]
// @Security BearerAuth
func (h *SkillHandler) GetSkills(c echo.Context) error {
	ctx := c.Request().Context()

	// 解析查询参数
	params := interfaces.SkillQueryParams{}

	if skillType := c.QueryParam("skill_type"); skillType != "" {
		params.SkillType = &skillType
	}

	if categoryID := c.QueryParam("category_id"); categoryID != "" {
		params.CategoryID = &categoryID
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

	// 查询列表
	skills, total, err := h.service.GetSkills(ctx, params)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为响应格式
	result := make([]SkillInfo, len(skills))
	for i, skill := range skills {
		result[i] = h.convertToSkillInfo(skill)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"list":  result,
		"total": total,
	})
}

// GetSkill 获取技能详情
// @Summary 获取技能详情
// @Description 根据技能ID获取单个技能的完整信息
// @Tags 技能系统
// @Accept json
// @Produce json
// @Param id path string true "技能ID (UUID格式)" example("01d132ed-6378-4e0b-bc16-a5b224e5b04a")
// @Success 200 {object} response.Response{data=SkillInfo} "成功返回技能详情"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "技能不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/skills/{id} [get]
// @Security BearerAuth
func (h *SkillHandler) GetSkill(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	skill, err := h.service.GetSkillByID(ctx, id)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToSkillInfo(skill))
}

// CreateSkill 创建技能
// @Summary 创建技能
// @Description 创建新技能,包含技能代码、名称、类型、特性标签等配置信息
// @Tags 技能系统
// @Accept json
// @Produce json
// @Param request body CreateSkillRequest true "创建技能请求,skill_code、skill_name、skill_type 为必填字段"
// @Success 200 {object} response.Response{data=SkillInfo} "成功创建技能,返回技能详情"
// @Failure 400 {object} response.Response "请求参数错误或验证失败"
// @Failure 409 {object} response.Response "技能代码已存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/skills [post]
// @Security BearerAuth
func (h *SkillHandler) CreateSkill(c echo.Context) error {
	ctx := c.Request().Context()

	var req CreateSkillRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 构造技能实体
	skill := &game_config.Skill{
		SkillCode: req.SkillCode,
		SkillName: req.SkillName,
		SkillType: req.SkillType,
	}

	if req.CategoryID != "" {
		skill.CategoryID.SetValid(req.CategoryID)
	}

	if req.MaxLevel > 0 {
		skill.MaxLevel.SetValid(req.MaxLevel)
	}

	if len(req.FeatureTags) > 0 {
		skill.FeatureTags = req.FeatureTags
	}

	if req.PassiveEffects != "" {
		// 验证 JSON 格式
		var jsonData interface{}
		if err := json.Unmarshal([]byte(req.PassiveEffects), &jsonData); err != nil {
			return response.EchoBadRequest(c, h.respWriter, "passive_effects 必须是有效的 JSON")
		}
		skill.PassiveEffects.UnmarshalJSON([]byte(req.PassiveEffects))
	}

	if req.Description != "" {
		skill.Description.SetValid(req.Description)
	}

	if req.DetailedDescription != "" {
		skill.DetailedDescription.SetValid(req.DetailedDescription)
	}

	if req.Icon != "" {
		skill.Icon.SetValid(req.Icon)
	}

	skill.IsActive.SetValid(req.IsActive)

	// 创建技能
	if err := h.service.CreateSkill(ctx, skill); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToSkillInfo(skill))
}

// UpdateSkill 更新技能
// @Summary 更新技能
// @Description 更新指定技能的配置信息,仅更新提供的字段
// @Tags 技能系统
// @Accept json
// @Produce json
// @Param id path string true "技能ID (UUID格式)" example("01d132ed-6378-4e0b-bc16-a5b224e5b04a")
// @Param request body UpdateSkillRequest true "更新技能请求,仅提供需要更新的字段"
// @Success 200 {object} response.Response{data=map[string]string} "成功更新技能,返回成功消息"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "技能不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/skills/{id} [put]
// @Security BearerAuth
func (h *SkillHandler) UpdateSkill(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	var req UpdateSkillRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	// 构造更新字段
	updates := make(map[string]interface{})

	if req.SkillCode != "" {
		updates["skill_code"] = req.SkillCode
	}
	if req.SkillName != "" {
		updates["skill_name"] = req.SkillName
	}
	if req.SkillType != "" {
		updates["skill_type"] = req.SkillType
	}
	updates["category_id"] = req.CategoryID // 允许清空
	if req.MaxLevel > 0 {
		updates["max_level"] = req.MaxLevel
	}
	updates["description"] = req.Description
	updates["detailed_description"] = req.DetailedDescription
	updates["icon"] = req.Icon
	updates["is_active"] = req.IsActive

	// 更新技能
	if err := h.service.UpdateSkill(ctx, id, updates); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": "技能更新成功",
	})
}

// DeleteSkill 删除技能
// @Summary 删除技能 (软删除)
// @Description 软删除指定技能,技能数据不会被物理删除,仅标记为已删除
// @Tags 技能系统
// @Accept json
// @Produce json
// @Param id path string true "技能ID (UUID格式)" example("01d132ed-6378-4e0b-bc16-a5b224e5b04a")
// @Success 200 {object} response.Response{data=map[string]string} "成功删除技能,返回成功消息"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "技能不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/skills/{id} [delete]
// @Security BearerAuth
func (h *SkillHandler) DeleteSkill(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	if err := h.service.DeleteSkill(ctx, id); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": "技能删除成功",
	})
}

// ==================== Helper Functions ====================

func (h *SkillHandler) convertToSkillInfo(skill *game_config.Skill) SkillInfo {
	info := SkillInfo{
		ID:        skill.ID,
		SkillCode: skill.SkillCode,
		SkillName: skill.SkillName,
		SkillType: skill.SkillType,
	}

	if skill.CategoryID.Valid {
		info.CategoryID = skill.CategoryID.String
	}

	if skill.MaxLevel.Valid {
		info.MaxLevel = skill.MaxLevel.Int
	}

	// types.StringArray 处理
	if len(skill.FeatureTags) > 0 {
		info.FeatureTags = skill.FeatureTags
	} else {
		info.FeatureTags = []string{}
	}

	// null.JSON 处理
	if skill.PassiveEffects.Valid {
		jsonBytes, _ := skill.PassiveEffects.MarshalJSON()
		info.PassiveEffects = string(jsonBytes)
	}

	if skill.Description.Valid {
		info.Description = skill.Description.String
	}

	if skill.DetailedDescription.Valid {
		info.DetailedDescription = skill.DetailedDescription.String
	}

	if skill.Icon.Valid {
		info.Icon = skill.Icon.String
	}

	if skill.IsActive.Valid {
		info.IsActive = skill.IsActive.Bool
	}

	if skill.CreatedAt.Valid {
		info.CreatedAt = skill.CreatedAt.Time.Unix()
	}

	if skill.UpdatedAt.Valid {
		info.UpdatedAt = skill.UpdatedAt.Time.Unix()
	}

	return info
}
