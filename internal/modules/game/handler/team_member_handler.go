package handler

import (
	"time"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/modules/game/service"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/xerrors"
)

// TeamMemberHandler 团队成员管理 Handler
type TeamMemberHandler struct {
	memberService *service.TeamMemberService
	respWriter    response.Writer
}

// NewTeamMemberHandler 创建团队成员管理 Handler
func NewTeamMemberHandler(serviceContainer *service.ServiceContainer, respWriter response.Writer) *TeamMemberHandler {
	return &TeamMemberHandler{
		memberService: serviceContainer.GetTeamMemberService(),
		respWriter:    respWriter,
	}
}

// getTeamIDAndInviterHeroID 从查询参数或请求体中获取团队ID和邀请人英雄ID
func (h *TeamMemberHandler) getTeamIDAndInviterHeroID(c echo.Context) (string, string, error) {
	// 优先从查询参数获取（中间件会从查询参数读取）
	teamID := c.QueryParam("team_id")
	inviterHeroID := c.QueryParam("hero_id")

	// 尝试从context获取hero_id（由认证中间件设置）
	if inviterHeroID == "" {
		if heroIDFromCtx := c.Get("hero_id"); heroIDFromCtx != nil {
			if heroIDStr, ok := heroIDFromCtx.(string); ok && heroIDStr != "" {
				inviterHeroID = heroIDStr
			}
		}
	}

	// 如果查询参数缺失，尝试从请求体获取（向后兼容）
	if teamID == "" || inviterHeroID == "" {
		var req InviteMemberRequest
		if err := c.Bind(&req); err == nil {
			if teamID == "" && req.TeamID != "" {
				teamID = req.TeamID
			}
			if inviterHeroID == "" && req.InviterHeroID != "" {
				inviterHeroID = req.InviterHeroID
			}
		}
	}

	if teamID == "" {
		return "", "", xerrors.New(xerrors.CodeInvalidParams, "团队ID不能为空，请提供查询参数 team_id 或请求体中的 team_id")
	}
	if inviterHeroID == "" {
		return "", "", xerrors.New(xerrors.CodeInvalidParams, "英雄ID不能为空，请提供查询参数 hero_id 或请求体中的 inviter_hero_id")
	}

	return teamID, inviterHeroID, nil
}

// ==================== HTTP Request/Response Models ====================

// ApplyToJoinRequest HTTP 申请加入团队请求
type ApplyToJoinRequest struct {
	TeamID  string  `json:"team_id" validate:"required" example:"team-uuid-001"` // 团队ID（必填）
	HeroID  string  `json:"hero_id" validate:"required" example:"hero-uuid-001"` // 英雄ID（必填）
	Message *string `json:"message,omitempty" example:"我想加入你们的团队！"`              // 申请留言（可选）
}

// ApproveJoinRequestRequest HTTP 审批加入申请请求
type ApproveJoinRequestRequest struct {
	RequestID string `json:"request_id" validate:"required" example:"request-uuid-001"` // 申请ID（必填）
	HeroID    string `json:"hero_id" validate:"required" example:"hero-uuid-001"`       // 审批人英雄ID（必填）
	Approved  bool   `json:"approved" example:"true"`                                   // 是否批准（true=批准, false=拒绝）
}

// InviteMemberRequest HTTP 邀请成员请求
type InviteMemberRequest struct {
	TeamID        string  `json:"team_id,omitempty" validate:"-" example:"team-uuid-001"`         // 团队ID（可选，支持查询参数）
	InviterHeroID string  `json:"inviter_hero_id,omitempty" validate:"-" example:"hero-uuid-001"` // 邀请人英雄ID（可选，支持查询参数）
	InviteeHeroID string  `json:"invitee_hero_id" validate:"required" example:"hero-uuid-002"`    // 被邀请人英雄ID（必填）
	Message       *string `json:"message,omitempty" example:"欢迎加入我们的团队！"`                         // 邀请留言（可选）
}

// InviteResponse HTTP 邀请响应
type InviteResponse struct {
	InvitationID  string `json:"invitation_id" example:"invitation-uuid-001"` // 邀请ID
	TeamID        string `json:"team_id" example:"team-uuid-001"`             // 团队ID
	InviterHeroID string `json:"inviter_hero_id" example:"hero-uuid-001"`     // 邀请人英雄ID
	InviteeHeroID string `json:"invitee_hero_id" example:"hero-uuid-002"`     // 被邀请人英雄ID
	Status        string `json:"status" example:"pending_approval"`           // 邀请状态
	ExpiresAt     string `json:"expires_at" example:"2025-12-02T00:00:00Z"`   // 过期时间
	Message       string `json:"message" example:"欢迎加入我们的团队！"`                // 邀请留言
}

// ApproveInvitationRequest HTTP 审批邀请请求
type ApproveInvitationRequest struct {
	InvitationID string `json:"invitation_id" validate:"required" example:"invitation-uuid-001"` // 邀请ID（必填）
	HeroID       string `json:"hero_id" validate:"required" example:"hero-uuid-001"`             // 审批人英雄ID（必填）
	Approved     bool   `json:"approved" example:"true"`                                         // 是否批准（true=批准, false=拒绝）
}

// AcceptInvitationRequest HTTP 接受邀请请求
type AcceptInvitationRequest struct {
	InvitationID string `json:"invitation_id" validate:"required" example:"invitation-uuid-001"` // 邀请ID（必填）
	HeroID       string `json:"hero_id" validate:"required" example:"hero-uuid-001"`             // 被邀请人英雄ID（必填）
}

// KickMemberRequest HTTP 踢出成员请求
type KickMemberRequest struct {
	TeamID       string  `json:"team_id" validate:"required" example:"team-uuid-001"`        // 团队ID（必填）
	TargetHeroID string  `json:"target_hero_id" validate:"required" example:"hero-uuid-002"` // 被踢出的英雄ID（必填）
	KickerHeroID string  `json:"kicker_hero_id" validate:"required" example:"hero-uuid-001"` // 踢出者英雄ID（必填）
	Reason       *string `json:"reason,omitempty" example:"违反团队规则"`                          // 踢出原因（可选）
}

// PromoteToAdminRequest HTTP 任命管理员请求
type PromoteToAdminRequest struct {
	TeamID       string `json:"team_id" validate:"required" example:"team-uuid-001"`        // 团队ID（必填）
	TargetHeroID string `json:"target_hero_id" validate:"required" example:"hero-uuid-002"` // 被任命的英雄ID（必填）
	LeaderHeroID string `json:"leader_hero_id" validate:"required" example:"hero-uuid-001"` // 队长英雄ID（必填）
}

// ==================== HTTP Handlers ====================

// ApplyToJoin 申请加入团队
// @Summary 申请加入团队
// @Description 申请加入指定团队，需要队长或管理员审批
// @Tags 团队成员
// @Accept json
// @Produce json
// @Param request body ApplyToJoinRequest true "申请加入团队请求"
// @Success 200 {object} response.Response "申请成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/teams/join/apply [post]
func (h *TeamMemberHandler) ApplyToJoin(c echo.Context) error {
	// 1. 绑定和验证 HTTP 请求
	var req ApplyToJoinRequest
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
	applyReq := &service.ApplyToJoinRequest{
		TeamID: req.TeamID,
		HeroID: req.HeroID,
		UserID: userID.(string),
	}
	if req.Message != nil {
		applyReq.Message = *req.Message
	}

	applicationID, err := h.memberService.ApplyToJoin(c.Request().Context(), applyReq)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"application_id": applicationID,
		"status":         "pending",
	})
}

// ApproveJoinRequest 审批加入申请
// @Summary 审批加入申请
// @Description 审批加入申请（队长或管理员）
// @Tags 团队成员
// @Accept json
// @Produce json
// @Param request body ApproveJoinRequestRequest true "审批加入申请请求"
// @Success 200 {object} response.Response "审批成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/teams/join/approve [post]
func (h *TeamMemberHandler) ApproveJoinRequest(c echo.Context) error {
	// 1. 绑定和验证 HTTP 请求
	var req ApproveJoinRequestRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 2. 调用 Service
	approveReq := &service.ApproveJoinRequestRequest{
		RequestID: req.RequestID,
		HeroID:    req.HeroID,
		Approved:  req.Approved,
	}

	if err := h.memberService.ApproveJoinRequest(c.Request().Context(), approveReq); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{})
}

// InviteMember 邀请成员
// @Summary 邀请成员
// @Description 邀请成员加入团队（需要团队成员权限）
// @Description 支持通过查询参数或请求体提供 team_id 和 inviter_hero_id
// @Tags 团队成员
// @Accept json
// @Produce json
// @Param team_id query string false "团队ID（可选）"
// @Param hero_id query string false "邀请人英雄ID（可选）"
// @Param request body InviteMemberRequest true "邀请成员请求"
// @Success 200 {object} response.Response{data=InviteResponse} "邀请成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/teams/invite [post]
func (h *TeamMemberHandler) InviteMember(c echo.Context) error {
	// 1. 获取团队ID和邀请人英雄ID（从查询参数或请求体）
	teamID, inviterHeroID, err := h.getTeamIDAndInviterHeroID(c)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 2. 绑定和验证 HTTP 请求（除了team_id和inviter_hero_id）
	var req InviteMemberRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}

	// 设置从查询参数获取的值
	req.TeamID = teamID
	req.InviterHeroID = inviterHeroID

	// 验证请求（只验证InviteeHeroID和Message）
	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 3. 调用 Service
	inviteReq := &service.InviteMemberRequest{
		TeamID:        req.TeamID,
		InviterHeroID: req.InviterHeroID,
		InviteeHeroID: req.InviteeHeroID,
	}
	if req.Message != nil {
		inviteReq.Message = *req.Message
	}

	// 3. 调用 Service
	invitation, err := h.memberService.InviteMember(c.Request().Context(), inviteReq)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 4. 返回邀请响应（包含邀请ID）
	return response.EchoOK(c, h.respWriter, InviteResponse{
		InvitationID:  invitation.ID,
		TeamID:        invitation.TeamID,
		InviterHeroID: invitation.InviterHeroID,
		InviteeHeroID: invitation.InviteeHeroID,
		Status:        invitation.Status,
		ExpiresAt:     invitation.CreatedAt.Add(24 * time.Hour).UTC().Format(time.RFC3339),
		Message:       "邀请发送成功",
	})
}

// ApproveInvitation 审批邀请
// @Summary 审批邀请
// @Description 审批邀请（队长或管理员）
// @Tags 团队成员
// @Accept json
// @Produce json
// @Param team_id query string true "团队ID（权限校验需要）"
// @Param request body ApproveInvitationRequest true "审批邀请请求"
// @Success 200 {object} response.Response "审批成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/teams/invite/approve [post]
func (h *TeamMemberHandler) ApproveInvitation(c echo.Context) error {
	// 1. 绑定和验证 HTTP 请求
	var req ApproveInvitationRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 2. 调用 Service
	approveReq := &service.ApproveInvitationRequest{
		InvitationID: req.InvitationID,
		HeroID:       req.HeroID,
		Approved:     req.Approved,
	}

	if err := h.memberService.ApproveInvitation(c.Request().Context(), approveReq); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{})
}

// AcceptInvitation 接受邀请
// @Summary 接受邀请
// @Description 接受邀请（被邀请人）
// @Tags 团队成员
// @Accept json
// @Produce json
// @Param request body AcceptInvitationRequest true "接受邀请请求"
// @Success 200 {object} response.Response "接受成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/teams/invite/accept [post]
func (h *TeamMemberHandler) AcceptInvitation(c echo.Context) error {
	// 1. 绑定和验证 HTTP 请求
	var req AcceptInvitationRequest
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
	acceptReq := &service.AcceptInvitationRequest{
		InvitationID: req.InvitationID,
		HeroID:       req.HeroID,
		UserID:       userID.(string),
	}

	if err := h.memberService.AcceptInvitation(c.Request().Context(), acceptReq); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{})
}

// RejectInvitation 拒绝邀请
// @Summary 拒绝邀请
// @Description 拒绝邀请（被邀请人）
// @Tags 团队成员
// @Accept json
// @Produce json
// @Param invitation_id query string true "邀请ID"
// @Param hero_id query string true "被邀请人英雄ID"
// @Success 200 {object} response.Response "拒绝成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/teams/invite/reject [post]
func (h *TeamMemberHandler) RejectInvitation(c echo.Context) error {
	// 1. 获取查询参数
	invitationID := c.QueryParam("invitation_id")
	if invitationID == "" {
		return response.EchoBadRequest(c, h.respWriter, "邀请ID不能为空")
	}

	heroID := c.QueryParam("hero_id")
	if heroID == "" {
		return response.EchoBadRequest(c, h.respWriter, "英雄ID不能为空")
	}

	// 2. 调用 Service
	if err := h.memberService.RejectInvitation(c.Request().Context(), invitationID, heroID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{})
}

// KickMember 踢出成员
// @Summary 踢出成员
// @Description 踢出成员（队长或管理员）
// @Tags 团队成员
// @Accept json
// @Produce json
// @Param request body KickMemberRequest true "踢出成员请求"
// @Success 200 {object} response.Response "踢出成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/teams/members/kick [post]
func (h *TeamMemberHandler) KickMember(c echo.Context) error {
	// 1. 绑定和验证 HTTP 请求
	var req KickMemberRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 2. 调用 Service
	kickReq := &service.KickMemberRequest{
		TeamID:       req.TeamID,
		TargetHeroID: req.TargetHeroID,
		KickerHeroID: req.KickerHeroID,
	}
	if req.Reason != nil {
		kickReq.Reason = *req.Reason
	}

	if err := h.memberService.KickMember(c.Request().Context(), kickReq); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{})
}

// PromoteToAdmin 任命管理员
// @Summary 任命管理员
// @Description 任命管理员（只有队长可以操作）
// @Tags 团队成员
// @Accept json
// @Produce json
// @Param request body PromoteToAdminRequest true "任命管理员请求"
// @Success 200 {object} response.Response "任命成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/teams/members/promote [post]
func (h *TeamMemberHandler) PromoteToAdmin(c echo.Context) error {
	// 1. 绑定和验证 HTTP 请求
	var req PromoteToAdminRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 2. 调用 Service
	promoteReq := &service.PromoteToAdminRequest{
		TeamID:       req.TeamID,
		TargetHeroID: req.TargetHeroID,
		LeaderHeroID: req.LeaderHeroID,
	}

	if err := h.memberService.PromoteToAdmin(c.Request().Context(), promoteReq); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{})
}

// DemoteAdmin 撤销管理员
// @Summary 撤销管理员
// @Description 撤销管理员（只有队长可以操作）
// @Tags 团队成员
// @Accept json
// @Produce json
// @Param team_id query string true "团队ID"
// @Param target_hero_id query string true "被撤销的英雄ID"
// @Param leader_hero_id query string true "队长英雄ID"
// @Success 200 {object} response.Response "撤销成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/teams/members/demote [post]
func (h *TeamMemberHandler) DemoteAdmin(c echo.Context) error {
	// 1. 获取查询参数
	teamID := c.QueryParam("team_id")
	if teamID == "" {
		return response.EchoBadRequest(c, h.respWriter, "团队ID不能为空")
	}

	targetHeroID := c.QueryParam("target_hero_id")
	if targetHeroID == "" {
		return response.EchoBadRequest(c, h.respWriter, "目标英雄ID不能为空")
	}

	leaderHeroID := c.QueryParam("leader_hero_id")
	if leaderHeroID == "" {
		return response.EchoBadRequest(c, h.respWriter, "队长英雄ID不能为空")
	}

	// 2. 调用 Service
	if err := h.memberService.DemoteAdmin(c.Request().Context(), teamID, targetHeroID, leaderHeroID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{})
}
