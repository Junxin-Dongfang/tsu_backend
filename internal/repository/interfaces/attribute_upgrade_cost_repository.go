package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// AttributeUpgradeCostRepository 属性升级消耗配置仓储接口
type AttributeUpgradeCostRepository interface {
	// GetByPointNumber 根据点数获取消耗配置
	GetByPointNumber(ctx context.Context, pointNumber int) (*game_config.AttributeUpgradeCost, error)

	// GetBatchByPointNumbers 批量获取消耗配置
	GetBatchByPointNumbers(ctx context.Context, pointNumbers []int) ([]*game_config.AttributeUpgradeCost, error)

	// GetAll 获取所有配置
	GetAll(ctx context.Context) ([]*game_config.AttributeUpgradeCost, error)

	// CalculateCost 计算从 fromPoint 到 toPoint 的总消耗
	CalculateCost(ctx context.Context, fromPoint, toPoint int) (int, error)
}

