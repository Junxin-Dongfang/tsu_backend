package handler

import (
	"database/sql"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/modules/game/service"
	"tsu-self/internal/pkg/response"
)

// HeroActivationHandler 处理英雄激活相关的HTTP请求
type HeroActivationHandler struct {
	heroActivationService *service.HeroActivationService
	respWriter            response.Writer
}

// NewHeroActivationHandler 创建新的英雄激活处理器
func NewHeroActivationHandler(db *sql.DB, respWriter response.Writer) *HeroActivationHandler {
	return &HeroActivationHandler{
		heroActivationService: service.NewHeroActivationService(db),
		respWriter:            respWriter,
	}
}

// ActivateHeroRequest 激活英雄请求
type ActivateHeroRequest struct {
	HeroID string `json:"hero_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// DeactivateHeroRequest 停用英雄请求
type DeactivateHeroRequest struct {
	HeroID string `json:"hero_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// SwitchCurrentHeroRequest 切换当前英雄请求
type SwitchCurrentHeroRequest struct {
	HeroID string `json:"hero_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// ActivatedHeroResponse 已激活英雄响应
type ActivatedHeroResponse struct {
	ID              string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	HeroName        string `json:"hero_name" example:"战士阿尔法"`
	IsActivated     bool   `json:"is_activated" example:"true"`
	IsCurrentHero   bool   `json:"is_current_hero" example:"true"`
	CreatedAt       string `json:"created_at" example:"2025-10-17 10:30:00"`
}

// ActivateHero 激活英雄
// @Summary 激活英雄
// @Description 激活指定的英雄。如果用户没有当前操作英雄，该英雄会自动成为当前英雄
// @Tags 英雄激活
// @Accept json
// @Produce json
// @Param request body ActivateHeroRequest true "激活英雄请求"
// @Success 200 {object} response.Response "激活成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未登录"
// @Failure 404 {object} response.Response "英雄不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/heroes/{hero_id}/activate [patch]
func (h *HeroActivationHandler) ActivateHero(c echo.Context) error {
	userID := c.Get("user_id")
	if userID == nil {
		return response.EchoUnauthorized(c, h.respWriter, "未登录")
	}

	var req ActivateHeroRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	err := h.heroActivationService.ActivateHero(c.Request().Context(), userID.(string), req.HeroID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]string{"status": "success"})
}

// DeactivateHero 停用英雄
// @Summary 停用英雄
// @Description 停用指定的英雄。不能停用当前操作英雄
// @Tags 英雄激活
// @Accept json
// @Produce json
// @Param request body DeactivateHeroRequest true "停用英雄请求"
// @Success 200 {object} response.Response "停用成功"
// @Failure 400 {object} response.Response "请求参数错误或无法停用当前英雄"
// @Failure 401 {object} response.Response "未登录"
// @Failure 404 {object} response.Response "英雄不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/heroes/{hero_id}/deactivate [patch]
func (h *HeroActivationHandler) DeactivateHero(c echo.Context) error {
	userID := c.Get("user_id")
	if userID == nil {
		return response.EchoUnauthorized(c, h.respWriter, "未登录")
	}

	var req DeactivateHeroRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	err := h.heroActivationService.DeactivateHero(c.Request().Context(), userID.(string), req.HeroID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]string{"status": "success"})
}

// SwitchCurrentHero 切换当前英雄
// @Summary 切换当前操作英雄
// @Description 将当前操作英雄切换到指定的已激活英雄
// @Tags 英雄激活
// @Accept json
// @Produce json
// @Param request body SwitchCurrentHeroRequest true "切换英雄请求"
// @Success 200 {object} response.Response "切换成功"
// @Failure 400 {object} response.Response "请求参数错误或目标英雄未激活"
// @Failure 401 {object} response.Response "未登录"
// @Failure 404 {object} response.Response "英雄不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/heroes/switch [patch]
func (h *HeroActivationHandler) SwitchCurrentHero(c echo.Context) error {
	userID := c.Get("user_id")
	if userID == nil {
		return response.EchoUnauthorized(c, h.respWriter, "未登录")
	}

	var req SwitchCurrentHeroRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	err := h.heroActivationService.SwitchCurrentHero(c.Request().Context(), userID.(string), req.HeroID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]string{"status": "success"})
}

// GetActivatedHeroes 获取已激活的英雄列表
// @Summary 获取已激活的英雄列表
// @Description 获取当前用户所有已激活的英雄，并标记哪个是当前操作英雄
// @Tags 英雄激活
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]ActivatedHeroResponse} "获取成功"
// @Failure 401 {object} response.Response "未登录"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/heroes/activated [get]
func (h *HeroActivationHandler) GetActivatedHeroes(c echo.Context) error {
	userID := c.Get("user_id")
	if userID == nil {
		return response.EchoUnauthorized(c, h.respWriter, "未登录")
	}

	heroes, currentHeroID, err := h.heroActivationService.GetActivatedHeroes(c.Request().Context(), userID.(string))
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	heroResponses := make([]ActivatedHeroResponse, 0)
	for _, hero := range heroes {
		resp := ActivatedHeroResponse{
			ID:            hero.ID,
			HeroName:      hero.HeroName,
			IsActivated:   hero.IsActivated,
			IsCurrentHero: hero.ID == currentHeroID,
			CreatedAt:     hero.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		heroResponses = append(heroResponses, resp)
	}

	return response.EchoOK(c, h.respWriter, &heroResponses)
}
