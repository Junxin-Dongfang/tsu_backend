package modules

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

func TestAdminCreateItemMinimal(t *testing.T) {
	ctx, cfg, client, factory := setup(t)
	token := adminToken(t, ctx, client, cfg)

	code := "auto-item-" + factory.RunID
	payload := map[string]interface{}{
		"item_code":    code,
		"item_name":    code,
		"item_level":   1,
		"item_quality": "normal",
		"item_type":    "equipment",
		"equip_slot":   "mainhand",
	}
	resp, httpResp, raw, err := apitest.PostJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/admin/items", payload, token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusOK, httpResp.StatusCode, string(raw))
	require.Equal(t, int(xerrors.CodeSuccess), resp.Code, string(raw))
}
