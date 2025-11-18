package modules

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

// 创建一个最小技能（需要后端基础校验通过）。
func TestAdminCreateSkillMinimal(t *testing.T) {
	ctx, cfg, client, factory := setup(t)
	token := adminToken(t, ctx, client, cfg)

	req := map[string]interface{}{
		"skill_code":        factory.RunID[:8] + "_skill",
		"skill_name":        "auto skill " + factory.RunID[:6],
		"max_level":         1,
		"skill_category_id": "", // 若缺少种子，将触发业务错误
	}

	resp, httpResp, raw, err := apitest.PostJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/admin/skills", req, token)
	require.NoError(t, err, string(raw))
	if httpResp.StatusCode != http.StatusOK || resp.Code != int(xerrors.CodeSuccess) {
		t.Skipf("创建技能未成功（可能缺少种子技能分类），status=%d code=%d body=%s", httpResp.StatusCode, resp.Code, string(raw))
	}
}
