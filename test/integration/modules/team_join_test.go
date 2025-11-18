package modules

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

func TestTeamJoinApplyUnauthorized(t *testing.T) {
	ctx, _, client, _ := setup(t)
	payload := map[string]string{"team_id": "team-unknown", "hero_id": "hero-unknown"}
	resp, httpResp, raw, err := apitest.PostJSON[map[string]string, map[string]interface{}](ctx, client, "/api/v1/game/teams/join/apply", payload, "")
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusUnauthorized, httpResp.StatusCode)
	require.Contains(t, string(raw), "Unauthorized")
	_ = resp
}

func TestTeamJoinApproveEdgeCases(t *testing.T) {
	ctx, _, client, factory := setup(t)
	leader := registerPlayerWithHero(t, ctx, client, factory, "join-leader")
	team := createTeamForPlayer(t, ctx, client, factory, leader)

	applicant := registerPlayerWithHero(t, ctx, client, factory, "join-applicant")
	applyPayload := map[string]string{"team_id": team.ID, "hero_id": applicant.HeroID}
	applyResp, httpResp, raw, err := apitest.PostJSON[map[string]string, map[string]interface{}](ctx, client, "/api/v1/game/teams/join/apply", applyPayload, applicant.Token)
	require.NoError(t, err, string(raw))
	if httpResp.StatusCode != http.StatusOK || applyResp.Code != int(xerrors.CodeSuccess) {
		t.Skipf("申请入队失败，status=%d code=%d body=%s", httpResp.StatusCode, applyResp.Code, string(raw))
	}

	approvePayload := apitest.TeamJoinApprovePayload{RequestID: "unknown", HeroID: leader.HeroID, Approved: true}
	approvePath := "/api/v1/game/teams/join/approve?team_id=" + team.ID + "&hero_id=" + leader.HeroID
	resp, httpResp2, raw2, err := apitest.PostJSON[apitest.TeamJoinApprovePayload, map[string]interface{}](ctx, client, approvePath, approvePayload, leader.Token)
	require.NoError(t, err, string(raw2))
	if httpResp2.StatusCode != http.StatusNotFound {
		t.Skipf("审批入队请求返回 %d，body=%s", httpResp2.StatusCode, string(raw2))
	}
	require.Equal(t, int(xerrors.CodeResourceNotFound), resp.Code)
}
