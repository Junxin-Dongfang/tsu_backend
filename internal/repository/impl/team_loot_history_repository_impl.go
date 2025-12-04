package impl

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/repository/interfaces"
)

type teamLootHistoryRepositoryImpl struct {
	db boil.ContextExecutor
}

// NewTeamLootHistoryRepository 创建分配历史仓储实例
func NewTeamLootHistoryRepository(db boil.ContextExecutor) interfaces.TeamLootHistoryRepository {
	return &teamLootHistoryRepositoryImpl{db: db}
}

func (r *teamLootHistoryRepositoryImpl) CreateDistribution(ctx context.Context, execer boil.ContextExecutor, req *interfaces.TeamLootHistoryCreateReq) error {
	if req == nil {
		return fmt.Errorf("请求不能为空")
	}
	if execer == nil {
		execer = r.db
	}

	record := &game_runtime.TeamLootDistributionHistory{
		ID:                uuid.NewString(),
		TeamID:            req.TeamID,
		WarehouseID:       req.WarehouseID,
		DistributorHeroID: req.DistributorHeroID,
		RecipientHeroID:   req.RecipientHeroID,
		ItemType:          req.ItemType,
		Quantity:          req.Quantity,
	}
	if req.ItemID != nil {
		record.ItemID.SetValid(*req.ItemID)
	}

	if err := record.Insert(ctx, execer, boil.Infer()); err != nil {
		return fmt.Errorf("写入分配历史失败: %w", err)
	}
	return nil
}

func (r *teamLootHistoryRepositoryImpl) ListByTeam(ctx context.Context, teamID string, startAt, endAt *string, limit, offset int) ([]*interfaces.TeamLootHistoryRow, int64, error) {
	if teamID == "" {
		return nil, 0, fmt.Errorf("team_id 不能为空")
	}
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	args := []interface{}{teamID}
	where := "team_id = $1"
	argPos := 2
	if startAt != nil {
		where += fmt.Sprintf(" AND distributed_at >= $%d", argPos)
		args = append(args, *startAt)
		argPos++
	}
	if endAt != nil {
		where += fmt.Sprintf(" AND distributed_at <= $%d", argPos)
		args = append(args, *endAt)
		argPos++
	}

	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM game_runtime.team_loot_distribution_history WHERE %s`, where)
	var total int64
	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("统计分配历史失败: %w", err)
	}

	querySQL := fmt.Sprintf(`
SELECT id, team_id, warehouse_id, distributor_hero_id, recipient_hero_id, item_type, item_id, quantity, distributed_at
FROM game_runtime.team_loot_distribution_history
WHERE %s
ORDER BY distributed_at DESC
LIMIT %d OFFSET %d
`, where, limit, offset)

	rows, err := r.db.QueryContext(ctx, querySQL, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询分配历史失败: %w", err)
	}
	defer rows.Close()

	var result []*interfaces.TeamLootHistoryRow
	for rows.Next() {
		row := &interfaces.TeamLootHistoryRow{}
		var itemID sql.NullString
		var distributedAt time.Time
		if err := rows.Scan(&row.ID, &row.TeamID, &row.WarehouseID, &row.DistributorHeroID, &row.RecipientHeroID, &row.ItemType, &itemID, &row.Quantity, &distributedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描分配历史失败: %w", err)
		}
		if itemID.Valid {
			val := itemID.String
			row.ItemID = &val
		}
		row.DistributedAt = distributedAt.Format(time.RFC3339)
		result = append(result, row)
	}

	return result, total, nil
}
