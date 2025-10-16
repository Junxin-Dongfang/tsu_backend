package handler

import (
	"github.com/labstack/echo/v4"

	"tsu-self/internal/modules/game/service"
	"tsu-self/internal/pkg/response"
)

// HeroHandler handles hero HTTP requests
type HeroHandler struct {
	heroService *service.HeroService
	respWriter  response.Writer
}

// NewHeroHandler creates a new hero handler
func NewHeroHandler(serviceContainer *service.ServiceContainer, respWriter response.Writer) *HeroHandler {
	return &HeroHandler{
		heroService: serviceContainer.GetHeroService(),
		respWriter:  respWriter,
	}
}

// ==================== HTTP Request/Response Models ====================

// HeroFullResponse HTTP hero full response (with class, attributes, skills)
type HeroFullResponse struct {
	// 英雄基本信息
	ID                  string  `json:"id"`
	UserID              string  `json:"user_id"`
	HeroName            string  `json:"hero_name"`
	Description         *string `json:"description,omitempty"`
	CurrentLevel        int     `json:"current_level"`
	ExperienceTotal     int64   `json:"experience_total"`
	ExperienceAvailable int64   `json:"experience_available"`
	ExperienceSpent     int64   `json:"experience_spent"`
	Status              string  `json:"status"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`

	// 职业信息
	Class *ClassInfo `json:"class"`

	// 属性列表
	Attributes []*AttributeInfo `json:"attributes"`

	// 技能列表
	Skills []*SkillInfo `json:"skills"`
}

// ClassInfo 职业信息
type ClassInfo struct {
	ID          string `json:"id"`
	ClassCode   string `json:"class_code"`
	ClassName   string `json:"class_name"`
	Tier        string `json:"tier"`
	Description string `json:"description,omitempty"`
}

// AttributeInfo 属性信息
type AttributeInfo struct {
	AttributeCode string `json:"attribute_code"`
	AttributeName string `json:"attribute_name"`
	BaseValue     int    `json:"base_value"`
	ClassBonus    int    `json:"class_bonus"`
	FinalValue    int    `json:"final_value"`
}

// SkillInfo 技能信息
type SkillInfo struct {
	HeroSkillID string `json:"hero_skill_id"`
	SkillID     string `json:"skill_id"`
	SkillName   string `json:"skill_name"`
	SkillCode   string `json:"skill_code"`
	SkillLevel  int    `json:"skill_level"`
	MaxLevel    int    `json:"max_level"`
}

// CreateHeroRequest HTTP create hero request
type CreateHeroRequest struct {
	ClassID     string `json:"class_id" validate:"required"`
	HeroName    string `json:"hero_name" validate:"required,min=2,max=20"`
	Description string `json:"description,omitempty"`
}

// AddExperienceRequest HTTP add experience request
type AddExperienceRequest struct {
	Amount int64 `json:"amount" validate:"required,min=1"`
}

// HeroResponse HTTP hero response
type HeroResponse struct {
	ID                  string                 `json:"id"`
	UserID              string                 `json:"user_id"`
	ClassID             string                 `json:"class_id"`
	HeroName            string                 `json:"hero_name"`
	Description         *string                `json:"description,omitempty"`
	CurrentLevel        int                    `json:"current_level"`
	ExperienceTotal     int64                  `json:"experience_total"`
	ExperienceAvailable int64                  `json:"experience_available"`
	ExperienceSpent     int64                  `json:"experience_spent"`
	AllocatedAttributes map[string]interface{} `json:"allocated_attributes,omitempty"`
	Status              string                 `json:"status"`
	CreatedAt           string                 `json:"created_at"`
	UpdatedAt           string                 `json:"updated_at"`
}

// ==================== HTTP Handlers ====================

// CreateHero handles hero creation
// @Summary 创建英雄
// @Description 创建新的英雄角色（仅限基础职业）
// @Tags 英雄
// @Accept json
// @Produce json
// @Param request body CreateHeroRequest true "创建英雄请求"
// @Success 200 {object} response.Response{data=HeroResponse} "创建成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/heroes [post]
func (h *HeroHandler) CreateHero(c echo.Context) error {
	// 1. 绑定和验证 HTTP 请求
	var req CreateHeroRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 2. 获取用户ID（从认证中间件）
	userID := c.Get("user_id")
	if userID == nil {
		return response.EchoUnauthorized(c, h.respWriter, "未登录")
	}

	// 3. 调用 Service
	createReq := &service.CreateHeroRequest{
		UserID:      userID.(string),
		ClassID:     req.ClassID,
		HeroName:    req.HeroName,
		Description: req.Description,
	}

	hero, err := h.heroService.CreateHero(c.Request().Context(), createReq)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 4. 转换为 HTTP 响应
	resp := &HeroResponse{
		ID:                  hero.ID,
		UserID:              hero.UserID,
		ClassID:             hero.ClassID,
		HeroName:            hero.HeroName,
		CurrentLevel:        int(hero.CurrentLevel),
		ExperienceTotal:     hero.ExperienceTotal,
		ExperienceAvailable: hero.ExperienceAvailable,
		ExperienceSpent:     hero.ExperienceSpent,
		Status:              hero.Status,
		CreatedAt:           hero.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:           hero.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if !hero.Description.IsZero() {
		desc := hero.Description.String
		resp.Description = &desc
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// GetHero handles getting hero details
// @Summary 获取英雄详情
// @Description 获取指定英雄的详细信息
// @Tags 英雄
// @Accept json
// @Produce json
// @Param hero_id path string true "英雄ID"
// @Success 200 {object} response.Response{data=HeroResponse} "获取成功"
// @Failure 404 {object} response.Response "英雄不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/heroes/{hero_id} [get]
func (h *HeroHandler) GetHero(c echo.Context) error {
	heroID := c.Param("hero_id")
	if heroID == "" {
		return response.EchoBadRequest(c, h.respWriter, "英雄ID不能为空")
	}

	hero, err := h.heroService.GetHeroByID(c.Request().Context(), heroID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为 HTTP 响应
	resp := &HeroResponse{
		ID:                  hero.ID,
		UserID:              hero.UserID,
		ClassID:             hero.ClassID,
		HeroName:            hero.HeroName,
		CurrentLevel:        int(hero.CurrentLevel),
		ExperienceTotal:     hero.ExperienceTotal,
		ExperienceAvailable: hero.ExperienceAvailable,
		ExperienceSpent:     hero.ExperienceSpent,
		Status:              hero.Status,
		CreatedAt:           hero.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:           hero.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if !hero.Description.IsZero() {
		desc := hero.Description.String
		resp.Description = &desc
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// GetUserHeroes handles getting user's hero list
// @Summary 获取用户的英雄列表
// @Description 获取当前用户的所有英雄
// @Tags 英雄
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]HeroResponse} "获取成功"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/heroes [get]
func (h *HeroHandler) GetUserHeroes(c echo.Context) error {
	// 获取用户ID
	userID := c.Get("user_id")
	if userID == nil {
		return response.EchoUnauthorized(c, h.respWriter, "未登录")
	}

	heroes, err := h.heroService.GetHeroesByUserID(c.Request().Context(), userID.(string))
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为 HTTP 响应
	respList := make([]*HeroResponse, len(heroes))
	for i, hero := range heroes {
		resp := &HeroResponse{
			ID:                  hero.ID,
			UserID:              hero.UserID,
			ClassID:             hero.ClassID,
			HeroName:            hero.HeroName,
			CurrentLevel:        int(hero.CurrentLevel),
			ExperienceTotal:     hero.ExperienceTotal,
			ExperienceAvailable: hero.ExperienceAvailable,
			ExperienceSpent:     hero.ExperienceSpent,
			Status:              hero.Status,
			CreatedAt:           hero.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:           hero.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		if !hero.Description.IsZero() {
			desc := hero.Description.String
			resp.Description = &desc
		}

		respList[i] = resp
	}

	return response.EchoOK(c, h.respWriter, respList)
}

// AdvanceClass handles class advancement
// @Summary 职业进阶
// @Description 将英雄的职业进阶到更高阶职业
// @Tags 英雄
// @Accept json
// @Produce json
// @Param hero_id path string true "英雄ID"
// @Param request body map[string]string true "进阶请求 {target_class_id}"
// @Success 200 {object} response.Response "进阶成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/heroes/{hero_id}/advance [post]
func (h *HeroHandler) AdvanceClass(c echo.Context) error {
	heroID := c.Param("hero_id")
	if heroID == "" {
		return response.EchoBadRequest(c, h.respWriter, "英雄ID不能为空")
	}

	var req struct {
		TargetClassID string `json:"target_class_id" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	if err := h.heroService.AdvanceClass(c.Request().Context(), heroID, req.TargetClassID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]string{
		"message": "职业进阶成功",
	})
}

// TransferClass handles class transfer
// @Summary 职业转职
// @Description 将英雄转职到另一个基础职业
// @Tags 英雄
// @Accept json
// @Produce json
// @Param hero_id path string true "英雄ID"
// @Param request body map[string]string true "转职请求 {target_class_id}"
// @Success 200 {object} response.Response "转职成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/heroes/{hero_id}/transfer [post]
func (h *HeroHandler) TransferClass(c echo.Context) error {
	heroID := c.Param("hero_id")
	if heroID == "" {
		return response.EchoBadRequest(c, h.respWriter, "英雄ID不能为空")
	}

	var req struct {
		TargetClassID string `json:"target_class_id" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	if err := h.heroService.TransferClass(c.Request().Context(), heroID, req.TargetClassID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]string{
		"message": "职业转职成功",
	})
}

// AddExperience handles adding experience to a hero
// @Summary 增加英雄经验
// @Description 为英雄增加经验值。用于测试和开发目的，生产环境应通过战斗/任务系统增加经验
// @Tags 英雄
// @Accept json
// @Produce json
// @Param hero_id path string true "英雄ID（UUID格式）"
// @Param request body AddExperienceRequest true "增加经验请求"
// @Success 200 {object} response.Response{data=object{message=string,hero=HeroResponse}} "增加成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "英雄不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/heroes/{hero_id}/experience [post]
func (h *HeroHandler) AddExperience(c echo.Context) error {
	heroID := c.Param("hero_id")
	if heroID == "" {
		return response.EchoBadRequest(c, h.respWriter, "英雄ID不能为空")
	}

	var req AddExperienceRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 调用 Service
	hero, err := h.heroService.AddExperience(c.Request().Context(), heroID, req.Amount)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为 HTTP 响应
	heroResp := &HeroResponse{
		ID:                  hero.ID,
		UserID:              hero.UserID,
		ClassID:             hero.ClassID,
		HeroName:            hero.HeroName,
		CurrentLevel:        int(hero.CurrentLevel),
		ExperienceTotal:     hero.ExperienceTotal,
		ExperienceAvailable: hero.ExperienceAvailable,
		ExperienceSpent:     hero.ExperienceSpent,
		Status:              hero.Status,
		CreatedAt:           hero.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:           hero.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if !hero.Description.IsZero() {
		desc := hero.Description.String
		heroResp.Description = &desc
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": "经验增加成功",
		"hero":    heroResp,
	})
}

// GetHeroFull handles getting hero full information
// @Summary 获取英雄完整信息
// @Description 获取英雄的完整信息，包括基本信息、职业详情、属性列表、技能列表
// @Tags 英雄
// @Accept json
// @Produce json
// @Param hero_id path string true "英雄ID"
// @Success 200 {object} response.Response{data=HeroFullResponse} "获取成功"
// @Failure 404 {object} response.Response "英雄不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/heroes/{hero_id}/full [get]
func (h *HeroHandler) GetHeroFull(c echo.Context) error {
	heroID := c.Param("hero_id")
	if heroID == "" {
		return response.EchoBadRequest(c, h.respWriter, "英雄ID不能为空")
	}

	// 调用 Service 获取完整信息
	fullInfo, err := h.heroService.GetHeroFullInfo(c.Request().Context(), heroID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为 HTTP 响应
	resp := &HeroFullResponse{
		ID:                  fullInfo.Hero.ID,
		UserID:              fullInfo.Hero.UserID,
		HeroName:            fullInfo.Hero.HeroName,
		CurrentLevel:        int(fullInfo.Hero.CurrentLevel),
		ExperienceTotal:     fullInfo.Hero.ExperienceTotal,
		ExperienceAvailable: fullInfo.Hero.ExperienceAvailable,
		ExperienceSpent:     fullInfo.Hero.ExperienceSpent,
		Status:              fullInfo.Hero.Status,
		CreatedAt:           fullInfo.Hero.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:           fullInfo.Hero.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if !fullInfo.Hero.Description.IsZero() {
		desc := fullInfo.Hero.Description.String
		resp.Description = &desc
	}

	// 职业信息
	resp.Class = &ClassInfo{
		ID:          fullInfo.ClassInfo.ID,
		ClassCode:   fullInfo.ClassInfo.ClassCode,
		ClassName:   fullInfo.ClassInfo.ClassName,
		Tier:        fullInfo.ClassInfo.Tier,
		Description: fullInfo.ClassInfo.Description,
	}

	// 属性列表
	resp.Attributes = make([]*AttributeInfo, len(fullInfo.Attributes))
	for i, attr := range fullInfo.Attributes {
		resp.Attributes[i] = &AttributeInfo{
			AttributeCode: attr.AttributeCode,
			AttributeName: attr.AttributeName,
			BaseValue:     attr.BaseValue,
			ClassBonus:    attr.ClassBonus,
			FinalValue:    attr.FinalValue,
		}
	}

	// 技能列表
	resp.Skills = make([]*SkillInfo, len(fullInfo.Skills))
	for i, skill := range fullInfo.Skills {
		resp.Skills[i] = &SkillInfo{
			HeroSkillID: skill.HeroSkillID,
			SkillID:     skill.SkillID,
			SkillName:   skill.SkillName,
			SkillCode:   skill.SkillCode,
			SkillLevel:  skill.SkillLevel,
			MaxLevel:    skill.MaxLevel,
		}
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// CheckAdvancement handles checking advancement requirements
// @Summary 检查职业进阶条件
// @Description 检查英雄是否满足指定职业的进阶条件，返回详细的检查结果
// @Tags 英雄
// @Accept json
// @Produce json
// @Param hero_id path string true "英雄ID"
// @Param target_class_id query string true "目标职业ID"
// @Success 200 {object} response.Response{data=service.AdvancementCheckResult} "检查成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "英雄不存在或进阶路径不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/heroes/{hero_id}/advancement-check [get]
func (h *HeroHandler) CheckAdvancement(c echo.Context) error {
	heroID := c.Param("hero_id")
	if heroID == "" {
		return response.EchoBadRequest(c, h.respWriter, "英雄ID不能为空")
	}

	targetClassID := c.QueryParam("target_class_id")
	if targetClassID == "" {
		return response.EchoBadRequest(c, h.respWriter, "目标职业ID不能为空")
	}

	// 调用 Service 检查进阶条件
	result, err := h.heroService.CheckAdvancementRequirements(c.Request().Context(), heroID, targetClassID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, result)
}
