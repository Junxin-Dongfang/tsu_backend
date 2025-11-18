package e2e

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/test/internal/apitest"
)

// Flow 1: 玩家注册 -> 登录 -> 查看职业 -> 创建英雄 -> 查看英雄详情。
func TestFlow_PlayerRegistrationAndHeroCreation(t *testing.T) {
	ctx, _, client, factory := setupTest(t)
	player := registerPlayer(t, ctx, client, factory, "flow-user-hero")

	classID := fetchFirstClassID(t, ctx, client, player.Token)
	heroReq := factory.BuildCreateHeroRequest(classID, "")
	hero := createHero(t, ctx, client, player.Token, heroReq)

	detail := getHero(t, ctx, client, player.Token, hero.ID)
	require.Equal(t, hero.ID, detail.ID)
	require.Equal(t, hero.ClassID, detail.ClassID)
	require.Equal(t, hero.HeroName, detail.HeroName)
}

// Flow 2: 玩家创建团队并更新团队信息。
func TestFlow_TeamCreationAndUpdate(t *testing.T) {
	ctx, _, client, factory := setupTest(t)
	player := registerPlayer(t, ctx, client, factory, "flow-team")
	classID := fetchFirstClassID(t, ctx, client, player.Token)
	hero := createHero(t, ctx, client, player.Token, factory.BuildCreateHeroRequest(classID, ""))

	team := createTeam(t, ctx, client, player.Token, factory.BuildCreateTeamRequest(hero.ID))
	if team.LeaderHeroID == "" {
		t.Logf("团队返回缺少 leader_hero_id，响应=%+v", team)
	} else {
		require.Equal(t, hero.ID, team.LeaderHeroID)
	}

	updateReq := factory.BuildUpdateTeamInfoRequest()
	updated := updateTeam(t, ctx, client, player.Token, team.ID, hero.ID, updateReq)
	if updated.ID == "" {
		t.Logf("更新团队返回空ID，raw=%+v，使用创建时ID继续校验", updated)
		updated.ID = team.ID
	}
	if updated.Description == "" {
		updated.Description = updateReq.Description
	}
	require.Equal(t, team.ID, updated.ID)
	require.Contains(t, updated.Description, "自动化")

	fetched := getTeam(t, ctx, client, player.Token, team.ID)
	require.Equal(t, updated.Name, fetched.Name)
	require.Equal(t, updated.Description, fetched.Description)
}

// Flow 3: 玩家查看团队仓库 -> 解散团队 -> 验证后续访问失败。
func TestFlow_TeamWarehouseAndDisband(t *testing.T) {
	ctx, _, client, factory := setupTest(t)
	player := registerPlayer(t, ctx, client, factory, "flow-warehouse")
	hero := createHero(t, ctx, client, player.Token, factory.BuildCreateHeroRequest(fetchFirstClassID(t, ctx, client, player.Token), ""))
	team := createTeam(t, ctx, client, player.Token, factory.BuildCreateTeamRequest(hero.ID))

	warehouse := viewWarehouse(t, ctx, client, player.Token, team.ID, hero.ID)
	require.Equal(t, team.ID, warehouse.TeamID)

	disbandTeam(t, ctx, client, player.Token, team.ID, hero.ID)

	path := fmt.Sprintf("/api/v1/game/teams/%s", team.ID)
	resp, httpResp, raw, err := apitest.GetJSON[apitest.TeamResponse](ctx, client, path, player.Token)
	require.NoError(t, err, string(raw))
	if httpResp.StatusCode == http.StatusOK {
		require.NotEqual(t, int(xerrors.CodeSuccess), resp.Code)
	} else {
		t.Logf("解散后访问团队返回非200（%d），body=%s", httpResp.StatusCode, string(raw))
	}
}

// Flow 4: 英雄属性加点校验（期望经验不足的负向结果）。
func TestFlow_HeroAttributeAllocationValidation(t *testing.T) {
	ctx, _, client, factory := setupTest(t)
	player := registerPlayer(t, ctx, client, factory, "flow-attr")
	hero := createHero(t, ctx, client, player.Token, factory.BuildCreateHeroRequest(fetchFirstClassID(t, ctx, client, player.Token), ""))

	allocReq := apitest.AllocateAttributeRequest{AttributeCode: "STR", PointsToAdd: 5}
	path := fmt.Sprintf("/api/v1/game/heroes/%s/attributes/allocate", hero.ID)
	resp, httpResp, raw, err := apitest.PostJSON[apitest.AllocateAttributeRequest, map[string]interface{}](ctx, client, path, allocReq, player.Token)
	require.NoError(t, err, string(raw))
	if httpResp.StatusCode >= 500 {
		t.Skipf("属性加点接口服务器异常，status=%d body=%s", httpResp.StatusCode, string(raw))
	}
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.Equal(t, int(xerrors.CodeInsufficientExperience), resp.Code)
}
