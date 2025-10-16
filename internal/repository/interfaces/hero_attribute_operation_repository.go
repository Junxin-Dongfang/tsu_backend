package interfaces

import (
	"context"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"tsu-self/internal/entity/game_runtime"
)

// HeroAttributeOperationRepository 英雄属性操作历史仓储接口
type HeroAttributeOperationRepository interface {
	// Create 创建属性操作记录
	Create(ctx context.Context, execer boil.ContextExecutor, operation *game_runtime.HeroAttributeOperation) error

	// GetLatestRollbackable 获取最近一次可回退的操作（栈顶）
	GetLatestRollbackable(ctx context.Context, heroID, attributeCode string) (*game_runtime.HeroAttributeOperation, error)

	// GetByHeroID 获取英雄的所有属性操作记录
	GetByHeroID(ctx context.Context, heroID string) ([]*game_runtime.HeroAttributeOperation, error)

	// MarkAsRolledBack 标记为已回退
	MarkAsRolledBack(ctx context.Context, execer boil.ContextExecutor, operationID string) error

	// GetTotalSpentXP 获取英雄在所有属性上花费的总经验
	GetTotalSpentXP(ctx context.Context, heroID string) (int, error)

	// DeleteExpiredOperations 删除过期的操作记录（已回退且超过保留期）
	DeleteExpiredOperations(ctx context.Context, expiryDate time.Time) error
}

