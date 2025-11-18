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

type teamDungeonRecordRepositoryImpl struct {
	db *sql.DB
}

// NewTeamDungeonRecordRepository 创建挑战记录仓储实例
func NewTeamDungeonRecordRepository(db *sql.DB) interfaces.TeamDungeonRecordRepository {
	return &teamDungeonRecordRepositoryImpl{db: db}
}

func (r *teamDungeonRecordRepositoryImpl) prepare(record *game_runtime.TeamDungeonRecord) {
	if record.ID == "" {
		record.ID = uuid.New().String()
	}
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now()
	}
	record.UpdatedAt = time.Now()
}

func (r *teamDungeonRecordRepositoryImpl) Create(ctx context.Context, execer boil.ContextExecutor, record *game_runtime.TeamDungeonRecord) error {
	r.prepare(record)
	if err := record.Insert(ctx, execer, boil.Infer()); err != nil {
		return fmt.Errorf("创建地城记录失败: %w", err)
	}
	return nil
}

func (r *teamDungeonRecordRepositoryImpl) Update(ctx context.Context, execer boil.ContextExecutor, record *game_runtime.TeamDungeonRecord, columns ...string) error {
	record.UpdatedAt = time.Now()
	if len(columns) == 0 {
		if _, err := record.Update(ctx, execer, boil.Infer()); err != nil {
			return fmt.Errorf("更新地城记录失败: %w", err)
		}
		return nil
	}
	cols := append(columns, "updated_at")
	if _, err := record.Update(ctx, execer, boil.Whitelist(cols...)); err != nil {
		return fmt.Errorf("更新地城记录失败: %w", err)
	}
	return nil
}

func (r *teamDungeonRecordRepositoryImpl) getByTeamAndDungeon(ctx context.Context, execer boil.ContextExecutor, teamID, dungeonID string, forUpdate bool) (*game_runtime.TeamDungeonRecord, error) {
	mods := []qm.QueryMod{
		qm.Where("team_id = ? AND dungeon_id = ?", teamID, dungeonID),
		qm.Limit(1),
	}
	if forUpdate {
		mods = append(mods, qm.For("UPDATE"))
	}
	query := game_runtime.TeamDungeonRecords(mods...)
	var (
		record *game_runtime.TeamDungeonRecord
		err    error
	)
	if execer == nil {
		record, err = query.One(ctx, r.db)
	} else {
		record, err = query.One(ctx, execer)
	}
	if err == sql.ErrNoRows {
		return nil, interfaces.ErrTeamDungeonRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("查询地城记录失败: %w", err)
	}
	return record, nil
}

func (r *teamDungeonRecordRepositoryImpl) GetByTeamAndDungeon(ctx context.Context, teamID, dungeonID string) (*game_runtime.TeamDungeonRecord, error) {
	return r.getByTeamAndDungeon(ctx, nil, teamID, dungeonID, false)
}

func (r *teamDungeonRecordRepositoryImpl) GetByTeamAndDungeonForUpdate(ctx context.Context, execer boil.ContextExecutor, teamID, dungeonID string) (*game_runtime.TeamDungeonRecord, error) {
	return r.getByTeamAndDungeon(ctx, execer, teamID, dungeonID, true)
}

func (r *teamDungeonRecordRepositoryImpl) ListByTeam(ctx context.Context, teamID string, limit, offset int) ([]*game_runtime.TeamDungeonRecord, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	mods := []qm.QueryMod{
		qm.Where("team_id = ?", teamID),
		qm.OrderBy("updated_at DESC"),
		qm.Limit(limit),
		qm.Offset(offset),
	}

	list, err := game_runtime.TeamDungeonRecords(mods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询地城记录失败: %w", err)
	}

	total, err := game_runtime.TeamDungeonRecords(
		qm.Where("team_id = ?", teamID),
		qm.Select("COUNT(*)"),
	).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("统计地城记录失败: %w", err)
	}

	return list, total, nil
}
