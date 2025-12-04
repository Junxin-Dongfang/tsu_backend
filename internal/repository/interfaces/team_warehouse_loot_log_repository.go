package interfaces

import (
	"context"

	"github.com/aarondl/sqlboiler/v4/boil"
)

// TeamWarehouseLootLogRepository 战利品入库审计日志仓储
type TeamWarehouseLootLogRepository interface {
	Log(ctx context.Context, execer boil.ContextExecutor, req *TeamWarehouseLootLog) error
}

// TeamWarehouseLootLog 入库日志记录
type TeamWarehouseLootLog struct {
	TeamID         string
	WarehouseID    string
	SourceDungeonID *string
	GoldAmount     int64
	ItemsJSON      string
	Result         string // success|failed
	Reason         *string
}
