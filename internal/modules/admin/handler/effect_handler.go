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

// EffectHandler 效果 HTTP 处理器
type EffectHandler struct {
	service    *service.EffectService
	respWriter response.Writer
}

// NewEffectHandler 创建效果处理器
func NewEffectHandler(db *sql.DB, respWriter response.Writer) *EffectHandler {
	return &EffectHandler{
		service:    service.NewEffectService(db),
		respWriter: respWriter,
	}
}

// ==================== HTTP Models ====================

// CreateEffectRequest 创建效果请求
type CreateEffectRequest struct {
	EffectCode         string   `json:"effect_code" validate:"required,max=50" example:"fire_damage"`                         // 效果唯一代码
	EffectName         string   `json:"effect_name" validate:"required,max=100" example:"火焰伤害"`                               // 效果名称
	EffectType         string   `json:"effect_type" validate:"required,max=50" example:"damage"`                              // 效果类型(damage/heal/buff/debuff等)
	Parameters         string   `json:"parameters" validate:"required" example:"{\"damage_type\":\"fire\",\"dice\":\"2d6\"}"` // 效果参数JSON(必需)
	CalculationFormula string   `json:"calculation_formula" example:"2d6 + @INT"`                                             // 计算公式
	TriggerCondition   string   `json:"trigger_condition" example:"{\"condition\":\"on_hit\",\"probability\":1.0}"`           // 触发条件JSON
	TriggerChance      float64  `json:"trigger_chance" example:"0.85"`                                                        // 触发概率(0-1)
	TargetFilter       string   `json:"target_filter" example:"{\"type\":\"enemy\",\"alive\":true}"`                          // 目标筛选JSON
	FeatureTags        []string `json:"feature_tags" example:"fire,elemental"`                                                // 特性标签数组
	VisualConfig       string   `json:"visual_config" example:"{\"effect\":\"fire_burst\",\"color\":\"#FF4500\"}"`            // 视觉效果配置JSON
	SoundConfig        string   `json:"sound_config" example:"{\"sound\":\"fire_impact.mp3\",\"volume\":0.8}"`                // 音效配置JSON
	Description        string   `json:"description" example:"造成火焰伤害"`                                                         // 效果描述
	TooltipTemplate    string   `json:"tooltip_template" example:"造成{damage}点火焰伤害"`                                           // 提示信息模板
	IsActive           bool     `json:"is_active" example:"true"`                                                             // 是否启用
}

// UpdateEffectRequest 更新效果请求
type UpdateEffectRequest struct {
	EffectCode         string  `json:"effect_code" example:"fire_damage"`                      // 效果唯一代码
	EffectName         string  `json:"effect_name" example:"火焰伤害"`                             // 效果名称
	EffectType         string  `json:"effect_type" example:"damage"`                           // 效果类型(damage/heal/buff/debuff等)
	Parameters         string  `json:"parameters" example:"{\"damage_type\":\"fire\"}"`        // 效果参数JSON
	CalculationFormula string  `json:"calculation_formula" example:"2d6 + @INT"`               // 计算公式
	TriggerCondition   string  `json:"trigger_condition" example:"{\"condition\":\"on_hit\"}"` // 触发条件JSON
	TriggerChance      float64 `json:"trigger_chance" example:"0.85"`                          // 触发概率(0-1)
	TargetFilter       string  `json:"target_filter" example:"{\"type\":\"enemy\"}"`           // 目标筛选JSON
	VisualConfig       string  `json:"visual_config" example:"{\"effect\":\"fire_burst\"}"`    // 视觉效果配置JSON
	SoundConfig        string  `json:"sound_config" example:"{\"sound\":\"fire_impact.mp3\"}"` // 音效配置JSON
	Description        string  `json:"description" example:"造成火焰伤害"`                           // 效果描述
	TooltipTemplate    string  `json:"tooltip_template" example:"造成{damage}点火焰伤害"`             // 提示信息模板
	IsActive           bool    `json:"is_active" example:"true"`                               // 是否启用
}

// EffectInfo 效果信息响应
type EffectInfo struct {
	ID                 string   `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`                // 效果ID
	EffectCode         string   `json:"effect_code" example:"fire_damage"`                                // 效果唯一代码
	EffectName         string   `json:"effect_name" example:"火焰伤害"`                                       // 效果名称
	EffectType         string   `json:"effect_type" example:"damage"`                                     // 效果类型(damage/heal/buff/debuff等)
	Parameters         string   `json:"parameters" example:"{\"damage_type\":\"fire\",\"dice\":\"2d6\"}"` // 效果参数JSON
	CalculationFormula string   `json:"calculation_formula" example:"2d6 + @INT"`                         // 计算公式
	TriggerCondition   string   `json:"trigger_condition" example:"{\"condition\":\"on_hit\"}"`           // 触发条件JSON
	TriggerChance      string   `json:"trigger_chance" example:"0.85"`                                    // 触发概率(0-1)
	TargetFilter       string   `json:"target_filter" example:"{\"type\":\"enemy\"}"`                     // 目标筛选JSON
	FeatureTags        []string `json:"feature_tags" example:"fire,elemental"`                            // 特性标签数组
	VisualConfig       string   `json:"visual_config" example:"{\"effect\":\"fire_burst\"}"`              // 视觉效果配置JSON
	SoundConfig        string   `json:"sound_config" example:"{\"sound\":\"fire_impact.mp3\"}"`           // 音效配置JSON
	Description        string   `json:"description" example:"造成火焰伤害"`                                     // 效果描述
	TooltipTemplate    string   `json:"tooltip_template" example:"造成{damage}点火焰伤害"`                       // 提示信息模板
	IsActive           bool     `json:"is_active" example:"true"`                                         // 是否启用
	CreatedAt          int64    `json:"created_at" example:"1633024800"`                                  // 创建时间戳
	UpdatedAt          int64    `json:"updated_at" example:"1633024800"`                                  // 更新时间戳
}

// ==================== HTTP Handlers ====================

// GetEffects 获取效果列表
// @Summary 获取效果列表
// @Description 分页查询效果列表，支持按效果类型和启用状态筛选
// @Tags 效果系统
// @Accept json
// @Produce json
// @Param effect_type query string false "效果类型，例如: damage, heal, buff"
// @Param is_active query bool false "是否启用，true或false"
// @Param limit query int false "每页数量，默认10"
// @Param offset query int false "偏移量，默认0"
// @Success 200 {object} response.Response{data=map[string]interface{}} "成功返回效果列表和总数"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/effects [get]
// @Security BearerAuth
func (h *EffectHandler) GetEffects(c echo.Context) error {
	ctx := c.Request().Context()

	// 解析查询参数
	params := interfaces.EffectQueryParams{}

	if effectType := c.QueryParam("effect_type"); effectType != "" {
		params.EffectType = &effectType
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
	effects, total, err := h.service.GetEffects(ctx, params)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为响应格式
	result := make([]EffectInfo, len(effects))
	for i, effect := range effects {
		result[i] = h.convertToEffectInfo(effect)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"list":  result,
		"total": total,
	})
}

// GetEffect 获取效果详情
// @Summary 获取效果详情
// @Description 根据效果ID获取效果的详细信息
// @Tags 效果系统
// @Accept json
// @Produce json
// @Param id path string true "效果ID (UUID格式)"
// @Success 200 {object} response.Response{data=EffectInfo} "成功返回效果详情"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "效果不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/effects/{id} [get]
// @Security BearerAuth
func (h *EffectHandler) GetEffect(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	effect, err := h.service.GetEffectByID(ctx, id)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToEffectInfo(effect))
}

// CreateEffect 创建效果
// @Summary 创建效果
// @Description 创建新的游戏效果，包括伤害、治疗、Buff等类型
// @Tags 效果系统
// @Accept json
// @Produce json
// @Param request body CreateEffectRequest true "创建效果请求参数"
// @Success 200 {object} response.Response{data=EffectInfo} "成功返回创建的效果信息"
// @Failure 400 {object} response.Response "参数错误或验证失败"
// @Failure 409 {object} response.Response "效果代码已存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/effects [post]
// @Security BearerAuth
func (h *EffectHandler) CreateEffect(c echo.Context) error {
	ctx := c.Request().Context()

	var req CreateEffectRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 构造效果实体
	effect := &game_config.Effect{
		EffectCode: req.EffectCode,
		EffectName: req.EffectName,
		EffectType: req.EffectType,
	}

	// 验证并设置 Parameters (必需字段)
	var parametersJSON interface{}
	if err := json.Unmarshal([]byte(req.Parameters), &parametersJSON); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "parameters 必须是有效的 JSON")
	}
	effect.Parameters.UnmarshalJSON([]byte(req.Parameters))

	// 设置可选字段
	if req.CalculationFormula != "" {
		effect.CalculationFormula.SetValid(req.CalculationFormula)
	}

	if req.TriggerCondition != "" {
		var triggerConditionJSON interface{}
		if err := json.Unmarshal([]byte(req.TriggerCondition), &triggerConditionJSON); err != nil {
			return response.EchoBadRequest(c, h.respWriter, "trigger_condition 必须是有效的 JSON")
		}
		effect.TriggerCondition.UnmarshalJSON([]byte(req.TriggerCondition))
	}

	if req.TriggerChance > 0 {
		if err := effect.TriggerChance.UnmarshalText([]byte(strconv.FormatFloat(req.TriggerChance, 'f', 2, 64))); err != nil {
			return response.EchoBadRequest(c, h.respWriter, "trigger_chance 格式错误")
		}
	}

	if req.TargetFilter != "" {
		var targetFilterJSON interface{}
		if err := json.Unmarshal([]byte(req.TargetFilter), &targetFilterJSON); err != nil {
			return response.EchoBadRequest(c, h.respWriter, "target_filter 必须是有效的 JSON")
		}
		effect.TargetFilter.UnmarshalJSON([]byte(req.TargetFilter))
	}

	if len(req.FeatureTags) > 0 {
		effect.FeatureTags = req.FeatureTags
	}

	if req.VisualConfig != "" {
		var visualConfigJSON interface{}
		if err := json.Unmarshal([]byte(req.VisualConfig), &visualConfigJSON); err != nil {
			return response.EchoBadRequest(c, h.respWriter, "visual_config 必须是有效的 JSON")
		}
		effect.VisualConfig.UnmarshalJSON([]byte(req.VisualConfig))
	}

	if req.SoundConfig != "" {
		var soundConfigJSON interface{}
		if err := json.Unmarshal([]byte(req.SoundConfig), &soundConfigJSON); err != nil {
			return response.EchoBadRequest(c, h.respWriter, "sound_config 必须是有效的 JSON")
		}
		effect.SoundConfig.UnmarshalJSON([]byte(req.SoundConfig))
	}

	if req.Description != "" {
		effect.Description.SetValid(req.Description)
	}

	if req.TooltipTemplate != "" {
		effect.TooltipTemplate.SetValid(req.TooltipTemplate)
	}

	effect.IsActive.SetValid(req.IsActive)

	// 创建效果
	if err := h.service.CreateEffect(ctx, effect); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToEffectInfo(effect))
}

// UpdateEffect 更新效果
// @Summary 更新效果
// @Description 更新已有效果的信息
// @Tags 效果系统
// @Accept json
// @Produce json
// @Param id path string true "效果ID (UUID格式)"
// @Param request body UpdateEffectRequest true "更新效果请求参数"
// @Success 200 {object} response.Response "成功返回更新结果"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "效果不存在"
// @Failure 409 {object} response.Response "效果代码已存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/effects/{id} [put]
// @Security BearerAuth
func (h *EffectHandler) UpdateEffect(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	var req UpdateEffectRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	// 构造更新字段
	updates := make(map[string]interface{})

	if req.EffectCode != "" {
		updates["effect_code"] = req.EffectCode
	}
	if req.EffectName != "" {
		updates["effect_name"] = req.EffectName
	}
	if req.EffectType != "" {
		updates["effect_type"] = req.EffectType
	}
	updates["description"] = req.Description
	updates["tooltip_template"] = req.TooltipTemplate
	updates["is_active"] = req.IsActive

	// 更新效果
	if err := h.service.UpdateEffect(ctx, id, updates); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": "效果更新成功",
	})
}

// DeleteEffect 删除效果
// @Summary 删除效果
// @Description 软删除指定的效果（设置deleted_at字段）
// @Tags 效果系统
// @Accept json
// @Produce json
// @Param id path string true "效果ID (UUID格式)"
// @Success 200 {object} response.Response "成功返回删除结果"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "效果不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/effects/{id} [delete]
// @Security BearerAuth
func (h *EffectHandler) DeleteEffect(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	if err := h.service.DeleteEffect(ctx, id); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": "效果删除成功",
	})
}

// ==================== Helper Functions ====================

func (h *EffectHandler) convertToEffectInfo(effect *game_config.Effect) EffectInfo {
	info := EffectInfo{
		ID:         effect.ID,
		EffectCode: effect.EffectCode,
		EffectName: effect.EffectName,
		EffectType: effect.EffectType,
	}

	// Parameters (必需字段, types.JSON 类型)
	jsonBytes, _ := effect.Parameters.MarshalJSON()
	info.Parameters = string(jsonBytes)

	// 可选字段
	if effect.CalculationFormula.Valid {
		info.CalculationFormula = effect.CalculationFormula.String
	}

	if effect.TriggerCondition.Valid {
		jsonBytes, _ := effect.TriggerCondition.MarshalJSON()
		info.TriggerCondition = string(jsonBytes)
	}

	if !effect.TriggerChance.IsZero() {
		chanceBytes, _ := effect.TriggerChance.MarshalText()
		info.TriggerChance = string(chanceBytes)
	}

	if effect.TargetFilter.Valid {
		jsonBytes, _ := effect.TargetFilter.MarshalJSON()
		info.TargetFilter = string(jsonBytes)
	}

	if len(effect.FeatureTags) > 0 {
		info.FeatureTags = effect.FeatureTags
	} else {
		info.FeatureTags = []string{}
	}

	if effect.VisualConfig.Valid {
		jsonBytes, _ := effect.VisualConfig.MarshalJSON()
		info.VisualConfig = string(jsonBytes)
	}

	if effect.SoundConfig.Valid {
		jsonBytes, _ := effect.SoundConfig.MarshalJSON()
		info.SoundConfig = string(jsonBytes)
	}

	if effect.Description.Valid {
		info.Description = effect.Description.String
	}

	if effect.TooltipTemplate.Valid {
		info.TooltipTemplate = effect.TooltipTemplate.String
	}

	if effect.IsActive.Valid {
		info.IsActive = effect.IsActive.Bool
	}

	if effect.CreatedAt.Valid {
		info.CreatedAt = effect.CreatedAt.Time.Unix()
	}

	if effect.UpdatedAt.Valid {
		info.UpdatedAt = effect.UpdatedAt.Time.Unix()
	}

	return info
}
