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

type actionRepositoryImpl struct {
	db *sql.DB
}

// NewActionRepository 创建动作仓储实例
func NewActionRepository(db *sql.DB) interfaces.ActionRepository {
	return &actionRepositoryImpl{db: db}
}

func (r *actionRepositoryImpl) GetByID(ctx context.Context, actionID string) (*game_config.Action, error) {
	action, err := game_config.Actions(
		qm.Where("id = ? AND deleted_at IS NULL", actionID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("动作不存在: %s", actionID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询动作失败: %w", err)
	}

	return action, nil
}

func (r *actionRepositoryImpl) GetByCode(ctx context.Context, code string) (*game_config.Action, error) {
	action, err := game_config.Actions(
		qm.Where("action_code = ? AND deleted_at IS NULL", code),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("动作不存在: %s", code)
	}
	if err != nil {
		return nil, fmt.Errorf("查询动作失败: %w", err)
	}

	return action, nil
}

func (r *actionRepositoryImpl) List(ctx context.Context, params interfaces.ActionQueryParams) ([]*game_config.Action, int64, error) {
	var baseQueryMods []qm.QueryMod
	baseQueryMods = append(baseQueryMods, qm.Where("deleted_at IS NULL"))

	if params.ActionType != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("action_type = ?", *params.ActionType))
	}
	if params.CategoryID != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("action_category_id = ?", *params.CategoryID))
	}
	if params.IsActive != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("is_active = ?", *params.IsActive))
	}

	count, err := game_config.Actions(baseQueryMods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询动作总数失败: %w", err)
	}

	var queryMods []qm.QueryMod
	queryMods = append(queryMods, baseQueryMods...)
	queryMods = append(queryMods, qm.OrderBy("action_code ASC"))

	if params.Limit > 0 {
		queryMods = append(queryMods, qm.Limit(params.Limit))
	}
	if params.Offset > 0 {
		queryMods = append(queryMods, qm.Offset(params.Offset))
	}

	actions, err := game_config.Actions(queryMods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询动作列表失败: %w", err)
	}

	return actions, count, nil
}

func (r *actionRepositoryImpl) Create(ctx context.Context, action *game_config.Action) error {
	if action.ID == "" {
		action.ID = uuid.New().String()
	}

	now := time.Now()
	action.CreatedAt.SetValid(now)
	action.UpdatedAt.SetValid(now)

	if err := action.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("创建动作失败: %w", err)
	}

	return nil
}

func (r *actionRepositoryImpl) Update(ctx context.Context, action *game_config.Action) error {
	action.UpdatedAt.SetValid(time.Now())

	if _, err := action.Update(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("更新动作失败: %w", err)
	}

	return nil
}

func (r *actionRepositoryImpl) Delete(ctx context.Context, actionID string) error {
	action, err := r.GetByID(ctx, actionID)
	if err != nil {
		return err
	}

	now := time.Now()
	action.DeletedAt.SetValid(now)
	action.UpdatedAt.SetValid(now)

	if _, err := action.Update(ctx, r.db, boil.Whitelist("deleted_at", "updated_at")); err != nil {
		return fmt.Errorf("删除动作失败: %w", err)
	}

	return nil
}

func (r *actionRepositoryImpl) Exists(ctx context.Context, code string) (bool, error) {
	count, err := game_config.Actions(
		qm.Where("action_code = ? AND deleted_at IS NULL", code),
	).Count(ctx, r.db)

	if err != nil {
		return false, fmt.Errorf("检查动作代码是否存在失败: %w", err)
	}

	return count > 0, nil
}
