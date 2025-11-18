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

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/repository/interfaces"
)

type teamRepositoryImpl struct {
	db *sql.DB
}

// NewTeamRepository 创建团队仓储实例
func NewTeamRepository(db *sql.DB) interfaces.TeamRepository {
	return &teamRepositoryImpl{db: db}
}

// Create 创建团队
func (r *teamRepositoryImpl) Create(ctx context.Context, team *game_runtime.Team) error {
	// 生成UUID
	if team.ID == "" {
		team.ID = uuid.New().String()
	}

	// 设置时间戳
	now := time.Now()
	team.CreatedAt = now
	team.UpdatedAt = now

	// 插入数据库
	if err := team.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("创建团队失败: %w", err)
	}

	return nil
}

// GetByID 根据ID获取团队
func (r *teamRepositoryImpl) GetByID(ctx context.Context, teamID string) (*game_runtime.Team, error) {
	team, err := game_runtime.Teams(
		qm.Where("id = ? AND deleted_at IS NULL", teamID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("团队不存在: %s", teamID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询团队失败: %w", err)
	}

	return team, nil
}

// Update 更新团队信息
func (r *teamRepositoryImpl) Update(ctx context.Context, team *game_runtime.Team) error {
	// 更新时间戳
	team.UpdatedAt = time.Now()

	// 更新数据库
	if _, err := team.Update(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("更新团队失败: %w", err)
	}

	return nil
}

// Delete 软删除团队
func (r *teamRepositoryImpl) Delete(ctx context.Context, teamID string) error {
	team, err := r.GetByID(ctx, teamID)
	if err != nil {
		return err
	}

	team.DeletedAt = null.TimeFromPtr(nullTimeNow())

	if _, err := team.Update(ctx, r.db, boil.Whitelist("deleted_at", "updated_at")); err != nil {
		return fmt.Errorf("软删除团队失败: %w", err)
	}

	return nil
}

// GetInactiveLeaderTeams 查询队长超过指定时间未活跃的团队
func (r *teamRepositoryImpl) GetInactiveLeaderTeams(ctx context.Context, inactiveDuration time.Duration) ([]*game_runtime.Team, error) {
	// 计算截止时间
	cutoffTime := time.Now().Add(-inactiveDuration)

	// 查询：加入 team_members 表，找出队长最后活跃时间早于截止时间的团队
	teams, err := game_runtime.Teams(
		qm.InnerJoin("game_runtime.team_members ON teams.id = team_members.team_id"),
		qm.Where("teams.deleted_at IS NULL"),
		qm.Where("team_members.role = ?", "leader"),
		qm.Where("team_members.last_active_at < ?", cutoffTime),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询不活跃队长的团队失败: %w", err)
	}

	return teams, nil
}

// List 分页查询团队列表
func (r *teamRepositoryImpl) List(ctx context.Context, params interfaces.TeamQueryParams) ([]*game_runtime.Team, int64, error) {
	// 构建查询条件
	queryMods := []qm.QueryMod{
		qm.Where("deleted_at IS NULL"),
	}

	// 名称模糊查询
	if params.Name != "" {
		queryMods = append(queryMods, qm.Where("name ILIKE ?", "%"+params.Name+"%"))
	}

	// 统计总数
	count, err := game_runtime.Teams(queryMods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("统计团队数量失败: %w", err)
	}

	// 添加排序和分页
	queryMods = append(queryMods, qm.OrderBy("created_at DESC"))
	if params.Limit > 0 {
		queryMods = append(queryMods, qm.Limit(params.Limit))
	}
	if params.Offset > 0 {
		queryMods = append(queryMods, qm.Offset(params.Offset))
	}

	// 查询列表
	teams, err := game_runtime.Teams(queryMods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询团队列表失败: %w", err)
	}

	return teams, count, nil
}

// Exists 检查团队名称是否存在
func (r *teamRepositoryImpl) Exists(ctx context.Context, name string) (bool, error) {
	count, err := game_runtime.Teams(
		qm.Where("name = ? AND deleted_at IS NULL", name),
	).Count(ctx, r.db)

	if err != nil {
		return false, fmt.Errorf("检查团队名称是否存在失败: %w", err)
	}

	return count > 0, nil
}

