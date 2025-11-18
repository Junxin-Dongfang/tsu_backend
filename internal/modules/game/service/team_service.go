package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// TeamService 团队服务
type TeamService struct {
	db                    *sql.DB
	teamRepo              interfaces.TeamRepository
	teamMemberRepo        interfaces.TeamMemberRepository
	teamWarehouseRepo     interfaces.TeamWarehouseRepository
	heroRepo              interfaces.HeroRepository
	teamPermissionService *TeamPermissionService
}

// NewTeamService 创建团队服务
func NewTeamService(db *sql.DB, teamPermissionService *TeamPermissionService) *TeamService {
	return &TeamService{
		db:                    db,
		teamRepo:              impl.NewTeamRepository(db),
		teamMemberRepo:        impl.NewTeamMemberRepository(db),
		teamWarehouseRepo:     impl.NewTeamWarehouseRepository(db),
		heroRepo:              impl.NewHeroRepository(db),
		teamPermissionService: teamPermissionService,
	}
}

// CreateTeamRequest 创建团队请求
type CreateTeamRequest struct {
	UserID      string
	HeroID      string
	TeamName    string
	Description string
}

// CreateTeam 创建团队
func (s *TeamService) CreateTeam(ctx context.Context, req *CreateTeamRequest) (*game_runtime.Team, error) {
	// 1. 验证参数
	if req.UserID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "用户ID不能为空")
	}
	if req.HeroID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "英雄ID不能为空")
	}
	if req.TeamName == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "团队名称不能为空")
	}

	// 2. 验证英雄是否存在且属于当前用户
	hero, err := s.heroRepo.GetByID(ctx, req.HeroID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "英雄不存在")
	}
	if hero.UserID != req.UserID {
		return nil, xerrors.New(xerrors.CodePermissionDenied, "该英雄不属于当前用户")
	}

	// 3. 检查英雄是否已担任其他团队的队长
	existingLeader, err := s.teamMemberRepo.GetLeaderTeam(ctx, req.HeroID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "检查队长状态失败")
	}
	if existingLeader != nil {
		return nil, xerrors.New(xerrors.CodeDuplicateResource, "该英雄已是其他团队的队长")
	}

	// 4. 检查团队名称是否已存在
	exists, err := s.teamRepo.Exists(ctx, req.TeamName)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "检查团队名称失败")
	}
	if exists {
		return nil, xerrors.New(xerrors.CodeDuplicateResource, "团队名称已存在")
	}

	// 5. 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			// 仅当 Rollback 失败且不是已提交的事务时，才表示有问题
		}
	}()

	// 6. 创建团队
	team := &game_runtime.Team{
		Name:          req.TeamName,
		LeaderHeroID:  req.HeroID,
		MaxMembers:    12,
	}
	if req.Description != "" {
		team.Description.SetValid(req.Description)
	}

	if err := s.teamRepo.Create(ctx, team); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "创建团队失败")
	}

	// 7. 创建队长成员记录
	member := &game_runtime.TeamMember{
		TeamID: team.ID,
		HeroID: req.HeroID,
		UserID: req.UserID,
		Role:   "leader",
	}
	if err := s.teamMemberRepo.Create(ctx, tx, member); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "创建队长成员记录失败")
	}

	// 8. 创建团队仓库
	warehouse := &game_runtime.TeamWarehouse{
		TeamID:     team.ID,
		GoldAmount: 0,
	}
	if err := s.teamWarehouseRepo.Create(ctx, tx, warehouse); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "创建团队仓库失败")
	}

	// 9. 提交事务
	if err := tx.Commit(); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	// 10. 同步队长权限到 Keto (异步,不阻塞主流程)
	if s.teamPermissionService != nil {
		if err := s.teamPermissionService.SyncMemberToKeto(ctx, member); err != nil {
			// 只记录警告,不阻止创建流程
			fmt.Printf("Warning: Failed to sync leader to Keto for team %s: %v\n", team.ID, err)
		}
	}

	return team, nil
}

// GetTeamByID 获取团队详情
func (s *TeamService) GetTeamByID(ctx context.Context, teamID string) (*game_runtime.Team, error) {
	team, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "团队不存在")
	}
	return team, nil
}

// UpdateTeamInfoRequest 更新团队信息请求
type UpdateTeamInfoRequest struct {
	TeamID      string
	HeroID      string // 操作者英雄ID
	Name        string
	Description *string
}

// UpdateTeamInfo 更新团队信息（只有队长可以操作）
func (s *TeamService) UpdateTeamInfo(ctx context.Context, req *UpdateTeamInfoRequest) error {
	// 1. 验证参数
	if req.TeamID == "" {
		return xerrors.New(xerrors.CodeInvalidParams, "团队ID不能为空")
	}
	if req.HeroID == "" {
		return xerrors.New(xerrors.CodeInvalidParams, "英雄ID不能为空")
	}

	// 2. 检查是否是队长
	member, err := s.teamMemberRepo.GetByTeamAndHero(ctx, req.TeamID, req.HeroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "您不是该团队成员")
	}
	if member.Role != "leader" {
		return xerrors.New(xerrors.CodePermissionDenied, "只有队长可以更新团队信息")
	}

	// 3. 获取团队
	team, err := s.teamRepo.GetByID(ctx, req.TeamID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "团队不存在")
	}

	// 4. 更新信息
	if req.Name != "" {
		// 检查新名称是否已存在
		if req.Name != team.Name {
			exists, err := s.teamRepo.Exists(ctx, req.Name)
			if err != nil {
				return xerrors.Wrap(err, xerrors.CodeInternalError, "检查团队名称失败")
			}
			if exists {
				return xerrors.New(xerrors.CodeDuplicateResource, "团队名称已存在")
			}
		}
		team.Name = req.Name
	}
	if req.Description != nil {
		team.Description.SetValid(*req.Description)
	}

	// 5. 保存更新
	if err := s.teamRepo.Update(ctx, team); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "更新团队信息失败")
	}

	return nil
}

// DisbandTeam 解散团队（只有队长可以操作，且仓库必须为空）
func (s *TeamService) DisbandTeam(ctx context.Context, teamID, heroID string) error {
	// 1. 验证参数
	if teamID == "" {
		return xerrors.New(xerrors.CodeInvalidParams, "团队ID不能为空")
	}
	if heroID == "" {
		return xerrors.New(xerrors.CodeInvalidParams, "英雄ID不能为空")
	}

	// 2. 检查是否是队长
	member, err := s.teamMemberRepo.GetByTeamAndHero(ctx, teamID, heroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "您不是该团队成员")
	}
	if member.Role != "leader" {
		return xerrors.New(xerrors.CodePermissionDenied, "只有队长可以解散团队")
	}

	// 3. 检查仓库是否为空
	warehouse, err := s.teamWarehouseRepo.GetByTeamID(ctx, teamID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "获取团队仓库失败")
	}
	if warehouse.GoldAmount > 0 {
		return xerrors.New(xerrors.CodeInvalidParams, "团队仓库不为空，请先分配所有金币")
	}

	// TODO: 检查仓库物品是否为空

	// 4. 获取所有成员列表 (用于清理 Keto)
	members, err := s.teamMemberRepo.ListByTeam(ctx, teamID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "获取成员列表失败")
	}

	// 5. 软删除团队
	if err := s.teamRepo.Delete(ctx, teamID); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "解散团队失败")
	}

	// 6. 删除所有成员的 Keto 权限关系
	if s.teamPermissionService != nil {
		for _, m := range members {
			if err := s.teamPermissionService.DeleteMemberFromKeto(ctx, teamID, m.HeroID); err != nil {
				fmt.Printf("Warning: Failed to delete member %s from Keto for team %s: %v\n", m.HeroID, teamID, err)
			}
		}
	}

	return nil
}

// LeaveTeam 离开团队（队长不能离开）
func (s *TeamService) LeaveTeam(ctx context.Context, teamID, heroID string) error {
	// 1. 验证参数
	if teamID == "" {
		return xerrors.New(xerrors.CodeInvalidParams, "团队ID不能为空")
	}
	if heroID == "" {
		return xerrors.New(xerrors.CodeInvalidParams, "英雄ID不能为空")
	}

	// 2. 获取成员记录
	member, err := s.teamMemberRepo.GetByTeamAndHero(ctx, teamID, heroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "您不是该团队成员")
	}

	// 3. 队长不能离开
	if member.Role == "leader" {
		return xerrors.New(xerrors.CodePermissionDenied, "队长不能离开团队，请先转移队长或解散团队")
	}

	// 4. 删除成员记录
	if err := s.teamMemberRepo.Delete(ctx, s.db, member.ID); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "离开团队失败")
	}

	// 5. 删除 Keto 权限关系
	if s.teamPermissionService != nil {
		if err := s.teamPermissionService.DeleteMemberFromKeto(ctx, teamID, heroID); err != nil {
			fmt.Printf("Warning: Failed to delete member %s from Keto for team %s: %v\n", heroID, teamID, err)
		}
	}

	return nil
}

// TransferInactiveLeaders 队长自动转移（定时任务调用）
func (s *TeamService) TransferInactiveLeaders(ctx context.Context) error {
	// 1. 查询7天未活跃的队长团队
	teams, err := s.teamRepo.GetInactiveLeaderTeams(ctx, 7*24*time.Hour)
	if err != nil {
		return fmt.Errorf("查询不活跃队长团队失败: %w", err)
	}

	// 2. 对每个团队进行队长转移
	for _, team := range teams {
		if err := s.transferLeader(ctx, team); err != nil {
			// 记录错误但继续处理其他团队
			fmt.Printf("转移团队 %s 的队长失败: %v\n", team.ID, err)
			continue
		}
	}

	return nil
}

// transferLeader 转移队长（内部方法）
func (s *TeamService) transferLeader(ctx context.Context, team *game_runtime.Team) error {
	// 1. 查找新队长候选人：优先管理员，其次成员
	newLeaderCandidate, err := s.teamMemberRepo.GetEarliestAdmin(ctx, team.ID)
	if err != nil {
		return fmt.Errorf("查找管理员失败: %w", err)
	}

	if newLeaderCandidate == nil {
		// 没有管理员，查找最早的成员
		newLeaderCandidate, err = s.teamMemberRepo.GetEarliestMember(ctx, team.ID)
		if err != nil {
			return fmt.Errorf("查找成员失败: %w", err)
		}
	}

	if newLeaderCandidate == nil {
		// 团队无其他成员，跳过
		return nil
	}

	// 2. 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback()

	// 3. 更新团队的队长ID
	team.LeaderHeroID = newLeaderCandidate.HeroID
	if err := s.teamRepo.Update(ctx, team); err != nil {
		return fmt.Errorf("更新团队队长失败: %w", err)
	}

	// 4. 更新原队长角色为 member
	if err := s.teamMemberRepo.UpdateRole(ctx, tx, team.ID, team.LeaderHeroID, "member"); err != nil {
		return fmt.Errorf("更新原队长角色失败: %w", err)
	}

	// 5. 更新新队长角色为 leader
	if err := s.teamMemberRepo.UpdateRole(ctx, tx, team.ID, newLeaderCandidate.HeroID, "leader"); err != nil {
		return fmt.Errorf("更新新队长角色失败: %w", err)
	}

	// 6. 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	// 7. 同步权限到 Keto
	if s.teamPermissionService != nil {
		// 获取原队长当前角色和新角色
		oldLeaderOldRole := "leader"
		oldLeaderNewRole := "member"

		// 更新原队长角色
		if err := s.teamPermissionService.UpdateMemberRoleInKeto(ctx, team.ID, team.LeaderHeroID, oldLeaderOldRole, oldLeaderNewRole); err != nil {
			fmt.Printf("Warning: Failed to update old leader role in Keto for team %s: %v\n", team.ID, err)
		}

		// 更新新队长角色 (从 admin 或 member 升级为 leader)
		newLeaderOldRole := newLeaderCandidate.Role // 原来的角色
		newLeaderNewRole := "leader"

		if err := s.teamPermissionService.UpdateMemberRoleInKeto(ctx, team.ID, newLeaderCandidate.HeroID, newLeaderOldRole, newLeaderNewRole); err != nil {
			fmt.Printf("Warning: Failed to update new leader role in Keto for team %s: %v\n", team.ID, err)
		}
	}

	return nil
}

