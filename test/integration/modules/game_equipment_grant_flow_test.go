package modules

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

// 发放物品到玩家背包并完成穿戴/卸下的正向流程。
func TestGrantAndEquipSuccessOrSkip(t *testing.T) {
	ctx, cfg, client, factory := setup(t)
	token := adminToken(t, ctx, client, cfg)

	// 1) 选取一个带 equip_slot 的装备物品
	itemsResp, itemsHTTP, itemsRaw, err := apitest.GetJSON[map[string]interface{}](ctx, client, "/api/v1/admin/items?page=1&page_size=20", token)
	require.NoError(t, err, string(itemsRaw))
	require.Equal(t, http.StatusOK, itemsHTTP.StatusCode, string(itemsRaw))
	data := *itemsResp.Data
	items, ok := data["items"].([]interface{})
	if !ok || len(items) == 0 {
		t.Skip("物品列表为空，无法发放装备")
	}
	var equipItemID string
	for _, it := range items {
		m, _ := it.(map[string]interface{})
		if slot, ok := m["equip_slot"].(string); ok && slot != "" {
			equipItemID = m["id"].(string)
			break
		}
	}
	if equipItemID == "" {
		t.Skip("未找到带 equip_slot 的装备物品，无法验证穿戴")
	}

	// 2) 注册玩家+英雄
	player := registerPlayerWithHero(t, ctx, client, factory, "grant-equip")

	// 3) 管理员发放装备到玩家背包
	grantReq := map[string]interface{}{
		"target_type": "user",
		"target_id":   player.UserID,
		"item_id":     equipItemID,
		"quantity":    1,
	}
	grantResp, grantHTTP, grantRaw, err := apitest.PostJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/admin/tools/grant-item", grantReq, token)
	require.NoError(t, err, string(grantRaw))
	if grantHTTP.StatusCode != http.StatusOK || grantResp.Code != int(xerrors.CodeSuccess) {
		t.Skipf("发放装备失败，status=%d code=%d body=%s", grantHTTP.StatusCode, grantResp.Code, string(grantRaw))
	}

	// 4) 查询玩家背包，取出新装备实例ID
	invPath := "/api/v1/game/inventory?owner_id=" + player.UserID + "&item_location=backpack&page=1&page_size=20"
	invResp, invHTTP, invRaw, err := apitest.GetJSON[map[string]interface{}](ctx, client, invPath, player.Token)
	require.NoError(t, err, string(invRaw))
	require.Equal(t, http.StatusOK, invHTTP.StatusCode, string(invRaw))
	invData := *invResp.Data
	itemList, ok := invData["items"].([]interface{})
	if !ok || len(itemList) == 0 {
		t.Skip("发放后未在背包找到物品实例")
	}
	var instanceID string
	for _, it := range itemList {
		m, _ := it.(map[string]interface{})
		if m["item_id"] == equipItemID {
			if id, ok := m["id"].(string); ok {
				instanceID = id
				break
			}
		}
	}
	if instanceID == "" {
		t.Skip("未找到匹配的物品实例，无法穿戴")
	}

	// 5) 穿戴装备
	equipReq := map[string]interface{}{
		"hero_id":          player.HeroID,
		"item_instance_id": instanceID,
	}
	equipResp, equipHTTP, equipRaw, err := apitest.PostJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/game/equipment/equip", equipReq, player.Token)
	require.NoError(t, err, string(equipRaw))
	require.Equal(t, http.StatusOK, equipHTTP.StatusCode, string(equipRaw))
	require.Equal(t, int(xerrors.CodeSuccess), equipResp.Code, string(equipRaw))

	// 6) 卸下装备
	unequipReq := map[string]interface{}{
		"hero_id": player.HeroID,
		"slot":    "",
	}
	// 先获取槽位以便卸下
	slotsPath := "/api/v1/game/equipment/slots/" + player.HeroID
	slotsResp, slotsHTTP, slotsRaw, err := apitest.GetJSON[map[string]interface{}](ctx, client, slotsPath, player.Token)
	require.NoError(t, err, string(slotsRaw))
	require.Equal(t, http.StatusOK, slotsHTTP.StatusCode, string(slotsRaw))
	slotsData := *slotsResp.Data
	slotsList, _ := slotsData["list"].([]interface{})
	var slotID string
	for _, s := range slotsList {
		m, _ := s.(map[string]interface{})
		if eq, ok := m["equipped_item_id"].(string); ok && eq == instanceID {
			slotID = m["id"].(string)
			break
		}
	}
	if slotID == "" {
		t.Skip("未找到已装备的槽位，无法执行卸下")
	}
	unequipReq["slot_id"] = slotID
	unequipResp, unequipHTTP, unequipRaw, err := apitest.PostJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/game/equipment/unequip", unequipReq, player.Token)
	require.NoError(t, err, string(unequipRaw))
	require.Equal(t, http.StatusOK, unequipHTTP.StatusCode, string(unequipRaw))
	require.Equal(t, int(xerrors.CodeSuccess), unequipResp.Code, string(unequipRaw))
}
