package modules

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

// 地城 CRUD：使用最小必填字段（无房间序列）。
func TestAdminDungeonCrud(t *testing.T) {
	ctx, cfg, client, factory := setup(t)
	token := adminToken(t, ctx, client, cfg)

	code := fmt.Sprintf("auto_dun_%s", factory.RunID[:6])
	createReq := map[string]interface{}{
		"dungeon_code":      code,
		"dungeon_name":      "自动地城" + factory.RunID[:4],
		"min_level":         1,
		"max_level":         2,
		"requires_attempts": false,
		"room_sequence": []map[string]interface{}{
			{"room_id": "d766ae32-8242-41bd-9134-f02d9c1db68c", "sort": 1},
		},
		"is_active": true,
	}

	createResp, createHTTP, createRaw, err := apitest.PostJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/admin/dungeons", createReq, token)
	require.NoError(t, err, string(createRaw))
	require.Equal(t, http.StatusOK, createHTTP.StatusCode, string(createRaw))
	require.Equal(t, int(xerrors.CodeSuccess), createResp.Code, string(createRaw))
	dungeonID := (*createResp.Data)["id"].(string)

	// 详情
	detailResp, detailHTTP, detailRaw, err := apitest.GetJSON[map[string]interface{}](ctx, client, "/api/v1/admin/dungeons/"+dungeonID, token)
	require.NoError(t, err, string(detailRaw))
	require.Equal(t, http.StatusOK, detailHTTP.StatusCode, string(detailRaw))
	require.Equal(t, int(xerrors.CodeSuccess), detailResp.Code, string(detailRaw))

	// 更新名称
	updReq := map[string]interface{}{"dungeon_name": "自动地城更新"}
	updResp, updHTTP, updRaw, err := apitest.PutJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/admin/dungeons/"+dungeonID, updReq, token)
	require.NoError(t, err, string(updRaw))
	require.Equal(t, http.StatusOK, updHTTP.StatusCode, string(updRaw))
	require.Equal(t, int(xerrors.CodeSuccess), updResp.Code, string(updRaw))

	// 删除
	delResp, delHTTP, delRaw, err := apitest.DeleteJSON[struct{}, map[string]interface{}](ctx, client, "/api/v1/admin/dungeons/"+dungeonID, nil, token)
	require.NoError(t, err, string(delRaw))
	require.Equal(t, http.StatusOK, delHTTP.StatusCode, string(delRaw))
	require.Equal(t, int(xerrors.CodeSuccess), delResp.Code, string(delRaw))
}
