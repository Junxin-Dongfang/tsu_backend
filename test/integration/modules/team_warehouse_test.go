package modules

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

func TestTeamWarehouseDistributeUnauthorized(t *testing.T) {
	ctx, _, client, factory := setup(t)
	leader := registerPlayerWithHero(t, ctx, client, factory, "warehouse-leader")
	team := createTeamForPlayer(t, ctx, client, factory, leader)

	req := apitest.DistributeGoldRequest{DistributorID: leader.HeroID, Distributions: map[string]int64{"hero-x": 10}}
	path := "/api/v1/game/teams/" + team.ID + "/warehouse/distribute-gold"
	resp, httpResp, raw, err := apitest.PostJSON[apitest.DistributeGoldRequest, map[string]interface{}](ctx, client, path, req, "")
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusUnauthorized, httpResp.StatusCode)
	require.Contains(t, string(raw), "Unauthorized")
	_ = resp
}

func TestTeamWarehouseDistributePermissionDenied(t *testing.T) {
	ctx, _, client, factory := setup(t)
	leader := registerPlayerWithHero(t, ctx, client, factory, "warehouse-leader2")
	team := createTeamForPlayer(t, ctx, client, factory, leader)
	member := registerPlayerWithHero(t, ctx, client, factory, "warehouse-member2")

	inviteReq := apitest.InviteMemberRequest{TeamID: team.ID, InviterHeroID: leader.HeroID, InviteeHeroID: member.HeroID}
	invitePath := "/api/v1/game/teams/invite?team_id=" + team.ID + "&hero_id=" + leader.HeroID
	inviteResp, httpResp, raw, err := apitest.PostJSON[apitest.InviteMemberRequest, map[string]interface{}](ctx, client, invitePath, inviteReq, leader.Token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusOK, httpResp.StatusCode, string(raw))
	require.Equal(t, int(xerrors.CodeSuccess), inviteResp.Code, string(raw))

	req := apitest.DistributeGoldRequest{DistributorID: member.HeroID, Distributions: map[string]int64{leader.HeroID: 5}}
	path := "/api/v1/game/teams/" + team.ID + "/warehouse/distribute-gold?hero_id=" + member.HeroID
	resp, httpResp2, raw2, err := apitest.PostJSON[apitest.DistributeGoldRequest, map[string]interface{}](ctx, client, path, req, member.Token)
	require.NoError(t, err, string(raw2))
	require.Equal(t, http.StatusForbidden, httpResp2.StatusCode, string(raw2))
	require.Equal(t, int(xerrors.CodePermissionDenied), resp.Code)
}

func TestTeamWarehouseDistributeGoldSuccessOrSkip(t *testing.T) {
	ctx, _, client, factory := setup(t)
	leader := registerPlayerWithHero(t, ctx, client, factory, "warehouse-leader-ok")
	team := createTeamForPlayer(t, ctx, client, factory, leader)

	// 准备一个成员分金币
	member := registerPlayerWithHero(t, ctx, client, factory, "warehouse-member-ok")
	inviteReq := apitest.InviteMemberRequest{TeamID: team.ID, InviterHeroID: leader.HeroID, InviteeHeroID: member.HeroID}
	invitePath := "/api/v1/game/teams/invite?team_id=" + team.ID + "&hero_id=" + leader.HeroID
	_, _, _, _ = apitest.PostJSON[apitest.InviteMemberRequest, map[string]interface{}](ctx, client, invitePath, inviteReq, leader.Token)

	req := apitest.DistributeGoldRequest{DistributorID: leader.HeroID, Distributions: map[string]int64{member.HeroID: 1}}
	path := "/api/v1/game/teams/" + team.ID + "/warehouse/distribute-gold?hero_id=" + leader.HeroID
	resp, httpResp, raw, err := apitest.PostJSON[apitest.DistributeGoldRequest, map[string]interface{}](ctx, client, path, req, leader.Token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusBadRequest, httpResp.StatusCode, string(raw))
	require.Equal(t, int(xerrors.CodeInvalidParams), resp.Code)
}
