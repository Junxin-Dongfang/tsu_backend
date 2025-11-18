package modules

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

// 怪物 CRUD：创建→详情→更新→删除
func TestAdminMonsterCrud(t *testing.T) {
	ctx, cfg, client, factory := setup(t)
	token := adminToken(t, ctx, client, cfg)

	code := fmt.Sprintf("auto_mon_%s", factory.RunID[:6])
	createReq := map[string]interface{}{
		"monster_code":  code,
		"monster_name":  "自动怪物" + factory.RunID[:4],
		"monster_level": 1,
		"max_hp":        10,
		"hp_recovery":   0,
		"max_mp":        0,
		"mp_recovery":   0,
		"base_str":      1,
		"base_agi":      1,
		"base_vit":      1,
		"base_wlp":      0,
		"base_int":      0,
		"base_wis":      0,
		"base_cha":      0,
		"drop_gold_min": 0,
		"drop_gold_max": 1,
		"drop_exp":      1,
		"is_active":     true,
	}

	createResp, createHTTP, createRaw, err := apitest.PostJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/admin/monsters", createReq, token)
	require.NoError(t, err, string(createRaw))
	require.Equal(t, http.StatusOK, createHTTP.StatusCode, string(createRaw))
	require.Equal(t, int(xerrors.CodeSuccess), createResp.Code, string(createRaw))
	monsterID := (*createResp.Data)["id"].(string)

	// 详情
	detailResp, detailHTTP, detailRaw, err := apitest.GetJSON[map[string]interface{}](ctx, client, "/api/v1/admin/monsters/"+monsterID, token)
	require.NoError(t, err, string(detailRaw))
	require.Equal(t, http.StatusOK, detailHTTP.StatusCode, string(detailRaw))
	require.Equal(t, int(xerrors.CodeSuccess), detailResp.Code, string(detailRaw))

	// 更新名称
	updReq := map[string]interface{}{"monster_name": "自动怪物更新"}
	updResp, updHTTP, updRaw, err := apitest.PutJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/admin/monsters/"+monsterID, updReq, token)
	require.NoError(t, err, string(updRaw))
	require.Equal(t, http.StatusOK, updHTTP.StatusCode, string(updRaw))
	require.Equal(t, int(xerrors.CodeSuccess), updResp.Code, string(updRaw))

	// 删除
	delResp, delHTTP, delRaw, err := apitest.DeleteJSON[struct{}, map[string]interface{}](ctx, client, "/api/v1/admin/monsters/"+monsterID, nil, token)
	require.NoError(t, err, string(delRaw))
	require.Equal(t, http.StatusOK, delHTTP.StatusCode, string(delRaw))
	require.Equal(t, int(xerrors.CodeSuccess), delResp.Code, string(delRaw))
}
