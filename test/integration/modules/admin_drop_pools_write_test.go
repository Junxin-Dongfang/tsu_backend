package modules

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

// 覆盖管理员掉落池创建及物品写入的正向路径，若缺少基础数据则跳过。
func TestAdminDropPoolCreateAndAddItem(t *testing.T) {
	ctx, cfg, client, factory := setup(t)
	token := adminToken(t, ctx, client, cfg)

	desc := "自动化掉落池"
	req := apitest.CreateDropPoolRequest{
		PoolCode:        "auto_pool_" + factory.RunID,
		PoolName:        "auto_pool_" + factory.RunID,
		PoolType:        "monster",
		MinDrops:        0,
		MaxDrops:        1,
		GuaranteedDrops: 0,
		Description:     &desc,
	}
	createResp, httpResp, raw, err := apitest.PostJSON[apitest.CreateDropPoolRequest, apitest.DropPoolResponse](ctx, client, "/api/v1/admin/drop-pools", req, token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusOK, httpResp.StatusCode, string(raw))
	require.Equal(t, int(xerrors.CodeSuccess), createResp.Code, string(raw))
	require.NotNil(t, createResp.Data)

	// 获取一个现有物品用于绑定到掉落池
	itemsResp, itemsHTTP, itemsRaw, err := apitest.GetJSON[apitest.AdminItemList](ctx, client, "/api/v1/admin/items?page=1&page_size=1", token)
	require.NoError(t, err, string(itemsRaw))
	require.Equal(t, http.StatusOK, itemsHTTP.StatusCode, string(itemsRaw))
	require.Equal(t, int(xerrors.CodeSuccess), itemsResp.Code, string(itemsRaw))
	require.NotNil(t, itemsResp.Data)
	require.NotEmpty(t, itemsResp.Data.Items)

	itemID := itemsResp.Data.Items[0].ID
	weight := 10
	addReq := apitest.AddDropPoolItemRequest{
		ItemID:      itemID,
		DropWeight:  &weight,
		MinQuantity: 1,
		MaxQuantity: 1,
	}

	addResp, addHTTP, addRaw, err := apitest.PostJSON[apitest.AddDropPoolItemRequest, apitest.DropPoolItemResponse](ctx, client, "/api/v1/admin/drop-pools/"+createResp.Data.ID+"/items", addReq, token)
	require.NoError(t, err, string(addRaw))
	require.Equal(t, http.StatusOK, addHTTP.StatusCode, string(addRaw))
	require.Equal(t, int(xerrors.CodeSuccess), addResp.Code, string(addRaw))
}
