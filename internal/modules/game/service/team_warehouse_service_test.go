package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tsu-self/internal/modules/game/testseed"
)

// TestTeamWarehouseService_GetWarehouse 测试获取团队仓库
func TestTeamWarehouseService_GetWarehouse(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupTestDB(t)
	defer db.Close()

	teamService := NewTeamService(db, nil)
	warehouseService := NewTeamWarehouseService(db)
	ctx := context.Background()

	// 创建测试团队
	leaderUserID := testseed.EnsureUser(t, db, "team-warehouse-leader")
	leaderHeroID := testseed.EnsureHero(t, db, leaderUserID, "team-warehouse-leader-hero")
	testseed.CleanupTeamsByHero(t, db, leaderHeroID)
	createReq := &CreateTeamRequest{
		UserID:      leaderUserID.String(),
		HeroID:      leaderHeroID.String(),
		TeamName:    "测试团队-" + time.Now().Format("20060102150405"),
		Description: "测试团队",
	}

	team, err := teamService.CreateTeam(ctx, createReq)
	require.NoError(t, err)
	defer cleanupTestData(t, db, team.ID)

	tests := []struct {
		name      string
		teamID    string
		heroID    string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "队长成功获取仓库",
			teamID:    team.ID,
			heroID:    leaderHeroID.String(),
			wantError: false,
		},
		{
			name:      "参数为空",
			teamID:    "",
			heroID:    "",
			wantError: true,
			errorMsg:  "参数不能为空",
		},
		{
			name:      "非团队成员",
			teamID:    team.ID,
			heroID:    testseed.StableUUID("team-warehouse-non-member").String(),
			wantError: true,
			errorMsg:  "您不是该团队成员",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warehouse, err := warehouseService.GetWarehouse(ctx, tt.teamID, tt.heroID)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, warehouse)
				assert.Equal(t, tt.teamID, warehouse.TeamID)
			}
		})
	}
}

// TestTeamWarehouseService_DistributeGold 测试分配金币
func TestTeamWarehouseService_DistributeGold(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupTestDB(t)
	defer db.Close()

	teamService := NewTeamService(db, nil)
	memberService := NewTeamMemberService(db, nil)
	warehouseService := NewTeamWarehouseService(db)
	ctx := context.Background()

	// 创建测试团队
	leaderUserID := testseed.EnsureUser(t, db, "team-warehouse-leader")
	leaderHeroID := testseed.EnsureHero(t, db, leaderUserID, "team-warehouse-leader-hero")
	testseed.CleanupTeamsByHero(t, db, leaderHeroID)
	createReq := &CreateTeamRequest{
		UserID:      leaderUserID.String(),
		HeroID:      leaderHeroID.String(),
		TeamName:    "测试团队-" + time.Now().Format("20060102150405"),
		Description: "测试团队",
	}

	team, err := teamService.CreateTeam(ctx, createReq)
	require.NoError(t, err)
	defer cleanupTestData(t, db, team.ID)

	// 添加一些金币到仓库
	_, err = db.Exec("UPDATE game_runtime.team_warehouses SET gold_amount = 10000 WHERE team_id = $1", team.ID)
	require.NoError(t, err)

	// 添加一个成员
	memberUserID := testseed.EnsureUser(t, db, "team-warehouse-member")
	memberHeroID := testseed.EnsureHero(t, db, memberUserID, "team-warehouse-member-hero")

	// 创建并批准加入申请
	applyReq := &ApplyToJoinRequest{
		TeamID:  team.ID,
		HeroID:  memberHeroID.String(),
		UserID:  memberUserID.String(),
		Message: "测试加入",
	}
	err = memberService.ApplyToJoin(ctx, applyReq)
	require.NoError(t, err)

	// 获取申请ID并批准
	var requestID string
	err = db.QueryRow("SELECT id FROM game_runtime.team_join_requests WHERE team_id = $1 AND hero_id = $2", team.ID, memberHeroID.String()).Scan(&requestID)
	require.NoError(t, err)

	approveReq := &ApproveJoinRequestRequest{
		RequestID: requestID,
		HeroID:    leaderHeroID.String(),
		Approved:  true,
	}
	err = memberService.ApproveJoinRequest(ctx, approveReq)
	require.NoError(t, err)

	defer func() {
		_, _ = db.Exec("DELETE FROM game_runtime.team_join_requests WHERE id = $1", requestID)
		_, _ = db.Exec("DELETE FROM game_runtime.team_members WHERE team_id = $1 AND hero_id = $2", team.ID, memberHeroID.String())
	}()

	tests := []struct {
		name          string
		distributorID string
		distributions map[string]int64
		wantError     bool
		errorMsg      string
	}{
		{
			name:          "队长成功分配金币",
			distributorID: leaderHeroID.String(),
			distributions: map[string]int64{
				memberHeroID.String(): 1000,
			},
			wantError: false,
		},
		{
			name:          "参数为空",
			distributorID: "",
			distributions: map[string]int64{},
			wantError:     true,
			errorMsg:      "参数不能为空",
		},
		{
			name:          "分配金额为负数",
			distributorID: leaderHeroID.String(),
			distributions: map[string]int64{
				memberHeroID.String(): -100,
			},
			wantError: true,
			errorMsg:  "分配金额必须大于0",
		},
		{
			name:          "仓库余额不足",
			distributorID: leaderHeroID.String(),
			distributions: map[string]int64{
				memberHeroID.String(): 100000,
			},
			wantError: true,
			errorMsg:  "仓库金币不足",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &DistributeGoldRequest{
				TeamID:        team.ID,
				DistributorID: tt.distributorID,
				Distributions: tt.distributions,
			}

			err := warehouseService.DistributeGold(ctx, req)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)

				// 验证金币是否被扣除
				var goldAmount int64
				err := db.QueryRow("SELECT gold_amount FROM game_runtime.team_warehouses WHERE team_id = $1", team.ID).Scan(&goldAmount)
				assert.NoError(t, err)
				assert.Less(t, goldAmount, int64(10000))
			}
		})
	}
}

// TestTeamWarehouseService_AddLootToWarehouse 测试战利品入库
func TestTeamWarehouseService_AddLootToWarehouse(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupTestDB(t)
	defer db.Close()

	teamService := NewTeamService(db, nil)
	warehouseService := NewTeamWarehouseService(db)
	ctx := context.Background()

	// 创建测试团队
	leaderUserID := testseed.EnsureUser(t, db, "team-warehouse-leader")
	leaderHeroID := testseed.EnsureHero(t, db, leaderUserID, "team-warehouse-leader-hero")
	testseed.CleanupTeamsByHero(t, db, leaderHeroID)
	createReq := &CreateTeamRequest{
		UserID:      leaderUserID.String(),
		HeroID:      leaderHeroID.String(),
		TeamName:    "测试团队-" + time.Now().Format("20060102150405"),
		Description: "测试团队",
	}

	team, err := teamService.CreateTeam(ctx, createReq)
	require.NoError(t, err)
	defer cleanupTestData(t, db, team.ID)

	item1 := testseed.StableUUID("item-001").String()
	item2 := testseed.StableUUID("item-002").String()
	dungeonID := testseed.StableUUID("dungeon-001").String()

	tests := []struct {
		name      string
		teamID    string
		gold      int64
		items     []LootItem
		wantError bool
	}{
		{
			name:   "成功添加战利品",
			teamID: team.ID,
			gold:   5000,
			items: []LootItem{
				{ItemID: item1, ItemType: "equipment", Quantity: 2},
				{ItemID: item2, ItemType: "consumable", Quantity: 10},
			},
			wantError: false,
		},
		{
			name:      "团队ID为空",
			teamID:    "",
			gold:      1000,
			items:     []LootItem{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &AddLootToWarehouseRequest{
				TeamID:          tt.teamID,
				SourceDungeonID: dungeonID,
				Gold:            tt.gold,
				Items:           tt.items,
			}

			err := warehouseService.AddLootToWarehouse(ctx, req)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// 验证金币是否添加成功
				var goldAmount int64
				err := db.QueryRow("SELECT gold_amount FROM game_runtime.team_warehouses WHERE team_id = $1", tt.teamID).Scan(&goldAmount)
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, goldAmount, tt.gold)
			}
		})
	}
}

// TestTeamWarehouseService_GetWarehouseItems 测试获取仓库物品列表
func TestTeamWarehouseService_GetWarehouseItems(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupTestDB(t)
	defer db.Close()

	teamService := NewTeamService(db, nil)
	warehouseService := NewTeamWarehouseService(db)
	ctx := context.Background()

	// 创建测试团队
	leaderUserID := testseed.EnsureUser(t, db, "team-warehouse-leader")
	leaderHeroID := testseed.EnsureHero(t, db, leaderUserID, "team-warehouse-leader-hero")
	testseed.CleanupTeamsByHero(t, db, leaderHeroID)
	createReq := &CreateTeamRequest{
		UserID:      leaderUserID.String(),
		HeroID:      leaderHeroID.String(),
		TeamName:    "测试团队-" + time.Now().Format("20060102150405"),
		Description: "测试团队",
	}

	team, err := teamService.CreateTeam(ctx, createReq)
	require.NoError(t, err)
	defer cleanupTestData(t, db, team.ID)

	item1 := testseed.StableUUID("item-001").String()
	item2 := testseed.StableUUID("item-002").String()
	dungeonID := testseed.StableUUID("dungeon-001").String()

	// 添加一些物品
	addReq := &AddLootToWarehouseRequest{
		TeamID:          team.ID,
		SourceDungeonID: dungeonID,
		Gold:            0,
		Items: []LootItem{
			{ItemID: item1, ItemType: "equipment", Quantity: 1},
			{ItemID: item2, ItemType: "consumable", Quantity: 5},
		},
	}
	err = warehouseService.AddLootToWarehouse(ctx, addReq)
	require.NoError(t, err)

	tests := []struct {
		name      string
		teamID    string
		heroID    string
		limit     int
		offset    int
		wantError bool
	}{
		{
			name:      "队长成功获取物品列表",
			teamID:    team.ID,
			heroID:    leaderHeroID.String(),
			limit:     10,
			offset:    0,
			wantError: false,
		},
		{
			name:      "参数为空",
			teamID:    "",
			heroID:    "",
			limit:     10,
			offset:    0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, total, err := warehouseService.GetWarehouseItems(ctx, tt.teamID, tt.heroID, tt.limit, tt.offset)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, items)
				assert.GreaterOrEqual(t, total, int64(2))
			}
		})
	}
}

// 运行测试：
// go test -v ./internal/modules/game/service -run TestTeamWarehouseService
