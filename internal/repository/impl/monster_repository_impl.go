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

type monsterRepositoryImpl struct {
	db *sql.DB
}

// NewMonsterRepository 创建怪物仓储实例
func NewMonsterRepository(db *sql.DB) interfaces.MonsterRepository {
	return &monsterRepositoryImpl{db: db}
}

// GetByID 根据ID获取怪物
func (r *monsterRepositoryImpl) GetByID(ctx context.Context, monsterID string) (*game_config.Monster, error) {
	monster, err := game_config.Monsters(
		qm.Where("id = ? AND deleted_at IS NULL", monsterID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("怪物不存在: %s", monsterID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询怪物失败: %w", err)
	}

	return monster, nil
}

// GetByCode 根据代码获取怪物
func (r *monsterRepositoryImpl) GetByCode(ctx context.Context, code string) (*game_config.Monster, error) {
	monster, err := game_config.Monsters(
		qm.Where("monster_code = ? AND deleted_at IS NULL", code),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("怪物不存在: %s", code)
	}
	if err != nil {
		return nil, fmt.Errorf("查询怪物失败: %w", err)
	}

	return monster, nil
}

// List 获取怪物列表
func (r *monsterRepositoryImpl) List(ctx context.Context, params interfaces.MonsterQueryParams) ([]*game_config.Monster, int64, error) {
	// 构建基础查询条件
	var baseQueryMods []qm.QueryMod
	baseQueryMods = append(baseQueryMods, qm.Where("deleted_at IS NULL"))

	// 筛选条件
	if params.MonsterCode != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("monster_code ILIKE ?", "%"+*params.MonsterCode+"%"))
	}
	if params.MonsterName != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("monster_name ILIKE ?", "%"+*params.MonsterName+"%"))
	}
	if params.MinLevel != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("monster_level >= ?", *params.MinLevel))
	}
	if params.MaxLevel != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("monster_level <= ?", *params.MaxLevel))
	}
	if params.IsActive != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("is_active = ?", *params.IsActive))
	}

	// 获取总数
	count, err := game_config.Monsters(baseQueryMods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询怪物总数失败: %w", err)
	}

	// 排序
	orderBy := "monster_level"
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
	monsters, err := game_config.Monsters(baseQueryMods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询怪物列表失败: %w", err)
	}

	return monsters, count, nil
}

// Create 创建怪物
func (r *monsterRepositoryImpl) Create(ctx context.Context, monster *game_config.Monster) error {
	// 生成UUID
	if monster.ID == "" {
		monster.ID = uuid.New().String()
	}

	// 设置时间戳
	now := time.Now()
	monster.CreatedAt = now
	monster.UpdatedAt = now

	// 插入数据库
	if err := monster.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("创建怪物失败: %w", err)
	}

	return nil
}

// Update 更新怪物
func (r *monsterRepositoryImpl) Update(ctx context.Context, monster *game_config.Monster) error {
	// 更新时间戳
	monster.UpdatedAt = time.Now()

	// 更新数据库
	if _, err := monster.Update(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("更新怪物失败: %w", err)
	}

	return nil
}

// Delete 软删除怪物
func (r *monsterRepositoryImpl) Delete(ctx context.Context, monsterID string) error {
	// 查询怪物
	monster, err := r.GetByID(ctx, monsterID)
	if err != nil {
		return err
	}

	// 设置删除时间
	now := time.Now()
	monster.DeletedAt = null.TimeFrom(now)
	monster.UpdatedAt = now

	// 更新数据库
	if _, err := monster.Update(ctx, r.db, boil.Whitelist("deleted_at", "updated_at")); err != nil {
		return fmt.Errorf("删除怪物失败: %w", err)
	}

	return nil
}

// Exists 检查代码是否存在
func (r *monsterRepositoryImpl) Exists(ctx context.Context, code string) (bool, error) {
	count, err := game_config.Monsters(
		qm.Where("monster_code = ? AND deleted_at IS NULL", code),
	).Count(ctx, r.db)

	if err != nil {
		return false, fmt.Errorf("检查怪物代码是否存在失败: %w", err)
	}

	return count > 0, nil
}

// ExistsExcludingID 检查代码是否存在（排除指定ID）
func (r *monsterRepositoryImpl) ExistsExcludingID(ctx context.Context, code string, excludeID string) (bool, error) {
	count, err := game_config.Monsters(
		qm.Where("monster_code = ? AND id != ? AND deleted_at IS NULL", code, excludeID),
	).Count(ctx, r.db)

	if err != nil {
		return false, fmt.Errorf("检查怪物代码是否存在失败: %w", err)
	}

	return count > 0, nil
}

