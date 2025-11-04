package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// MonsterDropRepository 怪物掉落仓储接口
type MonsterDropRepository interface {
	// Create 创建怪物掉落配置
	Create(ctx context.Context, monsterDrop *game_config.MonsterDrop) error

	// BatchCreate 批量创建怪物掉落配置
	BatchCreate(ctx context.Context, monsterDrops []*game_config.MonsterDrop) error

	// GetByMonsterID 获取怪物的所有掉落配置
	GetByMonsterID(ctx context.Context, monsterID string) ([]*game_config.MonsterDrop, error)

	// GetByMonsterAndPool 获取怪物的特定掉落池配置
	GetByMonsterAndPool(ctx context.Context, monsterID, dropPoolID string) (*game_config.MonsterDrop, error)

	// Update 更新怪物掉落配置
	Update(ctx context.Context, monsterDrop *game_config.MonsterDrop) error

	// Delete 软删除怪物掉落配置
	Delete(ctx context.Context, monsterID, dropPoolID string) error

	// DeleteByMonsterID 删除怪物的所有掉落配置（软删除）
	DeleteByMonsterID(ctx context.Context, monsterID string) error

	// Exists 检查怪物掉落配置是否存在
	Exists(ctx context.Context, monsterID, dropPoolID string) (bool, error)
}

