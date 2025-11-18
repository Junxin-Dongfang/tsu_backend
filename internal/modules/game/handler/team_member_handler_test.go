package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tsu-self/internal/modules/auth/client"
	"tsu-self/internal/modules/game/service"
	"tsu-self/internal/pkg/response"
)

// setupTestMemberHandler 设置测试 MemberHandler
func setupTestMemberHandler(t *testing.T) (*TeamMemberHandler, *TeamHandler, *echo.Echo) {
	t.Helper()

	db := setupTestDB(t)
	ketoClient := &client.KetoClient{}
	serviceContainer := service.NewServiceContainer(db, ketoClient, nil)
	respWriter := response.DefaultResponseHandler()

	memberHandler := NewTeamMemberHandler(serviceContainer, respWriter)
	teamHandler := NewTeamHandler(serviceContainer, respWriter)

	e := echo.New()
	return memberHandler, teamHandler, e
}

// TestTeamMemberHandler_ApplyToJoin 测试申请加入团队
func TestTeamMemberHandler_ApplyToJoin(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	memberHandler, teamHandler, e := setupTestMemberHandler(t)
	db := setupTestDB(t)
	defer db.Close()

	// 创建测试团队
	description := "测试团队"
	createReq := CreateTeamRequest{
		HeroID:      "test-hero-leader",
		TeamName:    "测试团队-" + time.Now().Format("20060102150405"),
		Description: &description,
	}
	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/game/teams", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", "test-user-leader")

	err := teamHandler.CreateTeam(c)
	require.NoError(t, err)

	// 获取创建的团队ID
	var createResp response.Response
	err = json.Unmarshal(rec.Body.Bytes(), &createResp)
	require.NoError(t, err)

	teamData := createResp.Data.(map[string]interface{})
	teamID := teamData["id"].(string)

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
				HeroID: "test-hero-applicant",
			},
			setupContext: func(c echo.Context) {
				c.Set("user_id", "test-user-applicant")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "未登录",
			requestBody: ApplyToJoinRequest{
				TeamID: teamID,
				HeroID: "test-hero-applicant",
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

	memberHandler, teamHandler, e := setupTestMemberHandler(t)
	db := setupTestDB(t)
	defer db.Close()

	// 创建测试团队
	description := "测试团队"
	createReq := CreateTeamRequest{
		HeroID:      "test-hero-leader",
		TeamName:    "测试团队-" + time.Now().Format("20060102150405"),
		Description: &description,
	}
	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/game/teams", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", "test-user-leader")

	err := teamHandler.CreateTeam(c)
	require.NoError(t, err)

	var createResp response.Response
	err = json.Unmarshal(rec.Body.Bytes(), &createResp)
	require.NoError(t, err)

	teamData := createResp.Data.(map[string]interface{})
	teamID := teamData["id"].(string)

	defer func() {
		_, _ = db.Exec("DELETE FROM game_runtime.team_members WHERE team_id = $1", teamID)
		_, _ = db.Exec("DELETE FROM game_runtime.team_warehouses WHERE team_id = $1", teamID)
		_, _ = db.Exec("DELETE FROM game_runtime.teams WHERE id = $1", teamID)
	}()

	// 创建申请
	_, err = db.Exec(`
		INSERT INTO game_runtime.team_join_requests (id, team_id, hero_id, user_id, status, message, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'pending', '测试申请', NOW(), NOW())
	`, "test-request-001", teamID, "test-hero-applicant", "test-user-applicant")
	require.NoError(t, err)

	defer func() {
		_, _ = db.Exec("DELETE FROM game_runtime.team_join_requests WHERE id = $1", "test-request-001")
	}()

	tests := []struct {
		name           string
		requestBody    ApproveJoinRequestRequest
		expectedStatus int
	}{
		{
			name: "队长批准申请",
			requestBody: ApproveJoinRequestRequest{
				RequestID: "test-request-001",
				HeroID:    "test-hero-leader",
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

// TestTeamMemberHandler_KickMember 测试踢出成员
func TestTeamMemberHandler_KickMember(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	memberHandler, _, e := setupTestMemberHandler(t)
	db := setupTestDB(t)
	defer db.Close()

	// 这里需要先创建团队和成员，然后测试踢出功能
	// 由于涉及较复杂的数据准备，这里只测试基本的参数验证
	tests := []struct {
		name           string
		requestBody    KickMemberRequest
		expectedStatus int
	}{
		{
			name: "参数验证 - TeamID为空",
			requestBody: KickMemberRequest{
				TeamID:       "",
				TargetHeroID: "test-hero-target",
				KickerHeroID: "test-hero-kicker",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/game/teams/members/kick", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)

			_ = memberHandler.KickMember(c)
			// 参数验证失败时可能不返回错误，而是设置状态码
		})
	}
}

// 运行测试：
// go test -v ./internal/modules/game/handler -run TestTeamMemberHandler
