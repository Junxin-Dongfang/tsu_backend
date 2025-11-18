package modules

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

func TestAdminWorldDropItemsCRUD(t *testing.T) {
	ctx, cfg, client, factory := setup(t)
	token := adminToken(t, ctx, client, cfg)

	baseItemID := createAdminItem(t, ctx, client, token, factory, "world-drop-item-base")
	createReq := map[string]interface{}{
		"item_id":          baseItemID,
		"base_drop_rate":   0.3,
		"total_drop_limit": 10,
	}
	createResp, createHTTP, createRaw, err := apitest.PostJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/admin/world-drops", createReq, token)
	require.NoError(t, err, string(createRaw))
	require.Equal(t, http.StatusOK, createHTTP.StatusCode, string(createRaw))
	require.Equal(t, int(xerrors.CodeSuccess), createResp.Code, string(createRaw))
	worldDropID := (*createResp.Data)["id"].(string)

	t.Cleanup(func() {
		cleanupAdminResource(t, client, token, "/api/v1/admin/world-drops/"+worldDropID)
	})

	detailResp, detailHTTP, detailRaw, err := apitest.GetJSON[map[string]interface{}](ctx, client, fmt.Sprintf("/api/v1/admin/world-drops/%s", worldDropID), token)
	require.NoError(t, err, string(detailRaw))
	require.Equal(t, http.StatusOK, detailHTTP.StatusCode, string(detailRaw))
	require.Equal(t, int(xerrors.CodeSuccess), detailResp.Code, string(detailRaw))

	extraItemID := createAdminItem(t, ctx, client, token, factory, "world-drop-item-extra")
	itemDetailResp, itemDetailHTTP, itemDetailRaw, err := apitest.GetJSON[map[string]interface{}](ctx, client, fmt.Sprintf("/api/v1/admin/items/%s", extraItemID), token)
	require.NoError(t, err, string(itemDetailRaw))
	require.Equal(t, http.StatusOK, itemDetailHTTP.StatusCode, string(itemDetailRaw))
	require.Equal(t, int(xerrors.CodeSuccess), itemDetailResp.Code, string(itemDetailRaw))

	addReq := map[string]interface{}{
		"item_id":      extraItemID,
		"drop_rate":    0.2,
		"min_quantity": 1,
		"max_quantity": 2,
	}
	addResp, addHTTP, addRaw, err := apitest.PostJSON[map[string]interface{}, map[string]interface{}](ctx, client, fmt.Sprintf("/api/v1/admin/world-drops/%s/items", worldDropID), addReq, token)
	require.NoError(t, err, string(addRaw))
	require.Equal(t, http.StatusOK, addHTTP.StatusCode, string(addRaw))
	require.Equal(t, int(xerrors.CodeSuccess), addResp.Code, string(addRaw))
	extraEntryID := (*addResp.Data)["id"].(string)

	listResp, listHTTP, listRaw, err := apitest.GetJSON[map[string]interface{}](ctx, client, fmt.Sprintf("/api/v1/admin/world-drops/%s/items?page=1&page_size=10", worldDropID), token)
	require.NoError(t, err, string(listRaw))
	require.Equal(t, http.StatusOK, listHTTP.StatusCode, string(listRaw))
	require.Equal(t, int(xerrors.CodeSuccess), listResp.Code, string(listRaw))

	updateReq := map[string]interface{}{
		"drop_rate":       0.1,
		"guaranteed_drop": true,
	}
	updResp, updHTTP, updRaw, err := apitest.PutJSON[map[string]interface{}, map[string]interface{}](ctx, client, fmt.Sprintf("/api/v1/admin/world-drops/%s/items/%s", worldDropID, extraEntryID), updateReq, token)
	require.NoError(t, err, string(updRaw))
	require.Equal(t, http.StatusOK, updHTTP.StatusCode, string(updRaw))
	require.Equal(t, int(xerrors.CodeSuccess), updResp.Code, string(updRaw))

	delResp, delHTTP, delRaw, err := apitest.DeleteJSON[struct{}, map[string]interface{}](ctx, client, fmt.Sprintf("/api/v1/admin/world-drops/%s/items/%s", worldDropID, extraEntryID), nil, token)
	require.NoError(t, err, string(delRaw))
	require.Equal(t, http.StatusOK, delHTTP.StatusCode, string(delRaw))
	require.Equal(t, int(xerrors.CodeSuccess), delResp.Code, string(delRaw))
}
