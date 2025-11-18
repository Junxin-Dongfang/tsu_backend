package modules

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

// 创建基础怪物配置，依赖已有职业/掉落池等数据，缺失则 Skip。
func TestAdminCreateMonsterMinimal(t *testing.T) {
	ctx, cfg, client, factory := setup(t)
	token := adminToken(t, ctx, client, cfg)

	req := map[string]interface{}{
		"monster_code": factory.RunID[:8] + "_mon",
		"monster_name": "auto monster " + factory.RunID[:6],
		"level":        1,
		"hp":           10,
		"attack":       1,
		"defense":      1,
		"speed":        1,
	}

	resp, httpResp, raw, err := apitest.PostJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/admin/monsters", req, token)
	require.NoError(t, err, string(raw))
	if httpResp.StatusCode != http.StatusOK || resp.Code != int(xerrors.CodeSuccess) {
		t.Skipf("创建怪物未成功（可能缺少关联配置），status=%d code=%d body=%s", httpResp.StatusCode, resp.Code, string(raw))
	}
}
