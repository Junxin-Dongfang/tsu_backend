package interfaces

import (
	"context"

	"github.com/aarondl/sqlboiler/v4/boil"

	"tsu-self/internal/entity/game_runtime"
)

// TeamDungeonRecordRepository 团队地城记录仓储接口
type TeamDungeonRecordRepository interface {
	Create(ctx context.Context, execer boil.ContextExecutor, record *game_runtime.TeamDungeonRecord) error
	Update(ctx context.Context, execer boil.ContextExecutor, record *game_runtime.TeamDungeonRecord, columns ...string) error

	GetByTeamAndDungeon(ctx context.Context, teamID, dungeonID string) (*game_runtime.TeamDungeonRecord, error)
	GetByTeamAndDungeonForUpdate(ctx context.Context, execer boil.ContextExecutor, teamID, dungeonID string) (*game_runtime.TeamDungeonRecord, error)

	ListByTeam(ctx context.Context, teamID string, limit, offset int) ([]*game_runtime.TeamDungeonRecord, int64, error)
}
