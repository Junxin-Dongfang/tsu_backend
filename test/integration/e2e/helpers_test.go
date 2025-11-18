package e2e

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

type playerSession struct {
	Username string
	Email    string
	Password string
	Token    string
	UserID   string
}

func setupTest(t *testing.T) (context.Context, apitest.Config, *apitest.Client, apitest.FixtureFactory) {
	t.Helper()
	cfg := apitest.LoadConfig()
	client := apitest.NewClient(cfg.BaseURL)
	factory := apitest.NewFixtureFactory(cfg)
	return context.Background(), cfg, client, factory
}

func registerPlayer(t *testing.T, ctx context.Context, client *apitest.Client, factory apitest.FixtureFactory, tag string) playerSession {
	t.Helper()
	username, email, password := factory.UniquePlayerCredentials(tag)
	regReq := apitest.RegisterRequest{Email: email, Password: password, Username: username}
	regResp, httpResp, raw, err := apitest.PostJSON[apitest.RegisterRequest, apitest.RegisterResponse](ctx, client, "/api/v1/game/auth/register", regReq, "")
	if err != nil || httpResp.StatusCode != http.StatusOK || regResp.Code != int(xerrors.CodeSuccess) {
		t.Skipf("注册玩家失败，跳过E2E：status=%v code=%v body=%s", apitest.Status(httpResp), regResp.Code, string(raw))
	}

	loginReq := apitest.LoginRequest{Identifier: username, Password: password}
	loginResp, httpResp2, raw2, err := apitest.PostJSON[apitest.LoginRequest, apitest.LoginResponse](ctx, client, "/api/v1/game/auth/login", loginReq, "")
	if err != nil || httpResp2.StatusCode != http.StatusOK || loginResp.Code != int(xerrors.CodeSuccess) {
		t.Skipf("登录玩家失败，跳过E2E：status=%v code=%v body=%s", apitest.Status(httpResp2), loginResp.Code, string(raw2))
	}

	return playerSession{
		Username: username,
		Email:    email,
		Password: password,
		Token:    loginResp.Data.SessionToken,
		UserID:   loginResp.Data.UserID,
	}
}

func fetchFirstClassID(t *testing.T, ctx context.Context, client *apitest.Client, token string) string {
	t.Helper()
	resp, httpResp, raw, err := apitest.GetJSON[apitest.ClassListData](ctx, client, "/api/v1/game/classes?page=1&page_size=5", token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.NotNil(t, resp.Data)
	require.NotEmpty(t, resp.Data.List)
	return resp.Data.List[0].ID
}

func createHero(t *testing.T, ctx context.Context, client *apitest.Client, token string, req apitest.CreateHeroRequest) apitest.HeroResponse {
	t.Helper()
	resp, httpResp, raw, err := apitest.PostJSON[apitest.CreateHeroRequest, apitest.HeroResponse](ctx, client, "/api/v1/game/heroes", req, token)
	require.NoError(t, err, string(raw))
	if httpResp.StatusCode != http.StatusOK {
		t.Fatalf("创建英雄失败，HTTP=%d Body=%s", httpResp.StatusCode, string(raw))
	}
	require.NotNil(t, resp.Data)
	return *resp.Data
}

func getHero(t *testing.T, ctx context.Context, client *apitest.Client, token, heroID string) apitest.HeroResponse {
	t.Helper()
	path := fmt.Sprintf("/api/v1/game/heroes/%s", heroID)
	resp, httpResp, raw, err := apitest.GetJSON[apitest.HeroResponse](ctx, client, path, token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.NotNil(t, resp.Data)
	return *resp.Data
}

func createTeam(t *testing.T, ctx context.Context, client *apitest.Client, token string, req apitest.CreateTeamRequest) apitest.TeamResponse {
	t.Helper()
	resp, httpResp, raw, err := apitest.PostJSON[apitest.CreateTeamRequest, apitest.TeamResponse](ctx, client, "/api/v1/game/teams", req, token)
	require.NoError(t, err, string(raw))
	if httpResp.StatusCode != http.StatusOK {
		t.Fatalf("创建团队失败，HTTP=%d Body=%s", httpResp.StatusCode, string(raw))
	}
	require.NotNil(t, resp.Data)
	return *resp.Data
}

func updateTeam(t *testing.T, ctx context.Context, client *apitest.Client, token, teamID, heroID string, req apitest.UpdateTeamInfoRequest) apitest.TeamResponse {
	t.Helper()
	path := fmt.Sprintf("/api/v1/game/teams/%s?hero_id=%s", teamID, heroID)
	resp, httpResp, raw, err := apitest.PutJSON[apitest.UpdateTeamInfoRequest, apitest.TeamResponse](ctx, client, path, req, token)
	require.NoError(t, err, string(raw))
	if httpResp.StatusCode != http.StatusOK {
		t.Fatalf("更新团队失败，HTTP=%d Body=%s", httpResp.StatusCode, string(raw))
	}
	require.NotNil(t, resp.Data)
	return *resp.Data
}

func getTeam(t *testing.T, ctx context.Context, client *apitest.Client, token, teamID string) apitest.TeamResponse {
	t.Helper()
	path := fmt.Sprintf("/api/v1/game/teams/%s", teamID)
	resp, httpResp, raw, err := apitest.GetJSON[apitest.TeamResponse](ctx, client, path, token)
	require.NoError(t, err, string(raw))
	if httpResp.StatusCode != http.StatusOK {
		t.Skipf("获取团队失败，status=%d body=%s", httpResp.StatusCode, string(raw))
	}
	require.NotNil(t, resp.Data)
	return *resp.Data
}

func viewWarehouse(t *testing.T, ctx context.Context, client *apitest.Client, token, teamID, heroID string) apitest.WarehouseResponse {
	t.Helper()
	path := fmt.Sprintf("/api/v1/game/teams/%s/warehouse?hero_id=%s", teamID, heroID)
	resp, httpResp, raw, err := apitest.GetJSON[apitest.WarehouseResponse](ctx, client, path, token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.NotNil(t, resp.Data)
	return *resp.Data
}

func disbandTeam(t *testing.T, ctx context.Context, client *apitest.Client, token, teamID, heroID string) *apitest.APIResponse[map[string]interface{}] {
	t.Helper()
	path := fmt.Sprintf("/api/v1/game/teams/%s/disband?hero_id=%s", teamID, heroID)
	var empty struct{}
	resp, httpResp, raw, err := apitest.PostJSON[struct{}, map[string]interface{}](ctx, client, path, empty, token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.Equal(t, int(xerrors.CodeSuccess), resp.Code)
	return resp
}

func adminLogin(t *testing.T, ctx context.Context, client *apitest.Client, cfg apitest.Config) string {
	t.Helper()
	loginReq := apitest.LoginRequest{Identifier: cfg.AdminUsername, Password: cfg.AdminPassword}
	resp, httpResp, raw, err := apitest.PostJSON[apitest.LoginRequest, apitest.LoginResponse](ctx, client, "/api/v1/admin/auth/login", loginReq, "")
	if err != nil || httpResp.StatusCode != http.StatusOK || resp.Code != int(xerrors.CodeSuccess) {
		t.Skipf("管理员登录失败，跳过E2E：status=%v code=%v body=%s", apitest.Status(httpResp), resp.Code, string(raw))
	}
	require.NotNil(t, resp.Data)
	return resp.Data.SessionToken
}
