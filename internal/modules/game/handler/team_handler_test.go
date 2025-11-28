package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tsu-self/internal/modules/game/service"
	"tsu-self/internal/modules/game/testseed"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/validator"
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

// setupTestHandler 设置测试 Handler
func setupTestHandler(t *testing.T, db *sql.DB) (*TeamHandler, *echo.Echo) {
	t.Helper()

	// 在单元/集成测试中跳过 Keto，同步路径由数据库降级逻辑覆盖
	serviceContainer := service.NewServiceContainer(db, nil, nil)
	respWriter := response.DefaultResponseHandler()
	handler := NewTeamHandler(serviceContainer, respWriter)

	e := echo.New()
	e.Validator = validator.New()
	return handler, e
}

// stringPtr 创建字符串指针的辅助函数
func stringPtr(s string) *string {
	return &s
}

// TestTeamHandler_CreateTeam 测试创建团队 API
func TestTeamHandler_CreateTeam(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupTestDB(t)
	defer db.Close()

	handler, e := setupTestHandler(t, db)

	userID := testseed.EnsureUser(t, db, "team-handler-user")
	heroID := testseed.EnsureHero(t, db, userID, "team-handler-hero")
	testseed.CleanupTeamsByHero(t, db, heroID)

	tests := []struct {
		name           string
		requestBody    CreateTeamRequest
		setupContext   func(c echo.Context)
		expectedStatus int
	}{
		{
			name: "成功创建团队",
			requestBody: CreateTeamRequest{
				HeroID:   stringPtr(heroID.String()),
				TeamName: "测试团队-" + time.Now().Format("20060102150405"),
			},
			setupContext: func(c echo.Context) {
				c.Set("user_id", userID.String())
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "团队名称为空",
			requestBody: CreateTeamRequest{
				HeroID:   stringPtr(heroID.String()),
				TeamName: "",
			},
			setupContext: func(c echo.Context) {
				c.Set("user_id", userID.String())
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "未登录",
			requestBody: CreateTeamRequest{
				HeroID:   stringPtr("test-hero-001"),
				TeamName: "测试团队",
			},
			setupContext: func(c echo.Context) {
				// 不设置 user_id
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 准备请求
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/game/teams", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			// 创建 Echo Context
			c := e.NewContext(req, rec)
			tt.setupContext(c)

			// 调用 Handler
			err := handler.CreateTeam(c)

			// 验证结果
			if tt.expectedStatus == http.StatusOK {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)

				// 解析响应
				var resp response.Response
				err = json.Unmarshal(rec.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, xerrors.CodeSuccess, resp.Code)

				// 清理测试数据
				if resp.Data != nil {
					dataMap, ok := resp.Data.(map[string]interface{})
					if ok {
						teamID, ok := dataMap["id"].(string)
						if ok {
							defer func() {
								_, _ = db.Exec("DELETE FROM game_runtime.team_members WHERE team_id = $1", teamID)
								_, _ = db.Exec("DELETE FROM game_runtime.team_warehouses WHERE team_id = $1", teamID)
								_, _ = db.Exec("DELETE FROM game_runtime.teams WHERE id = $1", teamID)
							}()
						}
					}
				}
			} else {
				// 对于错误情况，可能会返回错误或设置状态码
				if err == nil {
					assert.Equal(t, tt.expectedStatus, rec.Code)
				}
			}
		})
	}
}

// TestTeamHandler_DisbandTeam 覆盖仓库非空与成功解散
func TestTeamHandler_DisbandTeam(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupTestDB(t)
	defer db.Close()

	handler, e := setupTestHandler(t, db)

	userID := testseed.EnsureUser(t, db, "team-handler-disband-user")
	heroID := testseed.EnsureHero(t, db, userID, "team-handler-disband-hero")
	testseed.CleanupTeamsByHero(t, db, heroID)

	// 创建团队
	createReq := &service.CreateTeamRequest{
		UserID:      userID.String(),
		HeroID:      heroID.String(),
		TeamName:    "disband-" + time.Now().Format("150405"),
		Description: "for disband",
	}
	teamSvc := service.NewServiceContainer(db, nil, nil).GetTeamService()
	team, err := teamSvc.CreateTeam(context.Background(), createReq)
	require.NoError(t, err)
	defer func() {
		_, _ = db.Exec("DELETE FROM game_runtime.team_members WHERE team_id = $1", team.ID)
		_, _ = db.Exec("DELETE FROM game_runtime.team_warehouses WHERE team_id = $1", team.ID)
		_, _ = db.Exec("DELETE FROM game_runtime.teams WHERE id = $1", team.ID)
	}()

	// 仓库加金币，期望 400
	_, err = db.Exec(`UPDATE game_runtime.team_warehouses SET gold_amount = 100 WHERE team_id = $1`, team.ID)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/game/teams/"+team.ID+"/disband?hero_id="+heroID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("team_id")
	c.SetParamValues(team.ID)

	_ = handler.DisbandTeam(c)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// 清空仓库，再次解散应 200
	_, err = db.Exec(`UPDATE game_runtime.team_warehouses SET gold_amount = 0 WHERE team_id = $1`, team.ID)
	require.NoError(t, err)

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("team_id")
	c.SetParamValues(team.ID)

	err = handler.DisbandTeam(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestTeamHandler_LeaveTeam 覆盖成员可离队、队长不可离队
func TestTeamHandler_LeaveTeam(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupTestDB(t)
	defer db.Close()

	handler, e := setupTestHandler(t, db)
	svcContainer := service.NewServiceContainer(db, nil, nil)
	teamSvc := svcContainer.GetTeamService()

	leaderUserID := testseed.EnsureUser(t, db, "team-handler-leave-leader")
	leaderHeroID := testseed.EnsureHero(t, db, leaderUserID, "team-handler-leave-leader-hero")
	testseed.CleanupTeamsByHero(t, db, leaderHeroID)

	team, err := teamSvc.CreateTeam(context.Background(), &service.CreateTeamRequest{
		UserID:   leaderUserID.String(),
		HeroID:   leaderHeroID.String(),
		TeamName: "leave-" + time.Now().Format("150405"),
	})
	require.NoError(t, err)

	memberUserID := testseed.EnsureUser(t, db, "team-handler-leave-member")
	memberHeroID := testseed.EnsureHero(t, db, memberUserID, "team-handler-leave-member-hero")
	_, err = db.Exec(`
		INSERT INTO game_runtime.team_members (id, team_id, hero_id, user_id, role, joined_at)
		VALUES (gen_random_uuid(), $1, $2, $3, 'member', NOW())
	`, team.ID, memberHeroID.String(), memberUserID.String())
	require.NoError(t, err)

	// 普通成员离队
	req := httptest.NewRequest(http.MethodPost, "/api/v1/game/teams/"+team.ID+"/leave?hero_id="+memberHeroID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("team_id")
	c.SetParamValues(team.ID)

	err = handler.LeaveTeam(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// 队长离队应失败
	req = httptest.NewRequest(http.MethodPost, "/api/v1/game/teams/"+team.ID+"/leave?hero_id="+leaderHeroID.String(), nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("team_id")
	c.SetParamValues(team.ID)

	_ = handler.LeaveTeam(c)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// TestTeamHandler_GetTeam 测试获取团队详情 API
func TestTeamHandler_GetTeam(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupTestDB(t)
	defer db.Close()

	handler, e := setupTestHandler(t, db)

	// 先创建一个测试团队
	serviceContainer := service.NewServiceContainer(db, nil, nil)
	teamService := serviceContainer.GetTeamService()

	leaderUserID := testseed.EnsureUser(t, db, "team-handler-leader")
	leaderHeroID := testseed.EnsureHero(t, db, leaderUserID, "team-handler-leader-hero")
	testseed.CleanupTeamsByHero(t, db, leaderHeroID)

	createReq := &service.CreateTeamRequest{
		UserID:      leaderUserID.String(),
		HeroID:      leaderHeroID.String(),
		TeamName:    "测试团队-" + time.Now().Format("20060102150405"),
		Description: "测试团队",
	}

	team, err := teamService.CreateTeam(context.Background(), createReq)
	require.NoError(t, err)

	defer func() {
		_, _ = db.Exec("DELETE FROM game_runtime.team_members WHERE team_id = $1", team.ID)
		_, _ = db.Exec("DELETE FROM game_runtime.team_warehouses WHERE team_id = $1", team.ID)
		_, _ = db.Exec("DELETE FROM game_runtime.teams WHERE id = $1", team.ID)
	}()

	// 测试获取团队详情
	req := httptest.NewRequest(http.MethodGet, "/api/v1/game/teams/"+team.ID, nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetParamNames("team_id")
	c.SetParamValues(team.ID)

	err = handler.GetTeam(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// 运行测试：
// go test -v ./internal/modules/game/handler -run TestTeamHandler
