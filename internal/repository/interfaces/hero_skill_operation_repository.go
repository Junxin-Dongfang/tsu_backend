package interfaces

import (
	"context"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"tsu-self/internal/entity/game_runtime"
)

// HeroSkillOperationRepository 英雄技能操作历史仓储接口
type HeroSkillOperationRepository interface {
	// Create 创建技能操作记录
	Create(ctx context.Context, execer boil.ContextExecutor, operation *game_runtime.HeroSkillOperation) error

	// GetLatestRollbackable 获取最近一次可回退的操作（栈顶）
	GetLatestRollbackable(ctx context.Context, heroSkillID string) (*game_runtime.HeroSkillOperation, error)

	// GetByHeroSkillID 获取技能的所有操作记录
	GetByHeroSkillID(ctx context.Context, heroSkillID string) ([]*game_runtime.HeroSkillOperation, error)

	// MarkAsRolledBack 标记为已回退
	MarkAsRolledBack(ctx context.Context, execer boil.ContextExecutor, operationID string) error

	// GetTotalSpentXPByHeroID 获取英雄在所有技能上花费的总经验
	GetTotalSpentXPByHeroID(ctx context.Context, heroID string) (int, error)

	// DeleteExpiredOperations 删除过期的操作记录（已回退且超过保留期）
	DeleteExpiredOperations(ctx context.Context, expiryDate time.Time) (int64, error)
}

