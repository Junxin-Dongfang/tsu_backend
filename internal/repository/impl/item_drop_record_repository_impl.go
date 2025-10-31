package impl

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/repository/interfaces"
)

type itemDropRecordRepositoryImpl struct {
	db *sql.DB
}

// NewItemDropRecordRepository 创建物品掉落记录仓储实例
func NewItemDropRecordRepository(db *sql.DB) interfaces.ItemDropRecordRepository {
	return &itemDropRecordRepositoryImpl{db: db}
}

// Create 创建掉落记录
func (r *itemDropRecordRepositoryImpl) Create(ctx context.Context, execer boil.ContextExecutor, record *game_runtime.ItemDropRecord) error {
	if err := record.Insert(ctx, execer, boil.Infer()); err != nil {
		return fmt.Errorf("创建掉落记录失败: %w", err)
	}
	return nil
}

// GetByID 根据ID获取掉落记录
func (r *itemDropRecordRepositoryImpl) GetByID(ctx context.Context, recordID string) (*game_runtime.ItemDropRecord, error) {
	record, err := game_runtime.ItemDropRecords(
		qm.Where("id = ?", recordID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("掉落记录不存在: %s", recordID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询掉落记录失败: %w", err)
	}

	return record, nil
}

// GetByReceiver 查询玩家的掉落记录
func (r *itemDropRecordRepositoryImpl) GetByReceiver(ctx context.Context, receiverID string, limit int) ([]*game_runtime.ItemDropRecord, error) {
	mods := []qm.QueryMod{
		qm.Where("receiver_id = ?", receiverID),
		qm.OrderBy("dropped_at DESC"),
	}

	if limit > 0 {
		mods = append(mods, qm.Limit(limit))
	}

	records, err := game_runtime.ItemDropRecords(mods...).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("查询玩家掉落记录失败: %w", err)
	}

	return records, nil
}

// GetByItemConfig 查询特定物品的掉落记录
func (r *itemDropRecordRepositoryImpl) GetByItemConfig(ctx context.Context, itemConfigID string, limit int) ([]*game_runtime.ItemDropRecord, error) {
	mods := []qm.QueryMod{
		qm.Where("item_config_id = ?", itemConfigID),
		qm.OrderBy("dropped_at DESC"),
	}

	if limit > 0 {
		mods = append(mods, qm.Limit(limit))
	}

	records, err := game_runtime.ItemDropRecords(mods...).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("查询物品掉落记录失败: %w", err)
	}

	return records, nil
}

// GetBySource 查询特定来源的掉落记录
func (r *itemDropRecordRepositoryImpl) GetBySource(ctx context.Context, dropSource string, sourceID string, limit int) ([]*game_runtime.ItemDropRecord, error) {
	mods := []qm.QueryMod{
		qm.Where("drop_source = ?", dropSource),
		qm.OrderBy("dropped_at DESC"),
	}

	if sourceID != "" {
		mods = append(mods, qm.Where("source_id = ?", sourceID))
	}

	if limit > 0 {
		mods = append(mods, qm.Limit(limit))
	}

	records, err := game_runtime.ItemDropRecords(mods...).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("查询来源掉落记录失败: %w", err)
	}

	return records, nil
}

// GetRecentDrops 查询最近的掉落记录
func (r *itemDropRecordRepositoryImpl) GetRecentDrops(ctx context.Context, limit int) ([]*game_runtime.ItemDropRecord, error) {
	mods := []qm.QueryMod{
		qm.OrderBy("dropped_at DESC"),
	}

	if limit > 0 {
		mods = append(mods, qm.Limit(limit))
	}

	records, err := game_runtime.ItemDropRecords(mods...).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("查询最近掉落记录失败: %w", err)
	}

	return records, nil
}

// CountByItemConfigAndTimeRange 统计特定物品在时间范围内的掉落次数
func (r *itemDropRecordRepositoryImpl) CountByItemConfigAndTimeRange(ctx context.Context, itemConfigID string, startTime, endTime string) (int64, error) {
	count, err := game_runtime.ItemDropRecords(
		qm.Where("item_config_id = ?", itemConfigID),
		qm.Where("dropped_at >= ?", startTime),
		qm.Where("dropped_at <= ?", endTime),
	).Count(ctx, r.db)

	if err != nil {
		return 0, fmt.Errorf("统计掉落次数失败: %w", err)
	}

	return count, nil
}

