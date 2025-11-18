package modules

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

func TestTeamWarehouseItemsUnauthorized(t *testing.T) {
	ctx, _, client, factory := setup(t)
	leader := registerPlayerWithHero(t, ctx, client, factory, "warehouse-items-leader")
	team := createTeamForPlayer(t, ctx, client, factory, leader)

	path := "/api/v1/game/teams/" + team.ID + "/warehouse/items?hero_id=" + leader.HeroID
	resp, httpResp, raw, err := apitest.GetJSON[map[string]interface{}](ctx, client, path, "")
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusUnauthorized, httpResp.StatusCode)
	require.Contains(t, string(raw), "Unauthorized")
	_ = resp
}

func TestTeamWarehouseItemsSuccess(t *testing.T) {
	ctx, _, client, factory := setup(t)
	leader := registerPlayerWithHero(t, ctx, client, factory, "warehouse-items-leader2")
	team := createTeamForPlayer(t, ctx, client, factory, leader)

	path := "/api/v1/game/teams/" + team.ID + "/warehouse/items?hero_id=" + leader.HeroID
	resp, httpResp, raw, err := apitest.GetJSON[map[string]interface{}](ctx, client, path, leader.Token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.Equal(t, int(xerrors.CodeSuccess), resp.Code, string(raw))
}
