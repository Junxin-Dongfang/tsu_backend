package impl

import (
	"context"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"

	"tsu-self/internal/repository/interfaces"
)

type teamWarehouseLootLogRepositoryImpl struct {
	db boil.ContextExecutor
}

// NewTeamWarehouseLootLogRepository 创建入库审计仓储
func NewTeamWarehouseLootLogRepository(db boil.ContextExecutor) interfaces.TeamWarehouseLootLogRepository {
	return &teamWarehouseLootLogRepositoryImpl{db: db}
}

func (r *teamWarehouseLootLogRepositoryImpl) Log(ctx context.Context, execer boil.ContextExecutor, req *interfaces.TeamWarehouseLootLog) error {
	if req == nil {
		return fmt.Errorf("请求不能为空")
	}
	if execer == nil {
		execer = r.db
	}

	id := uuid.NewString()
	_, err := execer.ExecContext(ctx, `
INSERT INTO game_runtime.team_warehouse_loot_log
    (id, team_id, warehouse_id, source_dungeon_id, gold_amount, items, result, reason)
VALUES ($1, $2, $3, $4, $5, COALESCE($6, '[]'::jsonb), $7, $8)
`, id, req.TeamID, req.WarehouseID, req.SourceDungeonID, req.GoldAmount, req.ItemsJSON, req.Result, req.Reason)
	if err != nil {
		return fmt.Errorf("写入入库审计日志失败: %w", err)
	}
	return nil
}
