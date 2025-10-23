package handler

import (
	"database/sql"
	"encoding/json"
	"strconv"

	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/labstack/echo/v4"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/modules/admin/service"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/repository/interfaces"
)

// ActionHandler 动作 HTTP 处理器
type ActionHandler struct {
	service    *service.ActionService
	respWriter response.Writer
}

// NewActionHandler 创建动作处理器
func NewActionHandler(db *sql.DB, respWriter response.Writer) *ActionHandler {
	return &ActionHandler{
		service:    service.NewActionService(db),
		respWriter: respWriter,
	}
}

// ==================== HTTP Models ====================

// CreateActionRequest 创建动作请求
type CreateActionRequest struct {
	ActionCode         string   `json:"action_code" validate:"required,max=50" example:"attack"`                     // 动作唯一代码
	ActionName         string   `json:"action_name" validate:"required,max=100" example:"普通攻击"`                      // 动作名称
	ActionType         string   `json:"action_type" validate:"required" example:"main"`                              // 动作类型(main/minor/reaction)
	ActionCategoryID   string   `json:"action_category_id" example:"550e8400-e29b-41d4-a716-446655440000"`           // 动作类别ID
	RelatedSkillID     string   `json:"related_skill_id" example:"550e8400-e29b-41d4-a716-446655440000"`             // 关联技能ID
	FeatureTags        []string `json:"feature_tags" example:"melee,physical"`                                       // 特性标签数组
	RangeConfig        string   `json:"range_config" validate:"required" example:"{\"type\":\"melee\",\"range\":5}"` // 范围配置JSON(必需)
	TargetConfig       string   `json:"target_config" example:"{\"type\":\"single\",\"target\":\"enemy\"}"`          // 目标配置JSON
	AreaConfig         string   `json:"area_config" example:"{\"shape\":\"circle\",\"radius\":3}"`                   // 区域配置JSON
	ActionPointCost    int      `json:"action_point_cost" example:"1"`                                               // 行动点消耗
	ManaCost           int      `json:"mana_cost" example:"10"`                                                      // 法力消耗
	ManaCostFormula    string   `json:"mana_cost_formula" example:"base_cost + level * 2"`                           // 法力消耗公式
	CooldownTurns      int      `json:"cooldown_turns" example:"3"`                                                  // 冷却回合数
	UsesPerBattle      int      `json:"uses_per_battle" example:"5"`                                                 // 每场战斗可用次数
	HitRateConfig      string   `json:"hit_rate_config" example:"{\"base\":85,\"modifier\":\"dex\"}"`                // 命中率配置JSON
	LegacyEffectConfig string   `json:"legacy_effect_config" example:"{\"damage\":\"2d6+3\"}"`                       // Excel原始效果配置（用于兼容导入）
	Requirements       string   `json:"requirements" example:"{\"min_level\":5,\"class\":\"warrior\"}"`              // 需求配置JSON
	StartFlags         []string `json:"start_flags" example:"combat_start,first_round"`                              // 起始标记数组
	AnimationConfig    string   `json:"animation_config" example:"{\"animation\":\"swing\",\"duration\":1000}"`      // 动画配置JSON
	VisualEffects      string   `json:"visual_effects" example:"{\"effect\":\"slash\",\"color\":\"red\"}"`           // 视觉效果JSON
	SoundEffects       string   `json:"sound_effects" example:"{\"sound\":\"sword_swing.mp3\"}"`                     // 音效配置JSON
	Description        string   `json:"description" example:"使用武器进行近战攻击"`                                            // 动作描述
	IsActive           bool     `json:"is_active" example:"true"`                                                    // 是否启用
}

// UpdateActionRequest 更新动作请求
type UpdateActionRequest struct {
	ActionCode         string   `json:"action_code" example:"attack"`                                      // 动作唯一代码
	ActionName         string   `json:"action_name" example:"普通攻击"`                                        // 动作名称
	ActionType         string   `json:"action_type" example:"main"`                                        // 动作类型(main/minor/reaction)
	ActionCategoryID   string   `json:"action_category_id" example:"550e8400-e29b-41d4-a716-446655440000"` // 动作类别ID
	RelatedSkillID     string   `json:"related_skill_id" example:"550e8400-e29b-41d4-a716-446655440000"`   // 关联技能ID
	FeatureTags        []string `json:"feature_tags" example:"melee,physical"`                             // 特性标签数组
	RangeConfig        string   `json:"range_config" example:"{\"type\":\"melee\",\"range\":5}"`           // 范围配置JSON
	TargetConfig       string   `json:"target_config" example:"{\"type\":\"single\"}"`                     // 目标配置JSON
	AreaConfig         string   `json:"area_config" example:"{\"shape\":\"circle\",\"radius\":3}"`         // 区域配置JSON
	ActionPointCost    int      `json:"action_point_cost" example:"1"`                                     // 行动点消耗
	ManaCost           int      `json:"mana_cost" example:"10"`                                            // 法力消耗
	ManaCostFormula    string   `json:"mana_cost_formula" example:"base_cost + level * 2"`                 // 法力消耗公式
	CooldownTurns      int      `json:"cooldown_turns" example:"3"`                                        // 冷却回合数
	UsesPerBattle      int      `json:"uses_per_battle" example:"5"`                                       // 每场战斗可用次数
	HitRateConfig      string   `json:"hit_rate_config" example:"{\"base\":85}"`                           // 命中率配置JSON
	LegacyEffectConfig string   `json:"legacy_effect_config" example:"{\"damage\":\"2d6+3\"}"`             // Excel原始效果配置（用于兼容导入）
	Requirements       string   `json:"requirements" example:"{\"min_level\":5}"`                          // 需求配置JSON
	StartFlags         []string `json:"start_flags" example:"combat_start"`                                // 起始标记数组
	AnimationConfig    string   `json:"animation_config" example:"{\"animation\":\"swing\"}"`              // 动画配置JSON
	VisualEffects      string   `json:"visual_effects" example:"{\"effect\":\"slash\"}"`                   // 视觉效果JSON
	SoundEffects       string   `json:"sound_effects" example:"{\"sound\":\"sword_swing.mp3\"}"`           // 音效配置JSON
	Description        string   `json:"description" example:"使用武器进行近战攻击"`                                  // 动作描述
	IsActive           bool     `json:"is_active" example:"true"`                                          // 是否启用
}

// ActionInfo 动作信息响应
type ActionInfo struct {
	ID                 string   `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`                 // 动作ID
	ActionCode         string   `json:"action_code" example:"attack"`                                      // 动作唯一代码
	ActionName         string   `json:"action_name" example:"普通攻击"`                                        // 动作名称
	ActionType         string   `json:"action_type" example:"main"`                                        // 动作类型(main/minor/reaction)
	ActionCategoryID   string   `json:"action_category_id" example:"550e8400-e29b-41d4-a716-446655440000"` // 动作类别ID
	RelatedSkillID     string   `json:"related_skill_id" example:"550e8400-e29b-41d4-a716-446655440000"`   // 关联技能ID
	FeatureTags        []string `json:"feature_tags" example:"melee,physical"`                             // 特性标签数组
	RangeConfig        string   `json:"range_config" example:"{\"type\":\"melee\",\"range\":5}"`           // 范围配置JSON
	TargetConfig       string   `json:"target_config" example:"{\"type\":\"single\"}"`                     // 目标配置JSON
	AreaConfig         string   `json:"area_config" example:"{\"shape\":\"circle\",\"radius\":3}"`         // 区域配置JSON
	ActionPointCost    int      `json:"action_point_cost" example:"1"`                                     // 行动点消耗
	ManaCost           int      `json:"mana_cost" example:"10"`                                            // 法力消耗
	ManaCostFormula    string   `json:"mana_cost_formula" example:"base_cost + level * 2"`                 // 法力消耗公式
	CooldownTurns      int      `json:"cooldown_turns" example:"3"`                                        // 冷却回合数
	UsesPerBattle      int      `json:"uses_per_battle" example:"5"`                                       // 每场战斗可用次数
	HitRateConfig      string   `json:"hit_rate_config" example:"{\"base\":85}"`                           // 命中率配置JSON
	LegacyEffectConfig string   `json:"legacy_effect_config" example:"{\"damage\":\"2d6+3\"}"`             // Excel原始效果配置（用于兼容导入）
	Requirements       string   `json:"requirements" example:"{\"min_level\":5}"`                          // 需求配置JSON
	StartFlags         []string `json:"start_flags" example:"combat_start"`                                // 起始标记数组
	AnimationConfig    string   `json:"animation_config" example:"{\"animation\":\"swing\"}"`              // 动画配置JSON
	VisualEffects      string   `json:"visual_effects" example:"{\"effect\":\"slash\"}"`                   // 视觉效果JSON
	SoundEffects       string   `json:"sound_effects" example:"{\"sound\":\"sword_swing.mp3\"}"`           // 音效配置JSON
	Description        string   `json:"description" example:"使用武器进行近战攻击"`                                  // 动作描述
	IsActive           bool     `json:"is_active" example:"true"`                                          // 是否启用
	CreatedAt          int64    `json:"created_at" example:"1633024800"`                                   // 创建时间戳
	UpdatedAt          int64    `json:"updated_at" example:"1633024800"`                                   // 更新时间戳
}

// ==================== HTTP Handlers ====================

// GetActions 获取动作列表
// @Summary 获取动作列表
// @Description 获取动作列表,支持分页和多条件筛选(动作类型、分类ID、启用状态)
// @Tags 动作
// @Accept json
// @Produce json
// @Param action_type query string false "动作类型筛选(main/minor/reaction)"
// @Param action_category_id query string false "分类ID筛选"
// @Param is_active query bool false "是否启用筛选"
// @Param limit query int false "每页数量"
// @Param offset query int false "偏移量"
// @Success 200 {object} response.Response{data=object{list=[]ActionInfo,total=int}}
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth
// @Router /admin/actions [get]
func (h *ActionHandler) GetActions(c echo.Context) error {
	ctx := c.Request().Context()

	// 解析查询参数
	params := interfaces.ActionQueryParams{}

	if actionType := c.QueryParam("action_type"); actionType != "" {
		params.ActionType = &actionType
	}

	if categoryID := c.QueryParam("action_category_id"); categoryID != "" {
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
	actions, total, err := h.service.GetActions(ctx, params)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为响应格式
	result := make([]ActionInfo, len(actions))
	for i, action := range actions {
		result[i] = h.convertToActionInfo(action)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"list":  result,
		"total": total,
	})
}

// GetAction 获取动作详情
// @Summary 获取动作详情
// @Description 根据ID获取单个动作的详细信息,包括范围配置、目标配置等所有JSONB字段
// @Tags 动作
// @Accept json
// @Produce json
// @Param id path string true "动作的UUID"
// @Success 200 {object} response.Response{data=ActionInfo}
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "动作不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth
// @Router /admin/actions/{id} [get]
func (h *ActionHandler) GetAction(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	action, err := h.service.GetActionByID(ctx, id)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToActionInfo(action))
}

// CreateAction 创建动作
// @Summary 创建动作
// @Description 创建新的动作,action_code必须唯一,range_config为必需JSONB字段
// @Tags 动作
// @Accept json
// @Produce json
// @Param request body CreateActionRequest true "创建动作的请求参数,包含范围配置、目标配置等"
// @Success 200 {object} response.Response{data=ActionInfo}
// @Failure 400 {object} response.Response "参数错误或验证失败(包括JSONB格式错误)"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth
// @Router /admin/actions [post]
func (h *ActionHandler) CreateAction(c echo.Context) error {
	ctx := c.Request().Context()

	var req CreateActionRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 构造动作实体
	action := &game_config.Action{
		ActionCode: req.ActionCode,
		ActionName: req.ActionName,
		ActionType: req.ActionType,
	}

	// 设置可选外键
	if req.ActionCategoryID != "" {
		action.ActionCategoryID.SetValid(req.ActionCategoryID)
	}

	if req.RelatedSkillID != "" {
		action.RelatedSkillID.SetValid(req.RelatedSkillID)
	}

	// 设置数组字段
	if len(req.FeatureTags) > 0 {
		action.FeatureTags = types.StringArray(req.FeatureTags)
	}

	if len(req.StartFlags) > 0 {
		action.StartFlags = types.StringArray(req.StartFlags)
	}

	// 验证并设置 RangeConfig (必需字段)
	var rangeConfigJSON interface{}
	if err := json.Unmarshal([]byte(req.RangeConfig), &rangeConfigJSON); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "range_config 必须是有效的 JSON")
	}
	action.RangeConfig.UnmarshalJSON([]byte(req.RangeConfig))

	// 设置可选 JSONB 字段
	if req.TargetConfig != "" {
		var targetConfigJSON interface{}
		if err := json.Unmarshal([]byte(req.TargetConfig), &targetConfigJSON); err != nil {
			return response.EchoBadRequest(c, h.respWriter, "target_config 必须是有效的 JSON")
		}
		action.TargetConfig.UnmarshalJSON([]byte(req.TargetConfig))
	}

	if req.AreaConfig != "" {
		var areaConfigJSON interface{}
		if err := json.Unmarshal([]byte(req.AreaConfig), &areaConfigJSON); err != nil {
			return response.EchoBadRequest(c, h.respWriter, "area_config 必须是有效的 JSON")
		}
		action.AreaConfig.UnmarshalJSON([]byte(req.AreaConfig))
	}

	if req.HitRateConfig != "" {
		action.HitRateConfig.UnmarshalJSON([]byte(req.HitRateConfig))
	}

	if req.LegacyEffectConfig != "" {
		action.LegacyEffectConfig.UnmarshalJSON([]byte(req.LegacyEffectConfig))
	}

	if req.Requirements != "" {
		action.Requirements.UnmarshalJSON([]byte(req.Requirements))
	}

	if req.AnimationConfig != "" {
		action.AnimationConfig.UnmarshalJSON([]byte(req.AnimationConfig))
	}

	if req.VisualEffects != "" {
		action.VisualEffects.UnmarshalJSON([]byte(req.VisualEffects))
	}

	if req.SoundEffects != "" {
		action.SoundEffects.UnmarshalJSON([]byte(req.SoundEffects))
	}

	// 设置数值字段
	if req.ActionPointCost > 0 {
		action.ActionPointCost.SetValid(req.ActionPointCost)
	}

	if req.ManaCost > 0 {
		action.ManaCost.SetValid(req.ManaCost)
	}

	if req.ManaCostFormula != "" {
		action.ManaCostFormula.SetValid(req.ManaCostFormula)
	}

	if req.CooldownTurns > 0 {
		action.CooldownTurns.SetValid(req.CooldownTurns)
	}

	if req.UsesPerBattle > 0 {
		action.UsesPerBattle.SetValid(req.UsesPerBattle)
	}

	if req.Description != "" {
		action.Description.SetValid(req.Description)
	}

	action.IsActive.SetValid(req.IsActive)

	// 创建动作
	if err := h.service.CreateAction(ctx, action); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToActionInfo(action))
}

// UpdateAction 更新动作
// @Summary 更新动作
// @Description 更新指定ID的动作信息,支持部分字段更新
// @Tags 动作
// @Accept json
// @Produce json
// @Param id path string true "动作的UUID"
// @Param request body UpdateActionRequest true "更新动作的请求参数"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response "参数错误或验证失败"
// @Failure 404 {object} response.Response "动作不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth
// @Router /admin/actions/{id} [put]
func (h *ActionHandler) UpdateAction(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	var req UpdateActionRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	// 构造更新字段
	updates := make(map[string]interface{})

	if req.ActionCode != "" {
		updates["action_code"] = req.ActionCode
	}
	if req.ActionName != "" {
		updates["action_name"] = req.ActionName
	}
	if req.ActionType != "" {
		updates["action_type"] = req.ActionType
	}
	updates["action_category_id"] = req.ActionCategoryID
	updates["related_skill_id"] = req.RelatedSkillID
	updates["description"] = req.Description
	updates["mana_cost_formula"] = req.ManaCostFormula
	updates["legacy_effect_config"] = req.LegacyEffectConfig
	updates["is_active"] = req.IsActive

	// 更新数组字段
	if req.FeatureTags != nil {
		updates[game_config.ActionColumns.FeatureTags] = types.StringArray(req.FeatureTags)
	}
	if req.StartFlags != nil {
		updates[game_config.ActionColumns.StartFlags] = types.StringArray(req.StartFlags)
	}

	// 更新动作
	if err := h.service.UpdateAction(ctx, id, updates); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": "动作更新成功",
	})
}

// DeleteAction 删除动作
// @Summary 删除动作
// @Description 软删除指定ID的动作(设置deleted_at字段)
// @Tags 动作
// @Accept json
// @Produce json
// @Param id path string true "动作的UUID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "动作不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth
// @Router /admin/actions/{id} [delete]
func (h *ActionHandler) DeleteAction(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	if err := h.service.DeleteAction(ctx, id); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": "动作删除成功",
	})
}

// ==================== Helper Functions ====================

func (h *ActionHandler) convertToActionInfo(action *game_config.Action) ActionInfo {
	info := ActionInfo{
		ID:         action.ID,
		ActionCode: action.ActionCode,
		ActionName: action.ActionName,
		ActionType: action.ActionType,
	}

	// 外键字段
	if action.ActionCategoryID.Valid {
		info.ActionCategoryID = action.ActionCategoryID.String
	}

	if action.RelatedSkillID.Valid {
		info.RelatedSkillID = action.RelatedSkillID.String
	}

	// 数组字段
	if len(action.FeatureTags) > 0 {
		info.FeatureTags = action.FeatureTags
	} else {
		info.FeatureTags = []string{}
	}

	if len(action.StartFlags) > 0 {
		info.StartFlags = action.StartFlags
	} else {
		info.StartFlags = []string{}
	}

	// JSONB 字段
	rangeBytes, _ := action.RangeConfig.MarshalJSON()
	info.RangeConfig = string(rangeBytes)

	if action.TargetConfig.Valid {
		targetBytes, _ := action.TargetConfig.MarshalJSON()
		info.TargetConfig = string(targetBytes)
	}

	if action.AreaConfig.Valid {
		areaBytes, _ := action.AreaConfig.MarshalJSON()
		info.AreaConfig = string(areaBytes)
	}

	if action.HitRateConfig.Valid {
		hitRateBytes, _ := action.HitRateConfig.MarshalJSON()
		info.HitRateConfig = string(hitRateBytes)
	}

	if action.LegacyEffectConfig.Valid {
		legacyBytes, _ := action.LegacyEffectConfig.MarshalJSON()
		info.LegacyEffectConfig = string(legacyBytes)
	}

	if action.Requirements.Valid {
		reqBytes, _ := action.Requirements.MarshalJSON()
		info.Requirements = string(reqBytes)
	}

	if action.AnimationConfig.Valid {
		animBytes, _ := action.AnimationConfig.MarshalJSON()
		info.AnimationConfig = string(animBytes)
	}

	if action.VisualEffects.Valid {
		visualBytes, _ := action.VisualEffects.MarshalJSON()
		info.VisualEffects = string(visualBytes)
	}

	if action.SoundEffects.Valid {
		soundBytes, _ := action.SoundEffects.MarshalJSON()
		info.SoundEffects = string(soundBytes)
	}

	// 数值字段
	if action.ActionPointCost.Valid {
		info.ActionPointCost = action.ActionPointCost.Int
	}

	if action.ManaCost.Valid {
		info.ManaCost = action.ManaCost.Int
	}

	if action.ManaCostFormula.Valid {
		info.ManaCostFormula = action.ManaCostFormula.String
	}

	if action.CooldownTurns.Valid {
		info.CooldownTurns = action.CooldownTurns.Int
	}

	if action.UsesPerBattle.Valid {
		info.UsesPerBattle = action.UsesPerBattle.Int
	}

	if action.Description.Valid {
		info.Description = action.Description.String
	}

	if action.IsActive.Valid {
		info.IsActive = action.IsActive.Bool
	}

	if action.CreatedAt.Valid {
		info.CreatedAt = action.CreatedAt.Time.Unix()
	}

	if action.UpdatedAt.Valid {
		info.UpdatedAt = action.UpdatedAt.Time.Unix()
	}

	return info
}
