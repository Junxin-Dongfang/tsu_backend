package modules

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

func TestAdminAssignRolesUnauthorized(t *testing.T) {
	ctx, _, client, _ := setup(t)
	payload := apitest.AssignRolesRequest{RoleCodes: []string{"role-manager"}}
	resp, httpResp, raw, err := apitest.PostJSON[apitest.AssignRolesRequest, map[string]interface{}](ctx, client, "/api/v1/admin/users/some-user/roles", payload, "")
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusUnauthorized, httpResp.StatusCode)
	require.Contains(t, string(raw), "Unauthorized")
	_ = resp
}

func TestAdminAssignRolesInvalidUser(t *testing.T) {
	ctx, cfg, client, _ := setup(t)
	token := adminToken(t, ctx, client, cfg)

	payload := apitest.AssignRolesRequest{RoleCodes: []string{"role-manager"}}
	resp, httpResp, raw, err := apitest.PostJSON[apitest.AssignRolesRequest, map[string]interface{}](ctx, client, "/api/v1/admin/users/non-existent/roles", payload, token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusNotFound, httpResp.StatusCode)
	require.Equal(t, int(xerrors.CodeResourceNotFound), resp.Code)
}

func TestAdminAssignAndRemoveRolesSuccess(t *testing.T) {
	ctx, cfg, client, _ := setup(t)
	token := adminToken(t, ctx, client, cfg)

	// 使用现有 root 账户做角色赋权验证
	loginReq := apitest.LoginRequest{Identifier: cfg.AdminUsername, Password: cfg.AdminPassword}
	loginResp, httpRespLogin, rawLogin, err := apitest.PostJSON[apitest.LoginRequest, apitest.LoginResponse](ctx, client, "/api/v1/admin/auth/login", loginReq, "")
	require.NoError(t, err, string(rawLogin))
	require.Equal(t, http.StatusOK, httpRespLogin.StatusCode, string(rawLogin))
	require.Equal(t, int(xerrors.CodeSuccess), loginResp.Code, string(rawLogin))
	userID := loginResp.Data.UserID

	roleResp, httpRespRoles, rawRoles, err := apitest.GetJSON[apitest.AdminRoleListResponse](ctx, client, "/api/v1/admin/roles?page=1&page_size=1", token)
	require.NoError(t, err, string(rawRoles))
	require.Equal(t, http.StatusOK, httpRespRoles.StatusCode)
	require.NotNil(t, roleResp.Data)
	require.NotEmpty(t, roleResp.Data.Roles)
	roleCode := roleResp.Data.Roles[0].Code

	assignPayload := apitest.AssignRolesRequest{RoleCodes: []string{roleCode}}
	assignResp, httpRespAssign, rawAssign, err := apitest.PostJSON[apitest.AssignRolesRequest, map[string]interface{}](ctx, client, "/api/v1/admin/users/"+userID+"/roles", assignPayload, token)
	require.NoError(t, err, string(rawAssign))
	require.Equal(t, http.StatusOK, httpRespAssign.StatusCode, string(rawAssign))
	require.Equal(t, int(xerrors.CodeSuccess), assignResp.Code, string(rawAssign))

	removeResp, httpRespRemove, rawRemove, err := apitest.DeleteJSON[apitest.AssignRolesRequest, map[string]interface{}](ctx, client, "/api/v1/admin/users/"+userID+"/roles", &assignPayload, token)
	require.NoError(t, err, string(rawRemove))
	require.Equal(t, http.StatusOK, httpRespRemove.StatusCode, string(rawRemove))
	require.Equal(t, int(xerrors.CodeSuccess), removeResp.Code, string(rawRemove))
}
