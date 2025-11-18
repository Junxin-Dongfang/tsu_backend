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

// TestAdminLoginAndFetchUser exercises the admin login flow and ensures profile lookups succeed.
func TestAdminLoginAndFetchUser(t *testing.T) {
	cfg := apitest.LoadConfig()
	client := apitest.NewClient(cfg.BaseURL)
	ctx := context.Background()
	loginReq := apitest.LoginRequest{Identifier: cfg.AdminUsername, Password: cfg.AdminPassword}
	loginResp, httpResp, rawLogin, err := apitest.PostJSON[apitest.LoginRequest, apitest.LoginResponse](ctx, client, "/api/v1/admin/auth/login", loginReq, "")
	require.NoError(t, err, string(rawLogin))
	require.Equal(t, http.StatusOK, httpResp.StatusCode, string(rawLogin))
	require.Equal(t, int(xerrors.CodeSuccess), loginResp.Code, string(rawLogin))
	require.NotNil(t, loginResp.Data)
	token := loginResp.Data.SessionToken
	userID := loginResp.Data.UserID
	require.NotEmpty(t, token)
	require.NotEmpty(t, userID)

	userResp, httpResp2, raw2, err := apitest.GetJSON[apitest.GetUserResponse](ctx, client, fmt.Sprintf("/api/v1/admin/auth/users/%s", userID), token)
	require.NoError(t, err, string(raw2))
	require.Equal(t, http.StatusOK, httpResp2.StatusCode)
	require.Equal(t, int(xerrors.CodeSuccess), userResp.Code, string(raw2))
	require.NotNil(t, userResp.Data)
	require.Equal(t, userID, userResp.Data.UserID)
}
