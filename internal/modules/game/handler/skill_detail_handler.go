package handler

import (
	"strconv"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/modules/game/service"
	"tsu-self/internal/pkg/response"
)

// SkillDetailHandler 技能详情查询 Handler
type SkillDetailHandler struct {
	skillDetailService *service.SkillDetailService
	respWriter         response.Writer
}

// NewSkillDetailHandler 创建技能详情查询 Handler
func NewSkillDetailHandler(skillDetailService *service.SkillDetailService, respWriter response.Writer) *SkillDetailHandler {
	return &SkillDetailHandler{
		skillDetailService: skillDetailService,
		respWriter:         respWriter,
	}
}

// ==================== HTTP Response Models ====================

// SkillBasicResponse 技能基本信息响应
type SkillBasicResponse struct {
	ID           string  `json:"id" example:"skill-fireball-001"`                     // 技能ID（UUID格式）
	SkillCode    string  `json:"skill_code" example:"FIREBALL"`                       // 技能代码
	SkillName    string  `json:"skill_name" example:"火球术"`                            // 技能名称
	SkillType    string  `json:"skill_type" example:"active" enums:"active,passive"` // 技能类型：active=主动技能，passive=被动技能
	Description  *string `json:"description,omitempty" example:"发射一枚火球造成伤害"`          // 技能描述（可选）
	MaxLevel     int     `json:"max_level" example:"10"`                              // 技能最大等级
	CategoryID   string  `json:"category_id" example:"cat-fire-magic-001"`            // 分类ID
	CategoryName string  `json:"category_name" example:"火系魔法"`                        // 分类名称
	IsActive     bool    `json:"is_active" example:"true"`                            // 是否启用
}

// UnlockActionInfo 解锁动作信息
type UnlockActionInfo struct {
	ActionID           string                 `json:"action_id" example:"action-fireball-001"`    // 动作ID
	ActionCode         string                 `json:"action_code" example:"FIREBALL_CAST"`        // 动作代码
	ActionName         string                 `json:"action_name" example:"火球术施放"`               // 动作名称
	UnlockLevel        int                    `json:"unlock_level" example:"1"`                   // 解锁等级
	IsDefault          bool                   `json:"is_default" example:"true"`                  // 是否为默认动作
	LevelScalingConfig map[string]interface{} `json:"level_scaling_config,omitempty"`             // 等级成长配置（可选，JSONB格式）
}

// SkillStandardResponse 技能标准响应（含动作信息）
type SkillStandardResponse struct {
	SkillBasicResponse
	UnlockActions []*UnlockActionInfo `json:"unlock_actions"` // 解锁的动作列表
}

// EffectInfo 效果信息
type EffectInfo struct {
	EffectID           string                 `json:"effect_id" example:"effect-damage-001"`          // 效果ID
	EffectCode         string                 `json:"effect_code" example:"DMG_FIRE"`                 // 效果代码
	EffectName         string                 `json:"effect_name" example:"火焰伤害"`                     // 效果名称
	EffectType         string                 `json:"effect_type" example:"DMG_CALCULATION"`          // 效果类型
	ExecutionOrder     int                    `json:"execution_order" example:"1"`                    // 执行顺序
	Parameters         map[string]interface{} `json:"parameters"`                                     // 效果参数（JSONB格式）
	ParameterOverrides map[string]interface{} `json:"parameter_overrides,omitempty"`                  // 参数覆盖（可选，JSONB格式）
}

// ActionDetailInfo 动作详情
type ActionDetailInfo struct {
	ActionType     string                 `json:"action_type" example:"main" enums:"main,minor,reaction"` // 动作类型：main=主要动作，minor=次要动作，reaction=反应动作
	ActionCategory string                 `json:"action_category" example:"BASIC_ATTACK"`                 // 动作类别ID
	RangeConfig    map[string]interface{} `json:"range_config"`                                           // 射程配置（JSONB格式）
	TargetConfig   map[string]interface{} `json:"target_config,omitempty"`                                // 目标配置（可选，JSONB格式）
	HitRateConfig  map[string]interface{} `json:"hit_rate_config,omitempty"`                              // 命中率配置（可选，JSONB格式）
	Description    *string                `json:"description,omitempty" example:"对单个敌人发射火球"`               // 动作描述（可选）
	Effects        []*EffectInfo          `json:"effects"`                                                // 效果列表
}

// UnlockActionFullInfo 解锁动作完整信息
type UnlockActionFullInfo struct {
	ActionID           string                 `json:"action_id" example:"action-fireball-001"`    // 动作ID
	ActionCode         string                 `json:"action_code" example:"FIREBALL_CAST"`        // 动作代码
	ActionName         string                 `json:"action_name" example:"火球术施放"`               // 动作名称
	UnlockLevel        int                    `json:"unlock_level" example:"1"`                   // 解锁等级
	IsDefault          bool                   `json:"is_default" example:"true"`                  // 是否为默认动作
	LevelScalingConfig map[string]interface{} `json:"level_scaling_config,omitempty"`             // 等级成长配置（可选）
	ActionDetails      *ActionDetailInfo      `json:"action_details"`                             // 动作详情
}

// SkillFullResponse 技能完整响应（深度关联）
type SkillFullResponse struct {
	SkillBasicResponse
	UnlockActions []*UnlockActionFullInfo `json:"unlock_actions"` // 解锁的动作完整列表
}

// ==================== HTTP Handlers ====================

// GetSkillBasic 获取技能基本信息
// @Summary 获取技能基本信息（简化版）
// @Description 查询单个技能的基本信息，不包含关联的动作和效果。适合用于技能列表展示
// @Tags 技能查询
// @Accept json
// @Produce json
// @Param skill_id path string true "技能ID（UUID格式）"
// @Success 200 {object} response.Response{data=SkillBasicResponse} "查询成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "技能不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/skills/{skill_id}/basic [get]
func (h *SkillDetailHandler) GetSkillBasic(c echo.Context) error {
	skillID := c.Param("skill_id")
	if skillID == "" {
		return response.EchoBadRequest(c, h.respWriter, "技能ID不能为空")
	}

	detail, err := h.skillDetailService.GetSkillBasic(c.Request().Context(), skillID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, detail)
}

// GetSkillStandard 获取技能标准信息
// @Summary 获取技能标准信息（含动作）
// @Description 查询技能的详细信息，包含解锁的动作列表和等级成长配置。适合用于技能详情页展示
// @Tags 技能查询
// @Accept json
// @Produce json
// @Param skill_id path string true "技能ID（UUID格式）"
// @Success 200 {object} response.Response{data=SkillStandardResponse} "查询成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "技能不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/skills/{skill_id}/standard [get]
func (h *SkillDetailHandler) GetSkillStandard(c echo.Context) error {
	skillID := c.Param("skill_id")
	if skillID == "" {
		return response.EchoBadRequest(c, h.respWriter, "技能ID不能为空")
	}

	detail, err := h.skillDetailService.GetSkillStandard(c.Request().Context(), skillID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, detail)
}

// GetSkillFull 获取技能完整信息
// @Summary 获取技能完整信息（深度关联）
// @Description 查询技能的完整信息，包含动作、效果、参数等所有关联数据。适合用于技能编辑器或详细分析
// @Tags 技能查询
// @Accept json
// @Produce json
// @Param skill_id path string true "技能ID（UUID格式）"
// @Success 200 {object} response.Response{data=SkillFullResponse} "查询成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "技能不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/skills/{skill_id}/full [get]
func (h *SkillDetailHandler) GetSkillFull(c echo.Context) error {
	skillID := c.Param("skill_id")
	if skillID == "" {
		return response.EchoBadRequest(c, h.respWriter, "技能ID不能为空")
	}

	detail, err := h.skillDetailService.GetSkillFull(c.Request().Context(), skillID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, detail)
}

// ListSkillsBasic 获取技能列表（简化版）
// @Summary 获取技能列表（简化版）
// @Description 查询技能列表，仅返回基本信息，支持按类型、分类筛选和分页
// @Tags 技能查询
// @Accept json
// @Produce json
// @Param skill_type query string false "技能类型筛选" enums(active,passive)
// @Param category_id query string false "分类ID筛选（UUID格式）"
// @Param is_active query boolean false "是否启用筛选"
// @Param limit query int false "每页数量（默认20）" default(20) minimum(1) maximum(100)
// @Param offset query int false "偏移量（默认0）" default(0) minimum(0)
// @Success 200 {object} response.Response{data=object{list=[]SkillBasicResponse,total=int}} "查询成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/skills/basic [get]
func (h *SkillDetailHandler) ListSkillsBasic(c echo.Context) error {
	params := h.parseQueryParams(c)

	list, total, err := h.skillDetailService.ListSkillsBasic(c.Request().Context(), params)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"list":  list,
		"total": total,
	})
}

// ListSkillsStandard 获取技能列表（标准版）
// @Summary 获取技能列表（标准版）
// @Description 查询技能列表，包含动作信息，支持按类型、分类筛选和分页。适合用于技能库浏览
// @Tags 技能查询
// @Accept json
// @Produce json
// @Param skill_type query string false "技能类型筛选" enums(active,passive)
// @Param category_id query string false "分类ID筛选（UUID格式）"
// @Param is_active query boolean false "是否启用筛选"
// @Param limit query int false "每页数量（默认20）" default(20) minimum(1) maximum(100)
// @Param offset query int false "偏移量（默认0）" default(0) minimum(0)
// @Success 200 {object} response.Response{data=object{list=[]SkillStandardResponse,total=int}} "查询成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/skills/standard [get]
func (h *SkillDetailHandler) ListSkillsStandard(c echo.Context) error {
	params := h.parseQueryParams(c)

	list, total, err := h.skillDetailService.ListSkillsStandard(c.Request().Context(), params)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"list":  list,
		"total": total,
	})
}

// parseQueryParams 解析查询参数
func (h *SkillDetailHandler) parseQueryParams(c echo.Context) service.SkillQueryParams {
	params := service.SkillQueryParams{}

	// 技能类型
	if skillType := c.QueryParam("skill_type"); skillType != "" {
		params.SkillType = &skillType
	}

	// 分类ID
	if categoryID := c.QueryParam("category_id"); categoryID != "" {
		params.CategoryID = &categoryID
	}

	// 是否启用
	if isActiveStr := c.QueryParam("is_active"); isActiveStr != "" {
		isActive := isActiveStr == "true"
		params.IsActive = &isActive
	}

	// 分页参数
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 20 // 默认每页20条
	}
	params.Limit = limit

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}
	params.Offset = offset

	return params
}
