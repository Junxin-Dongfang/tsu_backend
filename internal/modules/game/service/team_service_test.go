package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tsu-self/internal/pkg/xerrors"
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

// cleanupTestData 清理测试数据
func cleanupTestData(t *testing.T, db *sql.DB, teamID string) {
	t.Helper()

	_, _ = db.Exec("DELETE FROM game_runtime.team_members WHERE team_id = $1", teamID)
	_, _ = db.Exec("DELETE FROM game_runtime.team_warehouses WHERE team_id = $1", teamID)
	_, _ = db.Exec("DELETE FROM game_runtime.teams WHERE id = $1", teamID)
}

// TestTeamService_CreateTeam 测试创建团队
func TestTeamService_CreateTeam(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupTestDB(t)
	defer db.Close()

	// 测试环境不需要 Keto,传入 nil
	teamService := NewTeamService(db, nil)
	ctx := context.Background()

	userID := "test-user-001"
	heroID := "test-hero-001"

	tests := []struct {
		name      string
		req       *CreateTeamRequest
		wantError bool
		errorCode xerrors.ErrorCode
	}{
		{
			name: "成功创建团队",
			req: &CreateTeamRequest{
				UserID:      userID,
				HeroID:      heroID,
				TeamName:    "测试团队-" + time.Now().Format("20060102150405"),
				Description: "这是一个测试团队",
			},
			wantError: false,
		},
		{
			name: "团队名称为空",
			req: &CreateTeamRequest{
				UserID:   userID,
				HeroID:   heroID,
				TeamName: "",
			},
			wantError: true,
			errorCode: xerrors.CodeInvalidParams,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			team, err := teamService.CreateTeam(ctx, tt.req)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorCode != 0 {
					xerr, ok := err.(*xerrors.AppError)
					if ok {
						assert.Equal(t, tt.errorCode, xerr.Code)
					}
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, team)
				assert.NotEmpty(t, team.ID)
				assert.Equal(t, tt.req.TeamName, team.Name)

				defer cleanupTestData(t, db, team.ID)

				// 验证仓库是否创建
				var warehouseID string
				err := db.QueryRow("SELECT id FROM game_runtime.team_warehouses WHERE team_id = $1", team.ID).Scan(&warehouseID)
				assert.NoError(t, err)
			}
		})
	}
}

// TestTeamService_UpdateTeamInfo 测试更新团队信息
func TestTeamService_UpdateTeamInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupTestDB(t)
	defer db.Close()

	// 测试环境不需要 Keto,传入 nil
	teamService := NewTeamService(db, nil)
	ctx := context.Background()

	userID := "test-user-001"
	heroID := "test-hero-001"
	createReq := &CreateTeamRequest{
		UserID:      userID,
		HeroID:      heroID,
		TeamName:    "测试团队-" + time.Now().Format("20060102150405"),
		Description: "原始描述",
	}

	team, err := teamService.CreateTeam(ctx, createReq)
	require.NoError(t, err)
	defer cleanupTestData(t, db, team.ID)

	newName := "新团队名称-" + time.Now().Format("20060102150405")
	newDesc := "新的描述"
	updateReq := &UpdateTeamInfoRequest{
		TeamID:      team.ID,
		HeroID:      heroID,
		Name:        newName,
		Description: &newDesc,
	}

	err = teamService.UpdateTeamInfo(ctx, updateReq)
	assert.NoError(t, err)

	updatedTeam, err := teamService.GetTeamByID(ctx, team.ID)
	require.NoError(t, err)
	assert.Equal(t, newName, updatedTeam.Name)
}

// 运行测试：
// go test -v ./internal/modules/game/service -run TestTeamService -short  # 跳过集成测试
// go test -v ./internal/modules/game/service -run TestTeamService         # 运行集成测试
