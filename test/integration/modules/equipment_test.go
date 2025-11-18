package modules

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

func TestEquipmentEquipUnauthorized(t *testing.T) {
	ctx, _, client, _ := setup(t)
	req := map[string]string{"hero_id": "hero-x", "slot": "weapon", "item_id": "item-x"}
	resp, httpResp, raw, err := apitest.PostJSON[map[string]string, map[string]interface{}](ctx, client, "/api/v1/game/equipment/equip", req, "")
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusUnauthorized, httpResp.StatusCode)
	require.Contains(t, string(raw), "Unauthorized")
	_ = resp
}

func TestEquipmentEquipValidation(t *testing.T) {
	ctx, _, client, factory := setup(t)
	player := registerPlayerWithHero(t, ctx, client, factory, "equip-player")
	req := map[string]string{"hero_id": player.HeroID, "slot": "", "item_id": ""}
	resp, httpResp, raw, err := apitest.PostJSON[map[string]string, map[string]interface{}](ctx, client, "/api/v1/game/equipment/equip", req, player.Token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusNotFound, httpResp.StatusCode, string(raw))
	require.NotEqual(t, int(xerrors.CodeSuccess), resp.Code)
}
