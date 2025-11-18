package modules

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

func TestTeamLeaderCannotLeave(t *testing.T) {
	ctx, _, client, factory := setup(t)
	leader := registerPlayerWithHero(t, ctx, client, factory, "leave-leader")
	team := createTeamForPlayer(t, ctx, client, factory, leader)

	path := "/api/v1/game/teams/" + team.ID + "/leave?hero_id=" + leader.HeroID
	resp, httpResp, raw, err := apitest.PostJSON[struct{}, map[string]interface{}](ctx, client, path, struct{}{}, leader.Token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusForbidden, httpResp.StatusCode, string(raw))
	require.Equal(t, int(xerrors.CodePermissionDenied), resp.Code)
}
