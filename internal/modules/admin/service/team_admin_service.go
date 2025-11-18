package service

import (
	"context"
	"database/sql"
	"time"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// TeamAdminService 团队管理服务（Admin Server）
type TeamAdminService struct {
	db             *sql.DB
	teamRepo       interfaces.TeamRepository
	teamMemberRepo interfaces.TeamMemberRepository
	// TODO: 添加 RPC Client 调用 Game Server
}

// NewTeamAdminService 创建团队管理服务
func NewTeamAdminService(db *sql.DB) *TeamAdminService {
	return &TeamAdminService{
		db:             db,
		teamRepo:       impl.NewTeamRepository(db),
		teamMemberRepo: impl.NewTeamMemberRepository(db),
	}
}

// ListTeamsRequest 查询团队列表请求
type ListTeamsRequest struct {
	Name   string // 团队名称（模糊查询）
	Limit  int    // 每页数量
	Offset int    // 偏移量
}

// ListTeams 查询团队列表
func (s *TeamAdminService) ListTeams(ctx context.Context, req *ListTeamsRequest) ([]*game_runtime.Team, int64, error) {
	// 设置默认值
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	// 查询团队列表
	params := interfaces.TeamQueryParams{
		Name:   req.Name,
		Limit:  req.Limit,
		Offset: req.Offset,
	}

	teams, total, err := s.teamRepo.List(ctx, params)
	if err != nil {
		return nil, 0, xerrors.Wrap(err, xerrors.CodeInternalError, "查询团队列表失败")
	}

	return teams, total, nil
}

// GetTeamByID 查询团队详情
func (s *TeamAdminService) GetTeamByID(ctx context.Context, teamID string) (*game_runtime.Team, error) {
	if teamID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "团队ID不能为空")
	}

	team, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "团队不存在")
	}

	return team, nil
}

// GetTeamMembers 查询团队成员列表
func (s *TeamAdminService) GetTeamMembers(ctx context.Context, teamID string) ([]*game_runtime.TeamMember, error) {
	if teamID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "团队ID不能为空")
	}

	// 检查团队是否存在
	_, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "团队不存在")
	}

	// 查询成员列表
	members, err := s.teamMemberRepo.ListByTeam(ctx, teamID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询团队成员失败")
	}

	return members, nil
}

// DisbandTeam 强制解散团队（管理员操作）
func (s *TeamAdminService) DisbandTeam(ctx context.Context, teamID string) error {
	if teamID == "" {
		return xerrors.New(xerrors.CodeInvalidParams, "团队ID不能为空")
	}

	// 检查团队是否存在
	_, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "团队不存在")
	}

	// 软删除团队（不检查仓库是否为空）
	if err := s.teamRepo.Delete(ctx, teamID); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "解散团队失败")
	}

	// TODO: 删除 Keto 权限关系
	// TODO: 通知所有成员

	return nil
}

// GetTeamStatistics 获取团队统计信息
func (s *TeamAdminService) GetTeamStatistics(ctx context.Context, teamID string) (*TeamStatistics, error) {
	if teamID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "团队ID不能为空")
	}

	// 检查团队是否存在
	team, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "团队不存在")
	}

	// 统计成员数量
	memberCount, err := s.teamMemberRepo.CountByTeam(ctx, teamID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "统计成员数量失败")
	}

	// 统计各角色数量
	members, err := s.teamMemberRepo.ListByTeam(ctx, teamID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询团队成员失败")
	}

	var leaderCount, adminCount, memberCountByRole int
	for _, member := range members {
		switch member.Role {
		case "leader":
			leaderCount++
		case "admin":
			adminCount++
		case "member":
			memberCountByRole++
		}
	}

	stats := &TeamStatistics{
		TeamID:       teamID,
		TeamName:     team.Name,
		TotalMembers: int(memberCount),
		LeaderCount:  leaderCount,
		AdminCount:   adminCount,
		MemberCount:  memberCountByRole,
		MaxMembers:   team.MaxMembers,
		CreatedAt:    team.CreatedAt,
	}

	return stats, nil
}

// TeamStatistics 团队统计信息
type TeamStatistics struct {
	TeamID       string
	TeamName     string
	TotalMembers int
	LeaderCount  int
	AdminCount   int
	MemberCount  int
	MaxMembers   int
	CreatedAt    time.Time
}
