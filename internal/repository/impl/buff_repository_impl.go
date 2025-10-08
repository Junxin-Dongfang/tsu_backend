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

type buffRepositoryImpl struct {
	db *sql.DB
}

// NewBuffRepository 创建Buff仓储实例
func NewBuffRepository(db *sql.DB) interfaces.BuffRepository {
	return &buffRepositoryImpl{db: db}
}

func (r *buffRepositoryImpl) GetByID(ctx context.Context, buffID string) (*game_config.Buff, error) {
	buff, err := game_config.Buffs(
		qm.Where("id = ? AND deleted_at IS NULL", buffID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("Buff不存在: %s", buffID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询Buff失败: %w", err)
	}

	return buff, nil
}

func (r *buffRepositoryImpl) GetByCode(ctx context.Context, code string) (*game_config.Buff, error) {
	buff, err := game_config.Buffs(
		qm.Where("buff_code = ? AND deleted_at IS NULL", code),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("Buff不存在: %s", code)
	}
	if err != nil {
		return nil, fmt.Errorf("查询Buff失败: %w", err)
	}

	return buff, nil
}

func (r *buffRepositoryImpl) List(ctx context.Context, params interfaces.BuffQueryParams) ([]*game_config.Buff, int64, error) {
	var baseQueryMods []qm.QueryMod
	baseQueryMods = append(baseQueryMods, qm.Where("deleted_at IS NULL"))

	if params.BuffType != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("buff_type = ?", *params.BuffType))
	}
	if params.Category != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("category = ?", *params.Category))
	}
	if params.IsActive != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("is_active = ?", *params.IsActive))
	}

	count, err := game_config.Buffs(baseQueryMods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询Buff总数失败: %w", err)
	}

	var queryMods []qm.QueryMod
	queryMods = append(queryMods, baseQueryMods...)
	queryMods = append(queryMods, qm.OrderBy("buff_code ASC"))

	if params.Limit > 0 {
		queryMods = append(queryMods, qm.Limit(params.Limit))
	}
	if params.Offset > 0 {
		queryMods = append(queryMods, qm.Offset(params.Offset))
	}

	buffs, err := game_config.Buffs(queryMods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询Buff列表失败: %w", err)
	}

	return buffs, count, nil
}

func (r *buffRepositoryImpl) Create(ctx context.Context, buff *game_config.Buff) error {
	if buff.ID == "" {
		buff.ID = uuid.New().String()
	}

	now := time.Now()
	buff.CreatedAt.SetValid(now)
	buff.UpdatedAt.SetValid(now)

	if err := buff.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("创建Buff失败: %w", err)
	}

	return nil
}

func (r *buffRepositoryImpl) Update(ctx context.Context, buff *game_config.Buff) error {
	buff.UpdatedAt.SetValid(time.Now())

	if _, err := buff.Update(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("更新Buff失败: %w", err)
	}

	return nil
}

func (r *buffRepositoryImpl) Delete(ctx context.Context, buffID string) error {
	buff, err := r.GetByID(ctx, buffID)
	if err != nil {
		return err
	}

	now := time.Now()
	buff.DeletedAt.SetValid(now)
	buff.UpdatedAt.SetValid(now)

	if _, err := buff.Update(ctx, r.db, boil.Whitelist("deleted_at", "updated_at")); err != nil {
		return fmt.Errorf("删除Buff失败: %w", err)
	}

	return nil
}

func (r *buffRepositoryImpl) Exists(ctx context.Context, code string) (bool, error) {
	count, err := game_config.Buffs(
		qm.Where("buff_code = ? AND deleted_at IS NULL", code),
	).Count(ctx, r.db)

	if err != nil {
		return false, fmt.Errorf("检查Buff代码是否存在失败: %w", err)
	}

	return count > 0, nil
}
