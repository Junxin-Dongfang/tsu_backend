package interfaces

import (
	"context"

	"github.com/aarondl/sqlboiler/v4/boil"
)

// TeamLootHistoryRepository 分配历史仓储
type TeamLootHistoryRepository interface {
	// CreateDistribution 记录一次分配
	CreateDistribution(ctx context.Context, execer boil.ContextExecutor, req *TeamLootHistoryCreateReq) error
	// ListByTeam 分页查询分配历史
	ListByTeam(ctx context.Context, teamID string, startAt, endAt *string, limit, offset int) ([]*TeamLootHistoryRow, int64, error)
}

// TeamLootHistoryCreateReq 创建分配记录请求
type TeamLootHistoryCreateReq struct {
	TeamID           string
	WarehouseID      string
	DistributorHeroID string
	RecipientHeroID  string
	ItemType         string // gold | item
	ItemID           *string
	Quantity         int64
}

// TeamLootHistoryRow 查询结果
type TeamLootHistoryRow struct {
	ID               string
	TeamID           string
	WarehouseID      string
	DistributorHeroID string
	RecipientHeroID  string
	ItemType         string
	ItemID           *string
	Quantity         int64
	DistributedAt    string
}
