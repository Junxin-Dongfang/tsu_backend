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

type teamJoinRequestRepositoryImpl struct {
	db *sql.DB
}

// NewTeamJoinRequestRepository 创建团队加入申请仓储实例
func NewTeamJoinRequestRepository(db *sql.DB) interfaces.TeamJoinRequestRepository {
	return &teamJoinRequestRepositoryImpl{db: db}
}

// Create 创建加入申请
func (r *teamJoinRequestRepositoryImpl) Create(ctx context.Context, request *game_runtime.TeamJoinRequest) error {
	// 生成UUID
	if request.ID == "" {
		request.ID = uuid.New().String()
	}

	// 设置时间戳
	request.CreatedAt = time.Now()

	// 插入数据库
	if err := request.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("创建加入申请失败: %w", err)
	}

	return nil
}

// GetByID 根据ID获取申请
func (r *teamJoinRequestRepositoryImpl) GetByID(ctx context.Context, requestID string) (*game_runtime.TeamJoinRequest, error) {
	request, err := game_runtime.TeamJoinRequests(
		qm.Where("id = ?", requestID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("加入申请不存在: %s", requestID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询加入申请失败: %w", err)
	}

	return request, nil
}

// Update 更新申请
func (r *teamJoinRequestRepositoryImpl) Update(ctx context.Context, execer boil.ContextExecutor, request *game_runtime.TeamJoinRequest) error {
	if _, err := request.Update(ctx, execer, boil.Infer()); err != nil {
		return fmt.Errorf("更新加入申请失败: %w", err)
	}
	return nil
}

// ListPendingByTeam 查询团队的待审批申请列表
func (r *teamJoinRequestRepositoryImpl) ListPendingByTeam(ctx context.Context, teamID string) ([]*game_runtime.TeamJoinRequest, error) {
	requests, err := game_runtime.TeamJoinRequests(
		qm.Where("team_id = ? AND status = ?", teamID, "pending"),
		qm.OrderBy("created_at DESC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询待审批申请列表失败: %w", err)
	}

	return requests, nil
}

// GetPendingByHeroAndTeam 查询英雄对团队的待审批申请
func (r *teamJoinRequestRepositoryImpl) GetPendingByHeroAndTeam(ctx context.Context, heroID, teamID string) (*game_runtime.TeamJoinRequest, error) {
	request, err := game_runtime.TeamJoinRequests(
		qm.Where("hero_id = ? AND team_id = ? AND status = ?", heroID, teamID, "pending"),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, nil // 没有待审批申请，返回 nil 而不是错误
	}
	if err != nil {
		return nil, fmt.Errorf("查询待审批申请失败: %w", err)
	}

	return request, nil
}

// ListByHero 查询英雄的申请列表
func (r *teamJoinRequestRepositoryImpl) ListByHero(ctx context.Context, heroID string) ([]*game_runtime.TeamJoinRequest, error) {
	requests, err := game_runtime.TeamJoinRequests(
		qm.Where("hero_id = ?", heroID),
		qm.OrderBy("created_at DESC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询申请列表失败: %w", err)
	}

	return requests, nil
}

