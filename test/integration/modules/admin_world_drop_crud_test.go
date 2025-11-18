package modules

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

func TestAdminWorldDropCrud(t *testing.T) {
	ctx, cfg, client, factory := setup(t)
	token := adminToken(t, ctx, client, cfg)

	itemID := createAdminItem(t, ctx, client, token, factory, "world-drop-base")
	createReq := map[string]interface{}{
		"item_id":           itemID,
		"base_drop_rate":    0.35,
		"total_drop_limit":  25,
		"daily_drop_limit":  5,
		"min_drop_interval": 60,
		"max_drop_interval": 600,
		"trigger_conditions": map[string]interface{}{
			"min_player_level": 1,
			"max_player_level": 60,
			"zone":             "automation",
		},
		"drop_rate_modifiers": map[string]interface{}{
			"time_of_day": map[string]interface{}{
				"day":   1.1,
				"night": 0.9,
			},
		},
	}

	createResp, createHTTP, createRaw, err := apitest.PostJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/admin/world-drops", createReq, token)
	require.NoError(t, err, string(createRaw))
	require.Equal(t, http.StatusOK, createHTTP.StatusCode, string(createRaw))
	require.Equal(t, int(xerrors.CodeSuccess), createResp.Code, string(createRaw))
	require.NotNil(t, createResp.Data, string(createRaw))
	wdData := *createResp.Data
	wdID, ok := wdData["id"].(string)
	require.True(t, ok, string(createRaw))
	require.NotEmpty(t, wdID)

	t.Cleanup(func() {
		if wdID == "" {
			return
		}
		cleanupAdminResource(t, client, token, "/api/v1/admin/world-drops/"+wdID)
	})

	detailResp, detailHTTP, detailRaw, err := apitest.GetJSON[map[string]interface{}](ctx, client, "/api/v1/admin/world-drops/"+wdID, token)
	require.NoError(t, err, string(detailRaw))
	require.Equal(t, http.StatusOK, detailHTTP.StatusCode, string(detailRaw))
	require.Equal(t, int(xerrors.CodeSuccess), detailResp.Code, string(detailRaw))
	require.Equal(t, wdID, (*detailResp.Data)["id"], string(detailRaw))

	isActive := false
	updReq := map[string]interface{}{
		"base_drop_rate":   0.45,
		"daily_drop_limit": 10,
		"is_active":        isActive,
	}
	updResp, updHTTP, updRaw, err := apitest.PutJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/admin/world-drops/"+wdID, updReq, token)
	require.NoError(t, err, string(updRaw))
	require.Equal(t, http.StatusOK, updHTTP.StatusCode, string(updRaw))
	require.Equal(t, int(xerrors.CodeSuccess), updResp.Code, string(updRaw))

	delResp, delHTTP, delRaw, err := apitest.DeleteJSON[struct{}, map[string]interface{}](ctx, client, "/api/v1/admin/world-drops/"+wdID, nil, token)
	require.NoError(t, err, string(delRaw))
	require.Equal(t, http.StatusOK, delHTTP.StatusCode, string(delRaw))
	require.Equal(t, int(xerrors.CodeSuccess), delResp.Code, string(delRaw))
}
