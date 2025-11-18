package impl

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"tsu-self/internal/repository/interfaces"
)

type worldDropItemRepositoryImpl struct {
	db *sql.DB
}

// NewWorldDropItemRepository 创建世界掉落物品仓储实例
func NewWorldDropItemRepository(db *sql.DB) interfaces.WorldDropItemRepository {
	return &worldDropItemRepositoryImpl{db: db}
}

func (r *worldDropItemRepositoryImpl) ListByConfig(ctx context.Context, params interfaces.ListWorldDropItemParams) ([]interfaces.WorldDropItemWithItem, int64, error) {
	page := params.Page
	if page <= 0 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	countQuery := `SELECT COUNT(1) FROM game_config.world_drop_items WHERE world_drop_config_id = $1 AND deleted_at IS NULL`
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, params.WorldDropConfigID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("统计世界掉落物品失败: %w", err)
	}

	query := `
SELECT wdi.id,
       wdi.world_drop_config_id,
       wdi.item_id,
       wdi.drop_rate,
       wdi.drop_weight,
       wdi.min_quantity,
       wdi.max_quantity,
       wdi.min_level,
       wdi.max_level,
       wdi.guaranteed_drop,
       COALESCE(wdi.metadata, '{}'::jsonb) as metadata,
       wdi.created_at,
       wdi.updated_at,
       i.item_code,
       i.item_name,
       i.item_quality
FROM game_config.world_drop_items wdi
         JOIN game_config.items i ON i.id = wdi.item_id
WHERE wdi.world_drop_config_id = $1
  AND wdi.deleted_at IS NULL
ORDER BY wdi.created_at DESC
LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, params.WorldDropConfigID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("查询世界掉落物品失败: %w", err)
	}
	defer rows.Close()

	items := make([]interfaces.WorldDropItemWithItem, 0)
	for rows.Next() {
		var entry interfaces.WorldDropItemWithItem
		var dropRate sql.NullFloat64
		var dropWeight sql.NullInt64
		var minLevel sql.NullInt64
		var maxLevel sql.NullInt64
		var metadataBytes []byte
		if err := rows.Scan(
			&entry.ID,
			&entry.WorldDropConfigID,
			&entry.ItemID,
			&dropRate,
			&dropWeight,
			&entry.MinQuantity,
			&entry.MaxQuantity,
			&minLevel,
			&maxLevel,
			&entry.GuaranteedDrop,
			&metadataBytes,
			&entry.CreatedAt,
			&entry.UpdatedAt,
			&entry.ItemCode,
			&entry.ItemName,
			&entry.ItemQuality,
		); err != nil {
			return nil, 0, fmt.Errorf("解析世界掉落物品失败: %w", err)
		}
		entry.DropRate = float64PtrFromNull(dropRate)
		entry.DropWeight = intPtrFromNull(dropWeight)
		entry.MinLevel = intPtrFromNull(minLevel)
		entry.MaxLevel = intPtrFromNull(maxLevel)
		entry.Metadata = metadataFromBytes(metadataBytes)
		items = append(items, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("遍历世界掉落物品失败: %w", err)
	}

	return items, total, nil
}

func (r *worldDropItemRepositoryImpl) GetByID(ctx context.Context, configID, itemEntryID string) (*interfaces.WorldDropItemWithItem, error) {
	query := `
SELECT wdi.id,
       wdi.world_drop_config_id,
       wdi.item_id,
       wdi.drop_rate,
       wdi.drop_weight,
       wdi.min_quantity,
       wdi.max_quantity,
       wdi.min_level,
       wdi.max_level,
       wdi.guaranteed_drop,
       COALESCE(wdi.metadata, '{}'::jsonb) as metadata,
       wdi.created_at,
       wdi.updated_at,
       i.item_code,
       i.item_name,
       i.item_quality
FROM game_config.world_drop_items wdi
         JOIN game_config.items i ON i.id = wdi.item_id
WHERE wdi.id = $1
  AND wdi.world_drop_config_id = $2
  AND wdi.deleted_at IS NULL`

	var entry interfaces.WorldDropItemWithItem
	var dropRate sql.NullFloat64
	var dropWeight sql.NullInt64
	var minLevel sql.NullInt64
	var maxLevel sql.NullInt64
	var metadataBytes []byte
	err := r.db.QueryRowContext(ctx, query, itemEntryID, configID).Scan(
		&entry.ID,
		&entry.WorldDropConfigID,
		&entry.ItemID,
		&dropRate,
		&dropWeight,
		&entry.MinQuantity,
		&entry.MaxQuantity,
		&minLevel,
		&maxLevel,
		&entry.GuaranteedDrop,
		&metadataBytes,
		&entry.CreatedAt,
		&entry.UpdatedAt,
		&entry.ItemCode,
		&entry.ItemName,
		&entry.ItemQuality,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("世界掉落物品不存在: %s", itemEntryID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询世界掉落物品失败: %w", err)
	}

	entry.DropRate = float64PtrFromNull(dropRate)
	entry.DropWeight = intPtrFromNull(dropWeight)
	entry.MinLevel = intPtrFromNull(minLevel)
	entry.MaxLevel = intPtrFromNull(maxLevel)
	entry.Metadata = metadataFromBytes(metadataBytes)

	return &entry, nil
}

func (r *worldDropItemRepositoryImpl) HasItemInConfig(ctx context.Context, configID, itemID string) (bool, error) {
	query := `SELECT 1 FROM game_config.world_drop_items WHERE world_drop_config_id = $1 AND item_id = $2 AND deleted_at IS NULL LIMIT 1`
	var dummy int
	err := r.db.QueryRowContext(ctx, query, configID, itemID).Scan(&dummy)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("检测世界掉落物品是否已存在失败: %w", err)
	}
	return true, nil
}

func (r *worldDropItemRepositoryImpl) Create(ctx context.Context, item *interfaces.WorldDropItem) error {
	if item.ID == "" {
		item.ID = uuid.NewString()
	}
	metadataBytes := metadataToBytes(item.Metadata)
	query := `
INSERT INTO game_config.world_drop_items (
    id, world_drop_config_id, item_id, drop_rate, drop_weight,
    min_quantity, max_quantity, min_level, max_level,
    guaranteed_drop, metadata)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`

	_, err := r.db.ExecContext(
		ctx,
		query,
		item.ID,
		item.WorldDropConfigID,
		item.ItemID,
		nullableFloat64Arg(item.DropRate),
		nullableIntArg(item.DropWeight),
		item.MinQuantity,
		item.MaxQuantity,
		nullableIntArg(item.MinLevel),
		nullableIntArg(item.MaxLevel),
		item.GuaranteedDrop,
		metadataBytes,
	)
	if err != nil {
		return fmt.Errorf("创建世界掉落物品失败: %w", err)
	}

	return nil
}

func (r *worldDropItemRepositoryImpl) Update(ctx context.Context, item *interfaces.WorldDropItem) error {
	metadataBytes := metadataToBytes(item.Metadata)
	query := `
UPDATE game_config.world_drop_items
SET item_id = $1,
    drop_rate = $2,
    drop_weight = $3,
    min_quantity = $4,
    max_quantity = $5,
    min_level = $6,
    max_level = $7,
    guaranteed_drop = $8,
    metadata = $9,
    updated_at = NOW(),
    deleted_at = NULL
WHERE id = $10 AND world_drop_config_id = $11`

	res, err := r.db.ExecContext(
		ctx,
		query,
		item.ItemID,
		nullableFloat64Arg(item.DropRate),
		nullableIntArg(item.DropWeight),
		item.MinQuantity,
		item.MaxQuantity,
		nullableIntArg(item.MinLevel),
		nullableIntArg(item.MaxLevel),
		item.GuaranteedDrop,
		metadataBytes,
		item.ID,
		item.WorldDropConfigID,
	)
	if err != nil {
		return fmt.Errorf("更新世界掉落物品失败: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("世界掉落物品不存在: %s", item.ID)
	}
	return nil
}

func (r *worldDropItemRepositoryImpl) SoftDelete(ctx context.Context, configID, itemEntryID string) error {
	query := `UPDATE game_config.world_drop_items SET deleted_at = NOW(), updated_at = NOW()
              WHERE id = $1 AND world_drop_config_id = $2 AND deleted_at IS NULL`
	res, err := r.db.ExecContext(ctx, query, itemEntryID, configID)
	if err != nil {
		return fmt.Errorf("删除世界掉落物品失败: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return fmt.Errorf("世界掉落物品不存在: %s", itemEntryID)
	}
	return nil
}

func (r *worldDropItemRepositoryImpl) ExistsActiveItem(ctx context.Context, itemID, excludeConfigID string) (bool, error) {
	query := `SELECT 1 FROM game_config.world_drop_items WHERE item_id = $1 AND deleted_at IS NULL`
	args := []interface{}{itemID}
	if excludeConfigID != "" {
		query += " AND world_drop_config_id <> $2"
		args = append(args, excludeConfigID)
	}
	query += " LIMIT 1"

	row := r.db.QueryRowContext(ctx, query, args...)
	var dummy int
	err := row.Scan(&dummy)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("检查世界掉落物品唯一性失败: %w", err)
	}
	return true, nil
}

func (r *worldDropItemRepositoryImpl) SumDropRates(ctx context.Context, configID string, excludeItemEntryID *string) (float64, error) {
	query := `SELECT COALESCE(SUM(drop_rate),0) FROM game_config.world_drop_items WHERE world_drop_config_id = $1 AND deleted_at IS NULL`
	args := []interface{}{configID}
	if excludeItemEntryID != nil && *excludeItemEntryID != "" {
		query += " AND id <> $2"
		args = append(args, *excludeItemEntryID)
	}

	var sum sql.NullFloat64
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&sum); err != nil {
		return 0, fmt.Errorf("统计世界掉落概率失败: %w", err)
	}
	if sum.Valid {
		return sum.Float64, nil
	}
	return 0, nil
}

func nullableFloat64Arg(v *float64) interface{} {
	if v == nil {
		return nil
	}
	return *v
}

func nullableIntArg(v *int) interface{} {
	if v == nil {
		return nil
	}
	return *v
}

func float64PtrFromNull(v sql.NullFloat64) *float64 {
	if !v.Valid {
		return nil
	}
	value := v.Float64
	return &value
}

func intPtrFromNull(v sql.NullInt64) *int {
	if !v.Valid {
		return nil
	}
	value := int(v.Int64)
	return &value
}

func metadataFromBytes(b []byte) json.RawMessage {
	if len(b) == 0 {
		return json.RawMessage("{}")
	}
	return json.RawMessage(b)
}

func metadataToBytes(raw json.RawMessage) []byte {
	if len(raw) == 0 {
		return []byte("{}")
	}
	return raw
}
