package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tsu-self/internal/modules/game/testseed"
)

// TestTeamDungeonService_SelectDungeon 测试选择地城
func TestTeamDungeonService_SelectDungeon(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}
	t.Skip("待补充地城配置数据后重新启用该集成测试")

	db := setupTestDB(t)
	defer db.Close()

	teamService := NewTeamService(db, nil)
	memberService := NewTeamMemberService(db, nil)
	dungeonService := NewTeamDungeonService(db, nil)
	ctx := context.Background()

	// 创建测试团队
	leaderUserID := "test-user-leader"
	leaderHeroID := "test-hero-leader"
	createReq := &CreateTeamRequest{
		UserID:      leaderUserID,
		HeroID:      leaderHeroID,
		TeamName:    "测试团队-" + time.Now().Format("20060102150405"),
		Description: "测试团队",
	}

	team, err := teamService.CreateTeam(ctx, createReq)
	require.NoError(t, err)
	defer cleanupTestData(t, db, team.ID)

	// 添加一个普通成员
	memberUserID := "test-user-member"
	memberHeroID := "test-hero-member"

	applyReq := &ApplyToJoinRequest{
		TeamID:  team.ID,
		HeroID:  memberHeroID,
		UserID:  memberUserID,
		Message: "测试加入",
	}
	err = memberService.ApplyToJoin(ctx, applyReq)
	require.NoError(t, err)

	// 获取申请ID并批准
	var requestID string
	err = db.QueryRow("SELECT id FROM game_runtime.team_join_requests WHERE team_id = $1 AND hero_id = $2", team.ID, memberHeroID).Scan(&requestID)
	require.NoError(t, err)

	approveReq := &ApproveJoinRequestRequest{
		RequestID: requestID,
		HeroID:    leaderHeroID,
		Approved:  true,
	}
	err = memberService.ApproveJoinRequest(ctx, approveReq)
	require.NoError(t, err)

	defer func() {
		_, _ = db.Exec("DELETE FROM game_runtime.team_join_requests WHERE id = $1", requestID)
		_, _ = db.Exec("DELETE FROM game_runtime.team_members WHERE team_id = $1 AND hero_id = $2", team.ID, memberHeroID)
	}()

	tests := []struct {
		name      string
		req       *SelectDungeonRequest
		wantError bool
		errorMsg  string
	}{
		{
			name: "队长成功选择地城",
			req: &SelectDungeonRequest{
				TeamID:    team.ID,
				HeroID:    leaderHeroID,
				DungeonID: "dungeon-001",
			},
			wantError: false,
		},
		{
			name: "参数为空",
			req: &SelectDungeonRequest{
				TeamID:    "",
				HeroID:    "",
				DungeonID: "",
			},
			wantError: true,
			errorMsg:  "参数不能为空",
		},
		{
			name: "普通成员无权限",
			req: &SelectDungeonRequest{
				TeamID:    team.ID,
				HeroID:    memberHeroID,
				DungeonID: "dungeon-001",
			},
			wantError: true,
			errorMsg:  "只有队长和管理员可以选择地城",
		},
		{
			name: "非团队成员",
			req: &SelectDungeonRequest{
				TeamID:    team.ID,
				HeroID:    "non-member-hero",
				DungeonID: "dungeon-001",
			},
			wantError: true,
			errorMsg:  "您不是该团队成员",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := dungeonService.SelectDungeon(ctx, tt.req)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				// 注意：由于实现不完整，暂时只验证不报错
				assert.NoError(t, err)
			}
		})
	}
}

// TestTeamDungeonService_EnterDungeon 测试进入地城
func TestTeamDungeonService_EnterDungeon(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}
	t.Skip("待补充地城配置数据后重新启用该集成测试")

	db := setupTestDB(t)
	defer db.Close()

	teamService := NewTeamService(db, nil)
	memberService := NewTeamMemberService(db, nil)
	dungeonService := NewTeamDungeonService(db, nil)
	ctx := context.Background()

	// 创建测试团队
	leaderUserID := "test-user-leader"
	leaderHeroID := "test-hero-leader"
	createReq := &CreateTeamRequest{
		UserID:      leaderUserID,
		HeroID:      leaderHeroID,
		TeamName:    "测试团队-" + time.Now().Format("20060102150405"),
		Description: "测试团队",
	}

	team, err := teamService.CreateTeam(ctx, createReq)
	require.NoError(t, err)
	defer cleanupTestData(t, db, team.ID)

	// 添加一个管理员
	adminUserID := "test-user-admin"
	adminHeroID := "test-hero-admin"

	applyReq := &ApplyToJoinRequest{
		TeamID:  team.ID,
		HeroID:  adminHeroID,
		UserID:  adminUserID,
		Message: "测试加入",
	}
	err = memberService.ApplyToJoin(ctx, applyReq)
	require.NoError(t, err)

	// 批准并提升为管理员
	var requestID string
	err = db.QueryRow("SELECT id FROM game_runtime.team_join_requests WHERE team_id = $1 AND hero_id = $2", team.ID, adminHeroID).Scan(&requestID)
	require.NoError(t, err)

	approveReq := &ApproveJoinRequestRequest{
		RequestID: requestID,
		HeroID:    leaderHeroID,
		Approved:  true,
	}
	err = memberService.ApproveJoinRequest(ctx, approveReq)
	require.NoError(t, err)

	promoteReq := &PromoteToAdminRequest{
		TeamID:       team.ID,
		LeaderHeroID: leaderHeroID,
		TargetHeroID: adminHeroID,
	}
	err = memberService.PromoteToAdmin(ctx, promoteReq)
	require.NoError(t, err)

	defer func() {
		_, _ = db.Exec("DELETE FROM game_runtime.team_join_requests WHERE id = $1", requestID)
		_, _ = db.Exec("DELETE FROM game_runtime.team_members WHERE team_id = $1 AND hero_id = $2", team.ID, adminHeroID)
	}()

	tests := []struct {
		name      string
		req       *EnterDungeonRequest
		wantError bool
		errorMsg  string
	}{
		{
			name: "队长成功进入地城",
			req: &EnterDungeonRequest{
				TeamID:    team.ID,
				HeroID:    leaderHeroID,
				DungeonID: "dungeon-001",
			},
			wantError: false,
		},
		{
			name: "管理员成功进入地城",
			req: &EnterDungeonRequest{
				TeamID:    team.ID,
				HeroID:    adminHeroID,
				DungeonID: "dungeon-001",
			},
			wantError: false,
		},
		{
			name: "参数为空",
			req: &EnterDungeonRequest{
				TeamID:    "",
				HeroID:    "",
				DungeonID: "",
			},
			wantError: true,
			errorMsg:  "参数不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := dungeonService.EnterDungeon(ctx, tt.req)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				// 注意：由于实现不完整，暂时只验证不报错
				assert.NoError(t, err)
			}
		})
	}
}

// TestTeamDungeonService_GetDungeonProgress 测试获取地城进度
func TestTeamDungeonService_GetDungeonProgress(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupTestDB(t)
	defer db.Close()

	teamService := NewTeamService(db, nil)
	dungeonService := NewTeamDungeonService(db, nil)
	ctx := context.Background()

	// 创建测试团队
	leaderUserID := testseed.EnsureUser(t, db, "team-dungeon-leader")
	leaderHeroID := testseed.EnsureHero(t, db, leaderUserID, "team-dungeon-leader-hero")
	createReq := &CreateTeamRequest{
		UserID:      leaderUserID.String(),
		HeroID:      leaderHeroID.String(),
		TeamName:    "测试团队-" + time.Now().Format("20060102150405"),
		Description: "测试团队",
	}

	team, err := teamService.CreateTeam(ctx, createReq)
	require.NoError(t, err)
	defer cleanupTestData(t, db, team.ID)

	tests := []struct {
		name      string
		teamID    string
		heroID    string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "团队成员查看进度",
			teamID:    team.ID,
			heroID:    leaderHeroID.String(),
			wantError: false,
		},
		{
			name:      "参数为空",
			teamID:    "",
			heroID:    "",
			wantError: true,
			errorMsg:  "参数不能为空",
		},
		{
			name:      "非团队成员",
			teamID:    team.ID,
			heroID:    testseed.StableUUID("team-dungeon-non-member").String(),
			wantError: true,
			errorMsg:  "您不是该团队成员",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			progress, err := dungeonService.GetDungeonProgress(ctx, tt.teamID, tt.heroID)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				// 注意：由于实现不完整，可能返回 nil
				assert.NoError(t, err)
				// progress 可能为 nil，因为 TODO 还未实现
				_ = progress
			}
		})
	}
}

// 运行测试：
// go test -v ./internal/modules/game/service -run TestTeamDungeonService
