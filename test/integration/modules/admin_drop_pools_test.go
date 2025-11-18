package modules

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

func TestAdminDropPoolsRequiresAuth(t *testing.T) {
	ctx, _, client, _ := setup(t)
	resp, httpResp, raw, err := apitest.GetJSON[map[string]interface{}](ctx, client, "/api/v1/admin/drop-pools", "")
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusUnauthorized, httpResp.StatusCode)
	require.Contains(t, string(raw), "Unauthorized")
	_ = resp
}

func TestAdminDropPoolsListAndItems(t *testing.T) {
	ctx, cfg, client, _ := setup(t)
	token := adminToken(t, ctx, client, cfg)

	resp, httpResp, raw, err := apitest.GetJSON[map[string]interface{}](ctx, client, "/api/v1/admin/drop-pools?page=1&page_size=5", token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.Equal(t, int(xerrors.CodeSuccess), resp.Code, string(raw))

	invalidPath := "/api/v1/admin/drop-pools/non-existent/items"
	resp2, httpResp2, raw2, err := apitest.GetJSON[map[string]interface{}](ctx, client, invalidPath, token)
	require.NoError(t, err, string(raw2))
	require.Equal(t, http.StatusNotFound, httpResp2.StatusCode, string(raw2))
	require.Equal(t, int(xerrors.CodeResourceNotFound), resp2.Code)
}
