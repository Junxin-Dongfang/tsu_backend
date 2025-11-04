package impl

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/google/uuid"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/interfaces"
)

type monsterDropRepositoryImpl struct {
	db *sql.DB
}

// NewMonsterDropRepository 创建怪物掉落仓储实例
func NewMonsterDropRepository(db *sql.DB) interfaces.MonsterDropRepository {
	return &monsterDropRepositoryImpl{db: db}
}

// Create 创建怪物掉落配置
func (r *monsterDropRepositoryImpl) Create(ctx context.Context, monsterDrop *game_config.MonsterDrop) error {
	// 生成UUID
	if monsterDrop.ID == "" {
		monsterDrop.ID = uuid.New().String()
	}

	// 设置时间戳
	now := time.Now()
	monsterDrop.CreatedAt = now
	monsterDrop.UpdatedAt = now

	// 插入数据库
	if err := monsterDrop.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("创建怪物掉落配置失败: %w", err)
	}

	return nil
}

// BatchCreate 批量创建怪物掉落配置
func (r *monsterDropRepositoryImpl) BatchCreate(ctx context.Context, monsterDrops []*game_config.MonsterDrop) error {
	if len(monsterDrops) == 0 {
		return nil
	}

	// 开启事务
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback()

	// 批量插入
	now := time.Now()
	for _, md := range monsterDrops {
		if md.ID == "" {
			md.ID = uuid.New().String()
		}
		md.CreatedAt = now
		md.UpdatedAt = now

		if err := md.Insert(ctx, tx, boil.Infer()); err != nil {
			return fmt.Errorf("批量创建怪物掉落配置失败: %w", err)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// GetByMonsterID 获取怪物的所有掉落配置
func (r *monsterDropRepositoryImpl) GetByMonsterID(ctx context.Context, monsterID string) ([]*game_config.MonsterDrop, error) {
	monsterDrops, err := game_config.MonsterDrops(
		qm.Where("monster_id = ? AND deleted_at IS NULL", monsterID),
		qm.OrderBy("display_order ASC, created_at ASC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询怪物掉落配置列表失败: %w", err)
	}

	return monsterDrops, nil
}

// GetByMonsterAndPool 获取怪物的特定掉落池配置
func (r *monsterDropRepositoryImpl) GetByMonsterAndPool(ctx context.Context, monsterID, dropPoolID string) (*game_config.MonsterDrop, error) {
	monsterDrop, err := game_config.MonsterDrops(
		qm.Where("monster_id = ? AND drop_pool_id = ? AND deleted_at IS NULL", monsterID, dropPoolID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("怪物掉落配置不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询怪物掉落配置失败: %w", err)
	}

	return monsterDrop, nil
}

// Update 更新怪物掉落配置
func (r *monsterDropRepositoryImpl) Update(ctx context.Context, monsterDrop *game_config.MonsterDrop) error {
	// 更新时间戳
	monsterDrop.UpdatedAt = time.Now()

	// 更新数据库
	if _, err := monsterDrop.Update(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("更新怪物掉落配置失败: %w", err)
	}

	return nil
}

// Delete 软删除怪物掉落配置
func (r *monsterDropRepositoryImpl) Delete(ctx context.Context, monsterID, dropPoolID string) error {
	// 查询怪物掉落配置
	monsterDrop, err := r.GetByMonsterAndPool(ctx, monsterID, dropPoolID)
	if err != nil {
		return err
	}

	// 设置删除时间
	now := time.Now()
	monsterDrop.DeletedAt = null.TimeFrom(now)
	monsterDrop.UpdatedAt = now

	// 更新数据库
	if _, err := monsterDrop.Update(ctx, r.db, boil.Whitelist("deleted_at", "updated_at")); err != nil {
		return fmt.Errorf("删除怪物掉落配置失败: %w", err)
	}

	return nil
}

// DeleteByMonsterID 删除怪物的所有掉落配置（软删除）
func (r *monsterDropRepositoryImpl) DeleteByMonsterID(ctx context.Context, monsterID string) error {
	// 查询怪物的所有掉落配置
	monsterDrops, err := r.GetByMonsterID(ctx, monsterID)
	if err != nil {
		return err
	}

	if len(monsterDrops) == 0 {
		return nil
	}

	// 开启事务
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback()

	// 批量软删除
	now := time.Now()
	for _, md := range monsterDrops {
		md.DeletedAt = null.TimeFrom(now)
		md.UpdatedAt = now

		if _, err := md.Update(ctx, tx, boil.Whitelist("deleted_at", "updated_at")); err != nil {
			return fmt.Errorf("删除怪物掉落配置失败: %w", err)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// Exists 检查怪物掉落配置是否存在
func (r *monsterDropRepositoryImpl) Exists(ctx context.Context, monsterID, dropPoolID string) (bool, error) {
	count, err := game_config.MonsterDrops(
		qm.Where("monster_id = ? AND drop_pool_id = ? AND deleted_at IS NULL", monsterID, dropPoolID),
	).Count(ctx, r.db)

	if err != nil {
		return false, fmt.Errorf("检查怪物掉落配置是否存在失败: %w", err)
	}

	return count > 0, nil
}

