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

type dungeonEventRepositoryImpl struct {
	db *sql.DB
}

// NewDungeonEventRepository 创建事件配置仓储实例
func NewDungeonEventRepository(db *sql.DB) interfaces.DungeonEventRepository {
	return &dungeonEventRepositoryImpl{db: db}
}

// GetByID 根据ID获取事件配置
func (r *dungeonEventRepositoryImpl) GetByID(ctx context.Context, eventID string) (*game_config.DungeonEvent, error) {
	event, err := game_config.DungeonEvents(
		qm.Where("id = ? AND deleted_at IS NULL", eventID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("事件配置不存在: %s", eventID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询事件配置失败: %w", err)
	}

	return event, nil
}

// GetByCode 根据代码获取事件配置
func (r *dungeonEventRepositoryImpl) GetByCode(ctx context.Context, code string) (*game_config.DungeonEvent, error) {
	event, err := game_config.DungeonEvents(
		qm.Where("event_code = ? AND deleted_at IS NULL", code),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("事件配置不存在: %s", code)
	}
	if err != nil {
		return nil, fmt.Errorf("查询事件配置失败: %w", err)
	}

	return event, nil
}

// Create 创建事件配置
func (r *dungeonEventRepositoryImpl) Create(ctx context.Context, event *game_config.DungeonEvent) error {
	// 生成UUID
	if event.ID == "" {
		event.ID = uuid.New().String()
	}

	// 设置时间戳
	now := time.Now()
	event.CreatedAt = now
	event.UpdatedAt = now

	// 插入数据库
	if err := event.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("创建事件配置失败: %w", err)
	}

	return nil
}

// Update 更新事件配置
func (r *dungeonEventRepositoryImpl) Update(ctx context.Context, event *game_config.DungeonEvent) error {
	// 更新时间戳
	event.UpdatedAt = time.Now()

	// 更新数据库
	if _, err := event.Update(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("更新事件配置失败: %w", err)
	}

	return nil
}

// Delete 软删除事件配置
func (r *dungeonEventRepositoryImpl) Delete(ctx context.Context, eventID string) error {
	// 查询事件配置
	event, err := r.GetByID(ctx, eventID)
	if err != nil {
		return err
	}

	// 设置删除时间
	now := time.Now()
	event.DeletedAt = null.TimeFrom(now)
	event.UpdatedAt = now

	// 更新数据库
	if _, err := event.Update(ctx, r.db, boil.Whitelist("deleted_at", "updated_at")); err != nil {
		return fmt.Errorf("删除事件配置失败: %w", err)
	}

	return nil
}

// Exists 检查代码是否存在
func (r *dungeonEventRepositoryImpl) Exists(ctx context.Context, code string) (bool, error) {
	count, err := game_config.DungeonEvents(
		qm.Where("event_code = ? AND deleted_at IS NULL", code),
	).Count(ctx, r.db)

	if err != nil {
		return false, fmt.Errorf("检查事件配置代码是否存在失败: %w", err)
	}

	return count > 0, nil
}

// ExistsExcludingID 检查代码是否存在（排除指定ID）
func (r *dungeonEventRepositoryImpl) ExistsExcludingID(ctx context.Context, code string, excludeID string) (bool, error) {
	count, err := game_config.DungeonEvents(
		qm.Where("event_code = ? AND id != ? AND deleted_at IS NULL", code, excludeID),
	).Count(ctx, r.db)

	if err != nil {
		return false, fmt.Errorf("检查事件配置代码是否存在失败: %w", err)
	}

	return count > 0, nil
}

