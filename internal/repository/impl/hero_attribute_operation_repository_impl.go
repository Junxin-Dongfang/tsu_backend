package impl

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/repository/interfaces"
)

type heroAttributeOperationRepositoryImpl struct {
	db *sql.DB
}

// NewHeroAttributeOperationRepository 创建属性操作历史仓储实例
func NewHeroAttributeOperationRepository(db *sql.DB) interfaces.HeroAttributeOperationRepository {
	return &heroAttributeOperationRepositoryImpl{db: db}
}

// Create 创建属性操作记录
func (r *heroAttributeOperationRepositoryImpl) Create(ctx context.Context, execer boil.ContextExecutor, operation *game_runtime.HeroAttributeOperation) error {
	if err := operation.Insert(ctx, execer, boil.Infer()); err != nil {
		return fmt.Errorf("创建属性操作记录失败: %w", err)
	}
	return nil
}

// GetLatestRollbackable 获取最近一次可回退的操作（栈顶）
func (r *heroAttributeOperationRepositoryImpl) GetLatestRollbackable(ctx context.Context, heroID, attributeCode string) (*game_runtime.HeroAttributeOperation, error) {
	operation, err := game_runtime.HeroAttributeOperations(
		qm.Where("hero_id = ? AND attribute_code = ? AND rolled_back_at IS NULL AND rollback_deadline > ?",
			heroID, attributeCode, time.Now()),
		qm.OrderBy("created_at DESC"),
		qm.Limit(1),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, nil // 没有可回退的操作
	}
	if err != nil {
		return nil, fmt.Errorf("查询可回退的属性操作失败: %w", err)
	}

	return operation, nil
}

// GetByHeroID 获取英雄的所有属性操作记录
func (r *heroAttributeOperationRepositoryImpl) GetByHeroID(ctx context.Context, heroID string) ([]*game_runtime.HeroAttributeOperation, error) {
	operations, err := game_runtime.HeroAttributeOperations(
		qm.Where("hero_id = ?", heroID),
		qm.OrderBy("created_at DESC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询属性操作记录失败: %w", err)
	}

	return operations, nil
}

// MarkAsRolledBack 标记为已回退
func (r *heroAttributeOperationRepositoryImpl) MarkAsRolledBack(ctx context.Context, execer boil.ContextExecutor, operationID string) error {
	operation, err := game_runtime.HeroAttributeOperations(
		qm.Where("id = ?", operationID),
	).One(ctx, execer)

	if err != nil {
		return fmt.Errorf("查询属性操作记录失败: %w", err)
	}

	operation.RolledBackAt = null.TimeFrom(time.Now())

	if _, err := operation.Update(ctx, execer, boil.Whitelist("rolled_back_at")); err != nil {
		return fmt.Errorf("标记属性操作为已回退失败: %w", err)
	}

	return nil
}

// GetTotalSpentXP 获取英雄在所有属性上花费的总经验（未回退的）
func (r *heroAttributeOperationRepositoryImpl) GetTotalSpentXP(ctx context.Context, heroID string) (int, error) {
	type Result struct {
		Total int `boil:"total"`
	}

	var result Result
	err := game_runtime.HeroAttributeOperations(
		qm.Select("COALESCE(SUM(xp_spent), 0) as total"),
		qm.Where("hero_id = ? AND rolled_back_at IS NULL", heroID),
	).Bind(ctx, r.db, &result)

	if err != nil {
		return 0, fmt.Errorf("计算属性花费总经验失败: %w", err)
	}

	return result.Total, nil
}

// DeleteExpiredOperations 删除过期的操作记录（已回退且超过保留期）
func (r *heroAttributeOperationRepositoryImpl) DeleteExpiredOperations(ctx context.Context, expiryDate time.Time) error {
	_, err := game_runtime.HeroAttributeOperations(
		qm.Where("rolled_back_at IS NOT NULL AND rolled_back_at < ?", expiryDate),
	).DeleteAll(ctx, r.db)

	if err != nil {
		return fmt.Errorf("删除过期属性操作记录失败: %w", err)
	}

	return nil
}

