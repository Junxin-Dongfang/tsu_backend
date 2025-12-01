package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// TeamMemberService 团队成员服务
type TeamMemberService struct {
	db                    *sql.DB
	teamRepo              interfaces.TeamRepository
	teamMemberRepo        interfaces.TeamMemberRepository
	teamJoinRequestRepo   interfaces.TeamJoinRequestRepository
	teamInvitationRepo    interfaces.TeamInvitationRepository
	teamKickedRecordRepo  interfaces.TeamKickedRecordRepository
	heroRepo              interfaces.HeroRepository
	teamPermissionService *TeamPermissionService
}

// NewTeamMemberService 创建团队成员服务
func NewTeamMemberService(db *sql.DB, teamPermissionService *TeamPermissionService) *TeamMemberService {
	return &TeamMemberService{
		db:                    db,
		teamRepo:              impl.NewTeamRepository(db),
		teamMemberRepo:        impl.NewTeamMemberRepository(db),
		teamJoinRequestRepo:   impl.NewTeamJoinRequestRepository(db),
		teamInvitationRepo:    impl.NewTeamInvitationRepository(db),
		teamKickedRecordRepo:  impl.NewTeamKickedRecordRepository(db),
		heroRepo:              impl.NewHeroRepository(db),
		teamPermissionService: teamPermissionService,
	}
}

// ApplyToJoinRequest 申请加入团队请求
type ApplyToJoinRequest struct {
	TeamID  string
	HeroID  string
	UserID  string
	Message string
}

// ApplyToJoin 申请加入团队
func (s *TeamMemberService) ApplyToJoin(ctx context.Context, req *ApplyToJoinRequest) (string, error) {
	// 1. 验证参数
	if req.TeamID == "" || req.HeroID == "" || req.UserID == "" {
		return "", xerrors.New(xerrors.CodeInvalidParams, "参数不能为空")
	}

	// 2. 检查团队是否存在
	team, err := s.teamRepo.GetByID(ctx, req.TeamID)
	if err != nil {
		return "", xerrors.Wrap(err, xerrors.CodeResourceNotFound, "团队不存在")
	}

	// 3. 检查是否已是成员
	existingMember, _ := s.teamMemberRepo.GetByTeamAndHero(ctx, req.TeamID, req.HeroID)
	if existingMember != nil {
		return "", xerrors.New(xerrors.CodeDuplicateResource, "您已是该团队成员")
	}

	// 4. 检查团队是否满员
	memberCount, err := s.teamMemberRepo.CountByTeam(ctx, req.TeamID)
	if err != nil {
		return "", xerrors.Wrap(err, xerrors.CodeInternalError, "统计成员数量失败")
	}
	if memberCount >= int64(team.MaxMembers) {
		return "", xerrors.New(xerrors.CodeInvalidParams, "团队已满员")
	}

	// 5. 检查是否在冷却期
	inCooldown, err := s.teamKickedRecordRepo.CheckCooldown(ctx, req.TeamID, req.HeroID)
	if err != nil {
		return "", xerrors.Wrap(err, xerrors.CodeInternalError, "检查冷却期失败")
	}
	if inCooldown {
		return "", xerrors.New(xerrors.CodeInvalidParams, "您在24小时内不能重新加入该团队")
	}

	// 6. 检查是否已有待审批的申请
	existingRequest, _ := s.teamJoinRequestRepo.GetPendingByHeroAndTeam(ctx, req.HeroID, req.TeamID)
	if existingRequest != nil {
		return "", xerrors.New(xerrors.CodeDuplicateResource, "您已有待审批的申请")
	}

	// 7. 创建申请记录
	joinRequest := &game_runtime.TeamJoinRequest{
		TeamID: req.TeamID,
		HeroID: req.HeroID,
		UserID: req.UserID,
		Status: "pending",
	}
	if req.Message != "" {
		joinRequest.Message.SetValid(req.Message)
	}

	if err := s.teamJoinRequestRepo.Create(ctx, joinRequest); err != nil {
		return "", xerrors.Wrap(err, xerrors.CodeInternalError, "创建申请失败")
	}

	// TODO: 通知队长和管理员

	return joinRequest.ID, nil
}

// ApproveJoinRequestRequest 审批加入申请请求
type ApproveJoinRequestRequest struct {
	RequestID string
	HeroID    string // 审批人英雄ID
	Approved  bool   // true=批准, false=拒绝
}

// ApproveJoinRequest 审批加入申请
func (s *TeamMemberService) ApproveJoinRequest(ctx context.Context, req *ApproveJoinRequestRequest) error {
	// 1. 获取申请记录
	joinRequest, err := s.teamJoinRequestRepo.GetByID(ctx, req.RequestID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "申请不存在")
	}

	// 2. 检查申请状态
	if joinRequest.Status != "pending" {
		return xerrors.New(xerrors.CodeInvalidParams, "申请已被处理")
	}

	// 3. 检查审批人权限（队长或管理员）
	member, err := s.teamMemberRepo.GetByTeamAndHero(ctx, joinRequest.TeamID, req.HeroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "您不是该团队成员")
	}
	if member.Role != "leader" && member.Role != "admin" {
		return xerrors.New(xerrors.CodePermissionDenied, "只有队长和管理员可以审批申请")
	}

	// 4. 更新申请状态
	if req.Approved {
		joinRequest.Status = "approved"
	} else {
		joinRequest.Status = "rejected"
	}
	joinRequest.ReviewedByHeroID.SetValid(req.HeroID)
	joinRequest.ReviewedAt.SetValid(time.Now())

	if err := s.teamJoinRequestRepo.Update(ctx, s.db, joinRequest); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "更新申请状态失败")
	}

	// 5. 如果批准，创建成员记录
	if req.Approved {
		newMember := &game_runtime.TeamMember{
			TeamID: joinRequest.TeamID,
			HeroID: joinRequest.HeroID,
			UserID: joinRequest.UserID,
			Role:   "member",
		}
		if err := s.teamMemberRepo.Create(ctx, s.db, newMember); err != nil {
			return xerrors.Wrap(err, xerrors.CodeInternalError, "创建成员记录失败")
		}

		// 同步权限到 Keto
		if s.teamPermissionService != nil {
			if err := s.teamPermissionService.SyncMemberToKeto(ctx, newMember); err != nil {
				fmt.Printf("Warning: Failed to sync member to Keto for team %s: %v\n", joinRequest.TeamID, err)
			}
		}
	}

	// TODO: 通知申请人

	return nil
}

// InviteMemberRequest 邀请成员请求
type InviteMemberRequest struct {
	TeamID        string
	InviterHeroID string // 邀请人英雄ID
	InviteeHeroID string // 被邀请人英雄ID
	Message       string
}

// InviteMember 邀请成员（需要队长/管理员审批），返回创建的邀请
func (s *TeamMemberService) InviteMember(ctx context.Context, req *InviteMemberRequest) (*game_runtime.TeamInvitation, error) {
	// 1. 验证参数
	if req.TeamID == "" || req.InviterHeroID == "" || req.InviteeHeroID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "参数不能为空")
	}

	// 2. 检查邀请人是否是团队成员
	if _, err := s.teamMemberRepo.GetByTeamAndHero(ctx, req.TeamID, req.InviterHeroID); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "您不是该团队成员")
	}

	// 3. 检查被邀请人是否已是成员
	existingMember, _ := s.teamMemberRepo.GetByTeamAndHero(ctx, req.TeamID, req.InviteeHeroID)
	if existingMember != nil {
		return nil, xerrors.New(xerrors.CodeDuplicateResource, "该英雄已是团队成员")
	}

	// 4. 检查团队是否满员
	team, err := s.teamRepo.GetByID(ctx, req.TeamID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "团队不存在")
	}
	memberCount, err := s.teamMemberRepo.CountByTeam(ctx, req.TeamID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "统计成员数量失败")
	}
	if memberCount >= int64(team.MaxMembers) {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "团队已满员")
	}

	// 5. 检查被邀请人是否在冷却期
	inCooldown, err := s.teamKickedRecordRepo.CheckCooldown(ctx, req.TeamID, req.InviteeHeroID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "检查冷却期失败")
	}
	if inCooldown {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "该英雄在24小时内不能重新加入该团队")
	}

	// 6. 创建邀请记录
	invitation := &game_runtime.TeamInvitation{
		TeamID:        req.TeamID,
		InviterHeroID: req.InviterHeroID,
		InviteeHeroID: req.InviteeHeroID,
		Status:        "pending_approval", // 等待队长/管理员审批
	}
	if req.Message != "" {
		invitation.Message.SetValid(req.Message)
	}

	if err := s.teamInvitationRepo.Create(ctx, invitation); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "创建邀请失败")
	}

	// TODO: 通知队长和管理员审批

	return invitation, nil
}

// ApproveInvitationRequest 审批邀请请求
type ApproveInvitationRequest struct {
	InvitationID string
	HeroID       string // 审批人英雄ID
	Approved     bool   // true=批准, false=拒绝
}

// ApproveInvitation 审批邀请（队长/管理员审批）
func (s *TeamMemberService) ApproveInvitation(ctx context.Context, req *ApproveInvitationRequest) error {
	// 1. 获取邀请记录
	invitation, err := s.teamInvitationRepo.GetByID(ctx, req.InvitationID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "邀请不存在")
	}

	// 2. 检查邀请状态
	if invitation.Status != "pending_approval" {
		return xerrors.New(xerrors.CodeInvalidParams, "邀请已被处理")
	}

	// 3. 检查审批人权限（队长或管理员）
	member, err := s.teamMemberRepo.GetByTeamAndHero(ctx, invitation.TeamID, req.HeroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "您不是该团队成员")
	}
	if member.Role != "leader" && member.Role != "admin" {
		return xerrors.New(xerrors.CodePermissionDenied, "只有队长和管理员可以审批邀请")
	}

	// 4. 更新邀请状态
	if req.Approved {
		invitation.Status = "pending_accept" // 等待被邀请人接受
	} else {
		invitation.Status = "rejected"
	}
	invitation.ApprovedByHeroID.SetValid(req.HeroID)
	invitation.ApprovedAt.SetValid(time.Now())

	if err := s.teamInvitationRepo.Update(ctx, s.db, invitation); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "更新邀请状态失败")
	}

	// TODO: 通知被邀请人或邀请人

	return nil
}

// AcceptInvitationRequest 接受邀请请求
type AcceptInvitationRequest struct {
	InvitationID string
	HeroID       string // 被邀请人英雄ID
	UserID       string // 被邀请人用户ID
}

// AcceptInvitation 接受邀请（被邀请人接受）
func (s *TeamMemberService) AcceptInvitation(ctx context.Context, req *AcceptInvitationRequest) error {
	// 1. 获取邀请记录
	invitation, err := s.teamInvitationRepo.GetByID(ctx, req.InvitationID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "邀请不存在")
	}

	// 2. 检查邀请状态
	if invitation.Status != "pending_accept" {
		return xerrors.New(xerrors.CodeInvalidParams, "邀请状态不正确")
	}

	// 3. 检查是否是被邀请人
	if invitation.InviteeHeroID != req.HeroID {
		return xerrors.New(xerrors.CodePermissionDenied, "您不是被邀请人")
	}

	// 4. 检查邀请是否过期
	if time.Now().After(invitation.ExpiresAt) {
		invitation.Status = "expired"
		s.teamInvitationRepo.Update(ctx, s.db, invitation)
		return xerrors.New(xerrors.CodeInvalidParams, "邀请已过期")
	}

	// 5. 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer tx.Rollback()

	// 6. 更新邀请状态
	invitation.Status = "accepted"
	invitation.RespondedAt.SetValid(time.Now())
	if err := s.teamInvitationRepo.Update(ctx, tx, invitation); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "更新邀请状态失败")
	}

	// 7. 创建成员记录
	newMember := &game_runtime.TeamMember{
		TeamID: invitation.TeamID,
		HeroID: req.HeroID,
		UserID: req.UserID,
		Role:   "member",
	}
	if err := s.teamMemberRepo.Create(ctx, tx, newMember); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "创建成员记录失败")
	}

	// 8. 提交事务
	if err := tx.Commit(); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	// 9. 同步权限到 Keto
	if s.teamPermissionService != nil {
		if err := s.teamPermissionService.SyncMemberToKeto(ctx, newMember); err != nil {
			fmt.Printf("Warning: Failed to sync member to Keto for team %s: %v\n", invitation.TeamID, err)
		}
	}

	// TODO: 通知团队成员

	return nil
}

// RejectInvitation 拒绝邀请
func (s *TeamMemberService) RejectInvitation(ctx context.Context, invitationID, heroID string) error {
	// 1. 获取邀请记录
	invitation, err := s.teamInvitationRepo.GetByID(ctx, invitationID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "邀请不存在")
	}

	// 2. 检查邀请状态
	if invitation.Status != "pending_accept" {
		return xerrors.New(xerrors.CodeInvalidParams, "邀请状态不正确")
	}

	// 3. 检查是否是被邀请人
	if invitation.InviteeHeroID != heroID {
		return xerrors.New(xerrors.CodePermissionDenied, "您不是被邀请人")
	}

	// 4. 更新邀请状态
	invitation.Status = "rejected"
	invitation.RespondedAt.SetValid(time.Now())
	if err := s.teamInvitationRepo.Update(ctx, s.db, invitation); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "更新邀请状态失败")
	}

	// TODO: 通知邀请人

	return nil
}

// KickMemberRequest 踢出成员请求
type KickMemberRequest struct {
	TeamID       string
	TargetHeroID string // 被踢出的英雄ID
	KickerHeroID string // 踢出者英雄ID
	Reason       string
}

// KickMember 踢出成员
func (s *TeamMemberService) KickMember(ctx context.Context, req *KickMemberRequest) error {
	// 1. 验证参数
	if req.TeamID == "" || req.TargetHeroID == "" || req.KickerHeroID == "" {
		return xerrors.New(xerrors.CodeInvalidParams, "参数不能为空")
	}

	// 2. 获取踢出者成员记录
	kicker, err := s.teamMemberRepo.GetByTeamAndHero(ctx, req.TeamID, req.KickerHeroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "您不是该团队成员")
	}

	// 3. 获取目标成员记录
	target, err := s.teamMemberRepo.GetByTeamAndHero(ctx, req.TeamID, req.TargetHeroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "目标成员不存在")
	}

	// 4. 检查权限
	if kicker.Role == "member" {
		return xerrors.New(xerrors.CodePermissionDenied, "普通成员不能踢出其他成员")
	}
	if kicker.Role == "admin" && (target.Role == "leader" || target.Role == "admin") {
		return xerrors.New(xerrors.CodePermissionDenied, "管理员不能踢出队长和其他管理员")
	}

	// 5. 不能踢出自己
	if req.TargetHeroID == req.KickerHeroID {
		return xerrors.New(xerrors.CodeInvalidParams, "不能踢出自己")
	}

	// 6. 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer tx.Rollback()

	// 7. 删除成员记录
	if err := s.teamMemberRepo.Delete(ctx, tx, target.ID); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "踢出成员失败")
	}

	// 8. 创建踢出记录（用于冷却期）
	kickedRecord := &game_runtime.TeamKickedRecord{
		TeamID:         req.TeamID,
		HeroID:         req.TargetHeroID,
		KickedByHeroID: req.KickerHeroID,
	}
	if req.Reason != "" {
		kickedRecord.Reason.SetValid(req.Reason)
	}

	if err := s.teamKickedRecordRepo.Create(ctx, kickedRecord); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "创建踢出记录失败")
	}

	// 9. 提交事务
	if err := tx.Commit(); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	// 10. 删除 Keto 权限关系
	if s.teamPermissionService != nil {
		if err := s.teamPermissionService.DeleteMemberFromKeto(ctx, req.TeamID, req.TargetHeroID); err != nil {
			fmt.Printf("Warning: Failed to delete member from Keto for team %s: %v\n", req.TeamID, err)
		}
	}

	// TODO: 通知被踢出的成员

	return nil
}

// PromoteToAdminRequest 任命管理员请求
type PromoteToAdminRequest struct {
	TeamID       string
	TargetHeroID string // 被任命的英雄ID
	LeaderHeroID string // 队长英雄ID
}

// PromoteToAdmin 任命管理员
func (s *TeamMemberService) PromoteToAdmin(ctx context.Context, req *PromoteToAdminRequest) error {
	// 1. 验证参数
	if req.TeamID == "" || req.TargetHeroID == "" || req.LeaderHeroID == "" {
		return xerrors.New(xerrors.CodeInvalidParams, "参数不能为空")
	}

	// 2. 检查操作者是否是队长
	leader, err := s.teamMemberRepo.GetByTeamAndHero(ctx, req.TeamID, req.LeaderHeroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "您不是该团队成员")
	}
	if leader.Role != "leader" {
		return xerrors.New(xerrors.CodePermissionDenied, "只有队长可以任命管理员")
	}

	// 3. 获取目标成员记录
	target, err := s.teamMemberRepo.GetByTeamAndHero(ctx, req.TeamID, req.TargetHeroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "目标成员不存在")
	}

	// 4. 检查目标成员角色
	if target.Role != "member" {
		return xerrors.New(xerrors.CodeInvalidParams, "只能任命普通成员为管理员")
	}

	// 5. 更新角色
	if err := s.teamMemberRepo.UpdateRole(ctx, s.db, req.TeamID, req.TargetHeroID, "admin"); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "任命管理员失败")
	}

	// 6. 同步权限到 Keto
	if s.teamPermissionService != nil {
		if err := s.teamPermissionService.UpdateMemberRoleInKeto(ctx, req.TeamID, req.TargetHeroID, "member", "admin"); err != nil {
			fmt.Printf("Warning: Failed to update member role to admin in Keto for team %s: %v\n", req.TeamID, err)
		}
	}

	// TODO: 通知被任命的成员

	return nil
}

// DemoteAdmin 撤销管理员
func (s *TeamMemberService) DemoteAdmin(ctx context.Context, teamID, targetHeroID, leaderHeroID string) error {
	// 1. 验证参数
	if teamID == "" || targetHeroID == "" || leaderHeroID == "" {
		return xerrors.New(xerrors.CodeInvalidParams, "参数不能为空")
	}

	// 2. 检查操作者是否是队长
	leader, err := s.teamMemberRepo.GetByTeamAndHero(ctx, teamID, leaderHeroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "您不是该团队成员")
	}
	if leader.Role != "leader" {
		return xerrors.New(xerrors.CodePermissionDenied, "只有队长可以撤销管理员")
	}

	// 3. 获取目标成员记录
	target, err := s.teamMemberRepo.GetByTeamAndHero(ctx, teamID, targetHeroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "目标成员不存在")
	}

	// 4. 检查目标成员角色
	if target.Role != "admin" {
		return xerrors.New(xerrors.CodeInvalidParams, "目标成员不是管理员")
	}

	// 5. 更新角色
	if err := s.teamMemberRepo.UpdateRole(ctx, s.db, teamID, targetHeroID, "member"); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "撤销管理员失败")
	}

	// 6. 同步权限到 Keto
	if s.teamPermissionService != nil {
		if err := s.teamPermissionService.UpdateMemberRoleInKeto(ctx, teamID, targetHeroID, "admin", "member"); err != nil {
			fmt.Printf("Warning: Failed to update member role from admin to member in Keto for team %s: %v\n", teamID, err)
		}
	}

	// TODO: 通知被撤销的成员

	return nil
}
