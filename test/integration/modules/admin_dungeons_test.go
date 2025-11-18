package modules

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

// 地城配置基础读取。
func TestAdminDungeonsList(t *testing.T) {
	ctx, cfg, client, _ := setup(t)
	token := adminToken(t, ctx, client, cfg)

	resp, httpResp, raw, err := apitest.GetJSON[map[string]interface{}](ctx, client, "/api/v1/admin/dungeons?page=1&page_size=5", token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusOK, httpResp.StatusCode, string(raw))
	require.Equal(t, int(xerrors.CodeSuccess), resp.Code, string(raw))
}

func TestAdminDungeonsListUnauthorized(t *testing.T) {
	ctx, _, client, _ := setup(t)
	resp, httpResp, raw, err := apitest.GetJSON[map[string]interface{}](ctx, client, "/api/v1/admin/dungeons?page=1&page_size=5", "")
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusUnauthorized, httpResp.StatusCode)
	require.NotEqual(t, int(xerrors.CodeSuccess), resp.Code)
}
