package modules

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

type player struct {
	Token    string
	UserID   string
	HeroID   string
	Username string
	Password string
}

func adminToken(t *testing.T, ctx context.Context, client *apitest.Client, cfg apitest.Config) string {
	loginReq := apitest.LoginRequest{Identifier: cfg.AdminUsername, Password: cfg.AdminPassword}
	resp, httpResp, raw, err := apitest.PostJSON[apitest.LoginRequest, apitest.LoginResponse](ctx, client, "/api/v1/admin/auth/login", loginReq, "")
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusOK, httpResp.StatusCode, string(raw))
	require.Equal(t, int(xerrors.CodeSuccess), resp.Code, string(raw))
	require.NotNil(t, resp.Data)
	return resp.Data.SessionToken
}

func setup(t *testing.T) (context.Context, apitest.Config, *apitest.Client, apitest.FixtureFactory) {
	t.Helper()
	cfg := apitest.LoadConfig()
	return context.Background(), cfg, apitest.NewClient(cfg.BaseURL), apitest.NewFixtureFactory(cfg)
}

func registerPlayerWithHero(t *testing.T, ctx context.Context, client *apitest.Client, factory apitest.FixtureFactory, tag string) player {
	t.Helper()
	username, email, password := factory.UniquePlayerCredentials(tag)
	regReq := apitest.RegisterRequest{Email: email, Password: password, Username: username}
	regResp, httpResp, raw, err := apitest.PostJSON[apitest.RegisterRequest, apitest.RegisterResponse](ctx, client, "/api/v1/game/auth/register", regReq, "")
	if err != nil || httpResp.StatusCode != http.StatusOK || regResp.Code != int(xerrors.CodeSuccess) {
		t.Skipf("注册玩家失败，跳过模块测试：status=%v code=%v body=%s", apitest.Status(httpResp), regResp.Code, string(raw))
	}

	loginReq := apitest.LoginRequest{Identifier: username, Password: password}
	loginResp, httpResp2, raw2, err := apitest.PostJSON[apitest.LoginRequest, apitest.LoginResponse](ctx, client, "/api/v1/game/auth/login", loginReq, "")
	if err != nil || httpResp2.StatusCode != http.StatusOK || loginResp.Code != int(xerrors.CodeSuccess) {
		t.Skipf("登录玩家失败，跳过模块测试：status=%v code=%v body=%s", apitest.Status(httpResp2), loginResp.Code, string(raw2))
	}

	classID := fetchClassID(t, ctx, client, loginResp.Data.SessionToken)
	heroReq := factory.BuildCreateHeroRequest(classID, "")
	heroResp, httpResp3, raw3, err := apitest.PostJSON[apitest.CreateHeroRequest, apitest.HeroResponse](ctx, client, "/api/v1/game/heroes", heroReq, loginResp.Data.SessionToken)
	if err != nil || httpResp3.StatusCode != http.StatusOK || heroResp.Code != int(xerrors.CodeSuccess) {
		t.Skipf("创建英雄失败，跳过模块测试：status=%v code=%v body=%s", apitest.Status(httpResp3), heroResp.Code, string(raw3))
	}

	return player{Token: loginResp.Data.SessionToken, UserID: loginResp.Data.UserID, HeroID: heroResp.Data.ID, Username: username, Password: password}
}

func fetchClassID(t *testing.T, ctx context.Context, client *apitest.Client, token string) string {
	t.Helper()
	resp, httpResp, raw, err := apitest.GetJSON[apitest.ClassListData](ctx, client, "/api/v1/game/classes?page=1&page_size=1", token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.NotNil(t, resp.Data)
	require.NotEmpty(t, resp.Data.List)
	return resp.Data.List[0].ID
}

func createTeamForPlayer(t *testing.T, ctx context.Context, client *apitest.Client, factory apitest.FixtureFactory, p player) apitest.TeamResponse {
	t.Helper()
	teamReq := factory.BuildCreateTeamRequest(p.HeroID)
	resp, httpResp, raw, err := apitest.PostJSON[apitest.CreateTeamRequest, apitest.TeamResponse](ctx, client, "/api/v1/game/teams", teamReq, p.Token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.Equal(t, int(xerrors.CodeSuccess), resp.Code, string(raw))
	require.NotNil(t, resp.Data)
	return *resp.Data
}

func createAdminItem(t *testing.T, ctx context.Context, client *apitest.Client, token string, factory apitest.FixtureFactory, tag string) string {
	t.Helper()
	code := fmt.Sprintf("%s-%s", tag, apitest.UniqueSuffix())
	if factory.RunID != "" {
		code = fmt.Sprintf("%s-%s", code, factory.RunID)
	}
	payload := map[string]interface{}{
		"item_code":    code,
		"item_name":    "自动化" + code,
		"item_level":   1,
		"item_quality": "normal",
		"item_type":    "equipment",
		"equip_slot":   "mainhand",
	}
	resp, httpResp, raw, err := apitest.PostJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/admin/items", payload, token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusOK, httpResp.StatusCode, string(raw))
	require.Equal(t, int(xerrors.CodeSuccess), resp.Code, string(raw))
	require.NotNil(t, resp.Data, string(raw))
	data := *resp.Data
	id, ok := data["id"].(string)
	require.True(t, ok, string(raw))
	require.NotEmpty(t, id)

	t.Cleanup(func() {
		cleanupAdminResource(t, client, token, "/api/v1/admin/items/"+id)
	})

	return id
}

func cleanupAdminResource(t *testing.T, client *apitest.Client, token, path string) {
	t.Helper()
	ctx := context.Background()
	resp, httpResp, raw, err := apitest.DeleteJSON[struct{}, map[string]interface{}](ctx, client, path, nil, token)
	if err != nil {
		t.Logf("清理 %s 失败: %v body=%s", path, err, string(raw))
		return
	}
	if httpResp.StatusCode == http.StatusNotFound {
		return
	}
	if httpResp.StatusCode != http.StatusOK {
		t.Logf("清理 %s 返回 %d body=%s", path, httpResp.StatusCode, string(raw))
		return
	}
	if resp != nil && resp.Code != int(xerrors.CodeSuccess) {
		t.Logf("清理 %s 返回业务码 %d body=%s", path, resp.Code, string(raw))
	}
}

func inviteAndAcceptMember(t *testing.T, ctx context.Context, client *apitest.Client, team apitest.TeamResponse, leader player, recruit player) {
	invitePayload := apitest.InviteMemberRequest{TeamID: team.ID, InviterHeroID: leader.HeroID, InviteeHeroID: recruit.HeroID}
	invitePath := fmt.Sprintf("/api/v1/game/teams/invite?team_id=%s&hero_id=%s", team.ID, leader.HeroID)
	resp, httpResp, raw, err := apitest.PostJSON[apitest.InviteMemberRequest, map[string]interface{}](ctx, client, invitePath, invitePayload, leader.Token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusOK, httpResp.StatusCode, string(raw))
	require.Equal(t, int(xerrors.CodeSuccess), resp.Code, string(raw))

	invitationID := fetchLatestInvitationID(t, team.ID, recruit.HeroID)
	approveReq := apitest.ApproveInvitationRequest{InvitationID: invitationID, HeroID: leader.HeroID, Approved: true}
	approvePath := fmt.Sprintf("/api/v1/game/teams/invite/approve?team_id=%s&hero_id=%s", team.ID, leader.HeroID)
	approveResp, approveHTTP, approveRaw, err := apitest.PostJSON[apitest.ApproveInvitationRequest, map[string]interface{}](ctx, client, approvePath, approveReq, leader.Token)
	require.NoError(t, err, string(approveRaw))
	require.Equal(t, http.StatusOK, approveHTTP.StatusCode, string(approveRaw))
	require.Equal(t, int(xerrors.CodeSuccess), approveResp.Code, string(approveRaw))

	acceptReq := apitest.AcceptInvitationRequest{InvitationID: invitationID, HeroID: recruit.HeroID}
	acceptPath := fmt.Sprintf("/api/v1/game/teams/invite/accept?team_id=%s&hero_id=%s", team.ID, recruit.HeroID)
	acceptResp, acceptHTTP, acceptRaw, err := apitest.PostJSON[apitest.AcceptInvitationRequest, map[string]interface{}](ctx, client, acceptPath, acceptReq, recruit.Token)
	require.NoError(t, err, string(acceptRaw))
	require.Equal(t, http.StatusOK, acceptHTTP.StatusCode, string(acceptRaw))
	require.Equal(t, int(xerrors.CodeSuccess), acceptResp.Code, string(acceptRaw))
}

func fetchActiveDungeonID(t *testing.T) string {
	db := openTestDB(t)
	defer db.Close()
	var id string
	query := `SELECT id FROM game_config.dungeons WHERE deleted_at IS NULL AND is_active = TRUE AND min_level <= 1 LIMIT 1`
	err := db.QueryRowContext(context.Background(), query).Scan(&id)
	if err == sql.ErrNoRows {
		t.Skip("未找到满足条件的地城配置")
	}
	require.NoError(t, err)
	return id
}

func fetchLatestInvitationID(t *testing.T, teamID, inviteeHeroID string) string {
	t.Helper()
	db := openTestDB(t)
	defer db.Close()
	var id string
	query := `SELECT id FROM game_runtime.team_invitations WHERE team_id=$1 AND invitee_hero_id=$2 ORDER BY created_at DESC LIMIT 1`
	err := db.QueryRowContext(context.Background(), query, teamID, inviteeHeroID).Scan(&id)
	require.NoErrorf(t, err, "未找到邀请记录 team=%s invitee=%s", teamID, inviteeHeroID)
	return id
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getenv("TEST_DB_HOST", "localhost"),
		getenv("TEST_DB_PORT", "5432"),
		getenv("TEST_DB_USER", "postgres"),
		getenv("TEST_DB_PASSWORD", "postgres"),
		getenv("TEST_DB_NAME", "tsu_db"),
	)
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	require.NoError(t, db.Ping())
	return db
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
