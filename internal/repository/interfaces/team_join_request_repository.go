package interfaces

import (
	"context"

	"github.com/aarondl/sqlboiler/v4/boil"
	"tsu-self/internal/entity/game_runtime"
)

// TeamJoinRequestRepository 团队加入申请仓储接口
type TeamJoinRequestRepository interface {
	// Create 创建加入申请
	Create(ctx context.Context, request *game_runtime.TeamJoinRequest) error

	// GetByID 根据ID获取申请
	GetByID(ctx context.Context, requestID string) (*game_runtime.TeamJoinRequest, error)

	// Update 更新申请
	Update(ctx context.Context, execer boil.ContextExecutor, request *game_runtime.TeamJoinRequest) error

	// ListPendingByTeam 查询团队的待审批申请列表
	ListPendingByTeam(ctx context.Context, teamID string) ([]*game_runtime.TeamJoinRequest, error)

	// GetPendingByHeroAndTeam 查询英雄对团队的待审批申请
	GetPendingByHeroAndTeam(ctx context.Context, heroID, teamID string) (*game_runtime.TeamJoinRequest, error)

	// ListByHero 查询英雄的申请列表
	ListByHero(ctx context.Context, heroID string) ([]*game_runtime.TeamJoinRequest, error)
}

