package interfaces

import (
	"context"

	"github.com/aarondl/sqlboiler/v4/boil"

	"tsu-self/internal/entity/game_runtime"
)

// TeamDungeonProgressRepository 团队地城进度仓储接口
type TeamDungeonProgressRepository interface {
	Create(ctx context.Context, execer boil.ContextExecutor, progress *game_runtime.TeamDungeonProgress) error
	Update(ctx context.Context, execer boil.ContextExecutor, progress *game_runtime.TeamDungeonProgress, columns ...string) error

	GetByID(ctx context.Context, progressID string) (*game_runtime.TeamDungeonProgress, error)
	GetActiveByTeam(ctx context.Context, teamID string) (*game_runtime.TeamDungeonProgress, error)
	GetActiveByTeamForUpdate(ctx context.Context, execer boil.ContextExecutor, teamID string) (*game_runtime.TeamDungeonProgress, error)
	GetByTeamAndDungeon(ctx context.Context, teamID, dungeonID string) (*game_runtime.TeamDungeonProgress, error)
}
