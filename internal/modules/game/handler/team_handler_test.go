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
				HeroID:   heroID.String(),
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
				HeroID:   heroID.String(),
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
				HeroID:   "test-hero-001",
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
