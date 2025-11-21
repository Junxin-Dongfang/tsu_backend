package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/modules/game/testseed"
)

// TestTeamPermissionService_SyncMemberToKeto 测试同步成员权限到 Keto
func TestTeamPermissionService_SyncMemberToKeto(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupTestDB(t)
	defer db.Close()

	// 测试没有 Keto 客户端的情况（优雅降级）
	permissionService := NewTeamPermissionService(db, nil, nil)
	ctx := context.Background()

	member := &game_runtime.TeamMember{
		ID:     "member-001",
		TeamID: "team-001",
		HeroID: "hero-001",
		Role:   "admin",
	}

	// 没有 Keto 客户端时应该静默成功
	err := permissionService.SyncMemberToKeto(ctx, member)
	assert.NoError(t, err)
}

// TestTeamPermissionService_DeleteMemberFromKeto 测试从 Keto 删除成员权限
func TestTeamPermissionService_DeleteMemberFromKeto(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupTestDB(t)
	defer db.Close()

	// 测试没有 Keto 客户端的情况（优雅降级）
	permissionService := NewTeamPermissionService(db, nil, nil)
	ctx := context.Background()

	// 没有 Keto 客户端时应该静默成功
	err := permissionService.DeleteMemberFromKeto(ctx, "team-001", "hero-001")
	assert.NoError(t, err)
}

// TestTeamPermissionService_UpdateMemberRoleInKeto 测试更新成员角色在 Keto 中
func TestTeamPermissionService_UpdateMemberRoleInKeto(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupTestDB(t)
	defer db.Close()

	// 测试没有 Keto 客户端的情况（优雅降级）
	permissionService := NewTeamPermissionService(db, nil, nil)
	ctx := context.Background()

	// 没有 Keto 客户端时应该静默成功
	err := permissionService.UpdateMemberRoleInKeto(ctx, "team-001", "hero-001", "member", "admin")
	assert.NoError(t, err)
}

// TestTeamPermissionService_CheckPermission 测试权限检查（数据库降级）
func TestTeamPermissionService_CheckPermission(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupTestDB(t)
	defer db.Close()

	teamService := NewTeamService(db, nil)
	memberService := NewTeamMemberService(db, nil)
	permissionService := NewTeamPermissionService(db, nil, nil)
	ctx := context.Background()

	// 创建测试团队
	leaderUserID := testseed.EnsureUser(t, db, "team-permission-leader")
	leaderHeroID := testseed.EnsureHero(t, db, leaderUserID, "team-permission-leader-hero")
	createReq := &CreateTeamRequest{
		UserID:      leaderUserID.String(),
		HeroID:      leaderHeroID.String(),
		TeamName:    "测试团队-" + time.Now().Format("20060102150405"),
		Description: "测试团队",
	}

	team, err := teamService.CreateTeam(ctx, createReq)
	require.NoError(t, err)
	defer cleanupTestData(t, db, team.ID)

	// 添加一个管理员
	adminUserID := testseed.EnsureUser(t, db, "team-permission-admin")
	adminHeroID := testseed.EnsureHero(t, db, adminUserID, "team-permission-admin-hero")

	applyReq := &ApplyToJoinRequest{
		TeamID:  team.ID,
		HeroID:  adminHeroID.String(),
		UserID:  adminUserID.String(),
		Message: "测试加入",
	}
	err = memberService.ApplyToJoin(ctx, applyReq)
	require.NoError(t, err)

	// 批准并提升为管理员
	var requestID string
	err = db.QueryRow("SELECT id FROM game_runtime.team_join_requests WHERE team_id = $1 AND hero_id = $2", team.ID, adminHeroID.String()).Scan(&requestID)
	require.NoError(t, err)

	approveReq := &ApproveJoinRequestRequest{
		RequestID: requestID,
		HeroID:    leaderHeroID.String(),
		Approved:  true,
	}
	err = memberService.ApproveJoinRequest(ctx, approveReq)
	require.NoError(t, err)

	promoteReq := &PromoteToAdminRequest{
		TeamID:       team.ID,
		LeaderHeroID: leaderHeroID.String(),
		TargetHeroID: adminHeroID.String(),
	}
	err = memberService.PromoteToAdmin(ctx, promoteReq)
	require.NoError(t, err)

	// 添加一个普通成员
	memberUserID := testseed.EnsureUser(t, db, "team-permission-member")
	memberHeroID := testseed.EnsureHero(t, db, memberUserID, "team-permission-member-hero")

	applyReq2 := &ApplyToJoinRequest{
		TeamID:  team.ID,
		HeroID:  memberHeroID.String(),
		UserID:  memberUserID.String(),
		Message: "测试加入",
	}
	err = memberService.ApplyToJoin(ctx, applyReq2)
	require.NoError(t, err)

	var requestID2 string
	err = db.QueryRow("SELECT id FROM game_runtime.team_join_requests WHERE team_id = $1 AND hero_id = $2", team.ID, memberHeroID.String()).Scan(&requestID2)
	require.NoError(t, err)

	approveReq2 := &ApproveJoinRequestRequest{
		RequestID: requestID2,
		HeroID:    leaderHeroID.String(),
		Approved:  true,
	}
	err = memberService.ApproveJoinRequest(ctx, approveReq2)
	require.NoError(t, err)

	defer func() {
		_, _ = db.Exec("DELETE FROM game_runtime.team_join_requests WHERE id IN ($1, $2)", requestID, requestID2)
		_, _ = db.Exec("DELETE FROM game_runtime.team_members WHERE team_id = $1 AND hero_id IN ($2, $3)", team.ID, adminHeroID.String(), memberHeroID.String())
	}()

	tests := []struct {
		name        string
		teamID      string
		heroID      string
		permission  string
		wantAllowed bool
		wantError   bool
	}{
		{
			name:        "队长拥有踢出成员权限",
			teamID:      team.ID,
			heroID:      leaderHeroID.String(),
			permission:  "kick_member",
			wantAllowed: true,
			wantError:   false,
		},
		{
			name:        "管理员拥有踢出成员权限",
			teamID:      team.ID,
			heroID:      adminHeroID.String(),
			permission:  "kick_member",
			wantAllowed: true,
			wantError:   false,
		},
		{
			name:        "普通成员无踢出成员权限",
			teamID:      team.ID,
			heroID:      memberHeroID.String(),
			permission:  "kick_member",
			wantAllowed: false,
			wantError:   false,
		},
		{
			name:        "队长拥有解散团队权限",
			teamID:      team.ID,
			heroID:      leaderHeroID.String(),
			permission:  "disband_team",
			wantAllowed: true,
			wantError:   false,
		},
		{
			name:        "管理员无解散团队权限",
			teamID:      team.ID,
			heroID:      adminHeroID.String(),
			permission:  "disband_team",
			wantAllowed: false,
			wantError:   false,
		},
		{
			name:        "队长拥有查看仓库权限",
			teamID:      team.ID,
			heroID:      leaderHeroID.String(),
			permission:  "view_warehouse",
			wantAllowed: true,
			wantError:   false,
		},
		{
			name:        "管理员拥有查看仓库权限",
			teamID:      team.ID,
			heroID:      adminHeroID.String(),
			permission:  "view_warehouse",
			wantAllowed: true,
			wantError:   false,
		},
		{
			name:        "普通成员无查看仓库权限",
			teamID:      team.ID,
			heroID:      memberHeroID.String(),
			permission:  "view_warehouse",
			wantAllowed: false,
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := permissionService.CheckPermission(ctx, tt.teamID, tt.heroID, tt.permission)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantAllowed, allowed, "权限检查结果不匹配")
			}
		})
	}
}

// TestTeamPermissionService_CheckPermission_NonMember 测试非成员的权限检查
func TestTeamPermissionService_CheckPermission_NonMember(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupTestDB(t)
	defer db.Close()

	teamService := NewTeamService(db, nil)
	permissionService := NewTeamPermissionService(db, nil, nil)
	ctx := context.Background()

	// 创建测试团队
	leaderUserID := testseed.EnsureUser(t, db, "team-permission-leader")
	leaderHeroID := testseed.EnsureHero(t, db, leaderUserID, "team-permission-leader-hero")
	createReq := &CreateTeamRequest{
		UserID:      leaderUserID.String(),
		HeroID:      leaderHeroID.String(),
		TeamName:    "测试团队-" + time.Now().Format("20060102150405"),
		Description: "测试团队",
	}

	team, err := teamService.CreateTeam(ctx, createReq)
	require.NoError(t, err)
	defer cleanupTestData(t, db, team.ID)

	// 测试非成员
	nonMemberHero := testseed.StableUUID("team-permission-non-member").String()
	allowed, err := permissionService.CheckPermission(ctx, team.ID, nonMemberHero, "view_warehouse")
	require.NoError(t, err)
	assert.False(t, allowed, "非成员不应拥有权限")
}

// 运行测试：
// go test -v ./internal/modules/game/service -run TestTeamPermissionService
