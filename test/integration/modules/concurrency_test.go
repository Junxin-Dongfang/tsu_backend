package modules

import (
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

type apiCallResult struct {
	status int
	code   int
	body   string
	err    error
}

// TestCreateTeamDuplicateSubmission ensures repeated提交会被幂等拒绝（用于模拟多次点击“创建团队”的场景）。
func TestCreateTeamDuplicateSubmission(t *testing.T) {
	ctx, _, client, factory := setup(t)
	player := registerPlayerWithHero(t, ctx, client, factory, "concurrent-team")

	req := factory.BuildCreateTeamRequest(player.HeroID)

	// 第一次提交——成功创建团队
	firstResp, firstHTTP, firstRaw, err := apitest.PostJSON[apitest.CreateTeamRequest, apitest.TeamResponse](ctx, client, "/api/v1/game/teams", req, player.Token)
	require.NoError(t, err, string(firstRaw))
	require.Equal(t, http.StatusOK, firstHTTP.StatusCode, string(firstRaw))
	require.Equal(t, int(xerrors.CodeSuccess), firstResp.Code, string(firstRaw))

	// 立即使用相同 payload 再提交一次，应返回重复错误，证明服务支持幂等
	secondResp, secondHTTP, secondRaw, err := apitest.PostJSON[apitest.CreateTeamRequest, apitest.TeamResponse](ctx, client, "/api/v1/game/teams", req, player.Token)
	require.NoError(t, err, string(secondRaw))
	require.Equal(t, http.StatusConflict, secondHTTP.StatusCode, string(secondRaw))
	require.Equal(t, int(xerrors.CodeDuplicateResource), secondResp.Code, string(secondRaw))
}

// TestWorldDropItemConcurrentCreate validates admin侧的世界掉落物品 API 在并发写入时保持幂等。
func TestWorldDropItemConcurrentCreate(t *testing.T) {
	ctx, cfg, client, factory := setup(t)
	token := adminToken(t, ctx, client, cfg)

	baseItemID := createAdminItem(t, ctx, client, token, factory, "world-drop-concurrent-base")
	createReq := map[string]interface{}{
		"item_id":        baseItemID,
		"base_drop_rate": 0.25,
	}
	createResp, httpResp, raw, err := apitest.PostJSON[map[string]interface{}, map[string]interface{}](ctx, client, "/api/v1/admin/world-drops", createReq, token)
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusOK, httpResp.StatusCode, string(raw))
	require.Equal(t, int(xerrors.CodeSuccess), createResp.Code, string(raw))
	worldDropID := (*createResp.Data)["id"].(string)

	t.Cleanup(func() {
		cleanupAdminResource(t, client, token, "/api/v1/admin/world-drops/"+worldDropID)
	})

	extraItemID := createAdminItem(t, ctx, client, token, factory, "world-drop-concurrent-extra")
	payload := map[string]interface{}{
		"item_id":      extraItemID,
		"drop_rate":    0.1,
		"min_quantity": 1,
		"max_quantity": 1,
	}

	var wg sync.WaitGroup
	results := make(chan apiCallResult, 2)
	path := "/api/v1/admin/world-drops/" + worldDropID + "/items"
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, httpResp, raw, err := apitest.PostJSON[map[string]interface{}, map[string]interface{}](ctx, client, path, payload, token)
			result := apiCallResult{body: string(raw), err: err}
			if httpResp != nil {
				result.status = httpResp.StatusCode
			}
			if resp != nil {
				result.code = resp.Code
			}
			results <- result
		}()
	}
	wg.Wait()
	close(results)

	var dropSuccess bool
	var duplicate bool
	for res := range results {
		require.NoError(t, res.err, res.body)
		switch {
		case res.status == http.StatusOK && res.code == int(xerrors.CodeSuccess):
			dropSuccess = true
		case res.code == int(xerrors.CodeDuplicateResource):
			duplicate = true
		default:
			t.Fatalf("unexpected world-drop item response: status=%d code=%d body=%s", res.status, res.code, res.body)
		}
	}
	require.True(t, dropSuccess, "expected one successful world-drop item creation")
	require.True(t, duplicate, "expected duplicate detection for concurrent world-drop item creation")
}
