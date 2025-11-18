package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	// 测试申请加入
	applicantUserID := "test-user-applicant"
	applicantHeroID := "test-hero-applicant"
	applyReq := &ApplyToJoinRequest{
		TeamID:  team.ID,
		HeroID:  applicantHeroID,
		UserID:  applicantUserID,
		Message: "我想加入你们的团队",
	}

	err = memberService.ApplyToJoin(ctx, applyReq)
	assert.NoError(t, err)

	// 验证申请记录是否创建
	var requestID string
	err = db.QueryRow("SELECT id FROM game_runtime.team_join_requests WHERE team_id = $1 AND hero_id = $2 AND status = 'pending'", team.ID, applicantHeroID).Scan(&requestID)
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

	// 创建申请
	applicantUserID := "test-user-applicant"
	applicantHeroID := "test-hero-applicant"
	applyReq := &ApplyToJoinRequest{
		TeamID:  team.ID,
		HeroID:  applicantHeroID,
		UserID:  applicantUserID,
		Message: "我想加入你们的团队",
	}

	err = memberService.ApplyToJoin(ctx, applyReq)
	require.NoError(t, err)

	// 获取申请ID
	var requestID string
	err = db.QueryRow("SELECT id FROM game_runtime.team_join_requests WHERE team_id = $1 AND hero_id = $2", team.ID, applicantHeroID).Scan(&requestID)
	require.NoError(t, err)

	defer func() {
		_, _ = db.Exec("DELETE FROM game_runtime.team_join_requests WHERE id = $1", requestID)
	}()

	// 测试批准申请
	approveReq := &ApproveJoinRequestRequest{
		RequestID: requestID,
		HeroID:    leaderHeroID,
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

	// TODO: 添加成员后再测试踢出
	// 这里需要先通过申请流程添加成员，然后再测试踢出功能
}

// 运行测试：
// go test -v ./internal/modules/game/service -run TestTeamMemberService

