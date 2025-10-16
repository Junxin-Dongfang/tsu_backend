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

	// 2. 获取用户ID（从上下文或 session）
	// TODO: 从认证中间件获取用户ID
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
