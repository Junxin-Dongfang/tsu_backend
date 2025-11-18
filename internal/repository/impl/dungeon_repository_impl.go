package impl

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

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

// buildListMods 构造查询条件
func buildListMods(params interfaces.DungeonQueryParams) []qm.QueryMod {
	mods := []qm.QueryMod{qm.Where("deleted_at IS NULL")}

	if params.DungeonCode != nil {
		mods = append(mods, qm.Where("dungeon_code ILIKE ?", "%"+*params.DungeonCode+"%"))
	}
	if params.DungeonName != nil {
		mods = append(mods, qm.Where("dungeon_name ILIKE ?", "%"+*params.DungeonName+"%"))
	}
	if params.MinLevel != nil {
		mods = append(mods, qm.Where("max_level >= ?", *params.MinLevel))
	}
	if params.MaxLevel != nil {
		mods = append(mods, qm.Where("min_level <= ?", *params.MaxLevel))
	}
	if params.IsActive != nil {
		mods = append(mods, qm.Where("is_active = ?", *params.IsActive))
	}

	return mods
}

// List 获取地城列表
func (r *dungeonRepositoryImpl) List(ctx context.Context, params interfaces.DungeonQueryParams) ([]*game_config.Dungeon, int64, error) {
	mods := buildListMods(params)

	count, err := game_config.Dungeons(mods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("统计地城数量失败: %w", err)
	}

	allowedOrders := map[string]string{
		"created_at":   "created_at",
		"updated_at":   "updated_at",
		"dungeon_code": "dungeon_code",
	}
	orderColumn := "created_at"
	if val, ok := allowedOrders[strings.ToLower(params.OrderBy)]; ok {
		orderColumn = val
	}
	orderDir := "DESC"
	if !params.OrderDesc {
		orderDir = "ASC"
	}
	orderBy := fmt.Sprintf("%s %s", orderColumn, orderDir)

	listMods := append(mods, qm.OrderBy(orderBy))
	if params.Limit > 0 {
		listMods = append(listMods, qm.Limit(params.Limit))
	}
	if params.Offset > 0 {
		listMods = append(listMods, qm.Offset(params.Offset))
	}

	dungeons, err := game_config.Dungeons(listMods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询地城失败: %w", err)
	}

	return dungeons, count, nil
}

// GetByID 根据 ID 获取地城配置
func (r *dungeonRepositoryImpl) GetByID(ctx context.Context, dungeonID string) (*game_config.Dungeon, error) {
	dungeon, err := game_config.Dungeons(
		qm.Where("id = ?", dungeonID),
		qm.Where("deleted_at IS NULL"),
	).One(ctx, r.db)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("地城不存在 (id=%s)", dungeonID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询地城失败: %w", err)
	}
	return dungeon, nil
}

// GetByCode 根据代码获取地城
func (r *dungeonRepositoryImpl) GetByCode(ctx context.Context, code string) (*game_config.Dungeon, error) {
	dungeon, err := game_config.Dungeons(
		qm.Where("dungeon_code = ?", code),
		qm.Where("deleted_at IS NULL"),
	).One(ctx, r.db)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("地城不存在 (code=%s)", code)
	}
	if err != nil {
		return nil, fmt.Errorf("查询地城失败: %w", err)
	}
	return dungeon, nil
}

// Exists 检查地城代码是否存在
func (r *dungeonRepositoryImpl) Exists(ctx context.Context, code string) (bool, error) {
	return game_config.Dungeons(
		qm.Where("dungeon_code = ?", code),
		qm.Where("deleted_at IS NULL"),
	).Exists(ctx, r.db)
}

// Create 创建地城
func (r *dungeonRepositoryImpl) Create(ctx context.Context, dungeon *game_config.Dungeon) error {
	if err := dungeon.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("创建地城失败: %w", err)
	}
	return nil
}

// Update 更新地城
func (r *dungeonRepositoryImpl) Update(ctx context.Context, dungeon *game_config.Dungeon) error {
	if _, err := dungeon.Update(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("更新地城失败: %w", err)
	}
	return nil
}

// Delete 软删除地城
func (r *dungeonRepositoryImpl) Delete(ctx context.Context, dungeonID string) error {
	dungeon, err := r.GetByID(ctx, dungeonID)
	if err != nil {
		return err
	}
	dungeon.DeletedAt = null.TimeFrom(time.Now())
	if _, err := dungeon.Update(ctx, r.db, boil.Whitelist("deleted_at")); err != nil {
		return fmt.Errorf("删除地城失败: %w", err)
	}
	return nil
}
