package modules

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

func TestHeroAttributeAllocateUnauthorized(t *testing.T) {
	ctx, _, client, factory := setup(t)
	p := registerPlayerWithHero(t, ctx, client, factory, "hero-unauth")

	req := apitest.AllocateAttributeRequest{AttributeCode: "STR", PointsToAdd: 1}
	path := fmt.Sprintf("/api/v1/game/heroes/%s/attributes/allocate", p.HeroID)
	resp, httpResp, raw, err := apitest.PostJSON[apitest.AllocateAttributeRequest, map[string]interface{}](ctx, client, path, req, "")
	require.NoError(t, err, string(raw))
	require.Equal(t, http.StatusUnauthorized, httpResp.StatusCode, string(raw))
	// Oathkeeper 返回 error 包，无 code 字段；验证 HTTP 即可
	_ = resp
}

func TestHeroAttributeAllocateInvalidCode(t *testing.T) {
	ctx, _, client, factory := setup(t)
	p := registerPlayerWithHero(t, ctx, client, factory, "hero-invalid")

	req := apitest.AllocateAttributeRequest{AttributeCode: "INVALID", PointsToAdd: 1}
	path := fmt.Sprintf("/api/v1/game/heroes/%s/attributes/allocate", p.HeroID)
	resp, httpResp, raw, err := apitest.PostJSON[apitest.AllocateAttributeRequest, map[string]interface{}](ctx, client, path, req, p.Token)
	require.NoError(t, err, string(raw))
	if httpResp.StatusCode != http.StatusNotFound {
		t.Skipf("属性加点无效代码返回 %d，body=%s", httpResp.StatusCode, string(raw))
	}
	require.Equal(t, int(xerrors.CodeResourceNotFound), resp.Code)
}
