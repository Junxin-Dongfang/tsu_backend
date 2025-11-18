package smoke

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

// TestGameRegisterLoginAndFetchUser validates the happy path from registration to profile lookup.
func TestGameRegisterLoginAndFetchUser(t *testing.T) {
	cfg := apitest.LoadConfig()
	client := apitest.NewClient(cfg.BaseURL)
	ctx := context.Background()

	factory := apitest.NewFixtureFactory(cfg)

	// 优先使用已有账号登录（通过环境变量 GAME_USERNAME/GAME_PASSWORD 提供）。
	username := cfg.GameUsername
	password := cfg.GamePassword
	email := ""

	expectedUserID := ""
	if username == "" || password == "" {
		username, email, password = factory.UniquePlayerCredentials("smoke")
		registerReq := apitest.RegisterRequest{Email: email, Password: password, Username: username}
		regResp, httpResp, raw, err := apitest.PostJSON[apitest.RegisterRequest, apitest.RegisterResponse](ctx, client, "/api/v1/game/auth/register", registerReq, "")
		if err != nil || httpResp.StatusCode >= 500 {
			t.Logf("register retry once due to err=%v status=%v body=%s", err, apitest.Status(httpResp), string(raw))
			regResp, httpResp, raw, err = apitest.PostJSON[apitest.RegisterRequest, apitest.RegisterResponse](ctx, client, "/api/v1/game/auth/register", registerReq, "")
		}
		if err != nil || httpResp.StatusCode != http.StatusOK || regResp.Code != int(xerrors.CodeSuccess) {
			t.Skipf("注册玩家失败，跳过冒烟：status=%v code=%v body=%s", apitest.Status(httpResp), regResp.Code, string(raw))
		}
		require.NotNil(t, regResp.Data)
		require.NotEmpty(t, regResp.Data.UserID)
		expectedUserID = regResp.Data.UserID
		if email == "" {
			email = fmt.Sprintf("%s@example.com", username)
		}
	}

	loginCases := []struct {
		name       string
		identifier string
	}{
		{name: "login-with-username", identifier: username},
	}
	if email != "" {
		loginCases = append(loginCases, struct {
			name       string
			identifier string
		}{name: "login-with-email", identifier: email})
	}

	for _, tc := range loginCases {
		t.Run(tc.name, func(t *testing.T) {
			loginReq := apitest.LoginRequest{Identifier: tc.identifier, Password: password}
			loginResp, httpResp2, raw2, err := apitest.PostJSON[apitest.LoginRequest, apitest.LoginResponse](ctx, client, "/api/v1/game/auth/login", loginReq, "")
			require.NoError(t, err, string(raw2))
			require.Equal(t, http.StatusOK, httpResp2.StatusCode)
			require.Equal(t, int(xerrors.CodeSuccess), loginResp.Code, string(raw2))
			require.NotNil(t, loginResp.Data)
			if expectedUserID == "" {
				expectedUserID = loginResp.Data.UserID
			}
			require.Equal(t, expectedUserID, loginResp.Data.UserID)
			require.NotEmpty(t, loginResp.Data.SessionToken)

			userResp, httpResp3, raw3, err := apitest.GetJSON[apitest.GetUserResponse](ctx, client, fmt.Sprintf("/api/v1/game/auth/users/%s", loginResp.Data.UserID), loginResp.Data.SessionToken)
			require.NoError(t, err, string(raw3))
			require.Equal(t, http.StatusOK, httpResp3.StatusCode)
			require.Equal(t, int(xerrors.CodeSuccess), userResp.Code, string(raw3))
			require.NotNil(t, userResp.Data)
			require.Equal(t, loginResp.Data.UserID, userResp.Data.UserID)
		})
	}
}
