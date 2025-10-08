package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// BuffQueryParams Buff查询参数
type BuffQueryParams struct {
	BuffType *string // Buff类型
	Category *string // 分类
	IsActive *bool   // 是否启用
	Limit    int     // 每页数量
	Offset   int     // 偏移量
}

// BuffRepository Buff仓储接口
type BuffRepository interface {
	// GetByID 根据ID获取Buff
	GetByID(ctx context.Context, buffID string) (*game_config.Buff, error)

	// GetByCode 根据代码获取Buff
	GetByCode(ctx context.Context, code string) (*game_config.Buff, error)

	// List 获取Buff列表
	List(ctx context.Context, params BuffQueryParams) ([]*game_config.Buff, int64, error)

	// Create 创建Buff
	Create(ctx context.Context, buff *game_config.Buff) error

	// Update 更新Buff
	Update(ctx context.Context, buff *game_config.Buff) error

	// Delete 软删除Buff
	Delete(ctx context.Context, buffID string) error

	// Exists 检查代码是否存在
	Exists(ctx context.Context, code string) (bool, error)
}
