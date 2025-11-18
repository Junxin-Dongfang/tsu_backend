package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/modules/admin/dto"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/interfaces"
)

// setupTestDB 设置测试数据库连接
func setupTestDB(t *testing.T) *sql.DB {
	// 使用环境变量或默认值连接测试数据库
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=tsu_db sslmode=disable"
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

// cleanupDropPools 清理测试数据
func cleanupDropPools(t *testing.T, db *sql.DB, poolCodes []string) {
	for _, code := range poolCodes {
		_, err := db.Exec("DELETE FROM game_config.drop_pools WHERE pool_code = $1", code)
		if err != nil {
			t.Logf("Warning: failed to cleanup drop pool %s: %v", code, err)
		}
	}
}

// cleanupItems 清理测试物品数据
func cleanupItems(t *testing.T, db *sql.DB, itemCodes []string) {
	for _, code := range itemCodes {
		_, err := db.Exec("DELETE FROM game_config.items WHERE item_code = $1", code)
		if err != nil {
			t.Logf("Warning: failed to cleanup item %s: %v", code, err)
		}
	}
}

// createTestItem 创建测试物品
func createTestItem(t *testing.T, db *sql.DB, itemCode string) *game_config.Item {
	item := &game_config.Item{
		ID:          generateUUID(),
		ItemCode:    itemCode,
		ItemName:    "Test Item " + itemCode,
		ItemType:    "equipment",
		ItemQuality: "normal", // 使用正确的枚举值
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := item.Insert(context.Background(), db, boil.Infer())
	require.NoError(t, err)
	return item
}

func TestDropPoolService_CreateDropPool(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewDropPoolService(db)
	ctx := context.Background()

	testCases := []struct {
		name        string
		request     *dto.CreateDropPoolRequest
		setup       func()
		cleanup     func()
		expectError bool
		errorCode   xerrors.ErrorCode
		validate    func(t *testing.T, resp *dto.DropPoolResponse)
	}{
		{
			name: "成功创建掉落池",
			request: &dto.CreateDropPoolRequest{
				PoolCode:        "TEST_POOL_001",
				PoolName:        "测试掉落池001",
				PoolType:        "monster",
				MinDrops:        1,
				MaxDrops:        3,
				GuaranteedDrops: 1,
			},
			cleanup: func() {
				cleanupDropPools(t, db, []string{"TEST_POOL_001"})
			},
			expectError: false,
			validate: func(t *testing.T, resp *dto.DropPoolResponse) {
				assert.Equal(t, "TEST_POOL_001", resp.PoolCode)
				assert.Equal(t, "测试掉落池001", resp.PoolName)
				assert.Equal(t, "monster", resp.PoolType)
				assert.Equal(t, int16(1), resp.MinDrops)
				assert.Equal(t, int16(3), resp.MaxDrops)
				assert.Equal(t, int16(1), resp.GuaranteedDrops)
				assert.True(t, resp.IsActive)
			},
		},
		{
			name: "pool_code重复应该失败",
			request: &dto.CreateDropPoolRequest{
				PoolCode:        "TEST_POOL_DUP",
				PoolName:        "重复测试",
				PoolType:        "dungeon",
				MinDrops:        1,
				MaxDrops:        2,
				GuaranteedDrops: 1,
			},
			setup: func() {
				// 先创建一个掉落池
				req := &dto.CreateDropPoolRequest{
					PoolCode:        "TEST_POOL_DUP",
					PoolName:        "原始掉落池",
					PoolType:        "monster",
					MinDrops:        1,
					MaxDrops:        2,
					GuaranteedDrops: 1,
				}
				_, err := service.CreateDropPool(ctx, req)
				require.NoError(t, err)
			},
			cleanup: func() {
				cleanupDropPools(t, db, []string{"TEST_POOL_DUP"})
			},
			expectError: true,
			errorCode:   xerrors.CodeDuplicateResource,
		},
		{
			name: "min_drops大于max_drops应该失败",
			request: &dto.CreateDropPoolRequest{
				PoolCode:        "TEST_POOL_INVALID_RANGE",
				PoolName:        "无效范围测试",
				PoolType:        "boss",
				MinDrops:        5,
				MaxDrops:        3,
				GuaranteedDrops: 1,
			},
			expectError: true,
			errorCode:   xerrors.CodeInvalidParams,
		},
		{
			name: "guaranteed_drops大于max_drops应该失败",
			request: &dto.CreateDropPoolRequest{
				PoolCode:        "TEST_POOL_INVALID_GUARANTEED",
				PoolName:        "无效保底测试",
				PoolType:        "quest",
				MinDrops:        1,
				MaxDrops:        3,
				GuaranteedDrops: 5,
			},
			expectError: true,
			errorCode:   xerrors.CodeInvalidParams,
		},
		{
			name: "带描述的掉落池",
			request: &dto.CreateDropPoolRequest{
				PoolCode:        "TEST_POOL_WITH_DESC",
				PoolName:        "带描述的掉落池",
				PoolType:        "activity",
				Description:     stringPtr("这是一个测试描述"),
				MinDrops:        2,
				MaxDrops:        5,
				GuaranteedDrops: 2,
			},
			cleanup: func() {
				cleanupDropPools(t, db, []string{"TEST_POOL_WITH_DESC"})
			},
			expectError: false,
			validate: func(t *testing.T, resp *dto.DropPoolResponse) {
				assert.NotNil(t, resp.Description)
				assert.Equal(t, "这是一个测试描述", *resp.Description)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup()
			}
			if tc.cleanup != nil {
				defer tc.cleanup()
			}

			resp, err := service.CreateDropPool(ctx, tc.request)

			if tc.expectError {
				require.Error(t, err)
				if tc.errorCode != 0 {
					var xerr *xerrors.AppError
					require.ErrorAs(t, err, &xerr, "error should be xerrors.AppError")
					assert.Equal(t, tc.errorCode, xerr.Code)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				if tc.validate != nil {
					tc.validate(t, resp)
				}
			}
		})
	}
}

func TestDropPoolService_GetDropPoolList(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewDropPoolService(db)
	ctx := context.Background()

	// 准备测试数据
	testPools := []string{"TEST_LIST_001", "TEST_LIST_002", "TEST_LIST_003"}
	for i, code := range testPools {
		req := &dto.CreateDropPoolRequest{
			PoolCode:        code,
			PoolName:        "列表测试" + code,
			PoolType:        []string{"monster", "dungeon", "boss"}[i%3],
			MinDrops:        1,
			MaxDrops:        3,
			GuaranteedDrops: 1,
		}
		_, err := service.CreateDropPool(ctx, req)
		require.NoError(t, err)
	}
	defer cleanupDropPools(t, db, testPools)

	t.Run("查询所有掉落池", func(t *testing.T) {
		params := interfaces.ListDropPoolParams{
			Page:     1,
			PageSize: 10,
		}
		resp, err := service.GetDropPoolList(ctx, params)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(resp.Items), 3)
		assert.GreaterOrEqual(t, resp.Total, int64(3))
	})

	t.Run("按pool_type筛选", func(t *testing.T) {
		poolType := "monster"
		params := interfaces.ListDropPoolParams{
			PoolType: &poolType,
			Page:     1,
			PageSize: 10,
		}
		resp, err := service.GetDropPoolList(ctx, params)
		require.NoError(t, err)
		for _, item := range resp.Items {
			assert.Equal(t, "monster", item.PoolType)
		}
	})

	t.Run("关键词搜索", func(t *testing.T) {
		keyword := "TEST_LIST_001"
		params := interfaces.ListDropPoolParams{
			Keyword:  &keyword,
			Page:     1,
			PageSize: 10,
		}
		resp, err := service.GetDropPoolList(ctx, params)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(resp.Items), 1)
	})

	t.Run("分页测试", func(t *testing.T) {
		params := interfaces.ListDropPoolParams{
			Page:     1,
			PageSize: 2,
		}
		resp, err := service.GetDropPoolList(ctx, params)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(resp.Items), 2)
		assert.Equal(t, 1, resp.Page)
		assert.Equal(t, 2, resp.PageSize)
	})
}

func TestDropPoolService_UpdateDropPool(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewDropPoolService(db)
	ctx := context.Background()

	// 创建测试掉落池
	createReq := &dto.CreateDropPoolRequest{
		PoolCode:        "TEST_UPDATE_001",
		PoolName:        "更新测试",
		PoolType:        "monster",
		MinDrops:        1,
		MaxDrops:        3,
		GuaranteedDrops: 1,
	}
	created, err := service.CreateDropPool(ctx, createReq)
	require.NoError(t, err)
	defer cleanupDropPools(t, db, []string{"TEST_UPDATE_001"})

	t.Run("成功更新掉落池", func(t *testing.T) {
		updateReq := &dto.UpdateDropPoolRequest{
			PoolName:        stringPtr("更新后的名称"),
			MaxDrops:        int16Ptr(5),
			GuaranteedDrops: int16Ptr(2),
		}
		resp, err := service.UpdateDropPool(ctx, created.ID, updateReq)
		require.NoError(t, err)
		assert.Equal(t, "更新后的名称", resp.PoolName)
		assert.Equal(t, int16(5), resp.MaxDrops)
		assert.Equal(t, int16(2), resp.GuaranteedDrops)
	})

	t.Run("更新为无效范围应该失败", func(t *testing.T) {
		updateReq := &dto.UpdateDropPoolRequest{
			MinDrops: int16Ptr(10),
			MaxDrops: int16Ptr(5),
		}
		_, err := service.UpdateDropPool(ctx, created.ID, updateReq)
		require.Error(t, err)
	})
}

// 辅助函数
func generateUUID() string {
	return uuid.New().String()
}
