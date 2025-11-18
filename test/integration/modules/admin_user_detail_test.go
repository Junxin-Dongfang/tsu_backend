package modules

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

func TestAdminUserDetailRequiresAuth(t *testing.T) {
	ctx, _, client, _ := setup(t)
	path := "/api/v1/admin/users/some-id"
	resp, httpResp, raw, err := apitest.GetJSON[apitest.AdminUserInfo](ctx, client, path, "")
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusUnauthorized, httpResp.StatusCode)
	require.Contains(t, string(raw), "Unauthorized")
	_ = resp
}

func TestAdminUserDetailSuccess(t *testing.T) {
	ctx, cfg, client, _ := setup(t)
	token := adminToken(t, ctx, client, cfg)

	listResp, httpResp, raw, err := apitest.GetJSON[apitest.AdminUserListResponse](ctx, client, "/api/v1/admin/users?page=1&page_size=1", token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.Equal(t, int(xerrors.CodeSuccess), listResp.Code, string(raw))
	require.NotNil(t, listResp.Data)
	require.NotEmpty(t, listResp.Data.Users)

	userID := listResp.Data.Users[0].ID
	detailPath := fmt.Sprintf("/api/v1/admin/users/%s", userID)
	detailResp, httpResp2, raw2, err := apitest.GetJSON[apitest.AdminUserInfo](ctx, client, detailPath, token)
	require.NoError(t, err, string(raw2))
	require.Equal(t, http.StatusOK, httpResp2.StatusCode)
	require.Equal(t, int(xerrors.CodeSuccess), detailResp.Code, string(raw2))
	require.NotNil(t, detailResp.Data)
	require.Equal(t, userID, detailResp.Data.ID)
}
