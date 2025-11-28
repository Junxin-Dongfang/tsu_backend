package service

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"tsu-self/internal/modules/game/testseed"
)

// setupHeroTestDB 设置测试数据库连接
func setupHeroTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := "host=localhost port=5432 user=tsu_user password=tsu_test dbname=tsu_db sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skipf("无法连接测试数据库: %v", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		t.Skipf("跳过依赖数据库的测试，原因: %v", err)
	}

	return db
}

func TestHeroActivationService_ActivateFirstHero(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupHeroTestDB(t)
	defer db.Close()

	service := NewHeroActivationService(db)
	ctx := context.Background()

	// 准备测试数据
	userID := testseed.EnsureUser(t, db, "activate-first-hero")
	heroID := testseed.EnsureHero(t, db, userID, "activate-first-hero-hero")

	// 激活第一个英雄
	err := service.ActivateHero(ctx, userID.String(), heroID.String())
	require.NoError(t, err)

	// 验证英雄已激活
	var isActivated bool
	err = db.QueryRowContext(ctx,
		`SELECT is_activated FROM game_runtime.heroes WHERE id = $1`,
		heroID.String()).Scan(&isActivated)
	require.NoError(t, err)
	require.True(t, isActivated)

	// 验证该英雄成为当前操作英雄
	var currentHeroID string
	err = db.QueryRowContext(ctx,
		`SELECT hero_id FROM game_runtime.current_hero_contexts WHERE user_id = $1`,
		userID.String()).Scan(&currentHeroID)
	require.NoError(t, err)
	require.Equal(t, heroID.String(), currentHeroID)
}

func TestHeroActivationService_ActivateSecondHero(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupHeroTestDB(t)
	defer db.Close()

	service := NewHeroActivationService(db)
	ctx := context.Background()

	// 准备测试数据
	userID := testseed.EnsureUser(t, db, "activate-second-hero")
	heroID1 := testseed.EnsureHero(t, db, userID, "activate-second-hero-1")
	heroID2 := testseed.EnsureHero(t, db, userID, "activate-second-hero-2")

	// 激活第一个英雄
	err := service.ActivateHero(ctx, userID.String(), heroID1.String())
	require.NoError(t, err)

	// 激活第二个英雄
	err = service.ActivateHero(ctx, userID.String(), heroID2.String())
	require.NoError(t, err)

	// 验证第二个英雄已激活
	var isActivated bool
	err = db.QueryRowContext(ctx,
		`SELECT is_activated FROM game_runtime.heroes WHERE id = $1`,
		heroID2.String()).Scan(&isActivated)
	require.NoError(t, err)
	require.True(t, isActivated)

	// 验证当前操作英雄仍是第一个英雄（不自动切换）
	var currentHeroID string
	err = db.QueryRowContext(ctx,
		`SELECT hero_id FROM game_runtime.current_hero_contexts WHERE user_id = $1`,
		userID.String()).Scan(&currentHeroID)
	require.NoError(t, err)
	require.Equal(t, heroID1.String(), currentHeroID)
}

func TestHeroActivationService_CannotDeactivateCurrentHero(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupHeroTestDB(t)
	defer db.Close()

	service := NewHeroActivationService(db)
	ctx := context.Background()

	// 准备测试数据
	userID := testseed.EnsureUser(t, db, "cannot-deactivate-current")
	heroID := testseed.EnsureHero(t, db, userID, "cannot-deactivate-current-hero")

	// 激活英雄
	err := service.ActivateHero(ctx, userID.String(), heroID.String())
	require.NoError(t, err)

	// 尝试停用当前英雄（应失败）
	err = service.DeactivateHero(ctx, userID.String(), heroID.String())
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot deactivate current hero")
}

func TestHeroActivationService_DeactivateNonCurrentHero(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupHeroTestDB(t)
	defer db.Close()

	service := NewHeroActivationService(db)
	ctx := context.Background()

	// 准备测试数据
	userID := testseed.EnsureUser(t, db, "deactivate-non-current")
	heroID1 := testseed.EnsureHero(t, db, userID, "deactivate-non-current-1")
	heroID2 := testseed.EnsureHero(t, db, userID, "deactivate-non-current-2")

	// 激活两个英雄
	err := service.ActivateHero(ctx, userID.String(), heroID1.String())
	require.NoError(t, err)

	err = service.ActivateHero(ctx, userID.String(), heroID2.String())
	require.NoError(t, err)

	// 停用非当前英雄（应成功）
	err = service.DeactivateHero(ctx, userID.String(), heroID2.String())
	require.NoError(t, err)

	// 验证英雄已停用
	var isActivated bool
	err = db.QueryRowContext(ctx,
		`SELECT is_activated FROM game_runtime.heroes WHERE id = $1`,
		heroID2.String()).Scan(&isActivated)
	require.NoError(t, err)
	require.False(t, isActivated)
}

func TestHeroActivationService_CannotSwitchToUnactivatedHero(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupHeroTestDB(t)
	defer db.Close()

	service := NewHeroActivationService(db)
	ctx := context.Background()

	// 准备测试数据
	userID := testseed.EnsureUser(t, db, "cannot-switch-unactivated")
	heroID1 := testseed.EnsureHero(t, db, userID, "cannot-switch-unactivated-1")
	heroID2 := testseed.EnsureHero(t, db, userID, "cannot-switch-unactivated-2")

	// 只激活第一个英雄
	err := service.ActivateHero(ctx, userID.String(), heroID1.String())
	require.NoError(t, err)

	// 尝试切换到未激活的英雄（应失败）
	err = service.SwitchCurrentHero(ctx, userID.String(), heroID2.String())
	require.Error(t, err)
	require.Contains(t, err.Error(), "not activated")
}

func TestHeroActivationService_SwitchToActivatedHero(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupHeroTestDB(t)
	defer db.Close()

	service := NewHeroActivationService(db)
	ctx := context.Background()

	// 准备测试数据
	userID := testseed.EnsureUser(t, db, "switch-activated")
	heroID1 := testseed.EnsureHero(t, db, userID, "switch-activated-1")
	heroID2 := testseed.EnsureHero(t, db, userID, "switch-activated-2")

	// 激活两个英雄
	err := service.ActivateHero(ctx, userID.String(), heroID1.String())
	require.NoError(t, err)

	err = service.ActivateHero(ctx, userID.String(), heroID2.String())
	require.NoError(t, err)

	// 切换到第二个英雄
	err = service.SwitchCurrentHero(ctx, userID.String(), heroID2.String())
	require.NoError(t, err)

	// 验证当前操作英雄已切换
	var currentHeroID string
	err = db.QueryRowContext(ctx,
		`SELECT hero_id FROM game_runtime.current_hero_contexts WHERE user_id = $1`,
		userID.String()).Scan(&currentHeroID)
	require.NoError(t, err)
	require.Equal(t, heroID2.String(), currentHeroID)
}

func TestHeroActivationService_GetActivatedHeroes(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupHeroTestDB(t)
	defer db.Close()

	service := NewHeroActivationService(db)
	ctx := context.Background()

	// 准备测试数据
	userID := testseed.EnsureUser(t, db, "get-activated-heroes")
	heroID1 := testseed.EnsureHero(t, db, userID, "get-activated-heroes-1")
	heroID2 := testseed.EnsureHero(t, db, userID, "get-activated-heroes-2")
	heroID3 := testseed.EnsureHero(t, db, userID, "get-activated-heroes-3")

	// 激活前两个英雄
	err := service.ActivateHero(ctx, userID.String(), heroID1.String())
	require.NoError(t, err)

	err = service.ActivateHero(ctx, userID.String(), heroID2.String())
	require.NoError(t, err)

	// 获取已激活的英雄
	heroes, currentHeroID, err := service.GetActivatedHeroes(ctx, userID.String())
	require.NoError(t, err)

	// 验证结果
	require.Len(t, heroes, 2)
	require.Equal(t, heroID1.String(), currentHeroID)

	// 验证第三个英雄不在列表中
	heroIDs := make([]string, 0)
	for _, h := range heroes {
		heroIDs = append(heroIDs, h.ID)
	}
	require.NotContains(t, heroIDs, heroID3.String())
}
