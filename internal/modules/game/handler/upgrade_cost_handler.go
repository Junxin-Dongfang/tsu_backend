package handler

import (
	"strconv"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/pkg/response"
	"tsu-self/internal/repository/interfaces"
)

// UpgradeCostHandler 升级消耗查询 Handler
type UpgradeCostHandler struct {
	heroLevelReqRepo     interfaces.HeroLevelRequirementRepository
	skillUpgradeCostRepo interfaces.SkillUpgradeCostRepository
	attrUpgradeCostRepo  interfaces.AttributeUpgradeCostRepository
	respWriter           response.Writer
}

// NewUpgradeCostHandler 创建升级消耗查询 Handler
func NewUpgradeCostHandler(
	heroLevelReqRepo interfaces.HeroLevelRequirementRepository,
	skillUpgradeCostRepo interfaces.SkillUpgradeCostRepository,
	attrUpgradeCostRepo interfaces.AttributeUpgradeCostRepository,
	respWriter response.Writer,
) *UpgradeCostHandler {
	return &UpgradeCostHandler{
		heroLevelReqRepo:     heroLevelReqRepo,
		skillUpgradeCostRepo: skillUpgradeCostRepo,
		attrUpgradeCostRepo:  attrUpgradeCostRepo,
		respWriter:           respWriter,
	}
}

// ==================== HTTP Response Models ====================

// HeroLevelRequirementResponse 英雄等级需求响应
type HeroLevelRequirementResponse struct {
	Level        int `json:"level"`         // 等级
	RequiredXP   int `json:"required_xp"`   // 从上一级升到该级所需增量经验
	CumulativeXP int `json:"cumulative_xp"` // 升到该等级需要的累计总经验
}

// SkillUpgradeCostResponse 技能升级消耗响应
type SkillUpgradeCostResponse struct {
	Level         int                      `json:"level"`          // 升到第N级
	CostXP        int                      `json:"cost_xp"`        // 消耗经验
	CostGold      int                      `json:"cost_gold"`      // 消耗金币
	CostMaterials []map[string]interface{} `json:"cost_materials"` // 消耗材料
}

// AttributeUpgradeCostResponse 属性升级消耗响应
type AttributeUpgradeCostResponse struct {
	PointNumber int `json:"point_number"` // 第N点属性加点
	CostXP      int `json:"cost_xp"`      // 该点所需经验值
}

// ==================== 英雄等级需求接口 ====================

// GetHeroLevelRequirements 获取所有英雄等级需求
// @Summary 获取英雄等级需求列表
// @Description 获取所有等级的升级需求，包含经验、技能点、属性点等信息
// @Tags 升级消耗查询
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]HeroLevelRequirementResponse} "查询成功"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/hero-level-requirements [get]
func (h *UpgradeCostHandler) GetHeroLevelRequirements(c echo.Context) error {
	requirements, err := h.heroLevelReqRepo.GetAll(c.Request().Context())
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为响应格式
	respList := make([]*HeroLevelRequirementResponse, len(requirements))
	for i, req := range requirements {
		respList[i] = &HeroLevelRequirementResponse{
			Level:        req.Level,
			RequiredXP:   req.RequiredXP,
			CumulativeXP: req.CumulativeXP,
		}
	}

	return response.EchoOK(c, h.respWriter, respList)
}

// GetHeroLevelRequirement 获取指定等级需求
// @Summary 获取指定等级需求
// @Description 获取升到指定等级所需的经验和奖励
// @Tags 升级消耗查询
// @Accept json
// @Produce json
// @Param level path int true "等级（2-40）"
// @Success 200 {object} response.Response{data=HeroLevelRequirementResponse} "查询成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "等级配置不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/hero-level-requirements/{level} [get]
func (h *UpgradeCostHandler) GetHeroLevelRequirement(c echo.Context) error {
	levelStr := c.Param("level")
	level, err := strconv.Atoi(levelStr)
	if err != nil || level < 1 || level > 40 {
		return response.EchoBadRequest(c, h.respWriter, "等级必须在 1-40 之间")
	}

	req, err := h.heroLevelReqRepo.GetByLevel(c.Request().Context(), level)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	resp := &HeroLevelRequirementResponse{
		Level:        req.Level,
		RequiredXP:   req.RequiredXP,
		CumulativeXP: req.CumulativeXP,
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// ==================== 技能升级消耗接口 ====================

// GetSkillUpgradeCosts 获取所有技能升级消耗
// @Summary 获取技能升级消耗列表
// @Description 获取所有等级的技能升级消耗配置（经验、金币、材料）
// @Tags 升级消耗查询
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]SkillUpgradeCostResponse} "查询成功"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/skill-upgrade-costs [get]
func (h *UpgradeCostHandler) GetSkillUpgradeCosts(c echo.Context) error {
	costs, err := h.skillUpgradeCostRepo.List(c.Request().Context())
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为响应格式
	respList := make([]*SkillUpgradeCostResponse, len(costs))
	for i, cost := range costs {
		resp := &SkillUpgradeCostResponse{
			Level:    cost.LevelNumber,
			CostXP:   int(cost.CostXP.Int),
			CostGold: int(cost.CostGold.Int),
		}

		// 解析 JSONB 材料字段
		if !cost.CostMaterials.IsZero() {
			var materials []map[string]interface{}
			if err := cost.CostMaterials.Unmarshal(&materials); err == nil {
				resp.CostMaterials = materials
			}
		}
		if resp.CostMaterials == nil {
			resp.CostMaterials = []map[string]interface{}{}
		}

		respList[i] = resp
	}

	return response.EchoOK(c, h.respWriter, respList)
}

// GetSkillUpgradeCost 获取指定等级升级消耗
// @Summary 获取指定等级技能升级消耗
// @Description 获取技能升到指定等级所需的消耗
// @Tags 升级消耗查询
// @Accept json
// @Produce json
// @Param level path int true "目标等级（2-10）"
// @Success 200 {object} response.Response{data=SkillUpgradeCostResponse} "查询成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "等级配置不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/skill-upgrade-costs/{level} [get]
func (h *UpgradeCostHandler) GetSkillUpgradeCost(c echo.Context) error {
	levelStr := c.Param("level")
	level, err := strconv.Atoi(levelStr)
	if err != nil || level < 2 || level > 10 {
		return response.EchoBadRequest(c, h.respWriter, "等级必须在 2-10 之间")
	}

	cost, err := h.skillUpgradeCostRepo.GetByLevel(c.Request().Context(), level)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	resp := &SkillUpgradeCostResponse{
		Level:    cost.LevelNumber,
		CostXP:   int(cost.CostXP.Int),
		CostGold: int(cost.CostGold.Int),
	}

	// 解析 JSONB 材料字段
	if !cost.CostMaterials.IsZero() {
		var materials []map[string]interface{}
		if err := cost.CostMaterials.Unmarshal(&materials); err == nil {
			resp.CostMaterials = materials
		}
	}
	if resp.CostMaterials == nil {
		resp.CostMaterials = []map[string]interface{}{}
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// ==================== 属性升级消耗接口 ====================

// GetAttributeUpgradeCosts 获取所有属性升级消耗
// @Summary 获取属性升级消耗列表
// @Description 获取所有属性值对应的升级消耗配置
// @Tags 升级消耗查询
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]AttributeUpgradeCostResponse} "查询成功"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/attribute-upgrade-costs [get]
func (h *UpgradeCostHandler) GetAttributeUpgradeCosts(c echo.Context) error {
	costs, err := h.attrUpgradeCostRepo.GetAll(c.Request().Context())
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为响应格式
	respList := make([]*AttributeUpgradeCostResponse, len(costs))
	for i, cost := range costs {
		respList[i] = &AttributeUpgradeCostResponse{
			PointNumber: cost.PointNumber,
			CostXP:      cost.CostXP,
		}
	}

	return response.EchoOK(c, h.respWriter, respList)
}

// GetAttributeUpgradeCost 获取指定点数升级消耗
// @Summary 获取指定点数属性升级消耗
// @Description 获取第N点属性加点所需的经验
// @Tags 升级消耗查询
// @Accept json
// @Produce json
// @Param point_number path int true "第几点属性"
// @Success 200 {object} response.Response{data=AttributeUpgradeCostResponse} "查询成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "点数配置不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/attribute-upgrade-costs/{point_number} [get]
func (h *UpgradeCostHandler) GetAttributeUpgradeCost(c echo.Context) error {
	pointStr := c.Param("point_number")
	pointNumber, err := strconv.Atoi(pointStr)
	if err != nil || pointNumber < 1 {
		return response.EchoBadRequest(c, h.respWriter, "点数必须大于 0")
	}

	cost, err := h.attrUpgradeCostRepo.GetByPointNumber(c.Request().Context(), pointNumber)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	resp := &AttributeUpgradeCostResponse{
		PointNumber: cost.PointNumber,
		CostXP:      cost.CostXP,
	}

	return response.EchoOK(c, h.respWriter, resp)
}
