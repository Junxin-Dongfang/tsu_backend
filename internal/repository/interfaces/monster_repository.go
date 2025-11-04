package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// MonsterQueryParams 怪物查询参数
type MonsterQueryParams struct {
	MonsterCode  *string // 怪物代码（模糊搜索）
	MonsterName  *string // 怪物名称（模糊搜索）
	MinLevel     *int16  // 最小等级
	MaxLevel     *int16  // 最大等级
	IsActive     *bool   // 是否启用
	Limit        int     // 每页数量
	Offset       int     // 偏移量
	OrderBy      string  // 排序字段（monster_level, created_at, updated_at）
	OrderDesc    bool    // 是否降序
}

// MonsterRepository 怪物仓储接口
type MonsterRepository interface {
	// GetByID 根据ID获取怪物
	GetByID(ctx context.Context, monsterID string) (*game_config.Monster, error)

	// GetByCode 根据代码获取怪物
	GetByCode(ctx context.Context, code string) (*game_config.Monster, error)

	// List 获取怪物列表
	List(ctx context.Context, params MonsterQueryParams) ([]*game_config.Monster, int64, error)

	// Create 创建怪物
	Create(ctx context.Context, monster *game_config.Monster) error

	// Update 更新怪物
	Update(ctx context.Context, monster *game_config.Monster) error

	// Delete 软删除怪物
	Delete(ctx context.Context, monsterID string) error

	// Exists 检查代码是否存在
	Exists(ctx context.Context, code string) (bool, error)

	// ExistsExcludingID 检查代码是否存在（排除指定ID）
	ExistsExcludingID(ctx context.Context, code string, excludeID string) (bool, error)
}

