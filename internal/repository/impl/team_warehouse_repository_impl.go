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

type teamWarehouseRepositoryImpl struct {
	db *sql.DB
}

// NewTeamWarehouseRepository 创建团队仓库仓储实例
func NewTeamWarehouseRepository(db *sql.DB) interfaces.TeamWarehouseRepository {
	return &teamWarehouseRepositoryImpl{db: db}
}

// Create 创建团队仓库
func (r *teamWarehouseRepositoryImpl) Create(ctx context.Context, execer boil.ContextExecutor, warehouse *game_runtime.TeamWarehouse) error {
	// 生成UUID
	if warehouse.ID == "" {
		warehouse.ID = uuid.New().String()
	}

	// 设置时间戳
	now := time.Now()
	warehouse.CreatedAt = now
	warehouse.UpdatedAt = now

	// 插入数据库
	if err := warehouse.Insert(ctx, execer, boil.Infer()); err != nil {
		return fmt.Errorf("创建团队仓库失败: %w", err)
	}

	return nil
}

// GetByTeamID 根据团队ID获取仓库
func (r *teamWarehouseRepositoryImpl) GetByTeamID(ctx context.Context, teamID string) (*game_runtime.TeamWarehouse, error) {
	warehouse, err := game_runtime.TeamWarehouses(
		qm.Where("team_id = ?", teamID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("团队仓库不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询团队仓库失败: %w", err)
	}

	return warehouse, nil
}

// Update 更新仓库信息
func (r *teamWarehouseRepositoryImpl) Update(ctx context.Context, execer boil.ContextExecutor, warehouse *game_runtime.TeamWarehouse) error {
	// 更新时间戳
	warehouse.UpdatedAt = time.Now()

	if _, err := warehouse.Update(ctx, execer, boil.Infer()); err != nil {
		return fmt.Errorf("更新团队仓库失败: %w", err)
	}
	return nil
}

// AddGold 增加金币
func (r *teamWarehouseRepositoryImpl) AddGold(ctx context.Context, execer boil.ContextExecutor, warehouseID string, amount int64) error {
	// 使用原生 SQL 更新，避免并发问题
	query := `UPDATE game_runtime.team_warehouses SET gold_amount = gold_amount + $1, updated_at = NOW() WHERE id = $2`
	_, err := execer.ExecContext(ctx, query, amount, warehouseID)
	if err != nil {
		return fmt.Errorf("增加金币失败: %w", err)
	}
	return nil
}

// DeductGold 扣除金币
func (r *teamWarehouseRepositoryImpl) DeductGold(ctx context.Context, execer boil.ContextExecutor, warehouseID string, amount int64) error {
	// 使用原生 SQL 更新，并检查余额
	query := `UPDATE game_runtime.team_warehouses SET gold_amount = gold_amount - $1, updated_at = NOW() WHERE id = $2 AND gold_amount >= $1`
	result, err := execer.ExecContext(ctx, query, amount, warehouseID)
	if err != nil {
		return fmt.Errorf("扣除金币失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("金币余额不足")
	}

	return nil
}

