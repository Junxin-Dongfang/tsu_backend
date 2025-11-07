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

type dungeonRepositoryImpl struct {
	db *sql.DB
}

// NewDungeonRepository 创建地城仓储实例
func NewDungeonRepository(db *sql.DB) interfaces.DungeonRepository {
	return &dungeonRepositoryImpl{db: db}
}

// GetByID 根据ID获取地城
func (r *dungeonRepositoryImpl) GetByID(ctx context.Context, dungeonID string) (*game_config.Dungeon, error) {
	dungeon, err := game_config.Dungeons(
		qm.Where("id = ? AND deleted_at IS NULL", dungeonID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("地城不存在: %s", dungeonID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询地城失败: %w", err)
	}

	return dungeon, nil
}

// GetByCode 根据代码获取地城
func (r *dungeonRepositoryImpl) GetByCode(ctx context.Context, code string) (*game_config.Dungeon, error) {
	dungeon, err := game_config.Dungeons(
		qm.Where("dungeon_code = ? AND deleted_at IS NULL", code),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("地城不存在: %s", code)
	}
	if err != nil {
		return nil, fmt.Errorf("查询地城失败: %w", err)
	}

	return dungeon, nil
}

// List 获取地城列表
func (r *dungeonRepositoryImpl) List(ctx context.Context, params interfaces.DungeonQueryParams) ([]*game_config.Dungeon, int64, error) {
	// 构建基础查询条件
	var baseQueryMods []qm.QueryMod
	baseQueryMods = append(baseQueryMods, qm.Where("deleted_at IS NULL"))

	// 筛选条件
	if params.DungeonCode != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("dungeon_code ILIKE ?", "%"+*params.DungeonCode+"%"))
	}
	if params.DungeonName != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("dungeon_name ILIKE ?", "%"+*params.DungeonName+"%"))
	}
	if params.MinLevel != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("max_level >= ?", *params.MinLevel))
	}
	if params.MaxLevel != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("min_level <= ?", *params.MaxLevel))
	}
	if params.IsActive != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("is_active = ?", *params.IsActive))
	}

	// 获取总数
	count, err := game_config.Dungeons(baseQueryMods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询地城总数失败: %w", err)
	}

	// 排序
	orderBy := "created_at"
	if params.OrderBy != "" {
		orderBy = params.OrderBy
	}
	if params.OrderDesc {
		baseQueryMods = append(baseQueryMods, qm.OrderBy(orderBy+" DESC"))
	} else {
		baseQueryMods = append(baseQueryMods, qm.OrderBy(orderBy+" ASC"))
	}

	// 分页
	if params.Limit > 0 {
		baseQueryMods = append(baseQueryMods, qm.Limit(params.Limit))
	}
	if params.Offset > 0 {
		baseQueryMods = append(baseQueryMods, qm.Offset(params.Offset))
	}

	// 查询列表
	dungeons, err := game_config.Dungeons(baseQueryMods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询地城列表失败: %w", err)
	}

	return dungeons, count, nil
}

// Create 创建地城
func (r *dungeonRepositoryImpl) Create(ctx context.Context, dungeon *game_config.Dungeon) error {
	// 生成UUID
	if dungeon.ID == "" {
		dungeon.ID = uuid.New().String()
	}

	// 设置时间戳
	now := time.Now()
	dungeon.CreatedAt = now
	dungeon.UpdatedAt = now

	// 插入数据库
	if err := dungeon.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("创建地城失败: %w", err)
	}

	return nil
}

// Update 更新地城
func (r *dungeonRepositoryImpl) Update(ctx context.Context, dungeon *game_config.Dungeon) error {
	// 更新时间戳
	dungeon.UpdatedAt = time.Now()

	// 更新数据库
	if _, err := dungeon.Update(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("更新地城失败: %w", err)
	}

	return nil
}

// Delete 软删除地城
func (r *dungeonRepositoryImpl) Delete(ctx context.Context, dungeonID string) error {
	// 查询地城
	dungeon, err := r.GetByID(ctx, dungeonID)
	if err != nil {
		return err
	}

	// 设置删除时间
	now := time.Now()
	dungeon.DeletedAt = null.TimeFrom(now)
	dungeon.UpdatedAt = now

	// 更新数据库
	if _, err := dungeon.Update(ctx, r.db, boil.Whitelist("deleted_at", "updated_at")); err != nil {
		return fmt.Errorf("删除地城失败: %w", err)
	}

	return nil
}

// Exists 检查代码是否存在
func (r *dungeonRepositoryImpl) Exists(ctx context.Context, code string) (bool, error) {
	count, err := game_config.Dungeons(
		qm.Where("dungeon_code = ? AND deleted_at IS NULL", code),
	).Count(ctx, r.db)

	if err != nil {
		return false, fmt.Errorf("检查地城代码是否存在失败: %w", err)
	}

	return count > 0, nil
}

// ExistsExcludingID 检查代码是否存在（排除指定ID）
func (r *dungeonRepositoryImpl) ExistsExcludingID(ctx context.Context, code string, excludeID string) (bool, error) {
	count, err := game_config.Dungeons(
		qm.Where("dungeon_code = ? AND id != ? AND deleted_at IS NULL", code, excludeID),
	).Count(ctx, r.db)

	if err != nil {
		return false, fmt.Errorf("检查地城代码是否存在失败: %w", err)
	}

	return count > 0, nil
}

