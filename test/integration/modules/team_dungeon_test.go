package modules

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

func TestTeamDungeonSelectUnauthorized(t *testing.T) {
	ctx, _, client, factory := setup(t)
	leader := registerPlayerWithHero(t, ctx, client, factory, "dungeon-leader")
	team := createTeamForPlayer(t, ctx, client, factory, leader)

	payload := map[string]string{"hero_id": leader.HeroID, "dungeon_id": "dungeon-sample"}
	resp, httpResp, raw, err := apitest.PostJSON[map[string]string, map[string]interface{}](ctx, client, "/api/v1/game/teams/"+team.ID+"/dungeons/select", payload, "")
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusUnauthorized, httpResp.StatusCode)
	require.Contains(t, string(raw), "Unauthorized")
	_ = resp
}

func TestTeamDungeonSelectForbiddenForMember(t *testing.T) {
	ctx, _, client, factory := setup(t)
	leader := registerPlayerWithHero(t, ctx, client, factory, "dungeon-leader2")
	team := createTeamForPlayer(t, ctx, client, factory, leader)

	member := registerPlayerWithHero(t, ctx, client, factory, "dungeon-member2")
	inviteReq := apitest.InviteMemberRequest{TeamID: team.ID, InviterHeroID: leader.HeroID, InviteeHeroID: member.HeroID}
	invitePath := "/api/v1/game/teams/invite?team_id=" + team.ID + "&hero_id=" + leader.HeroID
	inviteResp, httpResp, raw, err := apitest.PostJSON[apitest.InviteMemberRequest, map[string]interface{}](ctx, client, invitePath, inviteReq, leader.Token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusOK, httpResp.StatusCode, string(raw))
	require.Equal(t, int(xerrors.CodeSuccess), inviteResp.Code, string(raw))

	payload := map[string]string{"hero_id": member.HeroID, "dungeon_id": "dungeon-sample"}
	resp, httpResp2, raw2, err := apitest.PostJSON[map[string]string, map[string]interface{}](ctx, client, "/api/v1/game/teams/"+team.ID+"/dungeons/select", payload, member.Token)
	require.NoError(t, err, string(raw2))
	require.NotEqual(t, http.StatusOK, httpResp2.StatusCode)
	require.NotEqual(t, int(xerrors.CodeSuccess), resp.Code)
}
