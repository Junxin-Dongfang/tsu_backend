package impl

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/google/uuid"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/repository/interfaces"
)

type teamDungeonProgressRepositoryImpl struct {
	db *sql.DB
}

// NewTeamDungeonProgressRepository 创建团队地城进度仓储实例
func NewTeamDungeonProgressRepository(db *sql.DB) interfaces.TeamDungeonProgressRepository {
	return &teamDungeonProgressRepositoryImpl{db: db}
}

func (r *teamDungeonProgressRepositoryImpl) ensureDefaults(progress *game_runtime.TeamDungeonProgress) {
	if progress.ID == "" {
		progress.ID = uuid.New().String()
	}
	if len(progress.CompletedRooms) == 0 {
		progress.CompletedRooms = types.JSON([]byte("[]"))
	}
	if progress.Status == "" {
		progress.Status = "in_progress"
	}
	if progress.StartedAt.IsZero() {
		progress.StartedAt = time.Now()
	}
}

func (r *teamDungeonProgressRepositoryImpl) Create(ctx context.Context, execer boil.ContextExecutor, progress *game_runtime.TeamDungeonProgress) error {
	r.ensureDefaults(progress)
	progress.UpdatedAt = time.Now()

	if err := progress.Insert(ctx, execer, boil.Infer()); err != nil {
		return fmt.Errorf("创建地城进度失败: %w", err)
	}
	return nil
}

func (r *teamDungeonProgressRepositoryImpl) Update(ctx context.Context, execer boil.ContextExecutor, progress *game_runtime.TeamDungeonProgress, columns ...string) error {
	progress.UpdatedAt = time.Now()
	if len(columns) == 0 {
		if _, err := progress.Update(ctx, execer, boil.Infer()); err != nil {
			return fmt.Errorf("更新地城进度失败: %w", err)
		}
		return nil
	}

	cols := append(columns, "updated_at")
	if _, err := progress.Update(ctx, execer, boil.Whitelist(cols...)); err != nil {
		return fmt.Errorf("更新地城进度失败: %w", err)
	}
	return nil
}

func (r *teamDungeonProgressRepositoryImpl) GetByID(ctx context.Context, progressID string) (*game_runtime.TeamDungeonProgress, error) {
	progress, err := game_runtime.TeamDungeonProgresses(
		qm.Where("id = ?", progressID),
	).One(ctx, r.db)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("地城进度不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询地城进度失败: %w", err)
	}
	return progress, nil
}

func (r *teamDungeonProgressRepositoryImpl) getByTeamWithExecutor(ctx context.Context, execer boil.ContextExecutor, teamID string, forUpdate bool) (*game_runtime.TeamDungeonProgress, error) {
	mods := []qm.QueryMod{
		qm.Where("team_id = ? AND status = 'in_progress'", teamID),
		qm.OrderBy("started_at DESC"),
		qm.Limit(1),
	}
	if forUpdate {
		mods = append(mods, qm.For("UPDATE"))
	}

	query := game_runtime.TeamDungeonProgresses(mods...)
	var progress *game_runtime.TeamDungeonProgress
	var err error
	if execer == nil {
		progress, err = query.One(ctx, r.db)
	} else {
		progress, err = query.One(ctx, execer)
	}
	if err == sql.ErrNoRows {
		return nil, interfaces.ErrTeamDungeonProgressNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("查询地城进度失败: %w", err)
	}
	return progress, nil
}

func (r *teamDungeonProgressRepositoryImpl) GetActiveByTeam(ctx context.Context, teamID string) (*game_runtime.TeamDungeonProgress, error) {
	return r.getByTeamWithExecutor(ctx, nil, teamID, false)
}

func (r *teamDungeonProgressRepositoryImpl) GetActiveByTeamForUpdate(ctx context.Context, execer boil.ContextExecutor, teamID string) (*game_runtime.TeamDungeonProgress, error) {
	return r.getByTeamWithExecutor(ctx, execer, teamID, true)
}

func (r *teamDungeonProgressRepositoryImpl) GetByTeamAndDungeon(ctx context.Context, teamID, dungeonID string) (*game_runtime.TeamDungeonProgress, error) {
	progress, err := game_runtime.TeamDungeonProgresses(
		qm.Where("team_id = ? AND dungeon_id = ?", teamID, dungeonID),
		qm.OrderBy("started_at DESC"),
		qm.Limit(1),
	).One(ctx, r.db)
	if err == sql.ErrNoRows {
		return nil, interfaces.ErrTeamDungeonProgressNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("查询地城进度失败: %w", err)
	}
	return progress, nil
}
