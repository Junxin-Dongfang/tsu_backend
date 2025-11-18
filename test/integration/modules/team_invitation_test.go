package modules

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

func TestTeamInvitationFlow(t *testing.T) {
	ctx, _, client, factory := setup(t)
	leader := registerPlayerWithHero(t, ctx, client, factory, "invite-leader")
	team := createTeamForPlayer(t, ctx, client, factory, leader)

	invitee := registerPlayerWithHero(t, ctx, client, factory, "invite-target")
	inviteReq := apitest.InviteMemberRequest{TeamID: team.ID, InviterHeroID: leader.HeroID, InviteeHeroID: invitee.HeroID}
	invitePath := "/api/v1/game/teams/invite?team_id=" + team.ID + "&hero_id=" + leader.HeroID
	inviteResp, httpResp, raw, err := apitest.PostJSON[apitest.InviteMemberRequest, map[string]interface{}](ctx, client, invitePath, inviteReq, leader.Token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusOK, httpResp.StatusCode, string(raw))
	require.Equal(t, int(xerrors.CodeSuccess), inviteResp.Code, string(raw))

	acceptReq := apitest.AcceptInvitationRequest{InvitationID: "non-existent", HeroID: invitee.HeroID}
	resp, httpResp2, raw2, err := apitest.PostJSON[apitest.AcceptInvitationRequest, map[string]interface{}](ctx, client, "/api/v1/game/teams/invite/accept", acceptReq, invitee.Token)
	require.NoError(t, err, string(raw2))
	require.Equal(t, http.StatusNotFound, httpResp2.StatusCode)
	require.Equal(t, int(xerrors.CodeResourceNotFound), resp.Code)
}

func TestTeamInvitationApprovePermission(t *testing.T) {
	ctx, _, client, factory := setup(t)
	leader := registerPlayerWithHero(t, ctx, client, factory, "invite-leader2")
	team := createTeamForPlayer(t, ctx, client, factory, leader)
	member := registerPlayerWithHero(t, ctx, client, factory, "invite-member2")

	inviteReq := apitest.InviteMemberRequest{TeamID: team.ID, InviterHeroID: leader.HeroID, InviteeHeroID: member.HeroID}
	invitePath := "/api/v1/game/teams/invite?team_id=" + team.ID + "&hero_id=" + leader.HeroID
	inviteResp, httpResp, raw, err := apitest.PostJSON[apitest.InviteMemberRequest, map[string]interface{}](ctx, client, invitePath, inviteReq, leader.Token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusOK, httpResp.StatusCode, string(raw))
	require.Equal(t, int(xerrors.CodeSuccess), inviteResp.Code, string(raw))

	invitationID := fetchLatestInvitationID(t, team.ID, member.HeroID)
	approveReq := apitest.ApproveInvitationRequest{InvitationID: invitationID, HeroID: member.HeroID, Approved: true}
	path := "/api/v1/game/teams/invite/approve?team_id=" + team.ID + "&hero_id=" + member.HeroID
	resp, httpResp2, raw2, err := apitest.PostJSON[apitest.ApproveInvitationRequest, map[string]interface{}](ctx, client, path, approveReq, member.Token)
	require.NoError(t, err, string(raw2))
	require.Equal(t, http.StatusForbidden, httpResp2.StatusCode, string(raw2))
	require.Equal(t, int(xerrors.CodePermissionDenied), resp.Code)
}

func TestTeamInvitationApproveAndAccept(t *testing.T) {
	ctx, _, client, factory := setup(t)
	leader := registerPlayerWithHero(t, ctx, client, factory, "invite-leader-success")
	team := createTeamForPlayer(t, ctx, client, factory, leader)
	invitee := registerPlayerWithHero(t, ctx, client, factory, "invite-member-success")

	inviteReq := apitest.InviteMemberRequest{TeamID: team.ID, InviterHeroID: leader.HeroID, InviteeHeroID: invitee.HeroID}
	invitePath := "/api/v1/game/teams/invite?team_id=" + team.ID + "&hero_id=" + leader.HeroID
	inviteResp, httpResp, raw, err := apitest.PostJSON[apitest.InviteMemberRequest, map[string]interface{}](ctx, client, invitePath, inviteReq, leader.Token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusOK, httpResp.StatusCode, string(raw))
	require.Equal(t, int(xerrors.CodeSuccess), inviteResp.Code, string(raw))

	invitationID := fetchLatestInvitationID(t, team.ID, invitee.HeroID)
	approveReq := apitest.ApproveInvitationRequest{InvitationID: invitationID, HeroID: leader.HeroID, Approved: true}
	approvePath := "/api/v1/game/teams/invite/approve?team_id=" + team.ID + "&hero_id=" + leader.HeroID
	approveResp, approveHTTP, approveRaw, err := apitest.PostJSON[apitest.ApproveInvitationRequest, map[string]interface{}](ctx, client, approvePath, approveReq, leader.Token)
	require.NoError(t, err, string(approveRaw))
	require.Equal(t, http.StatusOK, approveHTTP.StatusCode, string(approveRaw))
	require.Equal(t, int(xerrors.CodeSuccess), approveResp.Code, string(approveRaw))

	acceptReq := apitest.AcceptInvitationRequest{InvitationID: invitationID, HeroID: invitee.HeroID}
	acceptPath := "/api/v1/game/teams/invite/accept?team_id=" + team.ID + "&hero_id=" + invitee.HeroID
	acceptResp, acceptHTTP, acceptRaw, err := apitest.PostJSON[apitest.AcceptInvitationRequest, map[string]interface{}](ctx, client, acceptPath, acceptReq, invitee.Token)
	require.NoError(t, err, string(acceptRaw))
	require.Equal(t, http.StatusOK, acceptHTTP.StatusCode, string(acceptRaw))
	require.Equal(t, int(xerrors.CodeSuccess), acceptResp.Code, string(acceptRaw))

	// 重复接受应提示邀请状态错误
	dupResp, dupHTTP, dupRaw, err := apitest.PostJSON[apitest.AcceptInvitationRequest, map[string]interface{}](ctx, client, acceptPath, acceptReq, invitee.Token)
	require.NoError(t, err, string(dupRaw))
	require.Equal(t, http.StatusBadRequest, dupHTTP.StatusCode, string(dupRaw))
	require.Equal(t, int(xerrors.CodeInvalidParams), dupResp.Code, string(dupRaw))
}

func TestTeamInvitationRejectFlow(t *testing.T) {
	ctx, _, client, factory := setup(t)
	leader := registerPlayerWithHero(t, ctx, client, factory, "invite-leader-reject")
	team := createTeamForPlayer(t, ctx, client, factory, leader)
	invitee := registerPlayerWithHero(t, ctx, client, factory, "invite-member-reject")

	inviteReq := apitest.InviteMemberRequest{TeamID: team.ID, InviterHeroID: leader.HeroID, InviteeHeroID: invitee.HeroID}
	invitePath := "/api/v1/game/teams/invite?team_id=" + team.ID + "&hero_id=" + leader.HeroID
	inviteResp, httpResp, raw, err := apitest.PostJSON[apitest.InviteMemberRequest, map[string]interface{}](ctx, client, invitePath, inviteReq, leader.Token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusOK, httpResp.StatusCode, string(raw))
	require.Equal(t, int(xerrors.CodeSuccess), inviteResp.Code, string(raw))

	invitationID := fetchLatestInvitationID(t, team.ID, invitee.HeroID)
	approveReq := apitest.ApproveInvitationRequest{InvitationID: invitationID, HeroID: leader.HeroID, Approved: true}
	approvePath := "/api/v1/game/teams/invite/approve?team_id=" + team.ID + "&hero_id=" + leader.HeroID
	_, approveHTTP, approveRaw, err := apitest.PostJSON[apitest.ApproveInvitationRequest, map[string]interface{}](ctx, client, approvePath, approveReq, leader.Token)
	require.NoError(t, err, string(approveRaw))
	require.Equal(t, http.StatusOK, approveHTTP.StatusCode, string(approveRaw))

	rejectPath := "/api/v1/game/teams/invite/reject?invitation_id=" + invitationID + "&hero_id=" + invitee.HeroID + "&team_id=" + team.ID
	rejectResp, rejectHTTP, rejectRaw, err := apitest.PostJSON[struct{}, map[string]interface{}](ctx, client, rejectPath, struct{}{}, invitee.Token)
	require.NoError(t, err, string(rejectRaw))
	require.Equal(t, http.StatusOK, rejectHTTP.StatusCode, string(rejectRaw))
	require.Equal(t, int(xerrors.CodeSuccess), rejectResp.Code, string(rejectRaw))

	// 被拒绝后再次接受应提示邀请不存在
	acceptReq := apitest.AcceptInvitationRequest{InvitationID: invitationID, HeroID: invitee.HeroID}
	rejectAcceptPath := "/api/v1/game/teams/invite/accept?team_id=" + team.ID + "&hero_id=" + invitee.HeroID
	acceptResp, acceptHTTP, acceptRaw, err := apitest.PostJSON[apitest.AcceptInvitationRequest, map[string]interface{}](ctx, client, rejectAcceptPath, acceptReq, invitee.Token)
	require.NoError(t, err, string(acceptRaw))
	require.Equal(t, http.StatusBadRequest, acceptHTTP.StatusCode, string(acceptRaw))
	require.Equal(t, int(xerrors.CodeInvalidParams), acceptResp.Code, string(acceptRaw))
}
