package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_runtime"
)

// TeamKickedRecordRepository 团队踢出记录仓储接口
type TeamKickedRecordRepository interface {
	// Create 创建踢出记录
	Create(ctx context.Context, record *game_runtime.TeamKickedRecord) error

	// CheckCooldown 检查冷却期（返回是否在冷却期内）
	CheckCooldown(ctx context.Context, teamID, heroID string) (bool, error)

	// GetLatestByTeamAndHero 获取最新的踢出记录
	GetLatestByTeamAndHero(ctx context.Context, teamID, heroID string) (*game_runtime.TeamKickedRecord, error)
}

