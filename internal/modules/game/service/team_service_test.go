package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tsu-self/internal/modules/game/testseed"
	"tsu-self/internal/pkg/xerrors"
)

// setupTestDB 设置测试数据库连接
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := "host=localhost port=5432 user=tsu_user password=tsu_test dbname=tsu_db sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skipf("无法连接测试数据库: %v", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		t.Skipf("跳过依赖数据库的测试，原因: %v", err)
	}

	return db
}

// cleanupTestData 清理测试数据
func cleanupTestData(t *testing.T, db *sql.DB, teamID string) {
	t.Helper()

	_, _ = db.Exec("DELETE FROM game_runtime.team_members WHERE team_id = $1", teamID)
	_, _ = db.Exec("DELETE FROM game_runtime.team_warehouses WHERE team_id = $1", teamID)
	_, _ = db.Exec("DELETE FROM game_runtime.teams WHERE id = $1", teamID)
}

// TestTeamService_CreateTeam 测试创建团队
func TestTeamService_CreateTeam(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupTestDB(t)
	defer db.Close()

	// 测试环境不需要 Keto,传入 nil
	teamService := NewTeamService(db, nil)
	ctx := context.Background()

	userID := testseed.EnsureUser(t, db, "team-service-leader")
	heroID := testseed.EnsureHero(t, db, userID, "team-service-leader-hero")

	tests := []struct {
		name      string
		req       *CreateTeamRequest
		wantError bool
		errorCode xerrors.ErrorCode
	}{
		{
			name: "成功创建团队",
			req: &CreateTeamRequest{
				UserID:      userID.String(),
				HeroID:      heroID.String(),
				TeamName:    "测试团队-" + time.Now().Format("20060102150405"),
				Description: "这是一个测试团队",
			},
			wantError: false,
		},
		{
			name: "团队名称为空",
			req: &CreateTeamRequest{
				UserID:   userID.String(),
				HeroID:   heroID.String(),
				TeamName: "",
			},
			wantError: true,
			errorCode: xerrors.CodeInvalidParams,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			team, err := teamService.CreateTeam(ctx, tt.req)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorCode != 0 {
					xerr, ok := err.(*xerrors.AppError)
					if ok {
						assert.Equal(t, tt.errorCode, xerr.Code)
					}
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, team)
				assert.NotEmpty(t, team.ID)
				assert.Equal(t, tt.req.TeamName, team.Name)

				defer cleanupTestData(t, db, team.ID)

				// 验证仓库是否创建
				var warehouseID string
				err := db.QueryRow("SELECT id FROM game_runtime.team_warehouses WHERE team_id = $1", team.ID).Scan(&warehouseID)
				assert.NoError(t, err)
			}
		})
	}
}

// TestTeamService_UpdateTeamInfo 测试更新团队信息
func TestTeamService_UpdateTeamInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupTestDB(t)
	defer db.Close()

	// 测试环境不需要 Keto,传入 nil
	teamService := NewTeamService(db, nil)
	ctx := context.Background()

	userID := testseed.EnsureUser(t, db, "team-service-leader")
	heroID := testseed.EnsureHero(t, db, userID, "team-service-leader-hero")
	createReq := &CreateTeamRequest{
		UserID:      userID.String(),
		HeroID:      heroID.String(),
		TeamName:    "测试团队-" + time.Now().Format("20060102150405"),
		Description: "原始描述",
	}

	team, err := teamService.CreateTeam(ctx, createReq)
	require.NoError(t, err)
	defer cleanupTestData(t, db, team.ID)

	newName := "新团队名称-" + time.Now().Format("20060102150405")
	newDesc := "新的描述"
	updateReq := &UpdateTeamInfoRequest{
		TeamID:      team.ID,
		HeroID:      heroID.String(),
		Name:        newName,
		Description: &newDesc,
	}

	err = teamService.UpdateTeamInfo(ctx, updateReq)
	assert.NoError(t, err)

	updatedTeam, err := teamService.GetTeamByID(ctx, team.ID)
	require.NoError(t, err)
	assert.Equal(t, newName, updatedTeam.Name)
}

// TestTeamService_DisbandTeam 覆盖仓库校验与软删除
func TestTeamService_DisbandTeam(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupTestDB(t)
	defer db.Close()

	teamService := NewTeamService(db, nil)
	ctx := context.Background()

	leaderUserID := testseed.EnsureUser(t, db, "team-service-disband-leader")
	leaderHeroID := testseed.EnsureHero(t, db, leaderUserID, "team-service-disband-hero")

	team, err := teamService.CreateTeam(ctx, &CreateTeamRequest{
		UserID:      leaderUserID.String(),
		HeroID:      leaderHeroID.String(),
		TeamName:    "disband-team-" + time.Now().Format("150405"),
		Description: "for disband test",
	})
	require.NoError(t, err)
	defer cleanupTestData(t, db, team.ID)

	// 1) 仓库非空应拒绝
	_, err = db.Exec(`UPDATE game_runtime.team_warehouses SET gold_amount = 123 WHERE team_id = $1`, team.ID)
	require.NoError(t, err)

	err = teamService.DisbandTeam(ctx, team.ID, leaderHeroID.String())
	assert.Error(t, err)

	// 2) 清空仓库后成功软删除
	_, err = db.Exec(`UPDATE game_runtime.team_warehouses SET gold_amount = 0 WHERE team_id = $1`, team.ID)
	require.NoError(t, err)

	err = teamService.DisbandTeam(ctx, team.ID, leaderHeroID.String())
	assert.NoError(t, err)

	var deletedAt sql.NullTime
	_ = db.QueryRow(`SELECT deleted_at FROM game_runtime.teams WHERE id = $1`, team.ID).Scan(&deletedAt)
	assert.True(t, deletedAt.Valid, "团队应被软删除")
}

// TestTeamService_LeaveTeam 覆盖离队权限
func TestTeamService_LeaveTeam(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupTestDB(t)
	defer db.Close()

	teamService := NewTeamService(db, nil)
	ctx := context.Background()

	leaderUserID := testseed.EnsureUser(t, db, "team-service-leave-leader")
	leaderHeroID := testseed.EnsureHero(t, db, leaderUserID, "team-service-leave-leader-hero")
	team, err := teamService.CreateTeam(ctx, &CreateTeamRequest{
		UserID:      leaderUserID.String(),
		HeroID:      leaderHeroID.String(),
		TeamName:    "leave-team-" + time.Now().Format("150405"),
		Description: "for leave test",
	})
	require.NoError(t, err)
	defer cleanupTestData(t, db, team.ID)

	memberUserID := testseed.EnsureUser(t, db, "team-service-leave-member")
	memberHeroID := testseed.EnsureHero(t, db, memberUserID, "team-service-leave-member-hero")
	_, err = db.Exec(`
		INSERT INTO game_runtime.team_members (id, team_id, hero_id, user_id, role, joined_at)
		VALUES (gen_random_uuid(), $1, $2, $3, 'member', NOW())
	`, team.ID, memberHeroID.String(), memberUserID.String())
	require.NoError(t, err)

	// 普通成员可离队
	err = teamService.LeaveTeam(ctx, team.ID, memberHeroID.String())
	assert.NoError(t, err)

	// 队长离队被拒绝
	err = teamService.LeaveTeam(ctx, team.ID, leaderHeroID.String())
	assert.Error(t, err)
}

// TestTeamService_TransferInactiveLeaders 覆盖自动转移逻辑
func TestTeamService_TransferInactiveLeaders(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupTestDB(t)
	defer db.Close()

	teamPermissionSvc := NewTeamPermissionService(db, nil, nil)
	teamService := NewTeamService(db, teamPermissionSvc)
	ctx := context.Background()

	leaderUserID := testseed.EnsureUser(t, db, "team-service-transfer-leader")
	leaderHeroID := testseed.EnsureHero(t, db, leaderUserID, "team-service-transfer-leader-hero")
	team, err := teamService.CreateTeam(ctx, &CreateTeamRequest{
		UserID:      leaderUserID.String(),
		HeroID:      leaderHeroID.String(),
		TeamName:    "transfer-team-" + time.Now().Format("150405"),
		Description: "for transfer test",
	})
	require.NoError(t, err)
	defer cleanupTestData(t, db, team.ID)

	adminUserID := testseed.EnsureUser(t, db, "team-service-transfer-admin")
	adminHeroID := testseed.EnsureHero(t, db, adminUserID, "team-service-transfer-admin-hero")
	_, err = db.Exec(`
		INSERT INTO game_runtime.team_members (id, team_id, hero_id, user_id, role, joined_at, last_active_at)
		VALUES (gen_random_uuid(), $1, $2, $3, 'admin', NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days')
	`, team.ID, adminHeroID.String(), adminUserID.String())
	require.NoError(t, err)

	// 标记队长长时间未活跃
	_, err = db.Exec(`
		UPDATE game_runtime.team_members
		SET last_active_at = NOW() - INTERVAL '8 days'
		WHERE team_id = $1 AND hero_id = $2
	`, team.ID, leaderHeroID.String())
	require.NoError(t, err)

	err = teamService.TransferInactiveLeaders(ctx)
	assert.NoError(t, err)

	var newLeaderID string
	_ = db.QueryRow(`SELECT leader_hero_id FROM game_runtime.teams WHERE id = $1`, team.ID).Scan(&newLeaderID)
	assert.Equal(t, adminHeroID.String(), newLeaderID, "应自动转移给最早的管理员")

	var oldLeaderRole, newLeaderRole string
	_ = db.QueryRow(`SELECT role FROM game_runtime.team_members WHERE team_id = $1 AND hero_id = $2`, team.ID, leaderHeroID.String()).Scan(&oldLeaderRole)
	_ = db.QueryRow(`SELECT role FROM game_runtime.team_members WHERE team_id = $1 AND hero_id = $2`, team.ID, adminHeroID.String()).Scan(&newLeaderRole)
	assert.Equal(t, "member", oldLeaderRole)
	assert.Equal(t, "leader", newLeaderRole)
}

// 运行测试：
// go test -v ./internal/modules/game/service -run TestTeamService -short  # 跳过集成测试
// go test -v ./internal/modules/game/service -run TestTeamService         # 运行集成测试
