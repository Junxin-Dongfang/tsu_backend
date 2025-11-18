package modules

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

type warehouseItemsPayload struct {
	Items  []warehouseItem `json:"items"`
	Total  int64           `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

type warehouseItem struct {
	ItemID          string  `json:"item_id"`
	ItemType        string  `json:"item_type"`
	Quantity        int     `json:"quantity"`
	SourceDungeonID *string `json:"source_dungeon_id"`
}

// TestTeamDungeonCompletionAwardsLoot 覆盖“战斗/地城→战利品入库”整链，模拟战斗结果回调。
func TestTeamDungeonCompletionAwardsLoot(t *testing.T) {
	ctx, cfg, client, factory := setup(t)
	leader := registerPlayerWithHero(t, ctx, client, factory, "dungeon-loot-leader")
	team := createTeamForPlayer(t, ctx, client, factory, leader)

	// 确保团队满足最低成员要求
	recruit := registerPlayerWithHero(t, ctx, client, factory, "dungeon-loot-member")
	inviteAndAcceptMember(t, ctx, client, team, leader, recruit)

	dungeonID := fetchActiveDungeonID(t)

	selectPath := fmt.Sprintf("/api/v1/game/teams/%s/dungeons/select?hero_id=%s", team.ID, leader.HeroID)
	selectReq := apitest.SelectDungeonRequest{HeroID: leader.HeroID, DungeonID: dungeonID}
	selResp, selHTTP, selRaw, err := apitest.PostJSON[apitest.SelectDungeonRequest, map[string]interface{}](ctx, client, selectPath, selectReq, leader.Token)
	require.NoError(t, err, string(selRaw))
	require.Equal(t, http.StatusOK, selHTTP.StatusCode, string(selRaw))
	require.Equal(t, int(xerrors.CodeSuccess), selResp.Code, string(selRaw))

	enterPath := fmt.Sprintf("/api/v1/game/teams/%s/dungeons/enter?hero_id=%s", team.ID, leader.HeroID)
	enterReq := apitest.EnterDungeonRequest{HeroID: leader.HeroID, DungeonID: dungeonID}
	enterResp, enterHTTP, enterRaw, err := apitest.PostJSON[apitest.EnterDungeonRequest, map[string]interface{}](ctx, client, enterPath, enterReq, leader.Token)
	require.NoError(t, err, string(enterRaw))
	require.Equal(t, http.StatusOK, enterHTTP.StatusCode, string(enterRaw))
	require.Equal(t, int(xerrors.CodeSuccess), enterResp.Code, string(enterRaw))

	adminToken := adminToken(t, ctx, client, cfg)
	lootItemID := createAdminItem(t, ctx, client, adminToken, factory, "dungeon-loot-item")

	lootGold := int64(77)
	completeReq := apitest.CompleteDungeonRequest{
		HeroID:    leader.HeroID,
		DungeonID: dungeonID,
		Loot: &apitest.LootPayload{
			Gold: lootGold,
			Items: []apitest.LootItem{
				{ItemID: lootItemID, ItemType: "equipment", Quantity: 1},
			},
		},
	}
	completePath := fmt.Sprintf("/api/v1/game/teams/%s/dungeons/complete?hero_id=%s", team.ID, leader.HeroID)
	compResp, compHTTP, compRaw, err := apitest.PostJSON[apitest.CompleteDungeonRequest, map[string]interface{}](ctx, client, completePath, completeReq, leader.Token)
	require.NoError(t, err, string(compRaw))
	require.Equal(t, http.StatusOK, compHTTP.StatusCode, string(compRaw))
	require.Equal(t, int(xerrors.CodeSuccess), compResp.Code, string(compRaw))

	warehousePath := fmt.Sprintf("/api/v1/game/teams/%s/warehouse?hero_id=%s", team.ID, leader.HeroID)
	warehouseResp, warehouseHTTP, warehouseRaw, err := apitest.GetJSON[apitest.WarehouseResponse](ctx, client, warehousePath, leader.Token)
	require.NoError(t, err, string(warehouseRaw))
	require.Equal(t, http.StatusOK, warehouseHTTP.StatusCode, string(warehouseRaw))
	require.Equal(t, int(xerrors.CodeSuccess), warehouseResp.Code, string(warehouseRaw))
	require.NotNil(t, warehouseResp.Data)
	require.GreaterOrEqual(t, warehouseResp.Data.Gold, lootGold)

	itemsPath := fmt.Sprintf("/api/v1/game/teams/%s/warehouse/items?hero_id=%s", team.ID, leader.HeroID)
	itemsResp, itemsHTTP, itemsRaw, err := apitest.GetJSON[warehouseItemsPayload](ctx, client, itemsPath, leader.Token)
	require.NoError(t, err, string(itemsRaw))
	require.Equal(t, http.StatusOK, itemsHTTP.StatusCode, string(itemsRaw))
	require.Equal(t, int(xerrors.CodeSuccess), itemsResp.Code, string(itemsRaw))
	require.NotNil(t, itemsResp.Data)

	found := false
	for _, item := range itemsResp.Data.Items {
		if item.ItemID == lootItemID {
			require.Equal(t, 1, item.Quantity)
			found = true
			break
		}
	}
	require.True(t, found, "团队仓库缺少通过地城发放的物品")
}
