package modules

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

func TestAdminDropPoolCrudFlow(t *testing.T) {
	ctx, cfg, client, factory := setup(t)
	token := adminToken(t, ctx, client, cfg)

	code := fmt.Sprintf("auto_pool_%s", factory.RunID[:8])
	createReq := map[string]interface{}{
		"pool_code":        code,
		"pool_name":        code,
		"pool_type":        "monster",
		"min_drops":        0,
		"max_drops":        1,
		"guaranteed_drops": 0,
	}

	createResp, createHTTP, createRaw, err := apitest.PostJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/admin/drop-pools", createReq, token)
	require.NoError(t, err, string(createRaw))
	require.Equal(t, http.StatusOK, createHTTP.StatusCode, string(createRaw))
	require.Equal(t, int(xerrors.CodeSuccess), createResp.Code, string(createRaw))
	poolID := (*createResp.Data)["id"].(string)

	// Update
	updReq := map[string]interface{}{"pool_name": code + "_upd"}
	updResp, updHTTP, updRaw, err := apitest.PutJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/admin/drop-pools/"+poolID, updReq, token)
	require.NoError(t, err, string(updRaw))
	require.Equal(t, http.StatusOK, updHTTP.StatusCode, string(updRaw))
	require.Equal(t, int(xerrors.CodeSuccess), updResp.Code, string(updRaw))

	// Get
	getResp, getHTTP, getRaw, err := apitest.GetJSON[map[string]interface{}](ctx, client, "/api/v1/admin/drop-pools/"+poolID, token)
	require.NoError(t, err, string(getRaw))
	require.Equal(t, http.StatusOK, getHTTP.StatusCode, string(getRaw))
	require.Equal(t, int(xerrors.CodeSuccess), getResp.Code, string(getRaw))

	// Delete
	delResp, delHTTP, delRaw, err := apitest.DeleteJSON[struct{}, map[string]interface{}](ctx, client, "/api/v1/admin/drop-pools/"+poolID, nil, token)
	require.NoError(t, err, string(delRaw))
	require.Equal(t, http.StatusOK, delHTTP.StatusCode, string(delRaw))
	require.Equal(t, int(xerrors.CodeSuccess), delResp.Code, string(delRaw))

	// Get again -> 404
	getResp2, getHTTP2, getRaw2, err := apitest.GetJSON[map[string]interface{}](ctx, client, "/api/v1/admin/drop-pools/"+poolID, token)
	require.NoError(t, err, string(getRaw2))
	require.Equal(t, http.StatusNotFound, getHTTP2.StatusCode, string(getRaw2))
	require.Equal(t, int(xerrors.CodeResourceNotFound), getResp2.Code)
}
