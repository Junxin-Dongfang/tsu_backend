package impl

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/queries/qm"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/interfaces"
)

type attributeUpgradeCostRepositoryImpl struct {
	db *sql.DB
}

// NewAttributeUpgradeCostRepository 创建属性升级消耗配置仓储实例
func NewAttributeUpgradeCostRepository(db *sql.DB) interfaces.AttributeUpgradeCostRepository {
	return &attributeUpgradeCostRepositoryImpl{db: db}
}

// GetByPointNumber 根据点数获取消耗配置
func (r *attributeUpgradeCostRepositoryImpl) GetByPointNumber(ctx context.Context, pointNumber int) (*game_config.AttributeUpgradeCost, error) {
	cost, err := game_config.AttributeUpgradeCosts(
		qm.Where("point_number = ?", pointNumber),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("属性加点消耗配置不存在: point_number=%d", pointNumber)
	}
	if err != nil {
		return nil, fmt.Errorf("查询属性加点消耗配置失败: %w", err)
	}

	return cost, nil
}

// GetBatchByPointNumbers 批量获取消耗配置
func (r *attributeUpgradeCostRepositoryImpl) GetBatchByPointNumbers(ctx context.Context, pointNumbers []int) ([]*game_config.AttributeUpgradeCost, error) {
	if len(pointNumbers) == 0 {
		return []*game_config.AttributeUpgradeCost{}, nil
	}

	costs, err := game_config.AttributeUpgradeCosts(
		qm.WhereIn("point_number IN ?", interfaceSlice(pointNumbers)...),
		qm.OrderBy("point_number ASC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("批量查询属性加点消耗配置失败: %w", err)
	}

	return costs, nil
}

// GetAll 获取所有配置
func (r *attributeUpgradeCostRepositoryImpl) GetAll(ctx context.Context) ([]*game_config.AttributeUpgradeCost, error) {
	costs, err := game_config.AttributeUpgradeCosts(
		qm.OrderBy("point_number ASC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询所有属性加点消耗配置失败: %w", err)
	}

	return costs, nil
}

// CalculateCost 计算从 fromPoint 到 toPoint 的总消耗
func (r *attributeUpgradeCostRepositoryImpl) CalculateCost(ctx context.Context, fromPoint, toPoint int) (int, error) {
	if fromPoint >= toPoint {
		return 0, nil
	}

	// 生成点数范围
	pointNumbers := make([]int, 0, toPoint-fromPoint)
	for i := fromPoint + 1; i <= toPoint; i++ {
		pointNumbers = append(pointNumbers, i)
	}

	costs, err := r.GetBatchByPointNumbers(ctx, pointNumbers)
	if err != nil {
		return 0, err
	}

	if len(costs) != len(pointNumbers) {
		return 0, fmt.Errorf("消耗配置不完整: 需要 %d 条记录，实际获取 %d 条", len(pointNumbers), len(costs))
	}

	totalCost := 0
	for _, cost := range costs {
		totalCost += cost.CostXP
	}

	return totalCost, nil
}

// interfaceSlice 辅助函数：将 int slice 转换为 interface{} slice
func interfaceSlice(nums []int) []interface{} {
	result := make([]interface{}, len(nums))
	for i, v := range nums {
		result[i] = v
	}
	return result
}

