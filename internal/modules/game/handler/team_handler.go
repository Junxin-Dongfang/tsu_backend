package handler

import (
	"github.com/labstack/echo/v4"

	"tsu-self/internal/modules/game/service"
	"tsu-self/internal/pkg/response"
)

// TeamHandler 团队管理 Handler
type TeamHandler struct {
	teamService *service.TeamService
	respWriter  response.Writer
}

// NewTeamHandler 创建团队管理 Handler
func NewTeamHandler(serviceContainer *service.ServiceContainer, respWriter response.Writer) *TeamHandler {
	return &TeamHandler{
		teamService: serviceContainer.GetTeamService(),
		respWriter:  respWriter,
	}
}

// ==================== HTTP Request/Response Models ====================

// CreateTeamRequest HTTP 创建团队请求
type CreateTeamRequest struct {
	HeroID      string  `json:"hero_id" validate:"required" example:"hero-uuid-001"`      // 英雄ID（必填）
	TeamName    string  `json:"team_name" validate:"required,min=2,max=20" example:"无敌战队"` // 团队名称（必填，2-20字符）
	Description *string `json:"description,omitempty" example:"我们是最强的！"`                  // 团队描述（可选）
}

// UpdateTeamInfoRequest HTTP 更新团队信息请求
type UpdateTeamInfoRequest struct {
	Name        string  `json:"name,omitempty" validate:"omitempty,min=2,max=20" example:"新团队名称"` // 团队名称（可选，2-20字符）
	Description *string `json:"description,omitempty" example:"新的团队描述"`                          // 团队描述（可选）
}

// TeamResponse HTTP 团队响应
type TeamResponse struct {
	ID            string  `json:"id" example:"team-uuid-001"`                  // 团队ID
	Name          string  `json:"name" example:"无敌战队"`                         // 团队名称
	LeaderHeroID  string  `json:"leader_hero_id" example:"hero-uuid-001"`      // 队长英雄ID
	MaxMembers    int     `json:"max_members" example:"12"`                    // 最大成员数
	Description   *string `json:"description,omitempty" example:"我们是最强的！"`     // 团队描述
	CreatedAt     string  `json:"created_at" example:"2025-01-01 12:00:00"`    // 创建时间
	UpdatedAt     string  `json:"updated_at" example:"2025-01-01 12:00:00"`    // 更新时间
}

// ==================== HTTP Handlers ====================

// CreateTeam 创建团队
// @Summary 创建团队
// @Description 创建新的团队。创建者自动成为队长，同时创建团队仓库
// @Description
// @Description **填写说明**：
// @Description - `hero_id`: 英雄ID，该英雄将成为队长
// @Description - `team_name`: 团队名称，2-20个字符，必须唯一
// @Description - `description`: 可选，团队描述
// @Description
// @Description **创建后**：
// @Description - 创建者自动成为队长
// @Description - 自动创建团队仓库
// @Description - 最大成员数为12人
// @Tags 团队
// @Accept json
// @Produce json
// @Param request body CreateTeamRequest true "创建团队请求"
// @Success 200 {object} response.Response{data=TeamResponse} "创建成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/teams [post]
func (h *TeamHandler) CreateTeam(c echo.Context) error {
	// 1. 绑定和验证 HTTP 请求
	var req CreateTeamRequest
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
	createReq := &service.CreateTeamRequest{
		UserID:   userID.(string),
		HeroID:   req.HeroID,
		TeamName: req.TeamName,
	}
	if req.Description != nil {
		createReq.Description = *req.Description
	}

	team, err := h.teamService.CreateTeam(c.Request().Context(), createReq)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 4. 转换为 HTTP 响应
	resp := &TeamResponse{
		ID:           team.ID,
		Name:         team.Name,
		LeaderHeroID: team.LeaderHeroID,
		MaxMembers:   team.MaxMembers,
		CreatedAt:    team.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    team.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if !team.Description.IsZero() {
		desc := team.Description.String
		resp.Description = &desc
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// GetTeam 获取团队详情
// @Summary 获取团队详情
// @Description 获取指定团队的详细信息
// @Tags 团队
// @Accept json
// @Produce json
// @Param team_id path string true "团队ID"
// @Success 200 {object} response.Response{data=TeamResponse} "获取成功"
// @Failure 404 {object} response.Response "团队不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/teams/{team_id} [get]
func (h *TeamHandler) GetTeam(c echo.Context) error {
	teamID := c.Param("team_id")
	if teamID == "" {
		return response.EchoBadRequest(c, h.respWriter, "团队ID不能为空")
	}

	team, err := h.teamService.GetTeamByID(c.Request().Context(), teamID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为 HTTP 响应
	resp := &TeamResponse{
		ID:           team.ID,
		Name:         team.Name,
		LeaderHeroID: team.LeaderHeroID,
		MaxMembers:   team.MaxMembers,
		CreatedAt:    team.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    team.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if !team.Description.IsZero() {
		desc := team.Description.String
		resp.Description = &desc
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// UpdateTeamInfo 更新团队信息
// @Summary 更新团队信息
// @Description 更新团队名称和描述（只有队长可以操作）
// @Tags 团队
// @Accept json
// @Produce json
// @Param team_id path string true "团队ID"
// @Param hero_id query string true "操作者英雄ID"
// @Param request body UpdateTeamInfoRequest true "更新团队信息请求"
// @Success 200 {object} response.Response "更新成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/teams/{team_id} [put]
func (h *TeamHandler) UpdateTeamInfo(c echo.Context) error {
	// 1. 获取路径参数
	teamID := c.Param("team_id")
	if teamID == "" {
		return response.EchoBadRequest(c, h.respWriter, "团队ID不能为空")
	}

	heroID := c.QueryParam("hero_id")
	if heroID == "" {
		return response.EchoBadRequest(c, h.respWriter, "英雄ID不能为空")
	}

	// 2. 绑定和验证 HTTP 请求
	var req UpdateTeamInfoRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 3. 调用 Service
	updateReq := &service.UpdateTeamInfoRequest{
		TeamID:      teamID,
		HeroID:      heroID,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.teamService.UpdateTeamInfo(c.Request().Context(), updateReq); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{})
}

// DisbandTeam 解散团队
// @Summary 解散团队
// @Description 解散团队（只有队长可以操作，且仓库必须为空）
// @Tags 团队
// @Accept json
// @Produce json
// @Param team_id path string true "团队ID"
// @Param hero_id query string true "操作者英雄ID"
// @Success 200 {object} response.Response "解散成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/teams/{team_id}/disband [post]
func (h *TeamHandler) DisbandTeam(c echo.Context) error {
	// 1. 获取路径参数
	teamID := c.Param("team_id")
	if teamID == "" {
		return response.EchoBadRequest(c, h.respWriter, "团队ID不能为空")
	}

	heroID := c.QueryParam("hero_id")
	if heroID == "" {
		return response.EchoBadRequest(c, h.respWriter, "英雄ID不能为空")
	}

	// 2. 调用 Service
	if err := h.teamService.DisbandTeam(c.Request().Context(), teamID, heroID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{})
}

// LeaveTeam 离开团队
// @Summary 离开团队
// @Description 离开团队（队长不能离开）
// @Tags 团队
// @Accept json
// @Produce json
// @Param team_id path string true "团队ID"
// @Param hero_id query string true "操作者英雄ID"
// @Success 200 {object} response.Response "离开成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/teams/{team_id}/leave [post]
func (h *TeamHandler) LeaveTeam(c echo.Context) error {
	// 1. 获取路径参数
	teamID := c.Param("team_id")
	if teamID == "" {
		return response.EchoBadRequest(c, h.respWriter, "团队ID不能为空")
	}

	heroID := c.QueryParam("hero_id")
	if heroID == "" {
		return response.EchoBadRequest(c, h.respWriter, "英雄ID不能为空")
	}

	// 2. 调用 Service
	if err := h.teamService.LeaveTeam(c.Request().Context(), teamID, heroID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{})
}

