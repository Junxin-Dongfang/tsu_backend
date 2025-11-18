package impl

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/google/uuid"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/repository/interfaces"
)

type teamMemberRepositoryImpl struct {
	db *sql.DB
}

// NewTeamMemberRepository 创建团队成员仓储实例
func NewTeamMemberRepository(db *sql.DB) interfaces.TeamMemberRepository {
	return &teamMemberRepositoryImpl{db: db}
}

// Create 创建团队成员记录
func (r *teamMemberRepositoryImpl) Create(ctx context.Context, execer boil.ContextExecutor, member *game_runtime.TeamMember) error {
	// 生成UUID
	if member.ID == "" {
		member.ID = uuid.New().String()
	}

	// 设置时间戳
	now := time.Now()
	member.JoinedAt = now
	member.LastActiveAt = now

	// 插入数据库
	if err := member.Insert(ctx, execer, boil.Infer()); err != nil {
		return fmt.Errorf("创建团队成员记录失败: %w", err)
	}

	return nil
}

// GetByID 根据ID获取成员记录
func (r *teamMemberRepositoryImpl) GetByID(ctx context.Context, memberID string) (*game_runtime.TeamMember, error) {
	member, err := game_runtime.TeamMembers(
		qm.Where("id = ?", memberID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("团队成员记录不存在: %s", memberID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询团队成员记录失败: %w", err)
	}

	return member, nil
}

// GetByTeamAndHero 根据团队ID和英雄ID获取成员记录
func (r *teamMemberRepositoryImpl) GetByTeamAndHero(ctx context.Context, teamID, heroID string) (*game_runtime.TeamMember, error) {
	member, err := game_runtime.TeamMembers(
		qm.Where("team_id = ? AND hero_id = ?", teamID, heroID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, interfaces.ErrTeamMemberNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("查询团队成员记录失败: %w", err)
	}

	return member, nil
}

// GetLeaderTeam 查询英雄担任队长的团队
func (r *teamMemberRepositoryImpl) GetLeaderTeam(ctx context.Context, heroID string) (*game_runtime.TeamMember, error) {
	member, err := game_runtime.TeamMembers(
		qm.Where("hero_id = ? AND role = ?", heroID, "leader"),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, nil // 未担任队长，返回 nil 而不是错误
	}
	if err != nil {
		return nil, fmt.Errorf("查询队长团队失败: %w", err)
	}

	return member, nil
}

// ListByTeam 查询团队成员列表
func (r *teamMemberRepositoryImpl) ListByTeam(ctx context.Context, teamID string) ([]*game_runtime.TeamMember, error) {
	members, err := game_runtime.TeamMembers(
		qm.Where("team_id = ?", teamID),
		qm.OrderBy("joined_at ASC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询团队成员列表失败: %w", err)
	}

	return members, nil
}

// ListByHero 查询英雄加入的团队列表
func (r *teamMemberRepositoryImpl) ListByHero(ctx context.Context, heroID string) ([]*game_runtime.TeamMember, error) {
	members, err := game_runtime.TeamMembers(
		qm.Where("hero_id = ?", heroID),
		qm.OrderBy("joined_at DESC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询英雄团队列表失败: %w", err)
	}

	return members, nil
}

// ListAll 查询所有团队成员
func (r *teamMemberRepositoryImpl) ListAll(ctx context.Context) ([]*game_runtime.TeamMember, error) {
	members, err := game_runtime.TeamMembers().All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("查询团队成员失败: %w", err)
	}
	return members, nil
}

// Update 更新成员信息
func (r *teamMemberRepositoryImpl) Update(ctx context.Context, execer boil.ContextExecutor, member *game_runtime.TeamMember) error {
	if _, err := member.Update(ctx, execer, boil.Infer()); err != nil {
		return fmt.Errorf("更新团队成员记录失败: %w", err)
	}
	return nil
}

// Delete 删除成员记录
func (r *teamMemberRepositoryImpl) Delete(ctx context.Context, execer boil.ContextExecutor, memberID string) error {
	member, err := r.GetByID(ctx, memberID)
	if err != nil {
		return err
	}

	if _, err := member.Delete(ctx, execer); err != nil {
		return fmt.Errorf("删除团队成员记录失败: %w", err)
	}

	return nil
}

// UpdateRole 更新成员角色
func (r *teamMemberRepositoryImpl) UpdateRole(ctx context.Context, execer boil.ContextExecutor, teamID, heroID, role string) error {
	member, err := r.GetByTeamAndHero(ctx, teamID, heroID)
	if err != nil {
		return err
	}

	member.Role = role

	if _, err := member.Update(ctx, execer, boil.Whitelist("role")); err != nil {
		return fmt.Errorf("更新成员角色失败: %w", err)
	}

	return nil
}

// UpdateLastActive 更新最后活跃时间
func (r *teamMemberRepositoryImpl) UpdateLastActive(ctx context.Context, execer boil.ContextExecutor, teamID, heroID string) error {
	member, err := r.GetByTeamAndHero(ctx, teamID, heroID)
	if err != nil {
		return err
	}

	member.LastActiveAt = time.Now()

	if _, err := member.Update(ctx, execer, boil.Whitelist("last_active_at")); err != nil {
		return fmt.Errorf("更新最后活跃时间失败: %w", err)
	}

	return nil
}

// GetEarliestAdmin 查询最早加入的管理员
func (r *teamMemberRepositoryImpl) GetEarliestAdmin(ctx context.Context, teamID string) (*game_runtime.TeamMember, error) {
	member, err := game_runtime.TeamMembers(
		qm.Where("team_id = ? AND role = ?", teamID, "admin"),
		qm.OrderBy("joined_at ASC"),
		qm.Limit(1),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, nil // 没有管理员，返回 nil 而不是错误
	}
	if err != nil {
		return nil, fmt.Errorf("查询最早管理员失败: %w", err)
	}

	return member, nil
}

// GetEarliestMember 查询最早加入的普通成员
func (r *teamMemberRepositoryImpl) GetEarliestMember(ctx context.Context, teamID string) (*game_runtime.TeamMember, error) {
	member, err := game_runtime.TeamMembers(
		qm.Where("team_id = ? AND role = ?", teamID, "member"),
		qm.OrderBy("joined_at ASC"),
		qm.Limit(1),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, nil // 没有普通成员，返回 nil 而不是错误
	}
	if err != nil {
		return nil, fmt.Errorf("查询最早成员失败: %w", err)
	}

	return member, nil
}

// CountByTeam 统计团队成员数量
func (r *teamMemberRepositoryImpl) CountByTeam(ctx context.Context, teamID string) (int64, error) {
	count, err := game_runtime.TeamMembers(
		qm.Where("team_id = ?", teamID),
	).Count(ctx, r.db)

	if err != nil {
		return 0, fmt.Errorf("统计团队成员数量失败: %w", err)
	}

	return count, nil
}

// GetByIDForUpdate 根据ID获取成员记录（带行锁）
func (r *teamMemberRepositoryImpl) GetByIDForUpdate(ctx context.Context, tx *sql.Tx, memberID string) (*game_runtime.TeamMember, error) {
	member, err := game_runtime.TeamMembers(
		qm.Where("id = ?", memberID),
		qm.For("UPDATE"),
	).One(ctx, tx)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("团队成员记录不存在: %s", memberID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询团队成员记录失败（带锁）: %w", err)
	}

	return member, nil
}
