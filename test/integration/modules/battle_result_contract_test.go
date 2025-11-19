package modules

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

type battleResultRequest struct {
	BattleID     string                    `json:"battle_id"`
	BattleCode   string                    `json:"battle_code"`
	Participants []battleResultParticipant `json:"participants"`
	Result       battleResultInfo          `json:"result"`
	Loot         *apitest.LootPayload      `json:"loot,omitempty"`
	Events       []map[string]interface{}  `json:"events,omitempty"`
}

type battleResultParticipant struct {
	HeroID string `json:"hero_id,omitempty"`
	TeamID string `json:"team_id,omitempty"`
	Role   string `json:"role,omitempty"`
}

type battleResultInfo struct {
	Status      string            `json:"status"`
	LootContext battleLootContext `json:"loot_context"`
}

type battleLootContext struct {
	Type      string `json:"type"`
	TeamID    string `json:"team_id"`
	DungeonID string `json:"dungeon_id"`
}

// TestBattleResultContractEndpoint 确保战斗契约入口能驱动掉落与仓库更新。
func TestBattleResultContractEndpoint(t *testing.T) {
	ctx, cfg, client, factory := setup(t)
	leader := registerPlayerWithHero(t, ctx, client, factory, "battle-result-leader")
	team := createTeamForPlayer(t, ctx, client, factory, leader)
	recruit := registerPlayerWithHero(t, ctx, client, factory, "battle-result-member")
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
	lootItemID := createAdminItem(t, ctx, client, adminToken, factory, "battle-result-item")

	payload := battleResultRequest{
		BattleID:   "battle-contract-" + team.ID,
		BattleCode: "contract-room",
		Participants: []battleResultParticipant{
			{HeroID: leader.HeroID, TeamID: team.ID, Role: "leader"},
		},
		Result: battleResultInfo{
			Status: "victory",
			LootContext: battleLootContext{
				Type:      "dungeon",
				TeamID:    team.ID,
				DungeonID: dungeonID,
			},
		},
		Loot: &apitest.LootPayload{
			Gold:  42,
			Items: []apitest.LootItem{{ItemID: lootItemID, ItemType: "equipment", Quantity: 1}},
		},
		Events: []map[string]interface{}{
			{"tick": 1, "actor": leader.HeroID, "action": "skill_fireball", "value": 200},
		},
	}

	battlePath := "/api/v1/internal/battles/result"
	battleResp, battleHTTP, battleRaw, err := apitest.PostJSON[battleResultRequest, map[string]interface{}](ctx, client, battlePath, payload, "")
	require.NoError(t, err, string(battleRaw))
	require.Equal(t, http.StatusOK, battleHTTP.StatusCode, string(battleRaw))
	require.Equal(t, int(xerrors.CodeSuccess), battleResp.Code, string(battleRaw))

	warehousePath := fmt.Sprintf("/api/v1/game/teams/%s/warehouse?hero_id=%s", team.ID, leader.HeroID)
	warehouseResp, warehouseHTTP, warehouseRaw, err := apitest.GetJSON[apitest.WarehouseResponse](ctx, client, warehousePath, leader.Token)
	require.NoError(t, err, string(warehouseRaw))
	require.Equal(t, http.StatusOK, warehouseHTTP.StatusCode, string(warehouseRaw))
	require.Equal(t, int(xerrors.CodeSuccess), warehouseResp.Code, string(warehouseRaw))

	itemsPath := fmt.Sprintf("/api/v1/game/teams/%s/warehouse/items?hero_id=%s", team.ID, leader.HeroID)
	itemsResp, itemsHTTP, itemsRaw, err := apitest.GetJSON[warehouseItemsPayload](ctx, client, itemsPath, leader.Token)
	require.NoError(t, err, string(itemsRaw))
	require.Equal(t, http.StatusOK, itemsHTTP.StatusCode, string(itemsRaw))
	require.Equal(t, int(xerrors.CodeSuccess), itemsResp.Code, string(itemsRaw))

	found := false
	for _, item := range itemsResp.Data.Items {
		if item.ItemID == lootItemID {
			found = true
			break
		}
	}
	require.True(t, found, "battle result loot not stored in warehouse")

	requireBattleReportStored(t, payload.BattleID, team.ID, dungeonID, lootItemID, payload.Loot.Gold)
}
