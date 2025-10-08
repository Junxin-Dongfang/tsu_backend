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

type skillLevelConfigRepositoryImpl struct {
	db *sql.DB
}

// NewSkillLevelConfigRepository 创建技能等级配置仓储实例
func NewSkillLevelConfigRepository(db *sql.DB) interfaces.SkillLevelConfigRepository {
	return &skillLevelConfigRepositoryImpl{db: db}
}

func (r *skillLevelConfigRepositoryImpl) GetByID(ctx context.Context, id string) (*game_config.SkillLevelConfig, error) {
	config, err := game_config.SkillLevelConfigs(
		qm.Where("id = ? AND deleted_at IS NULL", id),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("技能等级配置不存在: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("查询技能等级配置失败: %w", err)
	}

	return config, nil
}

func (r *skillLevelConfigRepositoryImpl) GetBySkillIDAndLevel(ctx context.Context, skillID string, levelNumber int) (*game_config.SkillLevelConfig, error) {
	config, err := game_config.SkillLevelConfigs(
		qm.Where("skill_id = ? AND level_number = ? AND deleted_at IS NULL", skillID, levelNumber),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("技能等级配置不存在: skillID=%s, level=%d", skillID, levelNumber)
	}
	if err != nil {
		return nil, fmt.Errorf("查询技能等级配置失败: %w", err)
	}

	return config, nil
}

func (r *skillLevelConfigRepositoryImpl) ListBySkillID(ctx context.Context, skillID string) ([]*game_config.SkillLevelConfig, error) {
	configs, err := game_config.SkillLevelConfigs(
		qm.Where("skill_id = ? AND deleted_at IS NULL", skillID),
		qm.OrderBy("level_number ASC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询技能等级配置列表失败: %w", err)
	}

	return configs, nil
}

func (r *skillLevelConfigRepositoryImpl) List(ctx context.Context, params interfaces.SkillLevelConfigQueryParams) ([]*game_config.SkillLevelConfig, int64, error) {
	var baseQueryMods []qm.QueryMod
	baseQueryMods = append(baseQueryMods, qm.Where("deleted_at IS NULL"))

	if params.SkillID != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("skill_id = ?", *params.SkillID))
	}
	if params.LevelNumber != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("level_number = ?", *params.LevelNumber))
	}

	count, err := game_config.SkillLevelConfigs(baseQueryMods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询技能等级配置总数失败: %w", err)
	}

	var queryMods []qm.QueryMod
	queryMods = append(queryMods, baseQueryMods...)
	queryMods = append(queryMods, qm.OrderBy("skill_id ASC, level_number ASC"))

	if params.Limit > 0 {
		queryMods = append(queryMods, qm.Limit(params.Limit))
	}
	if params.Offset > 0 {
		queryMods = append(queryMods, qm.Offset(params.Offset))
	}

	configs, err := game_config.SkillLevelConfigs(queryMods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询技能等级配置列表失败: %w", err)
	}

	return configs, count, nil
}

func (r *skillLevelConfigRepositoryImpl) Create(ctx context.Context, config *game_config.SkillLevelConfig) error {
	if config.ID == "" {
		config.ID = uuid.New().String()
	}

	now := time.Now()
	config.CreatedAt.SetValid(now)
	config.UpdatedAt.SetValid(now)

	if err := config.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("创建技能等级配置失败: %w", err)
	}

	return nil
}

func (r *skillLevelConfigRepositoryImpl) Update(ctx context.Context, config *game_config.SkillLevelConfig) error {
	config.UpdatedAt.SetValid(time.Now())

	if _, err := config.Update(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("更新技能等级配置失败: %w", err)
	}

	return nil
}

func (r *skillLevelConfigRepositoryImpl) Delete(ctx context.Context, id string) error {
	config, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()
	config.DeletedAt.SetValid(now)
	config.UpdatedAt.SetValid(now)

	if _, err := config.Update(ctx, r.db, boil.Whitelist("deleted_at", "updated_at")); err != nil {
		return fmt.Errorf("删除技能等级配置失败: %w", err)
	}

	return nil
}

func (r *skillLevelConfigRepositoryImpl) DeleteBySkillID(ctx context.Context, skillID string) error {
	_, err := game_config.SkillLevelConfigs(
		qm.Where("skill_id = ? AND deleted_at IS NULL", skillID),
	).UpdateAll(ctx, r.db, game_config.M{
		"deleted_at": sql.NullTime{Time: time.Now(), Valid: true},
		"updated_at": sql.NullTime{Time: time.Now(), Valid: true},
	})

	if err != nil {
		return fmt.Errorf("删除技能所有等级配置失败: %w", err)
	}

	return nil
}
