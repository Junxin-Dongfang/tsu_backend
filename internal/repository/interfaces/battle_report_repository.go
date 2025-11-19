package interfaces

import (
	"context"
	"encoding/json"
)

// BattleReport 描述战斗回调需要持久化的核心字段。
type BattleReport struct {
	BattleID     string          // 外部战斗引擎传入的唯一 ID
	BattleCode   string          // 战斗模板/房间编码
	TeamID       string          // 触发战斗的团队
	DungeonID    string          // 关联地城
	ResultStatus string          // 结果状态（victory/defeat/draw 等）
	LootGold     int64           // 奖励金币
	LootItems    json.RawMessage // 奖励物品 JSON
	Participants json.RawMessage // 参与者 JSON
	Events       json.RawMessage // 事件/日志 JSON
	RawPayload   json.RawMessage // 完整原始 payload
}

// BattleReportRepository 负责战斗回调的持久化。
type BattleReportRepository interface {
	Create(ctx context.Context, report *BattleReport) error
}
