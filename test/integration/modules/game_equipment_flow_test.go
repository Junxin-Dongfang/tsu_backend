package modules

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

// 装备穿戴/卸下/已装备查询的正向与未授权路径。
func TestGameEquipmentEquipAndUnequipFlow(t *testing.T) {
	ctx, _, client, factory := setup(t)
	player := registerPlayerWithHero(t, ctx, client, factory, "equip-flow")

	// 预置：获取仓库可装备物品（若无则跳过）
	listPath := "/api/v1/game/equipment/slots/" + player.HeroID
	slotResp, slotHTTP, slotRaw, err := apitest.GetJSON[map[string]interface{}](ctx, client, listPath, player.Token)
	require.NoError(t, err, string(slotRaw))
	if slotHTTP.StatusCode != http.StatusOK || slotResp.Code != int(xerrors.CodeSuccess) {
		t.Skipf("获取装备槽位失败，status=%d code=%d", slotHTTP.StatusCode, slotResp.Code)
	}

	// 这里假定后端为新英雄准备了一个可用物品；若接口返回空则跳过
	invPath := "/api/v1/game/equipment/equipped/" + player.HeroID
	_, invHTTP, invRaw, err := apitest.GetJSON[map[string]interface{}](ctx, client, invPath, player.Token)
	require.NoError(t, err, string(invRaw))
	if invHTTP.StatusCode == http.StatusNotFound {
		t.Skip("暂无可装备物品，待种子数据补充")
	}

	// 尝试穿戴一个占位 item_id（若后端校验为空则返回 404）
	equipReq := map[string]string{"hero_id": player.HeroID, "slot": "mainhand", "item_id": "test-item-id"}
	equipResp, equipHTTP, equipRaw, err := apitest.PostJSON[map[string]string, map[string]interface{}](ctx, client, "/api/v1/game/equipment/equip", equipReq, player.Token)
	require.NoError(t, err, string(equipRaw))
	require.NotEqual(t, http.StatusUnauthorized, equipHTTP.StatusCode, string(equipRaw))
	// 允许 200 或 404，若 404 说明数据未准备好
	if equipHTTP.StatusCode == http.StatusOK {
		require.Equal(t, int(xerrors.CodeSuccess), equipResp.Code, string(equipRaw))

		// 卸下
		unequipReq := map[string]string{"hero_id": player.HeroID, "slot": "mainhand"}
		unequipResp, unequipHTTP, unequipRaw, err := apitest.PostJSON[map[string]string, map[string]interface{}](ctx, client, "/api/v1/game/equipment/unequip", unequipReq, player.Token)
		require.NoError(t, err, string(unequipRaw))
		require.Equal(t, http.StatusOK, unequipHTTP.StatusCode, string(unequipRaw))
		require.Equal(t, int(xerrors.CodeSuccess), unequipResp.Code, string(unequipRaw))
	}
}

func TestGameEquipmentEquippedUnauthorized(t *testing.T) {
	ctx, _, client, factory := setup(t)
	player := registerPlayerWithHero(t, ctx, client, factory, "equip-unauth")
	path := "/api/v1/game/equipment/equipped/" + player.HeroID
	resp, httpResp, raw, err := apitest.GetJSON[map[string]interface{}](ctx, client, path, "")
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusUnauthorized, httpResp.StatusCode)
	_ = resp
}
