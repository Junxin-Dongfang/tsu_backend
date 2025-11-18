package modules

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

func TestAdminUsersListRequiresAuth(t *testing.T) {
	ctx, _, client, _ := setup(t)
	resp, httpResp, raw, err := apitest.GetJSON[apitest.AdminUserListResponse](ctx, client, "/api/v1/admin/users", "")
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusUnauthorized, httpResp.StatusCode)
	require.Contains(t, string(raw), "Unauthorized")
	_ = resp
}

func TestAdminUsersListSuccess(t *testing.T) {
	ctx, cfg, client, _ := setup(t)
	token := adminToken(t, ctx, client, cfg)

	resp, httpResp, raw, err := apitest.GetJSON[apitest.AdminUserListResponse](ctx, client, "/api/v1/admin/users", token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.Equal(t, int(xerrors.CodeSuccess), resp.Code, string(raw))
	require.NotNil(t, resp.Data)
}
