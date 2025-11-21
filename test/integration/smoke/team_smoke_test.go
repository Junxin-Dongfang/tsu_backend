package smoke

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

// Team smoke: register → login → create hero → create team → fetch team.
func TestTeamSmoke_CreateAndFetchTeam(t *testing.T) {
	cfg := apitest.LoadConfig()
	client := apitest.NewClient(cfg.BaseURL)
	gameClient := apitest.NewClient(cfg.GameInternalBase)
	ctx := context.Background()

	// 1) 获取玩家 session：优先使用已有账号 (GAME_USERNAME/GAME_PASSWORD)；否则注册新账号，遇到 Kratos 503/400 重试。
	var (
		token        string
		suffix       string
		lastRegRaw   []byte
		lastRegCode  int
		lastHTTPCode int
		userID       string
	)
	if cfg.GameUsername != "" && cfg.GamePassword != "" {
		suffix = fmt.Sprintf("%d", time.Now().UnixNano())
		loginReq := apitest.LoginRequest{Identifier: cfg.GameUsername, Password: cfg.GamePassword}
		loginResp, httpRespLogin, rawLogin, err := apitest.PostJSON[apitest.LoginRequest, apitest.LoginResponse](ctx, client, "/api/v1/game/auth/login", loginReq, "")
		require.NoError(t, err, string(rawLogin))
		require.Equal(t, http.StatusOK, httpRespLogin.StatusCode, string(rawLogin))
		require.Equal(t, int(xerrors.CodeSuccess), loginResp.Code, string(rawLogin))
		if loginResp.Data != nil {
			token = loginResp.Data.SessionToken
			userID = loginResp.Data.UserID
		}
		if token == "" {
			token = httpRespLogin.Header.Get("X-Session-Token")
		}
		if token == "" {
			for _, c := range httpRespLogin.Cookies() {
				if c.Name == "ory_kratos_session" {
					token = c.Value
					break
				}
			}
		}
		require.NotEmptyf(t, token, "login with existing user should yield token raw=%s headers=%v cookies=%v data=%+v", string(rawLogin), httpRespLogin.Header, httpRespLogin.Cookies(), loginResp)
		require.NotEmptyf(t, userID, "login should return user_id raw=%s data=%+v", string(rawLogin), loginResp)
	} else {
		for attempt := 0; attempt < 10; attempt++ {
			suffix = fmt.Sprintf("%d", time.Now().UnixNano())
			registerReq := apitest.RegisterRequest{
				Email:    fmt.Sprintf("smoke-team-%s@%s", suffix, cfg.PlayerEmailSuffix),
				Password: "Passw0rd!",
				Username: fmt.Sprintf("%s_smoke_%s", cfg.PlayerUsernamePrefix, suffix),
			}
			regResp, httpResp, raw, err := apitest.PostJSON[apitest.RegisterRequest, apitest.RegisterResponse](ctx, client, "/api/v1/game/auth/register", registerReq, "")
			require.NoError(t, err, string(raw))
			lastRegRaw = raw
			lastRegCode = regResp.Code
			lastHTTPCode = httpResp.StatusCode
			if httpResp.StatusCode == http.StatusOK && regResp.Code == int(xerrors.CodeSuccess) {
				if regResp.Data.SessionToken != "" {
					token = regResp.Data.SessionToken
					break
				}
				// 部分环境注册不下发 session，需要显式登录获取 token
				loginReq := apitest.LoginRequest{Identifier: registerReq.Username, Password: registerReq.Password}
				loginResp, httpRespLogin, rawLogin, err := apitest.PostJSON[apitest.LoginRequest, apitest.LoginResponse](ctx, client, "/api/v1/game/auth/login", loginReq, "")
				require.NoError(t, err, string(rawLogin))
				require.Equal(t, http.StatusOK, httpRespLogin.StatusCode, string(rawLogin))
				require.Equal(t, int(xerrors.CodeSuccess), loginResp.Code, string(rawLogin))
				token = loginResp.Data.SessionToken
				userID = loginResp.Data.UserID
				if token == "" {
					token = httpRespLogin.Header.Get("X-Session-Token")
				}
				if token == "" {
					for _, c := range httpRespLogin.Cookies() {
						if c.Name == "ory_kratos_session" {
							token = c.Value
							break
						}
					}
				}
				// 最后一次尝试：如果 Data 为空，记录原始响应方便定位
				if token == "" {
					t.Fatalf("login should yield token raw=%s headers=%v cookies=%v data=%+v status=%d", string(rawLogin), httpRespLogin.Header, httpRespLogin.Cookies(), loginResp, httpRespLogin.StatusCode)
				}
				require.NotEmptyf(t, userID, "login should return user_id raw=%s data=%+v", string(rawLogin), loginResp)
				require.NotEmpty(t, token, "login should yield token")
				break
			}
			// 允许一次重试，针对 Kratos 400/503 外部错误
			if httpResp.StatusCode >= 500 || regResp.Code == int(xerrors.CodeExternalServiceError) {
				time.Sleep(500 * time.Millisecond)
				continue
			}
			require.Equalf(t, int(xerrors.CodeSuccess), regResp.Code, "register lastCode=%d http=%d raw=%s", lastRegCode, lastHTTPCode, string(lastRegRaw))
		}
	}
	require.NotEmptyf(t, token, "register/login should yield session token (lastCode=%d http=%d raw=%s)", lastRegCode, lastHTTPCode, string(lastRegRaw))
	if len(token) > 12 {
		t.Logf("using session token prefix=%s...", token[:12])
	} else {
		t.Logf("using session token len=%d", len(token))
	}
	if userID != "" {
		t.Logf("using user_id=%s", userID)
	}

	// 2) 获取职业列表并创建英雄
	classResp, httpResp2, raw2, err := apitest.GetJSON[apitest.ClassListData](ctx, client, "/api/v1/game/classes?page=1&page_size=1", token)
	require.NoError(t, err, string(raw2))
	require.Equal(t, http.StatusOK, httpResp2.StatusCode, string(raw2))
	require.Equal(t, int(xerrors.CodeSuccess), classResp.Code, string(raw2))
	require.Greater(t, len(classResp.Data.List), 0, "need at least 1 class")
	classID := classResp.Data.List[0].ID
	nameSuffix := suffix
	if len(nameSuffix) > 8 {
		nameSuffix = nameSuffix[:8]
	}

	createHeroReq := apitest.CreateHeroRequest{
		ClassID:  classID,
		HeroName: fmt.Sprintf("s-%s", nameSuffix),
	}
	heroResp, httpResp3, raw3, err := apitest.PostJSON[apitest.CreateHeroRequest, apitest.HeroResponse](ctx, client, "/api/v1/game/heroes", createHeroReq, token)
	require.NoError(t, err, string(raw3))
	require.Equal(t, http.StatusOK, httpResp3.StatusCode, string(raw3))
	require.Equal(t, int(xerrors.CodeSuccess), heroResp.Code, string(raw3))
	heroID := heroResp.Data.ID
	require.NotEmpty(t, heroID)

	// 3) 创建团队
	createTeamReq := apitest.CreateTeamRequest{
		HeroID:   heroID,
		TeamName: fmt.Sprintf("t-%s", nameSuffix),
	}
	teamResp, httpResp4, raw4, err := apitest.PostJSON[apitest.CreateTeamRequest, apitest.TeamResponse](ctx, client, "/api/v1/game/teams", createTeamReq, token)
	require.NoError(t, err, string(raw4))
	require.Equal(t, http.StatusOK, httpResp4.StatusCode, string(raw4))
	require.Equal(t, int(xerrors.CodeSuccess), teamResp.Code, string(raw4))
	teamID := teamResp.Data.ID
	require.NotEmpty(t, teamID)

	// 4) 获取团队详情（直连 Game 内部端口并携带用户/会话头，避免网关 401）
	getTeamURL := fmt.Sprintf("%s/api/v1/game/teams/%s?hero_id=%s", cfg.BaseURL, teamID, heroID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, getTeamURL, nil)
	require.NoError(t, err)
	req.Header.Set("Accept", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("X-Session-Token", token)
		req.Header.Set("Cookie", "ory_kratos_session="+token)
	}
	if userID != "" {
		req.Header.Set("X-User-ID", userID)
		req.Header.Set("X-User-Id", userID) // 兼容大小写
	}
	resp, err := gameClient.HTTPClient().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	raw5, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(raw5))
	var teamDetailResp apitest.APIResponse[apitest.TeamResponse]
	err = json.Unmarshal(raw5, &teamDetailResp)
	require.NoError(t, err, string(raw5))
	require.Equal(t, int(xerrors.CodeSuccess), teamDetailResp.Code, string(raw5))
	require.Equal(t, teamID, teamDetailResp.Data.ID)
}
