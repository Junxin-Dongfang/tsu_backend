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

type effectRepositoryImpl struct {
	db *sql.DB
}

// NewEffectRepository 创建效果仓储实例
func NewEffectRepository(db *sql.DB) interfaces.EffectRepository {
	return &effectRepositoryImpl{db: db}
}

func (r *effectRepositoryImpl) GetByID(ctx context.Context, effectID string) (*game_config.Effect, error) {
	effect, err := game_config.Effects(
		qm.Where("id = ? AND deleted_at IS NULL", effectID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("效果不存在: %s", effectID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询效果失败: %w", err)
	}

	return effect, nil
}

func (r *effectRepositoryImpl) GetByCode(ctx context.Context, code string) (*game_config.Effect, error) {
	effect, err := game_config.Effects(
		qm.Where("effect_code = ? AND deleted_at IS NULL", code),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("效果不存在: %s", code)
	}
	if err != nil {
		return nil, fmt.Errorf("查询效果失败: %w", err)
	}

	return effect, nil
}

func (r *effectRepositoryImpl) List(ctx context.Context, params interfaces.EffectQueryParams) ([]*game_config.Effect, int64, error) {
	var baseQueryMods []qm.QueryMod
	baseQueryMods = append(baseQueryMods, qm.Where("deleted_at IS NULL"))

	if params.EffectType != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("effect_type = ?", *params.EffectType))
	}
	if params.IsActive != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("is_active = ?", *params.IsActive))
	}

	count, err := game_config.Effects(baseQueryMods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询效果总数失败: %w", err)
	}

	var queryMods []qm.QueryMod
	queryMods = append(queryMods, baseQueryMods...)
	queryMods = append(queryMods, qm.OrderBy("effect_code ASC"))

	if params.Limit > 0 {
		queryMods = append(queryMods, qm.Limit(params.Limit))
	}
	if params.Offset > 0 {
		queryMods = append(queryMods, qm.Offset(params.Offset))
	}

	effects, err := game_config.Effects(queryMods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询效果列表失败: %w", err)
	}

	return effects, count, nil
}

func (r *effectRepositoryImpl) Create(ctx context.Context, effect *game_config.Effect) error {
	if effect.ID == "" {
		effect.ID = uuid.New().String()
	}

	now := time.Now()
	effect.CreatedAt.SetValid(now)
	effect.UpdatedAt.SetValid(now)

	if err := effect.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("创建效果失败: %w", err)
	}

	return nil
}

func (r *effectRepositoryImpl) Update(ctx context.Context, effect *game_config.Effect) error {
	effect.UpdatedAt.SetValid(time.Now())

	if _, err := effect.Update(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("更新效果失败: %w", err)
	}

	return nil
}

func (r *effectRepositoryImpl) Delete(ctx context.Context, effectID string) error {
	effect, err := r.GetByID(ctx, effectID)
	if err != nil {
		return err
	}

	now := time.Now()
	effect.DeletedAt.SetValid(now)
	effect.UpdatedAt.SetValid(now)

	if _, err := effect.Update(ctx, r.db, boil.Whitelist("deleted_at", "updated_at")); err != nil {
		return fmt.Errorf("删除效果失败: %w", err)
	}

	return nil
}

func (r *effectRepositoryImpl) Exists(ctx context.Context, code string) (bool, error) {
	count, err := game_config.Effects(
		qm.Where("effect_code = ? AND deleted_at IS NULL", code),
	).Count(ctx, r.db)

	if err != nil {
		return false, fmt.Errorf("检查效果代码是否存在失败: %w", err)
	}

	return count > 0, nil
}
