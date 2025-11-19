package impl

import (
	"context"
	"database/sql"
	"fmt"

	"tsu-self/internal/repository/interfaces"
)

type battleReportRepositoryImpl struct {
	db *sql.DB
}

// NewBattleReportRepository 创建 BattleReport 仓储实例。
func NewBattleReportRepository(db *sql.DB) interfaces.BattleReportRepository {
	return &battleReportRepositoryImpl{db: db}
}

func (r *battleReportRepositoryImpl) Create(ctx context.Context, report *interfaces.BattleReport) error {
	if report == nil {
		return fmt.Errorf("battle report is nil")
	}

	query := `
		INSERT INTO game_runtime.battle_reports (
			battle_id, battle_code, team_id, dungeon_id, result_status,
			loot_gold, loot_items, participants, events, raw_payload
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (battle_id) DO UPDATE SET
			battle_code   = EXCLUDED.battle_code,
			team_id       = EXCLUDED.team_id,
			dungeon_id    = EXCLUDED.dungeon_id,
			result_status = EXCLUDED.result_status,
			loot_gold     = EXCLUDED.loot_gold,
			loot_items    = EXCLUDED.loot_items,
			participants  = EXCLUDED.participants,
			events        = EXCLUDED.events,
			raw_payload   = EXCLUDED.raw_payload,
			updated_at    = NOW()
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		report.BattleID,
		nullString(report.BattleCode),
		nullString(report.TeamID),
		nullString(report.DungeonID),
		report.ResultStatus,
		report.LootGold,
		nullJSON(report.LootItems),
		nullJSON(report.Participants),
		nullJSON(report.Events),
		nullJSON(report.RawPayload),
	)
	if err != nil {
		return fmt.Errorf("插入战斗回调记录失败: %w", err)
	}
	return nil
}

func nullString(val string) interface{} {
	if val == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: val, Valid: true}
}

func nullJSON(raw []byte) interface{} {
	if len(raw) == 0 {
		return nil
	}
	return raw
}
