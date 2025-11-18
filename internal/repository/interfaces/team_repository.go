package interfaces

import (
	"context"
	"time"

	"tsu-self/internal/entity/game_runtime"
)

// TeamQueryParams 团队查询参数
type TeamQueryParams struct {
	Name   string // 团队名称（模糊查询）
	Limit  int    // 每页数量
	Offset int    // 偏移量
}

// TeamRepository 团队仓储接口
type TeamRepository interface {
	// Create 创建团队
	Create(ctx context.Context, team *game_runtime.Team) error

	// GetByID 根据ID获取团队
	GetByID(ctx context.Context, teamID string) (*game_runtime.Team, error)

	// Update 更新团队信息
	Update(ctx context.Context, team *game_runtime.Team) error

	// Delete 软删除团队
	Delete(ctx context.Context, teamID string) error

	// GetInactiveLeaderTeams 查询队长超过指定时间未活跃的团队
	// inactiveDuration: 未活跃时长（如 7*24*time.Hour）
	GetInactiveLeaderTeams(ctx context.Context, inactiveDuration time.Duration) ([]*game_runtime.Team, error)

	// List 分页查询团队列表
	List(ctx context.Context, params TeamQueryParams) ([]*game_runtime.Team, int64, error)

	// Exists 检查团队名称是否存在
	Exists(ctx context.Context, name string) (bool, error)
}

