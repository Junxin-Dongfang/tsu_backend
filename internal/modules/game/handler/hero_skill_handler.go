package handler

import (
	"github.com/labstack/echo/v4"

	"tsu-self/internal/modules/game/service"
	"tsu-self/internal/pkg/response"
)

// HeroSkillHandler handles hero skill HTTP requests
type HeroSkillHandler struct {
	skillService *service.HeroSkillService
	respWriter   response.Writer
}

// NewHeroSkillHandler creates a new hero skill handler
func NewHeroSkillHandler(serviceContainer *service.ServiceContainer, respWriter response.Writer) *HeroSkillHandler {
	return &HeroSkillHandler{
		skillService: serviceContainer.GetHeroSkillService(),
		respWriter:   respWriter,
	}
}

// ==================== HTTP Request/Response Models ====================

// LearnSkillRequest HTTP learn skill request
type LearnSkillRequest struct {
	SkillID string `json:"skill_id" validate:"required" example:"skill-flame-slash-001"` // 技能ID（必填，从可学习技能列表获取）
}

// UpgradeSkillRequest HTTP upgrade skill request
type UpgradeSkillRequest struct {
	Levels int `json:"levels" validate:"required,min=1" example:"1"` // 升级等级数（必填，最小1，通常填1表示升1级）
}

// AvailableSkillResponse HTTP available skill response
type AvailableSkillResponse struct {
	SkillID           string `json:"skill_id" example:"skill-001"`                // 技能ID
	SkillName         string `json:"skill_name" example:"烈焰斩"`                    // 技能名称
	SkillCode         string `json:"skill_code" example:"FLAME_SLASH"`            // 技能代码
	MaxLevel          int    `json:"max_level" example:"10"`                      // 技能最大等级
	MaxLearnableLevel int    `json:"max_learnable_level" example:"5"`             // 当前可学习的最大等级（受英雄等级限制）
	CanLearn          bool   `json:"can_learn" example:"true"`                    // 是否可以学习（满足所有条件）
	Requirements      string `json:"requirements,omitempty" example:"需要等级5、力量15"` // 学习要求描述
}

// LearnedSkillResponse HTTP learned skill response
type LearnedSkillResponse struct {
	HeroSkillID    string `json:"hero_skill_id" example:"hero-skill-001"`                       // 英雄技能实例ID
	SkillID        string `json:"skill_id" example:"skill-001"`                                 // 技能配置ID
	SkillName      string `json:"skill_name" example:"烈焰斩"`                                     // 技能名称
	SkillCode      string `json:"skill_code" example:"FLAME_SLASH"`                             // 技能代码
	SkillLevel     int    `json:"skill_level" example:"3"`                                      // 当前等级
	MaxLevel       int    `json:"max_level" example:"10"`                                       // 最大等级
	LearnedMethod  string `json:"learned_method" example:"manual" enums:"initial,manual,quest"` // 学习方式：initial=初始技能，manual=手动学习，quest=任务奖励
	FirstLearnedAt string `json:"first_learned_at" example:"2025-10-17 10:30:00"`               // 首次学习时间
	CanUpgrade     bool   `json:"can_upgrade" example:"true"`                                   // 是否可以升级
	CanRollback    bool   `json:"can_rollback" example:"true"`                                  // 是否可以回退（1小时内）
}

// ==================== HTTP Handlers ====================

// GetAvailableSkills handles getting available skills
// @Summary 获取可学习技能
// @Description 获取英雄当前可学习的技能列表。包含职业初始技能和条件满足的可学习技能
// @Tags 英雄技能
// @Accept json
// @Produce json
// @Param hero_id path string true "英雄ID（UUID格式）"
// @Success 200 {object} response.Response{data=[]AvailableSkillResponse} "查询成功，返回可学习技能列表"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "英雄不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/heroes/{hero_id}/skills/available [get]
func (h *HeroSkillHandler) GetAvailableSkills(c echo.Context) error {
	heroID := c.Param("hero_id")
	if heroID == "" {
		return response.EchoBadRequest(c, h.respWriter, "英雄ID不能为空")
	}

	skills, err := h.skillService.GetAvailableSkills(c.Request().Context(), heroID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为 HTTP 响应
	respList := make([]*AvailableSkillResponse, len(skills))
	for i, skill := range skills {
		respList[i] = &AvailableSkillResponse{
			SkillID:           skill.SkillID,
			SkillName:         skill.SkillName,
			SkillCode:         skill.SkillCode,
			MaxLevel:          skill.MaxLevel,
			MaxLearnableLevel: skill.MaxLearnableLevel,
			CanLearn:          skill.CanLearn,
			Requirements:      skill.Requirements,
		}
	}

	return response.EchoOK(c, h.respWriter, respList)
}

// LearnSkill handles learning a skill
// @Summary 学习技能
// @Description 学习一个新技能。英雄必须满足技能的学习条件（等级、属性等）
// @Tags 英雄技能
// @Accept json
// @Produce json
// @Param hero_id path string true "英雄ID（UUID格式）"
// @Param request body LearnSkillRequest true "学习技能请求"
// @Success 200 {object} response.Response{data=object{message=string}} "学习成功，返回 {\"message\": \"技能学习成功\"}"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "英雄或技能不存在"
// @Failure 409 {object} response.Response "条件不满足或已学习该技能"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/heroes/{hero_id}/skills/learn [post]
func (h *HeroSkillHandler) LearnSkill(c echo.Context) error {
	heroID := c.Param("hero_id")
	if heroID == "" {
		return response.EchoBadRequest(c, h.respWriter, "英雄ID不能为空")
	}

	var req LearnSkillRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 调用 Service
	serviceReq := &service.LearnSkillRequest{
		HeroID:  heroID,
		SkillID: req.SkillID,
	}

	if err := h.skillService.LearnSkill(c.Request().Context(), serviceReq); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]string{
		"message": "技能学习成功",
	})
}

// UpgradeSkill handles upgrading a skill
// @Summary 升级技能
// @Description 升级已学习的技能。升级消耗经验，支持1小时内的回退
// @Tags 英雄技能
// @Accept json
// @Produce json
// @Param hero_id path string true "英雄ID（UUID格式）"
// @Param skill_id path string true "技能ID（UUID格式）"
// @Param request body UpgradeSkillRequest true "升级技能请求"
// @Success 200 {object} response.Response{data=object{message=string}} "升级成功，返回 {\"message\": \"技能升级成功\"}"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "英雄或技能未学习"
// @Failure 409 {object} response.Response "经验不足或达到最大等级"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/heroes/{hero_id}/skills/{skill_id}/upgrade [post]
func (h *HeroSkillHandler) UpgradeSkill(c echo.Context) error {
	heroID := c.Param("hero_id")
	skillID := c.Param("skill_id")

	if heroID == "" || skillID == "" {
		return response.EchoBadRequest(c, h.respWriter, "英雄ID和技能ID不能为空")
	}

	var req UpgradeSkillRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 调用 Service
	serviceReq := &service.UpgradeSkillRequest{
		HeroID:  heroID,
		SkillID: skillID,
		Levels:  req.Levels,
	}

	if err := h.skillService.UpgradeSkill(c.Request().Context(), serviceReq); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]string{
		"message": "技能升级成功",
	})
}

// RollbackSkill handles skill operation rollback
// @Summary 回退技能操作
// @Description 回退英雄最近一次技能操作（升级或学习）。支持堆栈式（LIFO）回退，仅限1小时内的操作
// @Tags 英雄技能
// @Accept json
// @Produce json
// @Param hero_id path string true "英雄ID（UUID格式）"
// @Param skill_id path string true "技能ID（UUID格式）"
// @Success 200 {object} response.Response{data=object{message=string}} "回退成功，返回 {\"message\": \"技能回退成功\"}"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "英雄或技能不存在，或没有可回退的操作"
// @Failure 409 {object} response.Response "回退时间已过期（超过1小时）"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/heroes/{hero_id}/skills/{skill_id}/rollback [post]
func (h *HeroSkillHandler) RollbackSkill(c echo.Context) error {
	heroID := c.Param("hero_id")
	skillID := c.Param("skill_id")

	if heroID == "" || skillID == "" {
		return response.EchoBadRequest(c, h.respWriter, "英雄ID和技能ID不能为空")
	}

	// 调用 Service（使用 heroID + skillID 组合）
	if err := h.skillService.RollbackSkillOperation(c.Request().Context(), heroID, skillID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]string{
		"message": "技能回退成功",
	})
}

// GetLearnedSkills handles getting learned skills
// @Summary 获取已学习技能列表
// @Description 获取英雄已学习的所有技能，包括技能详细信息、当前等级、是否可升级、是否可回退等状态
// @Tags 英雄技能
// @Accept json
// @Produce json
// @Param hero_id path string true "英雄ID（UUID格式）"
// @Success 200 {object} response.Response{data=[]LearnedSkillResponse} "查询成功，返回已学习技能列表"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "英雄不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/heroes/{hero_id}/skills/learned [get]
func (h *HeroSkillHandler) GetLearnedSkills(c echo.Context) error {
	heroID := c.Param("hero_id")
	if heroID == "" {
		return response.EchoBadRequest(c, h.respWriter, "英雄ID不能为空")
	}

	skills, err := h.skillService.GetLearnedSkills(c.Request().Context(), heroID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为 HTTP 响应
	respList := make([]*LearnedSkillResponse, len(skills))
	for i, skill := range skills {
		respList[i] = &LearnedSkillResponse{
			HeroSkillID:    skill.HeroSkillID,
			SkillID:        skill.SkillID,
			SkillName:      skill.SkillName,
			SkillCode:      skill.SkillCode,
			SkillLevel:     skill.SkillLevel,
			MaxLevel:       skill.MaxLevel,
			LearnedMethod:  skill.LearnedMethod,
			FirstLearnedAt: skill.FirstLearnedAt,
			CanUpgrade:     skill.CanUpgrade,
			CanRollback:    skill.CanRollback,
		}
	}

	return response.EchoOK(c, h.respWriter, respList)
}
