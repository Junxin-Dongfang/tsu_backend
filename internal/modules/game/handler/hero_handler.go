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
	ID                  string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`       // 英雄ID（UUID格式）
	UserID              string  `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`  // 用户ID（UUID格式）
	HeroName            string  `json:"hero_name" example:"艾泽拉斯勇士"`                              // 英雄名称（2-20字符）
	Description         *string `json:"description,omitempty" example:"一位来自北方的战士"`               // 英雄描述（可选）
	CurrentLevel        int     `json:"current_level" example:"5"`                               // 当前等级（1-40）
	ExperienceTotal     int64   `json:"experience_total" example:"750"`                          // 累计总经验
	ExperienceAvailable int64   `json:"experience_available" example:"250"`                      // 可用经验（用于加点、学技能）
	ExperienceSpent     int64   `json:"experience_spent" example:"500"`                          // 已消耗经验
	Status              string  `json:"status" example:"active" enums:"active,inactive,deleted"` // 英雄状态
	CreatedAt           string  `json:"created_at" example:"2025-10-17 10:30:00"`                // 创建时间
	UpdatedAt           string  `json:"updated_at" example:"2025-10-17 12:30:00"`                // 更新时间

	// 职业信息
	Class *ClassInfo `json:"class"` // 职业详情（包含名称、等阶等）

	// 属性列表
	Attributes []*AttributeInfo `json:"attributes"` // 属性列表

	// 技能列表
	Skills []*SkillInfo `json:"skills"` // 已学习技能列表
}

// ClassInfo 职业信息
type ClassInfo struct {
	ID          string `json:"id" example:"class-warrior-001"`                     // 职业ID
	ClassCode   string `json:"class_code" example:"WARRIOR"`                       // 职业代码
	ClassName   string `json:"class_name" example:"战士"`                            // 职业名称
	Tier        string `json:"tier" example:"basic" enums:"basic,advanced,master"` // 职业等阶：basic=基础职业，advanced=进阶职业，master=大师职业
	Description string `json:"description,omitempty" example:"擅长近战和物理攻击的职业"`       // 职业描述（可选）
}

// AttributeInfo 属性信息（来自 hero_computed_attributes 视图）
type AttributeInfo struct {
	AttributeCode string `json:"attribute_code" example:"STR" enums:"STR,DEX,CON,INT,WIS,CHA"` // 属性代码：STR=力量，DEX=敏捷，CON=体质，INT=智力，WIS=感知，CHA=魅力
	AttributeName string `json:"attribute_name" example:"力量"`                                  // 属性名称（中文）
	BaseValue     int    `json:"base_value" example:"15"`                                      // 基础值（职业初始值 + 玩家加点）
	ClassBonus    int    `json:"class_bonus" example:"2"`                                      // 职业加成（当前职业提供的加成）
	FinalValue    int    `json:"final_value" example:"17"`                                     // 最终值（base_value + class_bonus，未来可能包含装备、技能加成）
}

// SkillInfo 技能信息
type SkillInfo struct {
	HeroSkillID string `json:"hero_skill_id" example:"hero-skill-001"` // 英雄技能实例ID（UUID格式，用于升级或回退操作）
	SkillID     string `json:"skill_id" example:"skill-fireball-001"`  // 技能配置ID（UUID格式，关联 game_config.skills）
	SkillName   string `json:"skill_name" example:"烈焰斩"`               // 技能名称
	SkillCode   string `json:"skill_code" example:"FLAME_SLASH"`       // 技能代码（用于标识技能）
	SkillLevel  int    `json:"skill_level" example:"3"`                // 当前等级（1 到 max_level）
	MaxLevel    int    `json:"max_level" example:"10"`                 // 最大等级（来自技能配置）
}

// CreateHeroRequest HTTP create hero request
type CreateHeroRequest struct {
	ClassID     string `json:"class_id" validate:"required" example:"class-warrior-001"`    // 职业ID（必填，从职业列表接口获取）
	HeroName    string `json:"hero_name" validate:"required,min=2,max=20" example:"艾泽拉斯勇士"` // 英雄名称（必填，2-20字符）
	Description string `json:"description,omitempty" example:"来自北方的勇敢战士"`                   // 英雄描述（可选，最长200字符）
}

// AddExperienceRequest HTTP add experience request
type AddExperienceRequest struct {
	Amount int64 `json:"amount" validate:"required,min=1" example:"500"` // 增加的经验值（必填，最小1）
}

// AdvancementCheckResponse 进阶检查结果响应
type AdvancementCheckResponse struct {
	CanAdvance          bool     `json:"can_advance" example:"false"`                     // 是否可以进阶
	MissingRequirements []string `json:"missing_requirements,omitempty" example:"需要等级10"` // 缺少的条件列表（可选，仅当不满足时返回）
	RequiredLevel       int      `json:"required_level" example:"10"`                     // 要求等级
	CurrentLevel        int      `json:"current_level" example:"5"`                       // 当前等级
	RequiredHonor       int      `json:"required_honor" example:"100"`                    // 要求荣誉值
	CurrentHonor        int      `json:"current_honor" example:"0"`                       // 当前荣誉值
}

// HeroResponse HTTP hero response (基础信息，不含详细属性和技能)
type HeroResponse struct {
	ID                  string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`      // 英雄ID（UUID格式）
	UserID              string  `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"` // 用户ID（UUID格式）
	ClassID             string  `json:"class_id" example:"class-warrior-001"`                   // 当前职业ID
	HeroName            string  `json:"hero_name" example:"艾泽拉斯勇士"`                             // 英雄名称（2-20字符）
	Description         *string `json:"description,omitempty" example:"一位来自北方的战士"`              // 英雄描述（可选）
	CurrentLevel        int     `json:"current_level" example:"5"`                              // 当前等级（1-40）
	ExperienceTotal     int64   `json:"experience_total" example:"750"`                         // 累计总经验
	ExperienceAvailable int64   `json:"experience_available" example:"250"`                     // 可用经验（用于加点、学技能）
	ExperienceSpent     int64   `json:"experience_spent" example:"500"`                         // 已消耗经验
	Status              string  `json:"status" example:"active" enums:"active,inactive"`        // 英雄状态
	CreatedAt           string  `json:"created_at" example:"2025-10-17 10:30:00"`               // 创建时间
	UpdatedAt           string  `json:"updated_at" example:"2025-10-17 12:30:00"`               // 更新时间
}

// ==================== HTTP Handlers ====================

// CreateHero handles hero creation
// @Summary 创建英雄
// @Description 创建新的英雄角色。需要选择一个基础职业（战士/法师/盗贼等）
// @Description
// @Description **填写说明**：
// @Description - `class_id`: 从 `GET /api/v1/game/classes/basic` 接口获取基础职业列表，选择一个职业ID
// @Description - `hero_name`: 英雄名称，2-20个字符，支持中英文
// @Description - `description`: 可选，英雄背景描述
// @Description
// @Description **创建后**：
// @Description - 英雄初始等级为1，经验为0
// @Description - 继承选定职业的初始属性
// @Description - 自动学习职业的初始技能
// @Tags 英雄
// @Accept json
// @Produce json
// @Param request body CreateHeroRequest true "创建英雄请求"
// @Success 200 {object} response.Response{data=HeroResponse} "创建成功"
// @Failure 400 {object} response.Response "请求参数错误（如名称长度不符、职业ID无效）"
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
// @Description 检查英雄是否满足指定职业的进阶条件，返回详细的检查结果（等级、荣誉值、缺失的条件等）
// @Tags 英雄
// @Accept json
// @Produce json
// @Param hero_id path string true "英雄ID（UUID格式）"
// @Param target_class_id query string true "目标职业ID（UUID格式）"
// @Success 200 {object} response.Response{data=AdvancementCheckResponse} "检查成功"
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
	serviceResult, err := h.heroService.CheckAdvancementRequirements(c.Request().Context(), heroID, targetClassID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为 HTTP Response
	resp := &AdvancementCheckResponse{
		CanAdvance:          serviceResult.CanAdvance,
		MissingRequirements: serviceResult.MissingRequirements,
		RequiredLevel:       serviceResult.RequiredLevel,
		CurrentLevel:        serviceResult.CurrentLevel,
		RequiredHonor:       serviceResult.RequiredHonor,
		CurrentHonor:        serviceResult.CurrentHonor,
	}

	return response.EchoOK(c, h.respWriter, resp)
}
