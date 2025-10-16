package impl

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/repository/interfaces"
)

type heroRepositoryImpl struct {
	db *sql.DB
}

// NewHeroRepository 创建英雄仓储实例
func NewHeroRepository(db *sql.DB) interfaces.HeroRepository {
	return &heroRepositoryImpl{db: db}
}

// Create 创建英雄
func (r *heroRepositoryImpl) Create(ctx context.Context, hero *game_runtime.Hero) error {
	if err := hero.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("创建英雄失败: %w", err)
	}
	return nil
}

// GetByID 根据ID获取英雄
func (r *heroRepositoryImpl) GetByID(ctx context.Context, heroID string) (*game_runtime.Hero, error) {
	hero, err := game_runtime.Heroes(
		qm.Where("id = ? AND deleted_at IS NULL", heroID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("英雄不存在: %s", heroID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询英雄失败: %w", err)
	}

	return hero, nil
}

// GetByIDForUpdate 根据ID获取英雄（带行锁）
func (r *heroRepositoryImpl) GetByIDForUpdate(ctx context.Context, tx *sql.Tx, heroID string) (*game_runtime.Hero, error) {
	hero, err := game_runtime.Heroes(
		qm.Where("id = ? AND deleted_at IS NULL", heroID),
		qm.For("UPDATE"),
	).One(ctx, tx)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("英雄不存在: %s", heroID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询英雄失败（带锁）: %w", err)
	}

	return hero, nil
}

// GetByUserID 获取用户的英雄列表
func (r *heroRepositoryImpl) GetByUserID(ctx context.Context, userID string) ([]*game_runtime.Hero, error) {
	heroes, err := game_runtime.Heroes(
		qm.Where("user_id = ? AND deleted_at IS NULL", userID),
		qm.OrderBy("created_at DESC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询用户英雄列表失败: %w", err)
	}

	return heroes, nil
}

// Update 更新英雄信息
func (r *heroRepositoryImpl) Update(ctx context.Context, execer boil.ContextExecutor, hero *game_runtime.Hero) error {
	if _, err := hero.Update(ctx, execer, boil.Infer()); err != nil {
		return fmt.Errorf("更新英雄失败: %w", err)
	}
	return nil
}

// Delete 删除英雄（软删除）
func (r *heroRepositoryImpl) Delete(ctx context.Context, heroID string) error {
	hero, err := r.GetByID(ctx, heroID)
	if err != nil {
		return err
	}

	hero.DeletedAt = null.TimeFromPtr(nullTimeNow())

	if _, err := hero.Update(ctx, r.db, boil.Whitelist("deleted_at", "updated_at")); err != nil {
		return fmt.Errorf("软删除英雄失败: %w", err)
	}

	return nil
}

// CheckExistsByName 检查英雄名称是否已存在
func (r *heroRepositoryImpl) CheckExistsByName(ctx context.Context, userID, heroName string) (bool, error) {
	count, err := game_runtime.Heroes(
		qm.Where("user_id = ? AND hero_name = ? AND deleted_at IS NULL", userID, heroName),
	).Count(ctx, r.db)

	if err != nil {
		return false, fmt.Errorf("检查英雄名称是否存在失败: %w", err)
	}

	return count > 0, nil
}

