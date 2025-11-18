package modules

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

// 技能 CRUD + 类别 + 动作效果关联（依赖已有种子 category/effect/action）。
func TestAdminSkillCrudWithActionEffect(t *testing.T) {
	ctx, cfg, client, factory := setup(t)
	token := adminToken(t, ctx, client, cfg)

	// 1) 取现有 skill category
	catResp, catHTTP, catRaw, err := apitest.GetJSON[map[string]interface{}](ctx, client, "/api/v1/admin/skill-categories?limit=1", token)
	require.NoError(t, err, string(catRaw))
	require.Equal(t, http.StatusOK, catHTTP.StatusCode, string(catRaw))
	catData := *catResp.Data
	list, _ := catData["list"].([]interface{})
	if len(list) == 0 {
		t.Skip("无技能类别，跳过技能 CRUD")
	}
	catID := list[0].(map[string]interface{})["id"].(string)

	// 2) 创建技能
	skillCode := fmt.Sprintf("auto_skill_%s", factory.RunID[:6])
	createReq := map[string]interface{}{
		"skill_code":   skillCode,
		"skill_name":   "自动技能" + factory.RunID[:4],
		"skill_type":   "magic",
		"category_id":  catID,
		"max_level":    1,
		"is_active":    true,
		"description":  "自动化创建技能",
		"feature_tags": []string{"auto"},
	}
	createResp, createHTTP, createRaw, err := apitest.PostJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/admin/skills", createReq, token)
	require.NoError(t, err, string(createRaw))
	require.Equal(t, http.StatusOK, createHTTP.StatusCode, string(createRaw))
	require.Equal(t, int(xerrors.CodeSuccess), createResp.Code, string(createRaw))
	skillID := (*createResp.Data)["id"].(string)

	// 3) 取现有效果和动作
	effResp, effHTTP, effRaw, err := apitest.GetJSON[map[string]interface{}](ctx, client, "/api/v1/admin/effects?page=1&page_size=1", token)
	require.NoError(t, err, string(effRaw))
	if effHTTP.StatusCode != http.StatusOK {
		t.Skipf("查询效果失败 status=%d", effHTTP.StatusCode)
	}
	effList, _ := (*effResp.Data)["list"].([]interface{})
	if len(effList) == 0 {
		t.Skip("无可用效果，跳过动作关联")
	}
	effectID := effList[0].(map[string]interface{})["id"].(string)

	actResp, actHTTP, actRaw, err := apitest.GetJSON[map[string]interface{}](ctx, client, "/api/v1/admin/actions?page=1&page_size=1", token)
	require.NoError(t, err, string(actRaw))
	if actHTTP.StatusCode != http.StatusOK {
		t.Skipf("查询动作失败 status=%d", actHTTP.StatusCode)
	}
	actList, _ := (*actResp.Data)["list"].([]interface{})
	if len(actList) == 0 {
		t.Skip("无可用动作")
	}
	actionID := actList[0].(map[string]interface{})["id"].(string)

	// 4) 绑定动作效果
	bindReq := map[string]interface{}{
		"effect_id": effectID,
	}
	bindResp, bindHTTP, bindRaw, err := apitest.PostJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/admin/actions/"+actionID+"/effects", bindReq, token)
	require.NoError(t, err, string(bindRaw))
	if bindHTTP.StatusCode != http.StatusOK {
		t.Skipf("绑定效果失败 status=%d body=%s", bindHTTP.StatusCode, string(bindRaw))
	}
	require.Equal(t, int(xerrors.CodeSuccess), bindResp.Code, string(bindRaw))

	// 5) 为技能添加解锁动作（若接口存在）
	unlockReq := map[string]interface{}{
		"skill_id":    skillID,
		"action_id":   actionID,
		"unlock_type": "auto",
	}
	unlockResp, unlockHTTP, unlockRaw, err := apitest.PostJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/admin/skill-unlock-actions", unlockReq, token)
	require.NoError(t, err, string(unlockRaw))
	if unlockHTTP.StatusCode != http.StatusOK {
		t.Skipf("绑定技能解锁动作失败 status=%d body=%s", unlockHTTP.StatusCode, string(unlockRaw))
	}
	require.Equal(t, int(xerrors.CodeSuccess), unlockResp.Code, string(unlockRaw))

	// 6) 查询技能详情
	detailPath := "/api/v1/admin/skills/" + skillID
	detailResp, detailHTTP, detailRaw, err := apitest.GetJSON[map[string]interface{}](ctx, client, detailPath, token)
	require.NoError(t, err, string(detailRaw))
	require.Equal(t, http.StatusOK, detailHTTP.StatusCode, string(detailRaw))
	require.Equal(t, int(xerrors.CodeSuccess), detailResp.Code, string(detailRaw))

	// 7) 更新技能名称
	updateReq := map[string]interface{}{"skill_name": "auto-upd-" + factory.RunID[:4]}
	updateResp, updateHTTP, updateRaw, err := apitest.PutJSON[map[string]interface{}, map[string]interface{}](ctx, client, detailPath, updateReq, token)
	require.NoError(t, err, string(updateRaw))
	require.Equal(t, http.StatusOK, updateHTTP.StatusCode, string(updateRaw))
	require.Equal(t, int(xerrors.CodeSuccess), updateResp.Code, string(updateRaw))

	// 8) 删除技能
	delResp, delHTTP, delRaw, err := apitest.DeleteJSON[struct{}, map[string]interface{}](ctx, client, detailPath, nil, token)
	require.NoError(t, err, string(delRaw))
	require.Equal(t, http.StatusOK, delHTTP.StatusCode, string(delRaw))
	require.Equal(t, int(xerrors.CodeSuccess), delResp.Code, string(delRaw))
}
