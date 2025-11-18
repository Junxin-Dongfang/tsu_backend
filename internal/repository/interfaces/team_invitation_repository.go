package interfaces

import (
	"context"

	"github.com/aarondl/sqlboiler/v4/boil"
	"tsu-self/internal/entity/game_runtime"
)

// TeamInvitationRepository 团队邀请仓储接口
type TeamInvitationRepository interface {
	// Create 创建邀请
	Create(ctx context.Context, invitation *game_runtime.TeamInvitation) error

	// GetByID 根据ID获取邀请
	GetByID(ctx context.Context, invitationID string) (*game_runtime.TeamInvitation, error)

	// Update 更新邀请
	Update(ctx context.Context, execer boil.ContextExecutor, invitation *game_runtime.TeamInvitation) error

	// ListPendingApprovalByTeam 查询团队的待审批邀请列表
	ListPendingApprovalByTeam(ctx context.Context, teamID string) ([]*game_runtime.TeamInvitation, error)

	// ListPendingAcceptByHero 查询被邀请人的待接受邀请列表
	ListPendingAcceptByHero(ctx context.Context, heroID string) ([]*game_runtime.TeamInvitation, error)

	// ExpireInvitations 过期邀请处理（将过期的邀请状态更新为 expired）
	ExpireInvitations(ctx context.Context) (int64, error)
}

