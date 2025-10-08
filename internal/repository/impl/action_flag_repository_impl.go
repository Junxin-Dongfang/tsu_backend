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

type actionFlagRepositoryImpl struct {
	db *sql.DB
}

// NewActionFlagRepository 创建动作Flag仓储实例
func NewActionFlagRepository(db *sql.DB) interfaces.ActionFlagRepository {
	return &actionFlagRepositoryImpl{db: db}
}

func (r *actionFlagRepositoryImpl) GetByID(ctx context.Context, flagID string) (*game_config.ActionFlag, error) {
	flag, err := game_config.ActionFlags(
		qm.Where("id = ? AND deleted_at IS NULL", flagID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("动作Flag不存在: %s", flagID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询动作Flag失败: %w", err)
	}

	return flag, nil
}

func (r *actionFlagRepositoryImpl) GetByCode(ctx context.Context, code string) (*game_config.ActionFlag, error) {
	flag, err := game_config.ActionFlags(
		qm.Where("flag_code = ? AND deleted_at IS NULL", code),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("动作Flag不存在: %s", code)
	}
	if err != nil {
		return nil, fmt.Errorf("查询动作Flag失败: %w", err)
	}

	return flag, nil
}

func (r *actionFlagRepositoryImpl) List(ctx context.Context, params interfaces.ActionFlagQueryParams) ([]*game_config.ActionFlag, int64, error) {
	var baseQueryMods []qm.QueryMod
	baseQueryMods = append(baseQueryMods, qm.Where("deleted_at IS NULL"))

	if params.Category != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("category = ?", *params.Category))
	}
	if params.DurationType != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("duration_type = ?", *params.DurationType))
	}
	if params.IsActive != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("is_active = ?", *params.IsActive))
	}

	count, err := game_config.ActionFlags(baseQueryMods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询动作Flag总数失败: %w", err)
	}

	var queryMods []qm.QueryMod
	queryMods = append(queryMods, baseQueryMods...)
	queryMods = append(queryMods, qm.OrderBy("flag_code ASC"))

	if params.Limit > 0 {
		queryMods = append(queryMods, qm.Limit(params.Limit))
	}
	if params.Offset > 0 {
		queryMods = append(queryMods, qm.Offset(params.Offset))
	}

	flags, err := game_config.ActionFlags(queryMods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询动作Flag列表失败: %w", err)
	}

	return flags, count, nil
}

func (r *actionFlagRepositoryImpl) Create(ctx context.Context, flag *game_config.ActionFlag) error {
	if flag.ID == "" {
		flag.ID = uuid.New().String()
	}

	now := time.Now()
	flag.CreatedAt.SetValid(now)
	flag.UpdatedAt.SetValid(now)

	if err := flag.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("创建动作Flag失败: %w", err)
	}

	return nil
}

func (r *actionFlagRepositoryImpl) Update(ctx context.Context, flag *game_config.ActionFlag) error {
	flag.UpdatedAt.SetValid(time.Now())

	if _, err := flag.Update(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("更新动作Flag失败: %w", err)
	}

	return nil
}

func (r *actionFlagRepositoryImpl) Delete(ctx context.Context, flagID string) error {
	flag, err := r.GetByID(ctx, flagID)
	if err != nil {
		return err
	}

	now := time.Now()
	flag.DeletedAt.SetValid(now)
	flag.UpdatedAt.SetValid(now)

	if _, err := flag.Update(ctx, r.db, boil.Whitelist("deleted_at", "updated_at")); err != nil {
		return fmt.Errorf("删除动作Flag失败: %w", err)
	}

	return nil
}

func (r *actionFlagRepositoryImpl) Exists(ctx context.Context, code string) (bool, error) {
	count, err := game_config.ActionFlags(
		qm.Where("flag_code = ? AND deleted_at IS NULL", code),
	).Count(ctx, r.db)

	if err != nil {
		return false, fmt.Errorf("检查动作Flag代码是否存在失败: %w", err)
	}

	return count > 0, nil
}
