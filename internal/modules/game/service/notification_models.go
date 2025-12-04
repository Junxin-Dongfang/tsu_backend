package service

// DistributionEvent 分配通知事件
type DistributionEvent struct {
	TeamID      string            `json:"team_id"`
	WarehouseID string            `json:"warehouse_id"`
	Distributor string            `json:"distributor_hero_id"`
	Recipients  map[string]int64  `json:"recipients_gold,omitempty"` // heroID -> amount
	ItemPayload map[string]map[string]int `json:"recipients_items,omitempty"` // heroID -> itemID -> qty
	Result      string            `json:"result"` // success|failed
	Reason      string            `json:"reason,omitempty"`
}

// LootEvent 入库通知事件
type LootEvent struct {
	TeamID      string                   `json:"team_id"`
	WarehouseID string                   `json:"warehouse_id"`
	SourceDungeonID string               `json:"source_dungeon_id,omitempty"`
	Gold        int64                    `json:"gold"`
	Items       []LootItem               `json:"items"`
	Result      string                   `json:"result"` // success|failed
	Reason      string                   `json:"reason,omitempty"`
}
