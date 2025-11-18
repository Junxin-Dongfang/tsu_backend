package handler

import (
	"context"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/liangdas/mqant/module"
	mqrpc "github.com/liangdas/mqant/rpc"
	"google.golang.org/protobuf/proto"

	gamepb "tsu-self/internal/pb/game"
	commonpb "tsu-self/internal/pb/common"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/xerrors"
)

// TeamAdminHandler 团队管理后台 Handler
// 通过 RPC 调用 Game Server 实现团队管理功能
type TeamAdminHandler struct {
	rpcCaller  module.RPCModule
	respWriter response.Writer
}

// NewTeamAdminHandler 创建团队管理后台 Handler
func NewTeamAdminHandler(rpcCaller module.RPCModule, respWriter response.Writer) *TeamAdminHandler {
	return &TeamAdminHandler{
		rpcCaller:  rpcCaller,
		respWriter: respWriter,
	}
}

// ==================== HTTP Request/Response Models ====================

// TeamListResponse HTTP 团队列表响应
type TeamListResponse struct {
	Teams []TeamResponse `json:"teams"`  // 团队列表
	Total int64          `json:"total"`  // 总数
	Limit int            `json:"limit"`  // 每页数量
	Offset int           `json:"offset"` // 偏移量
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

// ListTeams 查询团队列表
// @Summary 查询团队列表
// @Description 查询团队列表(分页),通过RPC调用Game Server
// @Tags 团队管理(后台)
// @Accept json
// @Produce json
// @Param name query string false "团队名称(模糊查询)"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=TeamListResponse} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/teams [get]
func (h *TeamAdminHandler) ListTeams(c echo.Context) error {
	// 1. 获取查询参数
	name := c.QueryParam("name")

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page <= 0 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// 2. 构造 RPC 请求
	rpcReq := &gamepb.GetTeamListRequest{
		Name: name,
		Pagination: &commonpb.PaginationRequest{
			Page:     int32(page),
			PageSize: int32(pageSize),
		},
	}

	// 3. 调用 Game Server RPC
	rpcReqBytes, err := proto.Marshal(rpcReq)
	if err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "序列化RPC请求失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
	defer cancel()

	result, errStr := h.rpcCaller.Call(
		ctx,
		"game",
		"GetTeamList",
		mqrpc.Param(rpcReqBytes),
	)

	if errStr != "" {
		if ctx.Err() == context.DeadlineExceeded {
			appErr := xerrors.New(xerrors.CodeExternalServiceError, "Game服务超时")
			return response.EchoError(c, h.respWriter, appErr)
		}
		appErr := xerrors.New(xerrors.CodeExternalServiceError, errStr)
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 4. 解析响应
	rpcResp := &gamepb.GetTeamListResponse{}
	resultBytes, ok := result.([]byte)
	if !ok {
		appErr := xerrors.New(xerrors.CodeInternalError, "RPC响应类型错误")
		return response.EchoError(c, h.respWriter, appErr)
	}

	if err := proto.Unmarshal(resultBytes, rpcResp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 5. 转换为 HTTP 响应
	teamResponses := make([]TeamResponse, 0, len(rpcResp.Teams))
	for _, team := range rpcResp.Teams {
		teamResp := TeamResponse{
			ID:           team.Id,
			Name:         team.Name,
			LeaderHeroID: team.LeaderHeroId,
			MaxMembers:   int(team.MaxMembers),
			CreatedAt:    team.CreatedAt,
			UpdatedAt:    team.UpdatedAt,
		}

		if team.Description != "" {
			teamResp.Description = &team.Description
		}

		teamResponses = append(teamResponses, teamResp)
	}

	// 计算 offset
	offset := (page - 1) * pageSize

	resp := &TeamListResponse{
		Teams:  teamResponses,
		Total:  int64(rpcResp.Pagination.Total),
		Limit:  pageSize,
		Offset: offset,
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// GetTeam 查询团队详情
// @Summary 查询团队详情
// @Description 查询指定团队的详细信息,通过RPC调用Game Server
// @Tags 团队管理(后台)
// @Accept json
// @Produce json
// @Param team_id path string true "团队ID"
// @Success 200 {object} response.Response{data=TeamResponse} "获取成功"
// @Failure 404 {object} response.Response "团队不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/teams/{team_id} [get]
func (h *TeamAdminHandler) GetTeam(c echo.Context) error {
	teamID := c.Param("team_id")
	if teamID == "" {
		return response.EchoBadRequest(c, h.respWriter, "团队ID不能为空")
	}

	// 1. 构造 RPC 请求
	rpcReq := &gamepb.GetTeamDetailRequest{
		TeamId: teamID,
	}

	// 2. 调用 Game Server RPC
	rpcReqBytes, err := proto.Marshal(rpcReq)
	if err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "序列化RPC请求失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
	defer cancel()

	result, errStr := h.rpcCaller.Call(
		ctx,
		"game",
		"GetTeamDetail",
		mqrpc.Param(rpcReqBytes),
	)

	if errStr != "" {
		if ctx.Err() == context.DeadlineExceeded {
			appErr := xerrors.New(xerrors.CodeExternalServiceError, "Game服务超时")
			return response.EchoError(c, h.respWriter, appErr)
		}
		appErr := xerrors.New(xerrors.CodeExternalServiceError, errStr)
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 3. 解析响应
	rpcResp := &gamepb.GetTeamDetailResponse{}
	resultBytes, ok := result.([]byte)
	if !ok {
		appErr := xerrors.New(xerrors.CodeInternalError, "RPC响应类型错误")
		return response.EchoError(c, h.respWriter, appErr)
	}

	if err := proto.Unmarshal(resultBytes, rpcResp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 4. 转换为 HTTP 响应
	team := rpcResp.Team
	resp := &TeamResponse{
		ID:           team.Id,
		Name:         team.Name,
		LeaderHeroID: team.LeaderHeroId,
		MaxMembers:   int(team.MaxMembers),
		CreatedAt:    team.CreatedAt,
		UpdatedAt:    team.UpdatedAt,
	}

	if team.Description != "" {
		resp.Description = &team.Description
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// DisbandTeam 强制解散团队
// @Summary 强制解散团队
// @Description 强制解散团队(管理员操作),通过RPC调用Game Server
// @Tags 团队管理(后台)
// @Accept json
// @Produce json
// @Param team_id path string true "团队ID"
// @Success 200 {object} response.Response "解散成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/teams/{team_id}/disband [post]
func (h *TeamAdminHandler) DisbandTeam(c echo.Context) error {
	teamID := c.Param("team_id")
	if teamID == "" {
		return response.EchoBadRequest(c, h.respWriter, "团队ID不能为空")
	}

	// TODO: 获取管理员用户ID
	adminUserID := "admin" // 临时值，应从认证中间件获取

	// 1. 构造 RPC 请求
	rpcReq := &gamepb.ForceDisbandTeamRequest{
		TeamId:      teamID,
		AdminUserId: adminUserID,
	}

	// 2. 调用 Game Server RPC
	rpcReqBytes, err := proto.Marshal(rpcReq)
	if err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "序列化RPC请求失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
	defer cancel()

	result, errStr := h.rpcCaller.Call(
		ctx,
		"game",
		"ForceDisbandTeam",
		mqrpc.Param(rpcReqBytes),
	)

	if errStr != "" {
		if ctx.Err() == context.DeadlineExceeded {
			appErr := xerrors.New(xerrors.CodeExternalServiceError, "Game服务超时")
			return response.EchoError(c, h.respWriter, appErr)
		}
		appErr := xerrors.New(xerrors.CodeExternalServiceError, errStr)
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 3. 解析响应
	rpcResp := &gamepb.ForceDisbandTeamResponse{}
	resultBytes, ok := result.([]byte)
	if !ok {
		appErr := xerrors.New(xerrors.CodeInternalError, "RPC响应类型错误")
		return response.EchoError(c, h.respWriter, appErr)
	}

	if err := proto.Unmarshal(resultBytes, rpcResp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 4. 返回成功响应
	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": rpcResp.Message,
	})
}
