package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tsu-self/internal/modules/game/testseed"
)

// TestTeamFlow_CreateAndDistributeGold 覆盖从创建团队到仓库分配金币的闭环
func TestTeamFlow_CreateAndDistributeGold(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupTestDB(t)
	defer db.Close()

	teamSvc := NewTeamService(db, nil)
	warehouseSvc := NewTeamWarehouseService(db)
	ctx := context.Background()

	leaderUserID := testseed.EnsureUser(t, db, "team-flow-leader")
	leaderHeroID := testseed.EnsureHero(t, db, leaderUserID, "team-flow-leader-hero")
	team, err := teamSvc.CreateTeam(ctx, &CreateTeamRequest{
		UserID:      leaderUserID.String(),
		HeroID:      leaderHeroID.String(),
		TeamName:    "flow-team-" + time.Now().Format("150405"),
		Description: "flow test",
	})
	require.NoError(t, err)
	defer cleanupTestData(t, db, team.ID)

	// 添加一个成员作为分配对象
	memberUserID := testseed.EnsureUser(t, db, "team-flow-member")
	memberHeroID := testseed.EnsureHero(t, db, memberUserID, "team-flow-member-hero")
	_, err = db.Exec(`
		INSERT INTO game_runtime.team_members (id, team_id, hero_id, user_id, role, joined_at)
		VALUES (gen_random_uuid(), $1, $2, $3, 'member', NOW())
	`, team.ID, memberHeroID.String(), memberUserID.String())
	require.NoError(t, err)

	// 补充仓库金币以便分配
	_, err = db.Exec(`UPDATE game_runtime.team_warehouses SET gold_amount = 1000 WHERE team_id = $1`, team.ID)
	require.NoError(t, err)

	err = warehouseSvc.DistributeGold(ctx, &DistributeGoldRequest{
		TeamID:        team.ID,
		DistributorID: leaderHeroID.String(),
		Distributions: map[string]int64{
			memberHeroID.String(): 300,
		},
	})
	assert.NoError(t, err)

	var goldLeft int64
	_ = db.QueryRow(`SELECT gold_amount FROM game_runtime.team_warehouses WHERE team_id = $1`, team.ID).Scan(&goldLeft)
	assert.Equal(t, int64(700), goldLeft)
}
