package interfaces

import (
	"context"
	"database/sql"
	"errors"

	"github.com/aarondl/sqlboiler/v4/boil"
	"tsu-self/internal/entity/game_runtime"
)

var ErrTeamMemberNotFound = errors.New("team member not found")

// TeamMemberRepository 团队成员仓储接口
type TeamMemberRepository interface {
	// Create 创建团队成员记录
	Create(ctx context.Context, execer boil.ContextExecutor, member *game_runtime.TeamMember) error

	// GetByID 根据ID获取成员记录
	GetByID(ctx context.Context, memberID string) (*game_runtime.TeamMember, error)

	// GetByTeamAndHero 根据团队ID和英雄ID获取成员记录
	GetByTeamAndHero(ctx context.Context, teamID, heroID string) (*game_runtime.TeamMember, error)

	// GetLeaderTeam 查询英雄担任队长的团队
	GetLeaderTeam(ctx context.Context, heroID string) (*game_runtime.TeamMember, error)

	// ListByTeam 查询团队成员列表
	ListByTeam(ctx context.Context, teamID string) ([]*game_runtime.TeamMember, error)

	// ListByHero 查询英雄加入的团队列表
	ListByHero(ctx context.Context, heroID string) ([]*game_runtime.TeamMember, error)

	// ListAll 查询所有团队成员（用于一致性任务）
	ListAll(ctx context.Context) ([]*game_runtime.TeamMember, error)

	// Update 更新成员信息
	Update(ctx context.Context, execer boil.ContextExecutor, member *game_runtime.TeamMember) error

	// Delete 删除成员记录
	Delete(ctx context.Context, execer boil.ContextExecutor, memberID string) error

	// UpdateRole 更新成员角色
	UpdateRole(ctx context.Context, execer boil.ContextExecutor, teamID, heroID, role string) error

	// UpdateLastActive 更新最后活跃时间
	UpdateLastActive(ctx context.Context, execer boil.ContextExecutor, teamID, heroID string) error

	// GetEarliestAdmin 查询最早加入的管理员
	GetEarliestAdmin(ctx context.Context, teamID string) (*game_runtime.TeamMember, error)

	// GetEarliestMember 查询最早加入的普通成员
	GetEarliestMember(ctx context.Context, teamID string) (*game_runtime.TeamMember, error)

	// CountByTeam 统计团队成员数量
	CountByTeam(ctx context.Context, teamID string) (int64, error)

	// GetByIDForUpdate 根据ID获取成员记录（带行锁）
	GetByIDForUpdate(ctx context.Context, tx *sql.Tx, memberID string) (*game_runtime.TeamMember, error)
}
