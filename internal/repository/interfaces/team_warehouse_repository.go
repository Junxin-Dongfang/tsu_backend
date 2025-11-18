package interfaces

import (
	"context"

	"github.com/aarondl/sqlboiler/v4/boil"
	"tsu-self/internal/entity/game_runtime"
)

// TeamWarehouseRepository 团队仓库仓储接口
type TeamWarehouseRepository interface {
	// Create 创建团队仓库
	Create(ctx context.Context, execer boil.ContextExecutor, warehouse *game_runtime.TeamWarehouse) error

	// GetByTeamID 根据团队ID获取仓库
	GetByTeamID(ctx context.Context, teamID string) (*game_runtime.TeamWarehouse, error)

	// Update 更新仓库信息
	Update(ctx context.Context, execer boil.ContextExecutor, warehouse *game_runtime.TeamWarehouse) error

	// AddGold 增加金币
	AddGold(ctx context.Context, execer boil.ContextExecutor, warehouseID string, amount int64) error

	// DeductGold 扣除金币
	DeductGold(ctx context.Context, execer boil.ContextExecutor, warehouseID string, amount int64) error
}

