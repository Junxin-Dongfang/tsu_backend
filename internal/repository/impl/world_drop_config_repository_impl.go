package impl

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/interfaces"
)

type worldDropConfigRepositoryImpl struct {
	db *sql.DB
}

// NewWorldDropConfigRepository 创建世界掉落配置仓储实例
func NewWorldDropConfigRepository(db *sql.DB) interfaces.WorldDropConfigRepository {
	return &worldDropConfigRepositoryImpl{db: db}
}

// GetByID 根据ID获取世界掉落配置
func (r *worldDropConfigRepositoryImpl) GetByID(ctx context.Context, configID string) (*game_config.WorldDropConfig, error) {
	config, err := game_config.WorldDropConfigs(
		qm.Where("id = ? AND deleted_at IS NULL", configID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("世界掉落配置不存在: %s", configID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询世界掉落配置失败: %w", err)
	}

	return config, nil
}

// GetByItemID 根据物品ID获取世界掉落配置
func (r *worldDropConfigRepositoryImpl) GetByItemID(ctx context.Context, itemID string) (*game_config.WorldDropConfig, error) {
	config, err := game_config.WorldDropConfigs(
		qm.Where("item_id = ? AND deleted_at IS NULL", itemID),
		qm.Where("is_active = ?", true),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("世界掉落配置不存在: %s", itemID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询世界掉落配置失败: %w", err)
	}

	return config, nil
}

// GetActiveConfigs 获取所有激活的世界掉落配置
func (r *worldDropConfigRepositoryImpl) GetActiveConfigs(ctx context.Context) ([]*game_config.WorldDropConfig, error) {
	configs, err := game_config.WorldDropConfigs(
		qm.Where("deleted_at IS NULL"),
		qm.Where("is_active = ?", true),
		qm.OrderBy("base_drop_rate DESC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询激活的世界掉落配置失败: %w", err)
	}

	return configs, nil
}

// GetConfigsByTriggerCondition 根据触发条件获取世界掉落配置
func (r *worldDropConfigRepositoryImpl) GetConfigsByTriggerCondition(ctx context.Context, conditionType string, params map[string]interface{}) ([]*game_config.WorldDropConfig, error) {
	// 构建查询条件
	mods := []qm.QueryMod{
		qm.Where("deleted_at IS NULL"),
		qm.Where("is_active = ?", true),
	}

	// 根据条件类型构建JSON查询
	switch conditionType {
	case "level_range":
		// 查询等级范围内的配置
		// trigger_conditions->>'type' = 'level_range'
		// AND (trigger_conditions->>'min_level')::int <= playerLevel
		// AND (trigger_conditions->>'max_level')::int >= playerLevel
		if playerLevel, ok := params["player_level"].(int); ok {
			mods = append(mods,
				qm.Where("trigger_conditions->>'type' = ?", "level_range"),
				qm.Where("(trigger_conditions->>'min_level')::int <= ?", playerLevel),
				qm.Where("(trigger_conditions->>'max_level')::int >= ?", playerLevel),
			)
		}

	case "dungeon_type":
		// 查询特定地城类型的配置
		// trigger_conditions->>'type' = 'dungeon_type'
		// AND trigger_conditions->'dungeon_types' @> '["elite"]'
		if dungeonType, ok := params["dungeon_type"].(string); ok {
			mods = append(mods,
				qm.Where("trigger_conditions->>'type' = ?", "dungeon_type"),
				qm.Where("trigger_conditions->'dungeon_types' @> ?", fmt.Sprintf(`["%s"]`, dungeonType)),
			)
		}

	case "event":
		// 查询特定事件的配置
		// trigger_conditions->>'type' = 'event'
		// AND trigger_conditions->>'event_id' = eventID
		if eventID, ok := params["event_id"].(string); ok {
			mods = append(mods,
				qm.Where("trigger_conditions->>'type' = ?", "event"),
				qm.Where("trigger_conditions->>'event_id' = ?", eventID),
			)
		}

	default:
		return nil, fmt.Errorf("不支持的触发条件类型: %s", conditionType)
	}

	configs, err := game_config.WorldDropConfigs(mods...).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("查询世界掉落配置失败: %w", err)
	}

	return configs, nil
}

// List 查询世界掉落配置列表
func (r *worldDropConfigRepositoryImpl) List(ctx context.Context, params interfaces.ListWorldDropConfigParams) ([]*game_config.WorldDropConfig, int64, error) {
	// 构建查询条件
	mods := []qm.QueryMod{
		qm.Where("deleted_at IS NULL"),
	}

	// 添加筛选条件
	if params.ItemID != nil {
		mods = append(mods, qm.Where("item_id = ?", *params.ItemID))
	}
	if params.IsActive != nil {
		mods = append(mods, qm.Where("is_active = ?", *params.IsActive))
	}

	// 查询总数
	count, err := game_config.WorldDropConfigs(mods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询世界掉落配置总数失败: %w", err)
	}

	// 添加分页
	if params.PageSize > 0 {
		offset := (params.Page - 1) * params.PageSize
		mods = append(mods, qm.Limit(params.PageSize), qm.Offset(offset))
	}

	// 排序
	sortBy := params.SortBy
	if sortBy == "" {
		sortBy = "base_drop_rate"
	}
	sortOrder := params.SortOrder
	if sortOrder == "" {
		sortOrder = "DESC"
	}
	mods = append(mods, qm.OrderBy(fmt.Sprintf("%s %s, created_at DESC", sortBy, sortOrder)))

	// 查询列表
	configs, err := game_config.WorldDropConfigs(mods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询世界掉落配置列表失败: %w", err)
	}

	return configs, count, nil
}

// Create 创建世界掉落配置
func (r *worldDropConfigRepositoryImpl) Create(ctx context.Context, config *game_config.WorldDropConfig) error {
	return config.Insert(ctx, r.db, boil.Infer())
}

// Update 更新世界掉落配置
func (r *worldDropConfigRepositoryImpl) Update(ctx context.Context, config *game_config.WorldDropConfig) error {
	_, err := config.Update(ctx, r.db, boil.Infer())
	return err
}

// Delete 删除世界掉落配置（软删除）
func (r *worldDropConfigRepositoryImpl) Delete(ctx context.Context, configID string) error {
	config, err := r.GetByID(ctx, configID)
	if err != nil {
		return err
	}

	_, err = config.Delete(ctx, r.db, true) // soft delete
	return err
}

// Count 统计世界掉落配置数量
func (r *worldDropConfigRepositoryImpl) Count(ctx context.Context, params interfaces.ListWorldDropConfigParams) (int64, error) {
	mods := []qm.QueryMod{
		qm.Where("deleted_at IS NULL"),
	}

	if params.ItemID != nil {
		mods = append(mods, qm.Where("item_id = ?", *params.ItemID))
	}
	if params.IsActive != nil {
		mods = append(mods, qm.Where("is_active = ?", *params.IsActive))
	}

	return game_config.WorldDropConfigs(mods...).Count(ctx, r.db)
}

