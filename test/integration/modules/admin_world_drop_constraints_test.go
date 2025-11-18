package modules

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

func TestAdminWorldDropRejectsDuplicateItem(t *testing.T) {
	ctx, cfg, client, factory := setup(t)
	token := adminToken(t, ctx, client, cfg)
	itemID := createAdminItem(t, ctx, client, token, factory, "world-drop-dup")

	createReq := map[string]interface{}{
		"item_id":           itemID,
		"base_drop_rate":    0.2,
		"min_drop_interval": 30,
		"max_drop_interval": 300,
		"trigger_conditions": map[string]interface{}{
			"min_player_level": 1,
		},
	}

	firstResp, firstHTTP, firstRaw, err := apitest.PostJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/admin/world-drops", createReq, token)
	require.NoError(t, err, string(firstRaw))
	require.Equal(t, http.StatusOK, firstHTTP.StatusCode, string(firstRaw))
	require.Equal(t, int(xerrors.CodeSuccess), firstResp.Code, string(firstRaw))
	require.NotNil(t, firstResp.Data, string(firstRaw))
	primaryID := (*firstResp.Data)["id"].(string)
	require.NotEmpty(t, primaryID)

	t.Cleanup(func() {
		cleanupAdminResource(t, client, token, "/api/v1/admin/world-drops/"+primaryID)
	})

	secondResp, secondHTTP, secondRaw, err := apitest.PostJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/admin/world-drops", createReq, token)
	require.NoError(t, err, string(secondRaw))
	require.Equal(t, http.StatusConflict, secondHTTP.StatusCode, string(secondRaw))
	require.Equal(t, int(xerrors.CodeDuplicateResource), secondResp.Code, string(secondRaw))
}

func TestAdminWorldDropRejectsInvalidItemID(t *testing.T) {
	ctx, cfg, client, _ := setup(t)
	token := adminToken(t, ctx, client, cfg)
	invalidItem := uuid.NewString()

	createReq := map[string]interface{}{
		"item_id":        invalidItem,
		"base_drop_rate": 0.15,
	}

	resp, httpResp, raw, err := apitest.PostJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/admin/world-drops", createReq, token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusNotFound, httpResp.StatusCode, string(raw))
	require.Equal(t, int(xerrors.CodeResourceNotFound), resp.Code, string(raw))
}

func TestAdminWorldDropItemsRejectDuplicateWithinConfig(t *testing.T) {
	ctx, cfg, client, factory := setup(t)
	token := adminToken(t, ctx, client, cfg)
	baseItemID := createAdminItem(t, ctx, client, token, factory, "world-drop-dup-base")
	createResp, createHTTP, createRaw, err := apitest.PostJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/admin/world-drops", map[string]interface{}{
		"item_id":        baseItemID,
		"base_drop_rate": 0.25,
	}, token)
	require.NoError(t, err, string(createRaw))
	require.Equal(t, http.StatusOK, createHTTP.StatusCode, string(createRaw))
	require.Equal(t, int(xerrors.CodeSuccess), createResp.Code, string(createRaw))
	configID := (*createResp.Data)["id"].(string)

	t.Cleanup(func() {
		cleanupAdminResource(t, client, token, "/api/v1/admin/world-drops/"+configID)
	})

	detailResp, detailHTTP, detailRaw, err := apitest.GetJSON[map[string]interface{}](ctx, client, "/api/v1/admin/world-drops/"+configID, token)
	require.NoError(t, err, string(detailRaw))
	require.Equal(t, http.StatusOK, detailHTTP.StatusCode, string(detailRaw))
	require.Equal(t, int(xerrors.CodeSuccess), detailResp.Code, string(detailRaw))

	duplicateItemID := createAdminItem(t, ctx, client, token, factory, "world-drop-dup-extra")
	addReq := map[string]interface{}{
		"item_id":      duplicateItemID,
		"drop_rate":    0.1,
		"min_quantity": 1,
		"max_quantity": 1,
	}
	addResp, addHTTP, addRaw, err := apitest.PostJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/admin/world-drops/"+configID+"/items", addReq, token)
	require.NoError(t, err, string(addRaw))
	require.Equal(t, http.StatusOK, addHTTP.StatusCode, string(addRaw))
	require.Equal(t, int(xerrors.CodeSuccess), addResp.Code, string(addRaw))

	dupResp, dupHTTP, dupRaw, err := apitest.PostJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/admin/world-drops/"+configID+"/items", addReq, token)
	require.NoError(t, err, string(dupRaw))
	require.Equal(t, http.StatusConflict, dupHTTP.StatusCode, string(dupRaw))
	require.Equal(t, int(xerrors.CodeDuplicateResource), dupResp.Code, string(dupRaw))
}
