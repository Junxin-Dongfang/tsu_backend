package impl

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/google/uuid"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/repository/interfaces"
)

type teamInvitationRepositoryImpl struct {
	db *sql.DB
}

// NewTeamInvitationRepository 创建团队邀请仓储实例
func NewTeamInvitationRepository(db *sql.DB) interfaces.TeamInvitationRepository {
	return &teamInvitationRepositoryImpl{db: db}
}

// Create 创建邀请
func (r *teamInvitationRepositoryImpl) Create(ctx context.Context, invitation *game_runtime.TeamInvitation) error {
	// 生成UUID
	if invitation.ID == "" {
		invitation.ID = uuid.New().String()
	}

	// 设置时间戳
	now := time.Now()
	invitation.CreatedAt = now
	if invitation.ExpiresAt.IsZero() {
		invitation.ExpiresAt = now.Add(7 * 24 * time.Hour) // 默认7天过期
	}

	// 插入数据库
	if err := invitation.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("创建邀请失败: %w", err)
	}

	return nil
}

// GetByID 根据ID获取邀请
func (r *teamInvitationRepositoryImpl) GetByID(ctx context.Context, invitationID string) (*game_runtime.TeamInvitation, error) {
	invitation, err := game_runtime.TeamInvitations(
		qm.Where("id = ?", invitationID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("邀请不存在: %s", invitationID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询邀请失败: %w", err)
	}

	return invitation, nil
}

// Update 更新邀请
func (r *teamInvitationRepositoryImpl) Update(ctx context.Context, execer boil.ContextExecutor, invitation *game_runtime.TeamInvitation) error {
	if _, err := invitation.Update(ctx, execer, boil.Infer()); err != nil {
		return fmt.Errorf("更新邀请失败: %w", err)
	}
	return nil
}

// ListPendingApprovalByTeam 查询团队的待审批邀请列表
func (r *teamInvitationRepositoryImpl) ListPendingApprovalByTeam(ctx context.Context, teamID string) ([]*game_runtime.TeamInvitation, error) {
	invitations, err := game_runtime.TeamInvitations(
		qm.Where("team_id = ? AND status = ?", teamID, "pending_approval"),
		qm.OrderBy("created_at DESC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询待审批邀请列表失败: %w", err)
	}

	return invitations, nil
}

// ListPendingAcceptByHero 查询被邀请人的待接受邀请列表
func (r *teamInvitationRepositoryImpl) ListPendingAcceptByHero(ctx context.Context, heroID string) ([]*game_runtime.TeamInvitation, error) {
	invitations, err := game_runtime.TeamInvitations(
		qm.Where("invitee_hero_id = ? AND status = ?", heroID, "pending_accept"),
		qm.OrderBy("created_at DESC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询待接受邀请列表失败: %w", err)
	}

	return invitations, nil
}

// ExpireInvitations 过期邀请处理
func (r *teamInvitationRepositoryImpl) ExpireInvitations(ctx context.Context) (int64, error) {
	// 更新所有过期的邀请状态为 expired
	result, err := game_runtime.TeamInvitations(
		qm.Where("status IN (?, ?) AND expires_at < ?", "pending_approval", "pending_accept", time.Now()),
	).UpdateAll(ctx, r.db, game_runtime.M{
		"status": "expired",
	})

	if err != nil {
		return 0, fmt.Errorf("过期邀请处理失败: %w", err)
	}

	return result, nil
}

