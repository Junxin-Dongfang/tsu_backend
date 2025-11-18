package modules

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

func TestTeamDungeonSelectRequiresLeader(t *testing.T) {
	ctx, _, client, factory := setup(t)
	leader := registerPlayerWithHero(t, ctx, client, factory, "dungeon-leader")
	team := createTeamForPlayer(t, ctx, client, factory, leader)

	member := registerPlayerWithHero(t, ctx, client, factory, "dungeon-member")
	inviteReq := apitest.InviteMemberRequest{TeamID: team.ID, InviterHeroID: leader.HeroID, InviteeHeroID: member.HeroID}
	invitePath := "/api/v1/game/teams/invite?team_id=" + team.ID + "&hero_id=" + leader.HeroID
	inviteResp, httpResp, raw, err := apitest.PostJSON[apitest.InviteMemberRequest, map[string]interface{}](ctx, client, invitePath, inviteReq, leader.Token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusOK, httpResp.StatusCode, string(raw))
	require.Equal(t, int(xerrors.CodeSuccess), inviteResp.Code, string(raw))

	acceptReq := map[string]string{"invitation_id": "unknown", "hero_id": member.HeroID}
	_, _, _, _ = apitest.PostJSON[map[string]string, map[string]interface{}](ctx, client, "/api/v1/game/teams/invite/accept", acceptReq, member.Token)

	path := fmt.Sprintf("/api/v1/game/teams/%s/dungeons/select", team.ID)
	payload := map[string]string{"hero_id": member.HeroID, "dungeon_id": "dungeon-001"}
	resp, httpResp2, raw2, err := apitest.PostJSON[map[string]string, map[string]interface{}](ctx, client, path, payload, member.Token)
	require.NoError(t, err, string(raw2))
	require.Equal(t, http.StatusUnauthorized, httpResp2.StatusCode, string(raw2))
	require.NotEqual(t, int(xerrors.CodeSuccess), resp.Code)
}
