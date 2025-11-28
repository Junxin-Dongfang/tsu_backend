package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tsu-self/internal/modules/game/service"
	"tsu-self/internal/modules/game/testseed"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/validator"
)

// setupTestMemberHandler 设置测试 MemberHandler
func setupTestMemberHandler(t *testing.T) (*TeamMemberHandler, *TeamHandler, *echo.Echo, *sql.DB) {
	t.Helper()

	db := setupTestDB(t)
	serviceContainer := service.NewServiceContainer(db, nil, nil)
	respWriter := response.DefaultResponseHandler()

	memberHandler := NewTeamMemberHandler(serviceContainer, respWriter)
	teamHandler := NewTeamHandler(serviceContainer, respWriter)

	e := echo.New()
	e.Validator = validator.New()
	return memberHandler, teamHandler, e, db
}

// TestTeamMemberHandler_ApplyToJoin 测试申请加入团队
func TestTeamMemberHandler_ApplyToJoin(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	memberHandler, teamHandler, e, db := setupTestMemberHandler(t)
	defer db.Close()

	leaderUserID := testseed.EnsureUser(t, db, "team-member-handler-leader")
	leaderHeroID := testseed.EnsureHero(t, db, leaderUserID, "team-member-handler-leader-hero")
	testseed.CleanupTeamsByHero(t, db, leaderHeroID)

	// 创建测试团队
	description := "测试团队"
	createReq := CreateTeamRequest{
		HeroID:      stringPtr(leaderHeroID.String()),
		TeamName:    "测试团队-" + time.Now().Format("20060102150405"),
		Description: &description,
	}
	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/game/teams", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", leaderUserID.String())

	err := teamHandler.CreateTeam(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	// 获取创建的团队ID
	var createResp response.Response
	err = json.Unmarshal(rec.Body.Bytes(), &createResp)
	require.NoError(t, err)

	teamData, ok := createResp.Data.(map[string]interface{})
	require.True(t, ok)
	teamID, ok := teamData["id"].(string)
	require.True(t, ok)

	defer func() {
		_, _ = db.Exec("DELETE FROM game_runtime.team_members WHERE team_id = $1", teamID)
		_, _ = db.Exec("DELETE FROM game_runtime.team_warehouses WHERE team_id = $1", teamID)
		_, _ = db.Exec("DELETE FROM game_runtime.teams WHERE id = $1", teamID)
	}()

	tests := []struct {
		name           string
		requestBody    ApplyToJoinRequest
		setupContext   func(c echo.Context)
		expectedStatus int
	}{
		{
			name: "成功申请加入",
			requestBody: ApplyToJoinRequest{
				TeamID: teamID,
				HeroID: testseed.EnsureHero(t, db, testseed.EnsureUser(t, db, "team-member-handler-applicant"), "team-member-handler-applicant-hero").String(),
			},
			setupContext: func(c echo.Context) {
				c.Set("user_id", testseed.EnsureUser(t, db, "team-member-handler-applicant").String())
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "未登录",
			requestBody: ApplyToJoinRequest{
				TeamID: teamID,
				HeroID: testseed.EnsureHero(t, db, testseed.EnsureUser(t, db, "team-member-handler-applicant"), "team-member-handler-applicant-hero").String(),
			},
			setupContext: func(c echo.Context) {
				// 不设置 user_id
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/game/teams/join/apply", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			tt.setupContext(c)

			err := memberHandler.ApplyToJoin(c)

			if tt.expectedStatus == http.StatusOK {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)

				// 清理申请记录
				defer func() {
					_, _ = db.Exec("DELETE FROM game_runtime.team_join_requests WHERE team_id = $1 AND hero_id = $2",
						tt.requestBody.TeamID, tt.requestBody.HeroID)
				}()
			}
		})
	}
}

// TestTeamMemberHandler_ApproveJoinRequest 测试审批加入申请
func TestTeamMemberHandler_ApproveJoinRequest(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	memberHandler, teamHandler, e, db := setupTestMemberHandler(t)
	defer db.Close()

	leaderUserID := testseed.EnsureUser(t, db, "team-member-handler-leader")
	leaderHeroID := testseed.EnsureHero(t, db, leaderUserID, "team-member-handler-leader-hero")
	testseed.CleanupTeamsByHero(t, db, leaderHeroID)

	// 创建测试团队
	description := "测试团队"
	createReq := CreateTeamRequest{
		HeroID:      stringPtr(leaderHeroID.String()),
		TeamName:    "测试团队-" + time.Now().Format("20060102150405"),
		Description: &description,
	}
	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/game/teams", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", leaderUserID.String())

	err := teamHandler.CreateTeam(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	var createResp response.Response
	err = json.Unmarshal(rec.Body.Bytes(), &createResp)
	require.NoError(t, err)

	teamData, ok := createResp.Data.(map[string]interface{})
	require.True(t, ok)
	teamID, ok := teamData["id"].(string)
	require.True(t, ok)

	defer func() {
		_, _ = db.Exec("DELETE FROM game_runtime.team_members WHERE team_id = $1", teamID)
		_, _ = db.Exec("DELETE FROM game_runtime.team_warehouses WHERE team_id = $1", teamID)
		_, _ = db.Exec("DELETE FROM game_runtime.teams WHERE id = $1", teamID)
	}()

	// 创建申请
	applicantUserID := testseed.EnsureUser(t, db, "team-member-handler-applicant")
	applicantHeroID := testseed.EnsureHero(t, db, applicantUserID, "team-member-handler-applicant-hero")
	reqID := testseed.StableUUID("team-member-handler-request-1").String()

	_, err = db.Exec(`
		INSERT INTO game_runtime.team_join_requests (id, team_id, hero_id, user_id, status, message, created_at)
		VALUES ($1, $2, $3, $4, 'pending', '测试申请', NOW())
	`, reqID, teamID, applicantHeroID.String(), applicantUserID.String())
	require.NoError(t, err)

	defer func() {
		_, _ = db.Exec("DELETE FROM game_runtime.team_join_requests WHERE id = $1", reqID)
	}()

	tests := []struct {
		name           string
		requestBody    ApproveJoinRequestRequest
		expectedStatus int
	}{
		{
			name: "队长批准申请",
			requestBody: ApproveJoinRequestRequest{
				RequestID: reqID,
				HeroID:    leaderHeroID.String(),
				Approved:  true,
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/game/teams/join/approve", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)

			err := memberHandler.ApproveJoinRequest(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// TestTeamMemberHandler_KickMember 覆盖队长踢出成员
func TestTeamMemberHandler_KickMember(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	memberHandler, teamHandler, e, db := setupTestMemberHandler(t)
	defer db.Close()

	leaderUserID := testseed.EnsureUser(t, db, "team-member-handler-kick-leader")
	leaderHeroID := testseed.EnsureHero(t, db, leaderUserID, "team-member-handler-kick-leader-hero")
	testseed.CleanupTeamsByHero(t, db, leaderHeroID)

	// 创建团队
	description := "kick team"
	createReq := CreateTeamRequest{
		HeroID:      stringPtr(leaderHeroID.String()),
		TeamName:    "kick-" + time.Now().Format("150405"),
		Description: &description,
	}
	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/game/teams", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", leaderUserID.String())
	require.NoError(t, teamHandler.CreateTeam(c))
	if rec.Code != http.StatusOK {
		t.Fatalf("create team failed status=%d body=%s", rec.Code, rec.Body.String())
	}

	var createResp response.Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &createResp))
	teamData := createResp.Data.(map[string]interface{})
	teamID := teamData["id"].(string)
	defer func() {
		_, _ = db.Exec("DELETE FROM game_runtime.team_members WHERE team_id = $1", teamID)
		_, _ = db.Exec("DELETE FROM game_runtime.team_warehouses WHERE team_id = $1", teamID)
		_, _ = db.Exec("DELETE FROM game_runtime.teams WHERE id = $1", teamID)
	}()

	// 添加一个成员
	memberUserID := testseed.EnsureUser(t, db, "team-member-handler-kick-target")
	memberHeroID := testseed.EnsureHero(t, db, memberUserID, "team-member-handler-kick-target-hero")
	_, err := db.Exec(`
		INSERT INTO game_runtime.team_members (id, team_id, hero_id, user_id, role, joined_at)
		VALUES (gen_random_uuid(), $1, $2, $3, 'member', NOW())
	`, teamID, memberHeroID.String(), memberUserID.String())
	require.NoError(t, err)

	// 队长踢出成员
	kickReq := KickMemberRequest{
		TeamID:       teamID,
		TargetHeroID: memberHeroID.String(),
		KickerHeroID: leaderHeroID.String(),
	}
	kickBody, _ := json.Marshal(kickReq)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/game/teams/members/kick", bytes.NewReader(kickBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err = memberHandler.KickMember(c)
	if rec.Code != http.StatusOK {
		t.Fatalf("kick member failed status=%d body=%s err=%v", rec.Code, rec.Body.String(), err)
	}
}

// 运行测试：
// go test -v ./internal/modules/game/handler -run TestTeamMemberHandler
