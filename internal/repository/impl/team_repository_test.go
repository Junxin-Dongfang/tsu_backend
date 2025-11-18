package impl

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tsu-self/internal/entity/game_runtime"
)

// setupTestDB 设置测试数据库连接
func setupTestDB(t *testing.T) *sql.DB {
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

// TestTeamRepository_Create 测试创建团队
func TestTeamRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := NewTeamRepository(db)
	ctx := context.Background()

	// 生成测试用的UUID
	testHeroID := uuid.New().String()

	team := &game_runtime.Team{
		Name:         "测试团队-" + time.Now().Format("20060102150405"),
		LeaderHeroID: testHeroID,
		MaxMembers:   12,
	}
	team.Description.SetValid("测试团队描述")

	err := repo.Create(ctx, team)
	require.NoError(t, err)
	assert.NotEmpty(t, team.ID)

	// 清理测试数据
	defer func() {
		_, _ = db.Exec("DELETE FROM game_runtime.teams WHERE id = $1", team.ID)
	}()

	// 验证数据是否插入
	var name string
	err = db.QueryRow("SELECT name FROM game_runtime.teams WHERE id = $1", team.ID).Scan(&name)
	assert.NoError(t, err)
	assert.Equal(t, team.Name, name)
}

// TestTeamRepository_GetByID 测试根据ID获取团队
func TestTeamRepository_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := NewTeamRepository(db)
	ctx := context.Background()

	// 先创建一个团队
	team := &game_runtime.Team{
		Name:         "测试团队-" + time.Now().Format("20060102150405"),
		LeaderHeroID: uuid.New().String(),
		MaxMembers:   12,
	}

	err := repo.Create(ctx, team)
	require.NoError(t, err)

	defer func() {
		_, _ = db.Exec("DELETE FROM game_runtime.teams WHERE id = $1", team.ID)
	}()

	// 测试获取
	foundTeam, err := repo.GetByID(ctx, team.ID)
	require.NoError(t, err)
	assert.Equal(t, team.ID, foundTeam.ID)
	assert.Equal(t, team.Name, foundTeam.Name)
	assert.Equal(t, team.LeaderHeroID, foundTeam.LeaderHeroID)
}

// TestTeamRepository_Update 测试更新团队
func TestTeamRepository_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := NewTeamRepository(db)
	ctx := context.Background()

	// 先创建一个团队
	team := &game_runtime.Team{
		Name:         "测试团队-" + time.Now().Format("20060102150405"),
		LeaderHeroID: uuid.New().String(),
		MaxMembers:   12,
	}

	err := repo.Create(ctx, team)
	require.NoError(t, err)

	defer func() {
		_, _ = db.Exec("DELETE FROM game_runtime.teams WHERE id = $1", team.ID)
	}()

	// 测试更新
	newName := "新团队名称-" + time.Now().Format("20060102150405")
	team.Name = newName

	err = repo.Update(ctx, team)
	require.NoError(t, err)

	// 验证更新结果
	var name string
	err = db.QueryRow("SELECT name FROM game_runtime.teams WHERE id = $1", team.ID).Scan(&name)
	assert.NoError(t, err)
	assert.Equal(t, newName, name)
}

// TestTeamRepository_Exists 测试检查团队名称是否存在
func TestTeamRepository_Exists(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := NewTeamRepository(db)
	ctx := context.Background()

	teamName := "测试团队-" + time.Now().Format("20060102150405")

	// 先检查不存在
	exists, err := repo.Exists(ctx, teamName)
	require.NoError(t, err)
	assert.False(t, exists)

	// 创建团队
	team := &game_runtime.Team{
		Name:         teamName,
		LeaderHeroID: uuid.New().String(),
		MaxMembers:   12,
	}

	err = repo.Create(ctx, team)
	require.NoError(t, err)

	defer func() {
		_, _ = db.Exec("DELETE FROM game_runtime.teams WHERE id = $1", team.ID)
	}()

	// 再检查存在
	exists, err = repo.Exists(ctx, teamName)
	require.NoError(t, err)
	assert.True(t, exists)
}

// 运行测试：
// go test -v ./internal/repository/impl -run TestTeamRepository
