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

type heroSkillOperationRepositoryImpl struct {
	db *sql.DB
}

// NewHeroSkillOperationRepository 创建技能操作历史仓储实例
func NewHeroSkillOperationRepository(db *sql.DB) interfaces.HeroSkillOperationRepository {
	return &heroSkillOperationRepositoryImpl{db: db}
}

// Create 创建技能操作记录
func (r *heroSkillOperationRepositoryImpl) Create(ctx context.Context, execer boil.ContextExecutor, operation *game_runtime.HeroSkillOperation) error {
	if err := operation.Insert(ctx, execer, boil.Infer()); err != nil {
		return fmt.Errorf("创建技能操作记录失败: %w", err)
	}
	return nil
}

// GetLatestRollbackable 获取最近一次可回退的操作（栈顶）
func (r *heroSkillOperationRepositoryImpl) GetLatestRollbackable(ctx context.Context, heroSkillID string) (*game_runtime.HeroSkillOperation, error) {
	operation, err := game_runtime.HeroSkillOperations(
		qm.Where("hero_skill_id = ? AND rolled_back_at IS NULL AND rollback_deadline > ?",
			heroSkillID, time.Now()),
		qm.OrderBy("created_at DESC"),
		qm.Limit(1),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, nil // 没有可回退的操作
	}
	if err != nil {
		return nil, fmt.Errorf("查询可回退的技能操作失败: %w", err)
	}

	return operation, nil
}

// GetByHeroSkillID 获取技能的所有操作记录
func (r *heroSkillOperationRepositoryImpl) GetByHeroSkillID(ctx context.Context, heroSkillID string) ([]*game_runtime.HeroSkillOperation, error) {
	operations, err := game_runtime.HeroSkillOperations(
		qm.Where("hero_skill_id = ?", heroSkillID),
		qm.OrderBy("created_at DESC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询技能操作记录失败: %w", err)
	}

	return operations, nil
}

// MarkAsRolledBack 标记为已回退
func (r *heroSkillOperationRepositoryImpl) MarkAsRolledBack(ctx context.Context, execer boil.ContextExecutor, operationID string) error {
	operation, err := game_runtime.HeroSkillOperations(
		qm.Where("id = ?", operationID),
	).One(ctx, execer)

	if err != nil {
		return fmt.Errorf("查询技能操作记录失败: %w", err)
	}

	operation.RolledBackAt = null.TimeFrom(time.Now())

	if _, err := operation.Update(ctx, execer, boil.Whitelist("rolled_back_at")); err != nil {
		return fmt.Errorf("标记技能操作为已回退失败: %w", err)
	}

	return nil
}

// GetTotalSpentXPByHeroID 获取英雄在所有技能上花费的总经验（未回退的）
func (r *heroSkillOperationRepositoryImpl) GetTotalSpentXPByHeroID(ctx context.Context, heroID string) (int, error) {
	type Result struct {
		Total int `boil:"total"`
	}

	var result Result
	query := `
		SELECT COALESCE(SUM(hso.xp_spent), 0) as total
		FROM game_runtime.hero_skill_operations hso
		JOIN game_runtime.hero_skills hs ON hs.id = hso.hero_skill_id
		WHERE hs.hero_id = $1 AND hso.rolled_back_at IS NULL
	`

	err := game_runtime.NewQuery(qm.SQL(query, heroID)).Bind(ctx, r.db, &result)
	if err != nil {
		return 0, fmt.Errorf("计算技能花费总经验失败: %w", err)
	}

	return result.Total, nil
}

// DeleteExpiredOperations 删除过期的操作记录（已回退且超过保留期）
func (r *heroSkillOperationRepositoryImpl) DeleteExpiredOperations(ctx context.Context, expiryDate time.Time) (int64, error) {
	count, err := game_runtime.HeroSkillOperations(
		qm.Where("rolled_back_at IS NOT NULL AND rolled_back_at < ?", expiryDate),
	).DeleteAll(ctx, r.db)

	if err != nil {
		return 0, fmt.Errorf("删除过期技能操作记录失败: %w", err)
	}

	return count, nil
}

