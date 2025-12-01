package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tsu-self/internal/modules/game/testseed"
)

// TestTeamMemberService_ApplyToJoin 测试申请加入团队
func TestTeamMemberService_ApplyToJoin(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupTestDB(t)
	defer db.Close()

	// 测试环境不需要 Keto,传入 nil
	teamService := NewTeamService(db, nil)
	memberService := NewTeamMemberService(db, nil)
	ctx := context.Background()

	// 创建测试团队
	leaderUserID := testseed.EnsureUser(t, db, "team-member-leader")
	leaderHeroID := testseed.EnsureHero(t, db, leaderUserID, "team-member-leader-hero")
	createReq := &CreateTeamRequest{
		UserID:      leaderUserID.String(),
		HeroID:      leaderHeroID.String(),
		TeamName:    "测试团队-" + time.Now().Format("20060102150405"),
		Description: "测试团队",
	}

	team, err := teamService.CreateTeam(ctx, createReq)
	require.NoError(t, err)
	defer cleanupTestData(t, db, team.ID)

	// 测试申请加入
	applicantUserID := testseed.EnsureUser(t, db, "team-member-applicant")
	applicantHeroID := testseed.EnsureHero(t, db, applicantUserID, "team-member-applicant-hero")
	applyReq := &ApplyToJoinRequest{
		TeamID:  team.ID,
		HeroID:  applicantHeroID.String(),
		UserID:  applicantUserID.String(),
		Message: "我想加入你们的团队",
	}

	err = memberService.ApplyToJoin(ctx, applyReq)
	assert.NoError(t, err)

	// 验证申请记录是否创建
	var requestID string
	err = db.QueryRow("SELECT id FROM game_runtime.team_join_requests WHERE team_id = $1 AND hero_id = $2 AND status = 'pending'", team.ID, applicantHeroID.String()).Scan(&requestID)
	assert.NoError(t, err)
	assert.NotEmpty(t, requestID)

	// 清理申请记录
	defer func() {
		_, _ = db.Exec("DELETE FROM game_runtime.team_join_requests WHERE id = $1", requestID)
	}()
}

// TestTeamMemberService_ApproveJoinRequest 测试审批加入申请
func TestTeamMemberService_ApproveJoinRequest(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupTestDB(t)
	defer db.Close()

	// 测试环境不需要 Keto,传入 nil
	teamService := NewTeamService(db, nil)
	memberService := NewTeamMemberService(db, nil)
	ctx := context.Background()

	// 创建测试团队
	leaderUserID := testseed.EnsureUser(t, db, "team-member-leader")
	leaderHeroID := testseed.EnsureHero(t, db, leaderUserID, "team-member-leader-hero")
	createReq := &CreateTeamRequest{
		UserID:      leaderUserID.String(),
		HeroID:      leaderHeroID.String(),
		TeamName:    "测试团队-" + time.Now().Format("20060102150405"),
		Description: "测试团队",
	}

	team, err := teamService.CreateTeam(ctx, createReq)
	require.NoError(t, err)
	defer cleanupTestData(t, db, team.ID)

	// 创建申请
	applicantUserID := testseed.EnsureUser(t, db, "team-member-applicant")
	applicantHeroID := testseed.EnsureHero(t, db, applicantUserID, "team-member-applicant-hero")
	applyReq := &ApplyToJoinRequest{
		TeamID:  team.ID,
		HeroID:  applicantHeroID.String(),
		UserID:  applicantUserID.String(),
		Message: "我想加入你们的团队",
	}

	err = memberService.ApplyToJoin(ctx, applyReq)
	require.NoError(t, err)

	// 获取申请ID
	var requestID string
	err = db.QueryRow("SELECT id FROM game_runtime.team_join_requests WHERE team_id = $1 AND hero_id = $2", team.ID, applicantHeroID.String()).Scan(&requestID)
	require.NoError(t, err)

	defer func() {
		_, _ = db.Exec("DELETE FROM game_runtime.team_join_requests WHERE id = $1", requestID)
	}()

	// 测试批准申请
	approveReq := &ApproveJoinRequestRequest{
		RequestID: requestID,
		HeroID:    leaderHeroID.String(),
		Approved:  true,
	}

	err = memberService.ApproveJoinRequest(ctx, approveReq)
	assert.NoError(t, err)

	// 验证申请状态是否更新
	var status string
	err = db.QueryRow("SELECT status FROM game_runtime.team_join_requests WHERE id = $1", requestID).Scan(&status)
	assert.NoError(t, err)
	assert.Equal(t, "approved", status)

	// 验证成员记录是否创建
	var memberID string
	err = db.QueryRow("SELECT id FROM game_runtime.team_members WHERE team_id = $1 AND hero_id = $2", team.ID, applicantHeroID).Scan(&memberID)
	assert.NoError(t, err)
	assert.NotEmpty(t, memberID)

	// 清理成员记录
	defer func() {
		_, _ = db.Exec("DELETE FROM game_runtime.team_members WHERE id = $1", memberID)
	}()
}

// TestTeamMemberService_KickMember 测试踢出成员
func TestTeamMemberService_KickMember(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupTestDB(t)
	defer db.Close()

	// 测试环境不需要 Keto,传入 nil
	teamService := NewTeamService(db, nil)
	_ = NewTeamMemberService(db, nil) // TODO: 在实现踢出测试时使用
	ctx := context.Background()

	// 创建测试团队
	leaderUserID := testseed.EnsureUser(t, db, "team-member-leader")
	leaderHeroID := testseed.EnsureHero(t, db, leaderUserID, "team-member-leader-hero")
	createReq := &CreateTeamRequest{
		UserID:      leaderUserID.String(),
		HeroID:      leaderHeroID.String(),
		TeamName:    "测试团队-" + time.Now().Format("20060102150405"),
		Description: "测试团队",
	}

	team, err := teamService.CreateTeam(ctx, createReq)
	require.NoError(t, err)
	defer cleanupTestData(t, db, team.ID)

	// 先添加一个成员
	memberUserID := testseed.EnsureUser(t, db, "team-member-kick-target-user")
	memberHeroID := testseed.EnsureHero(t, db, memberUserID, "team-member-kick-target-hero")
	_, err = db.Exec(`
		INSERT INTO game_runtime.team_members (id, team_id, hero_id, user_id, role, joined_at)
		VALUES (gen_random_uuid(), $1, $2, $3, 'member', NOW())
	`, team.ID, memberHeroID.String(), memberUserID.String())
	require.NoError(t, err)

	memberService := NewTeamMemberService(db, nil)

	// 队长踢出成员应成功
	err = memberService.KickMember(ctx, &KickMemberRequest{
		TeamID:       team.ID,
		TargetHeroID: memberHeroID.String(),
		KickerHeroID: leaderHeroID.String(),
		Reason:       "长期不活跃",
	})
	assert.NoError(t, err)

	// 验证成员被删除并记录冷却
	var count int
	_ = db.QueryRow(`SELECT COUNT(1) FROM game_runtime.team_members WHERE team_id = $1 AND hero_id = $2`, team.ID, memberHeroID.String()).Scan(&count)
	assert.Equal(t, 0, count)

	var cooldownExists int
	_ = db.QueryRow(`SELECT COUNT(1) FROM game_runtime.team_kicked_records WHERE team_id = $1 AND hero_id = $2`, team.ID, memberHeroID.String()).Scan(&cooldownExists)
	assert.Equal(t, 1, cooldownExists)

	// 管理员不能踢出队长
	adminUserID := testseed.EnsureUser(t, db, "team-member-kick-admin-user")
	adminHeroID := testseed.EnsureHero(t, db, adminUserID, "team-member-kick-admin-hero")
	_, err = db.Exec(`
		INSERT INTO game_runtime.team_members (id, team_id, hero_id, user_id, role, joined_at)
		VALUES (gen_random_uuid(), $1, $2, $3, 'admin', NOW())
		ON CONFLICT DO NOTHING
	`, team.ID, adminHeroID.String(), adminUserID.String())
	require.NoError(t, err)

	err = memberService.KickMember(ctx, &KickMemberRequest{
		TeamID:       team.ID,
		TargetHeroID: leaderHeroID.String(),
		KickerHeroID: adminHeroID.String(),
		Reason:       "尝试踢出队长应失败",
	})
	assert.Error(t, err)
}

// 运行测试：
// go test -v ./internal/modules/game/service -run TestTeamMemberService
