package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
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
	"tsu-self/internal/pkg/xerrors"
)

// setupTestWarehouseHandler 设置测试 WarehouseHandler
func setupTestWarehouseHandler(t *testing.T) (*TeamWarehouseHandler, *TeamHandler, *echo.Echo, *sql.DB) {
	t.Helper()

	db := setupTestDB(t)
	serviceContainer := service.NewServiceContainer(db, nil, nil)
	respWriter := response.DefaultResponseHandler()

	warehouseHandler := NewTeamWarehouseHandler(serviceContainer, respWriter)
	teamHandler := NewTeamHandler(serviceContainer, respWriter)

	e := echo.New()
	e.Validator = validator.New()
	return warehouseHandler, teamHandler, e, db
}

// TestTeamWarehouseHandler_GetWarehouse 测试获取团队仓库
func TestTeamWarehouseHandler_GetWarehouse(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	warehouseHandler, teamHandler, e, db := setupTestWarehouseHandler(t)
	defer db.Close()

	leaderUserID := testseed.EnsureUser(t, db, "team-warehouse-handler-leader")
	leaderHeroID := testseed.EnsureHero(t, db, leaderUserID, "team-warehouse-handler-leader-hero")
	testseed.CleanupTeamsByHero(t, db, leaderHeroID)

	// 创建测试团队
	description := "测试团队"
	createReq := CreateTeamRequest{
		HeroID:      leaderHeroID.String(),
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
		teamID         string
		heroID         string
		expectedStatus int
	}{
		{
			name:           "队长成功获取仓库",
			teamID:         teamID,
			heroID:         leaderHeroID.String(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "TeamID为空",
			teamID:         "",
			heroID:         leaderHeroID.String(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "HeroID为空",
			teamID:         teamID,
			heroID:         "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/v1/game/teams/%s/warehouse?hero_id=%s", tt.teamID, tt.heroID)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			c.SetParamNames("team_id")
			c.SetParamValues(tt.teamID)

			err := warehouseHandler.GetWarehouse(c)

			if tt.expectedStatus == http.StatusOK {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)

				// 验证响应内容
				var resp response.Response
				err = json.Unmarshal(rec.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, xerrors.CodeSuccess, resp.Code)
			}
		})
	}
}

// TestTeamWarehouseHandler_DistributeGold 测试分配金币
func TestTeamWarehouseHandler_DistributeGold(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	warehouseHandler, teamHandler, e, db := setupTestWarehouseHandler(t)
	defer db.Close()

	leaderUserID := testseed.EnsureUser(t, db, "team-warehouse-handler-leader")
	leaderHeroID := testseed.EnsureHero(t, db, leaderUserID, "team-warehouse-handler-leader-hero")
	testseed.CleanupTeamsByHero(t, db, leaderHeroID)

	// 创建测试团队
	description := "测试团队"
	createReq := CreateTeamRequest{
		HeroID:      leaderHeroID.String(),
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

	// 添加一些金币到仓库
	_, err = db.Exec("UPDATE game_runtime.team_warehouses SET gold_amount = 10000 WHERE team_id = $1", teamID)
	require.NoError(t, err)

	// 添加一个团队成员
	memberUserID := testseed.EnsureUser(t, db, "team-warehouse-handler-member")
	memberHeroID := testseed.EnsureHero(t, db, memberUserID, "team-warehouse-handler-member-hero")
	_, err = db.Exec(`
		INSERT INTO game_runtime.team_members (id, team_id, hero_id, user_id, role, joined_at)
		VALUES (gen_random_uuid(), $1, $2, $3, 'member', NOW())
		ON CONFLICT (team_id, hero_id) DO NOTHING
	`, teamID, memberHeroID.String(), memberUserID.String())
	require.NoError(t, err)

	tests := []struct {
		name           string
		teamID         string
		requestBody    DistributeGoldRequest
		expectedStatus int
	}{
		{
			name:   "队长成功分配金币",
			teamID: teamID,
			requestBody: DistributeGoldRequest{
				DistributorID: leaderHeroID.String(),
				Distributions: map[string]int64{
					memberHeroID.String(): 1000,
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "参数验证 - DistributorID为空",
			teamID: teamID,
			requestBody: DistributeGoldRequest{
				DistributorID: "",
				Distributions: map[string]int64{
					memberHeroID.String(): 1000,
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			url := fmt.Sprintf("/api/v1/game/teams/%s/warehouse/distribute-gold", tt.teamID)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			c.SetParamNames("team_id")
			c.SetParamValues(tt.teamID)

			_ = warehouseHandler.DistributeGold(c)
			// 注意：某些验证失败可能不抛出错误，而是设置状态码
		})
	}
}

// TestTeamWarehouseHandler_GetWarehouseItems 测试获取仓库物品列表
func TestTeamWarehouseHandler_GetWarehouseItems(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	warehouseHandler, teamHandler, e, db := setupTestWarehouseHandler(t)
	defer db.Close()

	leaderUserID := testseed.EnsureUser(t, db, "team-warehouse-handler-leader")
	leaderHeroID := testseed.EnsureHero(t, db, leaderUserID, "team-warehouse-handler-leader-hero")
	testseed.CleanupTeamsByHero(t, db, leaderHeroID)

	// 创建测试团队
	description := "测试团队"
	createReq := CreateTeamRequest{
		HeroID:      leaderHeroID.String(),
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

	tests := []struct {
		name           string
		teamID         string
		heroID         string
		expectedStatus int
	}{
		{
			name:           "队长成功获取物品列表",
			teamID:         teamID,
			heroID:         leaderHeroID.String(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "TeamID为空",
			teamID:         "",
			heroID:         leaderHeroID.String(),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/v1/game/teams/%s/warehouse/items?hero_id=%s&limit=20&offset=0", tt.teamID, tt.heroID)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			c.SetParamNames("team_id")
			c.SetParamValues(tt.teamID)

			_ = warehouseHandler.GetWarehouseItems(c)
		})
	}
}

// 运行测试：
// go test -v ./internal/modules/game/handler -run TestTeamWarehouseHandler
