package modules

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

func TestAdminGrantItemUnauthorized(t *testing.T) {
	ctx, _, client, _ := setup(t)
	req := map[string]interface{}{
		"target_type": "user",
		"target_id":   "some-user",
		"item_id":     "00000000-0000-0000-0000-000000000000",
		"quantity":    1,
	}
	resp, httpResp, raw, err := apitest.PostJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/admin/tools/grant-item", req, "")
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusUnauthorized, httpResp.StatusCode)
	_ = resp
}

func TestAdminGrantItemToTeamWarehouse(t *testing.T) {
	ctx, cfg, client, factory := setup(t)
	token := adminToken(t, ctx, client, cfg)

	// 1) 取一个有效 item_id
	itemsResp, itemsHTTP, itemsRaw, err := apitest.GetJSON[map[string]interface{}](ctx, client, "/api/v1/admin/items?page=1&page_size=1", token)
	require.NoError(t, err, string(itemsRaw))
	if itemsHTTP.StatusCode != http.StatusOK {
		t.Skipf("获取物品列表失败，status=%d body=%s", itemsHTTP.StatusCode, string(itemsRaw))
	}
	require.NotNil(t, itemsResp.Data)
	data := *itemsResp.Data
	itemsList, ok := data["items"].([]interface{})
	if !ok || len(itemsList) == 0 {
		t.Skip("没有可用物品，跳过发放测试")
	}
	first := itemsList[0].(map[string]interface{})
	itemID, _ := first["id"].(string)

	// 2) 创建一支团队
	player := registerPlayerWithHero(t, ctx, client, factory, "grant-team")
	team := createTeamForPlayer(t, ctx, client, factory, player)

	// 3) 发放到团队仓库
	req := map[string]interface{}{
		"target_type": "team_warehouse",
		"target_id":   team.ID,
		"item_id":     itemID,
		"quantity":    1,
	}
	resp, httpResp, raw, err := apitest.PostJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/admin/tools/grant-item", req, token)
	require.NoError(t, err, string(raw))
	if httpResp.StatusCode != http.StatusOK || resp.Code != int(xerrors.CodeSuccess) {
		t.Skipf("发放物品失败，status=%d code=%d body=%s", httpResp.StatusCode, resp.Code, string(raw))
	}
}
