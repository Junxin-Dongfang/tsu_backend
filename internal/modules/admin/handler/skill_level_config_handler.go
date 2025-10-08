package handler

import (
	"database/sql"
	"encoding/json"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/modules/admin/service"
	"tsu-self/internal/pkg/response"
)

// SkillLevelConfigHandler 技能等级配置 HTTP 处理器
type SkillLevelConfigHandler struct {
	service    *service.SkillLevelConfigService
	respWriter response.Writer
}

// NewSkillLevelConfigHandler 创建技能等级配置处理器
func NewSkillLevelConfigHandler(db *sql.DB, respWriter response.Writer) *SkillLevelConfigHandler {
	return &SkillLevelConfigHandler{
		service:    service.NewSkillLevelConfigService(db),
		respWriter: respWriter,
	}
}

// ==================== HTTP Models ====================

// CreateSkillLevelConfigRequest 创建技能等级配置请求
type CreateSkillLevelConfigRequest struct {
	LevelNumber       int    `json:"level_number" validate:"required,min=1"`
	DamageMultiplier  string `json:"damage_multiplier"`  // Decimal as string
	HealingMultiplier string `json:"healing_multiplier"` // Decimal as string
	DurationModifier  int    `json:"duration_modifier"`
	RangeModifier     int    `json:"range_modifier"`
	CooldownModifier  int    `json:"cooldown_modifier"`
	ManaCostModifier  int    `json:"mana_cost_modifier"`
	EffectModifiers   string `json:"effect_modifiers"` // JSON string
	UpgradeCostXP     int    `json:"upgrade_cost_xp"`
	UpgradeCostGold   int    `json:"upgrade_cost_gold"`
	UpgradeMaterials  string `json:"upgrade_materials"` // JSON string
}

// UpdateSkillLevelConfigRequest 更新技能等级配置请求
type UpdateSkillLevelConfigRequest struct {
	DamageMultiplier  string `json:"damage_multiplier"`
	HealingMultiplier string `json:"healing_multiplier"`
	DurationModifier  int    `json:"duration_modifier"`
	RangeModifier     int    `json:"range_modifier"`
	CooldownModifier  int    `json:"cooldown_modifier"`
	ManaCostModifier  int    `json:"mana_cost_modifier"`
	UpgradeCostXP     int    `json:"upgrade_cost_xp"`
	UpgradeCostGold   int    `json:"upgrade_cost_gold"`
}

// SkillLevelConfigInfo 技能等级配置信息响应
type SkillLevelConfigInfo struct {
	ID                string `json:"id"`
	SkillID           string `json:"skill_id"`
	LevelNumber       int    `json:"level_number"`
	DamageMultiplier  string `json:"damage_multiplier"`
	HealingMultiplier string `json:"healing_multiplier"`
	DurationModifier  int    `json:"duration_modifier"`
	RangeModifier     int    `json:"range_modifier"`
	CooldownModifier  int    `json:"cooldown_modifier"`
	ManaCostModifier  int    `json:"mana_cost_modifier"`
	EffectModifiers   string `json:"effect_modifiers"`
	UpgradeCostXP     int    `json:"upgrade_cost_xp"`
	UpgradeCostGold   int    `json:"upgrade_cost_gold"`
	UpgradeMaterials  string `json:"upgrade_materials"`
	CreatedAt         int64  `json:"created_at"`
	UpdatedAt         int64  `json:"updated_at"`
}

// ==================== HTTP Handlers ====================

// GetSkillLevelConfigs 获取技能的所有等级配置
// @Summary 获取技能的所有等级配置
// @Description 获取指定技能的所有等级配置列表,包含伤害倍率、治疗倍率、升级成本等信息
// @Tags 技能系统
// @Accept json
// @Produce json
// @Param id path string true "技能ID (UUID格式)" example("01d132ed-6378-4e0b-bc16-a5b224e5b04a")
// @Success 200 {object} response.Response{data=[]SkillLevelConfigInfo} "成功返回等级配置列表"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "技能不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/skills/{id}/level-configs [get]
// @Security BearerAuth
func (h *SkillLevelConfigHandler) GetSkillLevelConfigs(c echo.Context) error {
	ctx := c.Request().Context()
	skillID := c.Param("id")

	configs, err := h.service.GetSkillLevelConfigsBySkillID(ctx, skillID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	result := make([]SkillLevelConfigInfo, len(configs))
	for i, config := range configs {
		result[i] = h.convertToInfo(config)
	}

	return response.EchoOK(c, h.respWriter, result)
}

// GetSkillLevelConfig 获取单个等级配置详情
// @Summary 获取单个等级配置详情
// @Description 根据配置ID获取指定技能等级配置的详细信息
// @Tags 技能系统
// @Accept json
// @Produce json
// @Param id path string true "技能ID (UUID格式)" example("01d132ed-6378-4e0b-bc16-a5b224e5b04a")
// @Param config_id path string true "配置ID (UUID格式)" example("f3a8b1c2-4d5e-6f7a-8b9c-0d1e2f3a4b5c")
// @Success 200 {object} response.Response{data=SkillLevelConfigInfo} "成功返回等级配置详情"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "配置不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/skills/{id}/level-configs/{config_id} [get]
// @Security BearerAuth
func (h *SkillLevelConfigHandler) GetSkillLevelConfig(c echo.Context) error {
	ctx := c.Request().Context()
	configID := c.Param("config_id")

	config, err := h.service.GetSkillLevelConfigByID(ctx, configID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToInfo(config))
}

// CreateSkillLevelConfig 创建技能等级配置
// @Summary 创建技能等级配置
// @Description 为指定技能创建新的等级配置,包含伤害/治疗倍率、各类修正值、升级成本等
// @Tags 技能系统
// @Accept json
// @Produce json
// @Param id path string true "技能ID (UUID格式)" example("01d132ed-6378-4e0b-bc16-a5b224e5b04a")
// @Param request body CreateSkillLevelConfigRequest true "创建请求,level_number为必填字段,damage_multiplier等为字符串格式的十进制数"
// @Success 200 {object} response.Response{data=SkillLevelConfigInfo} "成功创建等级配置,返回配置详情"
// @Failure 400 {object} response.Response "请求参数错误或验证失败"
// @Failure 404 {object} response.Response "技能不存在"
// @Failure 409 {object} response.Response "该等级配置已存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/skills/{id}/level-configs [post]
// @Security BearerAuth
func (h *SkillLevelConfigHandler) CreateSkillLevelConfig(c echo.Context) error {
	ctx := c.Request().Context()
	skillID := c.Param("id")

	var req CreateSkillLevelConfigRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 构造配置实体
	config := &game_config.SkillLevelConfig{
		SkillID:     skillID,
		LevelNumber: req.LevelNumber,
	}

	if req.DamageMultiplier != "" {
		if err := config.DamageMultiplier.UnmarshalText([]byte(req.DamageMultiplier)); err != nil {
			return response.EchoBadRequest(c, h.respWriter, "damage_multiplier 格式错误")
		}
	}

	if req.HealingMultiplier != "" {
		if err := config.HealingMultiplier.UnmarshalText([]byte(req.HealingMultiplier)); err != nil {
			return response.EchoBadRequest(c, h.respWriter, "healing_multiplier 格式错误")
		}
	}

	if req.DurationModifier != 0 {
		config.DurationModifier.SetValid(req.DurationModifier)
	}

	if req.RangeModifier != 0 {
		config.RangeModifier.SetValid(req.RangeModifier)
	}

	if req.CooldownModifier != 0 {
		config.CooldownModifier.SetValid(req.CooldownModifier)
	}

	if req.ManaCostModifier != 0 {
		config.ManaCostModifier.SetValid(req.ManaCostModifier)
	}

	if req.EffectModifiers != "" {
		var jsonData interface{}
		if err := json.Unmarshal([]byte(req.EffectModifiers), &jsonData); err != nil {
			return response.EchoBadRequest(c, h.respWriter, "effect_modifiers 必须是有效的 JSON")
		}
		config.EffectModifiers.UnmarshalJSON([]byte(req.EffectModifiers))
	}

	if req.UpgradeCostXP > 0 {
		config.UpgradeCostXP.SetValid(req.UpgradeCostXP)
	}

	if req.UpgradeCostGold > 0 {
		config.UpgradeCostGold.SetValid(req.UpgradeCostGold)
	}

	if req.UpgradeMaterials != "" {
		var jsonData interface{}
		if err := json.Unmarshal([]byte(req.UpgradeMaterials), &jsonData); err != nil {
			return response.EchoBadRequest(c, h.respWriter, "upgrade_materials 必须是有效的 JSON")
		}
		config.UpgradeMaterials.UnmarshalJSON([]byte(req.UpgradeMaterials))
	}

	// 创建配置
	if err := h.service.CreateSkillLevelConfig(ctx, config); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToInfo(config))
}

// UpdateSkillLevelConfig 更新技能等级配置
// @Summary 更新技能等级配置
// @Description 更新指定技能等级配置的信息,仅更新提供的字段
// @Tags 技能系统
// @Accept json
// @Produce json
// @Param id path string true "技能ID (UUID格式)" example("01d132ed-6378-4e0b-bc16-a5b224e5b04a")
// @Param config_id path string true "配置ID (UUID格式)" example("f3a8b1c2-4d5e-6f7a-8b9c-0d1e2f3a4b5c")
// @Param request body UpdateSkillLevelConfigRequest true "更新请求,仅提供需要更新的字段"
// @Success 200 {object} response.Response{data=map[string]string} "成功更新等级配置,返回成功消息"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "配置不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/skills/{id}/level-configs/{config_id} [put]
// @Security BearerAuth
func (h *SkillLevelConfigHandler) UpdateSkillLevelConfig(c echo.Context) error {
	ctx := c.Request().Context()
	configID := c.Param("config_id")

	var req UpdateSkillLevelConfigRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	// 构造更新字段
	updates := make(map[string]interface{})

	if req.DamageMultiplier != "" {
		updates["damage_multiplier"] = req.DamageMultiplier
	}
	if req.HealingMultiplier != "" {
		updates["healing_multiplier"] = req.HealingMultiplier
	}
	if req.DurationModifier != 0 {
		updates["duration_modifier"] = req.DurationModifier
	}
	if req.RangeModifier != 0 {
		updates["range_modifier"] = req.RangeModifier
	}
	if req.CooldownModifier != 0 {
		updates["cooldown_modifier"] = req.CooldownModifier
	}
	if req.ManaCostModifier != 0 {
		updates["mana_cost_modifier"] = req.ManaCostModifier
	}
	if req.UpgradeCostXP > 0 {
		updates["upgrade_cost_xp"] = req.UpgradeCostXP
	}
	if req.UpgradeCostGold > 0 {
		updates["upgrade_cost_gold"] = req.UpgradeCostGold
	}

	// 更新配置
	if err := h.service.UpdateSkillLevelConfig(ctx, configID, updates); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": "等级配置更新成功",
	})
}

// DeleteSkillLevelConfig 删除技能等级配置
// @Summary 删除技能等级配置 (软删除)
// @Description 软删除指定技能等级配置,配置数据不会被物理删除,仅标记为已删除
// @Tags 技能系统
// @Accept json
// @Produce json
// @Param id path string true "技能ID (UUID格式)" example("01d132ed-6378-4e0b-bc16-a5b224e5b04a")
// @Param config_id path string true "配置ID (UUID格式)" example("f3a8b1c2-4d5e-6f7a-8b9c-0d1e2f3a4b5c")
// @Success 200 {object} response.Response{data=map[string]string} "成功删除等级配置,返回成功消息"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "配置不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/skills/{id}/level-configs/{config_id} [delete]
// @Security BearerAuth
func (h *SkillLevelConfigHandler) DeleteSkillLevelConfig(c echo.Context) error {
	ctx := c.Request().Context()
	configID := c.Param("config_id")

	if err := h.service.DeleteSkillLevelConfig(ctx, configID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": "等级配置删除成功",
	})
}

// ==================== Helper Functions ====================

func (h *SkillLevelConfigHandler) convertToInfo(config *game_config.SkillLevelConfig) SkillLevelConfigInfo {
	info := SkillLevelConfigInfo{
		ID:          config.ID,
		SkillID:     config.SkillID,
		LevelNumber: config.LevelNumber,
	}

	// Decimal 类型处理 (NullDecimal 通过 IsZero 判断)
	if !config.DamageMultiplier.IsZero() {
		damageBytes, _ := config.DamageMultiplier.MarshalText()
		info.DamageMultiplier = string(damageBytes)
	}

	if !config.HealingMultiplier.IsZero() {
		healingBytes, _ := config.HealingMultiplier.MarshalText()
		info.HealingMultiplier = string(healingBytes)
	}

	if config.DurationModifier.Valid {
		info.DurationModifier = config.DurationModifier.Int
	}

	if config.RangeModifier.Valid {
		info.RangeModifier = config.RangeModifier.Int
	}

	if config.CooldownModifier.Valid {
		info.CooldownModifier = config.CooldownModifier.Int
	}

	if config.ManaCostModifier.Valid {
		info.ManaCostModifier = config.ManaCostModifier.Int
	}

	// JSONB 处理
	if config.EffectModifiers.Valid {
		jsonBytes, _ := config.EffectModifiers.MarshalJSON()
		info.EffectModifiers = string(jsonBytes)
	}

	if config.UpgradeCostXP.Valid {
		info.UpgradeCostXP = config.UpgradeCostXP.Int
	}

	if config.UpgradeCostGold.Valid {
		info.UpgradeCostGold = config.UpgradeCostGold.Int
	}

	if config.UpgradeMaterials.Valid {
		jsonBytes, _ := config.UpgradeMaterials.MarshalJSON()
		info.UpgradeMaterials = string(jsonBytes)
	}

	if config.CreatedAt.Valid {
		info.CreatedAt = config.CreatedAt.Time.Unix()
	}

	if config.UpdatedAt.Valid {
		info.UpdatedAt = config.UpdatedAt.Time.Unix()
	}

	return info
}
