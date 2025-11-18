package modules

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

// 负例：未登录创建团队
func TestTeamCreateRequiresAuth(t *testing.T) {
	ctx, _, client, factory := setup(t)
	heroReq := factory.BuildCreateTeamRequest("hero-missing")
	resp, httpResp, raw, err := apitest.PostJSON[apitest.CreateTeamRequest, map[string]interface{}](ctx, client, "/api/v1/game/teams", heroReq, "")
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusUnauthorized, httpResp.StatusCode)
	require.Contains(t, string(raw), "Unauthorized")
	_ = resp
}

// 负例：非队长邀请成员应被拒绝
func TestTeamInvitePermissionDeniedForMember(t *testing.T) {
	ctx, _, client, factory := setup(t)
	leader := registerPlayerWithHero(t, ctx, client, factory, "leader")
	team := createTeamForPlayer(t, ctx, client, factory, leader)

	member := registerPlayerWithHero(t, ctx, client, factory, "member")
	inviteReq := apitest.InviteMemberRequest{
		TeamID:        team.ID,
		InviterHeroID: member.HeroID,
		InviteeHeroID: leader.HeroID,
	}
	path := "/api/v1/game/teams/invite?team_id=" + team.ID + "&hero_id=" + member.HeroID
	resp, httpResp, raw, err := apitest.PostJSON[apitest.InviteMemberRequest, map[string]interface{}](ctx, client, path, inviteReq, member.Token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusForbidden, httpResp.StatusCode, string(raw))
	require.Equal(t, int(xerrors.CodePermissionDenied), resp.Code)
}

// 正例：队长创建团队并查看仓库
func TestTeamLeaderCanViewWarehouse(t *testing.T) {
	ctx, _, client, factory := setup(t)
	leader := registerPlayerWithHero(t, ctx, client, factory, "view-warehouse")
	team := createTeamForPlayer(t, ctx, client, factory, leader)

	path := fmt.Sprintf("/api/v1/game/teams/%s/warehouse?hero_id=%s", team.ID, leader.HeroID)
	resp, httpResp, raw, err := apitest.GetJSON[apitest.WarehouseResponse](ctx, client, path, leader.Token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.Equal(t, int(xerrors.CodeSuccess), resp.Code)
	require.NotNil(t, resp.Data)
	require.Equal(t, team.ID, resp.Data.TeamID)
}
