package modules

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

func TestAdminDropPoolItemsCrud(t *testing.T) {
	ctx, cfg, client, factory := setup(t)
	token := adminToken(t, ctx, client, cfg)

	// 准备一个掉落池
	createReq := map[string]interface{}{
		"pool_code":        "items_pool_" + factory.RunID[:6],
		"pool_name":        "items_pool_" + factory.RunID[:6],
		"pool_type":        "monster",
		"min_drops":        0,
		"max_drops":        1,
		"guaranteed_drops": 0,
	}
	poolResp, poolHTTP, poolRaw, err := apitest.PostJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/admin/drop-pools", createReq, token)
	require.NoError(t, err, string(poolRaw))
	require.Equal(t, http.StatusOK, poolHTTP.StatusCode, string(poolRaw))
	require.Equal(t, int(xerrors.CodeSuccess), poolResp.Code, string(poolRaw))
	poolID := (*poolResp.Data)["id"].(string)

	// 取一个物品 ID
	itemsResp, itemsHTTP, itemsRaw, err := apitest.GetJSON[map[string]interface{}](ctx, client, "/api/v1/admin/items?page=1&page_size=1", token)
	require.NoError(t, err, string(itemsRaw))
	require.Equal(t, http.StatusOK, itemsHTTP.StatusCode, string(itemsRaw))
	itemsList := (*itemsResp.Data)["items"].([]interface{})
	require.NotEmpty(t, itemsList, "没有可用物品")
	itemID := itemsList[0].(map[string]interface{})["id"].(string)

	// 添加掉落物品
	addReq := map[string]interface{}{
		"item_id":      itemID,
		"min_quantity": 1,
		"max_quantity": 1,
	}
	addResp, addHTTP, addRaw, err := apitest.PostJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/admin/drop-pools/"+poolID+"/items", addReq, token)
	require.NoError(t, err, string(addRaw))
	require.Equal(t, http.StatusOK, addHTTP.StatusCode, string(addRaw))
	require.Equal(t, int(xerrors.CodeSuccess), addResp.Code, string(addRaw))

	// 更新掉落物品（数量）
	updReq := map[string]interface{}{
		"min_quantity": 1,
		"max_quantity": 2,
	}
	updResp, updHTTP, updRaw, err := apitest.PutJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/admin/drop-pools/"+poolID+"/items/"+itemID, updReq, token)
	require.NoError(t, err, string(updRaw))
	require.Equal(t, http.StatusOK, updHTTP.StatusCode, string(updRaw))
	require.Equal(t, int(xerrors.CodeSuccess), updResp.Code, string(updRaw))

	// 删除掉落物品
	delResp, delHTTP, delRaw, err := apitest.DeleteJSON[struct{}, map[string]interface{}](ctx, client, "/api/v1/admin/drop-pools/"+poolID+"/items/"+itemID, nil, token)
	require.NoError(t, err, string(delRaw))
	require.Equal(t, http.StatusOK, delHTTP.StatusCode, string(delRaw))
	require.Equal(t, int(xerrors.CodeSuccess), delResp.Code, string(delRaw))
}
