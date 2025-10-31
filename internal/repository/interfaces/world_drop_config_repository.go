package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// WorldDropConfigRepository 世界掉落配置仓储接口
type WorldDropConfigRepository interface {
	// GetByID 根据ID获取世界掉落配置
	GetByID(ctx context.Context, configID string) (*game_config.WorldDropConfig, error)

	// GetByItemID 根据物品ID获取世界掉落配置
	GetByItemID(ctx context.Context, itemID string) (*game_config.WorldDropConfig, error)

	// GetActiveConfigs 获取所有激活的世界掉落配置
	GetActiveConfigs(ctx context.Context) ([]*game_config.WorldDropConfig, error)

	// GetConfigsByTriggerCondition 根据触发条件获取世界掉落配置
	// 例如: 查询等级范围内的世界掉落配置
	GetConfigsByTriggerCondition(ctx context.Context, conditionType string, params map[string]interface{}) ([]*game_config.WorldDropConfig, error)

	// List 查询世界掉落配置列表
	List(ctx context.Context, params ListWorldDropConfigParams) ([]*game_config.WorldDropConfig, int64, error)

	// Create 创建世界掉落配置
	Create(ctx context.Context, config *game_config.WorldDropConfig) error

	// Update 更新世界掉落配置
	Update(ctx context.Context, config *game_config.WorldDropConfig) error

	// Delete 删除世界掉落配置（软删除）
	Delete(ctx context.Context, configID string) error

	// Count 统计世界掉落配置数量
	Count(ctx context.Context, params ListWorldDropConfigParams) (int64, error)
}

// ListWorldDropConfigParams 查询世界掉落配置列表参数
type ListWorldDropConfigParams struct {
	ItemID    *string // 物品ID筛选
	IsActive  *bool   // 是否启用
	SortBy    string  // 排序字段
	SortOrder string  // 排序方向
	Page      int     // 页码(从1开始)
	PageSize  int     // 每页数量
}

