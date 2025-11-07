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

type dungeonBattleRepositoryImpl struct {
	db *sql.DB
}

// NewDungeonBattleRepository 创建战斗配置仓储实例
func NewDungeonBattleRepository(db *sql.DB) interfaces.DungeonBattleRepository {
	return &dungeonBattleRepositoryImpl{db: db}
}

// GetByID 根据ID获取战斗配置
func (r *dungeonBattleRepositoryImpl) GetByID(ctx context.Context, battleID string) (*game_config.DungeonBattle, error) {
	battle, err := game_config.DungeonBattles(
		qm.Where("id = ? AND deleted_at IS NULL", battleID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("战斗配置不存在: %s", battleID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询战斗配置失败: %w", err)
	}

	return battle, nil
}

// GetByCode 根据代码获取战斗配置
func (r *dungeonBattleRepositoryImpl) GetByCode(ctx context.Context, code string) (*game_config.DungeonBattle, error) {
	battle, err := game_config.DungeonBattles(
		qm.Where("battle_code = ? AND deleted_at IS NULL", code),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("战斗配置不存在: %s", code)
	}
	if err != nil {
		return nil, fmt.Errorf("查询战斗配置失败: %w", err)
	}

	return battle, nil
}

// Create 创建战斗配置
func (r *dungeonBattleRepositoryImpl) Create(ctx context.Context, battle *game_config.DungeonBattle) error {
	// 生成UUID
	if battle.ID == "" {
		battle.ID = uuid.New().String()
	}

	// 设置时间戳
	now := time.Now()
	battle.CreatedAt = now
	battle.UpdatedAt = now

	// 插入数据库
	if err := battle.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("创建战斗配置失败: %w", err)
	}

	return nil
}

// Update 更新战斗配置
func (r *dungeonBattleRepositoryImpl) Update(ctx context.Context, battle *game_config.DungeonBattle) error {
	// 更新时间戳
	battle.UpdatedAt = time.Now()

	// 更新数据库
	if _, err := battle.Update(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("更新战斗配置失败: %w", err)
	}

	return nil
}

// Delete 软删除战斗配置
func (r *dungeonBattleRepositoryImpl) Delete(ctx context.Context, battleID string) error {
	// 查询战斗配置
	battle, err := r.GetByID(ctx, battleID)
	if err != nil {
		return err
	}

	// 设置删除时间
	now := time.Now()
	battle.DeletedAt = null.TimeFrom(now)
	battle.UpdatedAt = now

	// 更新数据库
	if _, err := battle.Update(ctx, r.db, boil.Whitelist("deleted_at", "updated_at")); err != nil {
		return fmt.Errorf("删除战斗配置失败: %w", err)
	}

	return nil
}

// Exists 检查代码是否存在
func (r *dungeonBattleRepositoryImpl) Exists(ctx context.Context, code string) (bool, error) {
	count, err := game_config.DungeonBattles(
		qm.Where("battle_code = ? AND deleted_at IS NULL", code),
	).Count(ctx, r.db)

	if err != nil {
		return false, fmt.Errorf("检查战斗配置代码是否存在失败: %w", err)
	}

	return count > 0, nil
}

// ExistsExcludingID 检查代码是否存在（排除指定ID）
func (r *dungeonBattleRepositoryImpl) ExistsExcludingID(ctx context.Context, code string, excludeID string) (bool, error) {
	count, err := game_config.DungeonBattles(
		qm.Where("battle_code = ? AND id != ? AND deleted_at IS NULL", code, excludeID),
	).Count(ctx, r.db)

	if err != nil {
		return false, fmt.Errorf("检查战斗配置代码是否存在失败: %w", err)
	}

	return count > 0, nil
}

