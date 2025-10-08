package impl

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/google/uuid"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/interfaces"
)

type attributeBonusRepositoryImpl struct {
	db *sql.DB
}

// NewAttributeBonusRepository 创建属性加成仓储实例
func NewAttributeBonusRepository(db *sql.DB) interfaces.AttributeBonusRepository {
	return &attributeBonusRepositoryImpl{db: db}
}

// GetByClassID 获取职业的所有属性加成
func (r *attributeBonusRepositoryImpl) GetByClassID(ctx context.Context, classID string) ([]*game_config.ClassAttributeBonuse, error) {
	bonuses, err := game_config.ClassAttributeBonuses(
		qm.Where("class_id = ?", classID),
		qm.OrderBy("created_at ASC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询职业属性加成失败: %w", err)
	}

	return bonuses, nil
}

// GetByID 根据ID获取属性加成
func (r *attributeBonusRepositoryImpl) GetByID(ctx context.Context, bonusID string) (*game_config.ClassAttributeBonuse, error) {
	bonus, err := game_config.ClassAttributeBonuses(
		qm.Where("id = ?", bonusID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("属性加成不存在: %s", bonusID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询属性加成失败: %w", err)
	}

	return bonus, nil
}

// Create 创建属性加成
func (r *attributeBonusRepositoryImpl) Create(ctx context.Context, bonus *game_config.ClassAttributeBonuse) error {
	// 生成 UUID
	if bonus.ID == "" {
		bonus.ID = uuid.New().String()
	}

	// 设置时间戳
	now := time.Now()
	bonus.CreatedAt = now
	bonus.UpdatedAt = now

	// 插入数据库
	if err := bonus.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("创建属性加成失败: %w", err)
	}

	return nil
}

// Update 更新属性加成
func (r *attributeBonusRepositoryImpl) Update(ctx context.Context, bonus *game_config.ClassAttributeBonuse) error {
	// 更新时间戳
	bonus.UpdatedAt = time.Now()

	// 更新数据库
	if _, err := bonus.Update(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("更新属性加成失败: %w", err)
	}

	return nil
}

// Delete 删除属性加成
func (r *attributeBonusRepositoryImpl) Delete(ctx context.Context, bonusID string) error {
	bonus, err := r.GetByID(ctx, bonusID)
	if err != nil {
		return err
	}

	if _, err := bonus.Delete(ctx, r.db); err != nil {
		return fmt.Errorf("删除属性加成失败: %w", err)
	}

	return nil
}

// BatchCreate 批量创建属性加成
func (r *attributeBonusRepositoryImpl) BatchCreate(ctx context.Context, bonuses []*game_config.ClassAttributeBonuse) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	for _, bonus := range bonuses {
		if bonus.ID == "" {
			bonus.ID = uuid.New().String()
		}
		bonus.CreatedAt = now
		bonus.UpdatedAt = now

		if err := bonus.Insert(ctx, tx, boil.Infer()); err != nil {
			return fmt.Errorf("批量创建属性加成失败: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// DeleteByClassID 删除职业的所有属性加成
func (r *attributeBonusRepositoryImpl) DeleteByClassID(ctx context.Context, classID string) error {
	_, err := game_config.ClassAttributeBonuses(
		qm.Where("class_id = ?", classID),
	).DeleteAll(ctx, r.db)

	if err != nil {
		return fmt.Errorf("删除职业属性加成失败: %w", err)
	}

	return nil
}

// Exists 检查职业-属性组合是否已存在
func (r *attributeBonusRepositoryImpl) Exists(ctx context.Context, classID, attributeID string) (bool, error) {
	count, err := game_config.ClassAttributeBonuses(
		qm.Where("class_id = ? AND attribute_id = ?", classID, attributeID),
	).Count(ctx, r.db)

	if err != nil {
		return false, fmt.Errorf("检查属性加成是否存在失败: %w", err)
	}

	return count > 0, nil
}
