package handler

import (
	"context"
	"database/sql"
	"fmt"

	"google.golang.org/protobuf/proto"

	pb "tsu-self/internal/pb/game"
	commonpb "tsu-self/internal/pb/common"
	"tsu-self/internal/modules/game/service"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
	"tsu-self/internal/pkg/xerrors"
)

// TeamRPCHandler 团队 RPC 处理器
// 提供给 Admin Server 调用的团队管理接口
type TeamRPCHandler struct {
	db          *sql.DB
	teamService *service.TeamService
}

// NewTeamRPCHandler 创建团队 RPC Handler
func NewTeamRPCHandler(serviceContainer *service.ServiceContainer, db *sql.DB) *TeamRPCHandler {
	return &TeamRPCHandler{
		db:          db,
		teamService: serviceContainer.GetTeamService(),
	}
}

// ==================== RPC Methods ====================

// GetTeamList 获取团队列表
// 供 Admin Server 查询团队列表（分页）
func (h *TeamRPCHandler) GetTeamList(data []byte) ([]byte, error) {
	req := &pb.GetTeamListRequest{}
	if err := proto.Unmarshal(data, req); err != nil {
		return nil, xerrors.NewInvalidArgumentError("request", "invalid protobuf data")
	}

	ctx := context.Background()

	// 处理分页参数
	offset := 0
	limit := 20 // 默认每页 20 条
	if req.Pagination != nil {
		if req.Pagination.Page > 0 && req.Pagination.PageSize > 0 {
			offset = int(req.Pagination.Page-1) * int(req.Pagination.PageSize)
			limit = int(req.Pagination.PageSize)
		}
	}

	// 调用 Repository 查询
	teamRepo := impl.NewTeamRepository(h.db)
	teams, total, err := teamRepo.List(ctx, interfaces.TeamQueryParams{
		Name:   req.Name,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "failed to list teams")
	}

	// 转换为 Protobuf
	pbTeams := make([]*pb.TeamInfo, len(teams))
	for i, team := range teams {
		pbTeams[i] = &pb.TeamInfo{
			Id:           team.ID,
			Name:         team.Name,
			LeaderHeroId: team.LeaderHeroID,
			MaxMembers:   int32(team.MaxMembers),
			CreatedAt:    team.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    team.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		if !team.Description.IsZero() {
			pbTeams[i].Description = team.Description.String
		}
	}

	// 计算分页元数据
	totalPages := int32(total) / int32(limit)
	if int32(total)%int32(limit) > 0 {
		totalPages++
	}

	resp := &pb.GetTeamListResponse{
		Teams: pbTeams,
		Pagination: &commonpb.PaginationMetadata{
			Page:       req.Pagination.GetPage(),
			PageSize:   req.Pagination.GetPageSize(),
			Total:      int32(total),
			TotalPages: totalPages,
		},
	}

	return proto.Marshal(resp)
}

// GetTeamDetail 获取团队详情
// 供 Admin Server 查询团队详细信息，包括成员列表和统计信息
func (h *TeamRPCHandler) GetTeamDetail(data []byte) ([]byte, error) {
	req := &pb.GetTeamDetailRequest{}
	if err := proto.Unmarshal(data, req); err != nil {
		return nil, xerrors.NewInvalidArgumentError("request", "invalid protobuf data")
	}

	ctx := context.Background()

	// 1. 获取团队基本信息
	team, err := h.teamService.GetTeamByID(ctx, req.TeamId)
	if err != nil {
		return nil, err
	}

	pbTeam := &pb.TeamInfo{
		Id:           team.ID,
		Name:         team.Name,
		LeaderHeroId: team.LeaderHeroID,
		MaxMembers:   int32(team.MaxMembers),
		CreatedAt:    team.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    team.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if !team.Description.IsZero() {
		pbTeam.Description = team.Description.String
	}

	// 2. 获取团队成员列表
	memberRepo := impl.NewTeamMemberRepository(h.db)
	members, err := memberRepo.ListByTeam(ctx, req.TeamId)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "failed to list team members")
	}

	pbMembers := make([]*pb.TeamMemberInfo, len(members))
	for i, member := range members {
		pbMembers[i] = &pb.TeamMemberInfo{
			Id:           member.ID,
			TeamId:       member.TeamID,
			HeroId:       member.HeroID,
			Role:         member.Role,
			JoinedAt:     member.JoinedAt.Format("2006-01-02 15:04:05"),
			LastActiveAt: member.LastActiveAt.Format("2006-01-02 15:04:05"),
		}
	}

	// 3. 获取团队统计信息
	statistics := &pb.TeamStatistics{
		MemberCount: int32(len(members)),
	}

	// 获取仓库信息
	warehouseRepo := impl.NewTeamWarehouseRepository(h.db)
	warehouse, err := warehouseRepo.GetByTeamID(ctx, req.TeamId)
	if err == nil && warehouse != nil {
		statistics.WarehouseGold = int32(warehouse.GoldAmount)

		// 获取仓库物品数量
		warehouseItemRepo := impl.NewTeamWarehouseItemRepository(h.db)
		itemCount, err := warehouseItemRepo.CountDistinctItems(ctx, warehouse.ID)
		if err == nil {
			statistics.WarehouseItems = int32(itemCount)
		}
	}

	resp := &pb.GetTeamDetailResponse{
		Team:       pbTeam,
		Members:    pbMembers,
		Statistics: statistics,
	}

	return proto.Marshal(resp)
}

// ForceDisbandTeam 强制解散团队
// 供 Admin Server 强制解散团队
func (h *TeamRPCHandler) ForceDisbandTeam(data []byte) ([]byte, error) {
	req := &pb.ForceDisbandTeamRequest{}
	if err := proto.Unmarshal(data, req); err != nil {
		return nil, xerrors.NewInvalidArgumentError("request", "invalid protobuf data")
	}

	ctx := context.Background()

	// 调用 Service 层解散团队
	// 注意：这里使用空字符串作为 heroID，因为是管理员强制解散
	err := h.teamService.DisbandTeam(ctx, req.TeamId, "")
	if err != nil {
		return nil, err
	}

	// 记录管理员操作日志
	fmt.Printf("[Team RPC] Team %s force disbanded by admin %s\n", req.TeamId, req.AdminUserId)

	resp := &pb.ForceDisbandTeamResponse{
		Success: true,
		Message: "团队已强制解散",
	}

	return proto.Marshal(resp)
}

// GetTeamMembers 获取团队成员列表
// 供 Admin Server 查询团队成员列表
func (h *TeamRPCHandler) GetTeamMembers(data []byte) ([]byte, error) {
	req := &pb.GetTeamMembersRequest{}
	if err := proto.Unmarshal(data, req); err != nil {
		return nil, xerrors.NewInvalidArgumentError("request", "invalid protobuf data")
	}

	ctx := context.Background()

	// 获取团队成员列表
	memberRepo := impl.NewTeamMemberRepository(h.db)
	members, err := memberRepo.ListByTeam(ctx, req.TeamId)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "failed to list team members")
	}

	pbMembers := make([]*pb.TeamMemberInfo, len(members))
	for i, member := range members {
		pbMembers[i] = &pb.TeamMemberInfo{
			Id:           member.ID,
			TeamId:       member.TeamID,
			HeroId:       member.HeroID,
			Role:         member.Role,
			JoinedAt:     member.JoinedAt.Format("2006-01-02 15:04:05"),
			LastActiveAt: member.LastActiveAt.Format("2006-01-02 15:04:05"),
		}
	}

	resp := &pb.GetTeamMembersResponse{
		Members: pbMembers,
	}

	return proto.Marshal(resp)
}
