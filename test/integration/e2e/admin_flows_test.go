package e2e

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/test/internal/apitest"
)

// Flow 5: 管理员登录 -> 列表用户 -> 查看用户详情。
func TestFlow_AdminUserListingAndDetail(t *testing.T) {
	ctx, cfg, client, _ := setupTest(t)
	token := adminLogin(t, ctx, client, cfg)

    listPath := "/api/v1/admin/users?page=1&page_size=5"
	listResp, httpResp, raw, err := apitest.GetJSON[apitest.AdminUserListResponse](ctx, client, listPath, token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.NotNil(t, listResp.Data)

	if len(listResp.Data.Users) == 0 {
		t.Skip("admin 用户列表为空，无法继续详情校验")
	}

	target := listResp.Data.Users[0]
    detailPath := fmt.Sprintf("/api/v1/admin/users/%s", target.ID)
	detailResp, httpResp2, raw2, err := apitest.GetJSON[apitest.AdminUserInfo](ctx, client, detailPath, token)
	require.NoError(t, err, string(raw2))
	require.Equal(t, http.StatusOK, httpResp2.StatusCode)
	require.NotNil(t, detailResp.Data)
	require.Equal(t, target.ID, detailResp.Data.ID)
	require.Equal(t, target.Username, detailResp.Data.Username)
}
