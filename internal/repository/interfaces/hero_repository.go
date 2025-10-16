package interfaces

import (
	"context"
	"database/sql"

	"github.com/aarondl/sqlboiler/v4/boil"
	"tsu-self/internal/entity/game_runtime"
)

// HeroRepository 英雄仓储接口
type HeroRepository interface {
	// Create 创建英雄
	Create(ctx context.Context, hero *game_runtime.Hero) error

	// GetByID 根据ID获取英雄
	GetByID(ctx context.Context, heroID string) (*game_runtime.Hero, error)

	// GetByIDForUpdate 根据ID获取英雄（带行锁）
	GetByIDForUpdate(ctx context.Context, tx *sql.Tx, heroID string) (*game_runtime.Hero, error)

	// GetByUserID 获取用户的英雄列表
	GetByUserID(ctx context.Context, userID string) ([]*game_runtime.Hero, error)

	// Update 更新英雄信息
	Update(ctx context.Context, execer boil.ContextExecutor, hero *game_runtime.Hero) error

	// Delete 删除英雄（软删除）
	Delete(ctx context.Context, heroID string) error

	// CheckExistsByName 检查英雄名称是否已存在
	CheckExistsByName(ctx context.Context, userID, heroName string) (bool, error)
}

