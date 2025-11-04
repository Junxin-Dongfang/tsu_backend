package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tsu-self/internal/modules/admin/dto"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/interfaces"
)

func TestDropPoolService_AddDropPoolItem(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewDropPoolService(db)
	ctx := context.Background()

	// 创建测试掉落池
	poolReq := &dto.CreateDropPoolRequest{
		PoolCode:        "TEST_ITEM_POOL_001",
		PoolName:        "物品测试池",
		PoolType:        "monster",
		MinDrops:        1,
		MaxDrops:        3,
		GuaranteedDrops: 1,
	}
	pool, err := service.CreateDropPool(ctx, poolReq)
	require.NoError(t, err)
	defer cleanupDropPools(t, db, []string{"TEST_ITEM_POOL_001"})

	// 创建测试物品
	item := createTestItem(t, db, "TEST_ITEM_001")
	defer cleanupItems(t, db, []string{"TEST_ITEM_001"})

	testCases := []struct {
		name        string
		request     *dto.AddDropPoolItemRequest
		setup       func()
		expectError bool
		errorCode   xerrors.ErrorCode
		validate    func(t *testing.T, resp *dto.DropPoolItemResponse)
	}{
		{
			name: "成功添加掉落物品",
			request: &dto.AddDropPoolItemRequest{
				ItemID:      item.ID,
				MinQuantity: 1,
				MaxQuantity: 3,
				DropWeight:  int64Ptr(100),
			},
			expectError: false,
			validate: func(t *testing.T, resp *dto.DropPoolItemResponse) {
				assert.Equal(t, pool.ID, resp.DropPoolID)
				assert.Equal(t, item.ID, resp.ItemID)
				assert.Equal(t, item.ItemCode, resp.ItemCode)
				assert.Equal(t, int16(1), resp.MinQuantity)
				assert.Equal(t, int16(3), resp.MaxQuantity)
				assert.NotNil(t, resp.DropWeight)
				assert.Equal(t, 100, *resp.DropWeight)
			},
		},
		{
			name: "添加带固定概率的物品",
			request: &dto.AddDropPoolItemRequest{
				ItemID:      item.ID,
				MinQuantity: 1,
				MaxQuantity: 1,
				DropRate:    float64Ptr(0.15),
			},
			setup: func() {
				// 先清理之前添加的物品
				_, _ = db.Exec("DELETE FROM game_config.drop_pool_items WHERE drop_pool_id = $1 AND item_id = $2", pool.ID, item.ID)
			},
			expectError: false,
			validate: func(t *testing.T, resp *dto.DropPoolItemResponse) {
				assert.NotNil(t, resp.DropRate)
				assert.InDelta(t, 0.15, *resp.DropRate, 0.001)
			},
		},
		{
			name: "添加带品质权重的物品",
			request: &dto.AddDropPoolItemRequest{
				ItemID:         item.ID,
				MinQuantity:    1,
				MaxQuantity:    5,
				DropWeight:     int64Ptr(50),
				QualityWeights: json.RawMessage(`{"common": 50, "rare": 30, "epic": 15, "legendary": 5}`),
			},
			setup: func() {
				_, _ = db.Exec("DELETE FROM game_config.drop_pool_items WHERE drop_pool_id = $1 AND item_id = $2", pool.ID, item.ID)
			},
			expectError: false,
			validate: func(t *testing.T, resp *dto.DropPoolItemResponse) {
				assert.NotNil(t, resp.QualityWeights)
				var weights map[string]int
				err := json.Unmarshal(resp.QualityWeights, &weights)
				require.NoError(t, err)
				assert.Equal(t, 50, weights["common"])
				assert.Equal(t, 5, weights["legendary"])
			},
		},
		{
			name: "添加带等级限制的物品",
			request: &dto.AddDropPoolItemRequest{
				ItemID:      item.ID,
				MinQuantity: 1,
				MaxQuantity: 2,
				DropWeight:  int64Ptr(80),
				MinLevel:    int16Ptr(10),
				MaxLevel:    int16Ptr(20),
			},
			setup: func() {
				_, _ = db.Exec("DELETE FROM game_config.drop_pool_items WHERE drop_pool_id = $1 AND item_id = $2", pool.ID, item.ID)
			},
			expectError: false,
			validate: func(t *testing.T, resp *dto.DropPoolItemResponse) {
				assert.NotNil(t, resp.MinLevel)
				assert.NotNil(t, resp.MaxLevel)
				assert.Equal(t, int16(10), *resp.MinLevel)
				assert.Equal(t, int16(20), *resp.MaxLevel)
			},
		},
		{
			name: "物品不存在应该失败",
			request: &dto.AddDropPoolItemRequest{
				ItemID:      uuid.New().String(),
				MinQuantity: 1,
				MaxQuantity: 1,
			},
			expectError: true,
			errorCode:   xerrors.CodeResourceNotFound,
		},
		{
			name: "min_quantity大于max_quantity应该失败",
			request: &dto.AddDropPoolItemRequest{
				ItemID:      item.ID,
				MinQuantity: 5,
				MaxQuantity: 3,
			},
			setup: func() {
				_, _ = db.Exec("DELETE FROM game_config.drop_pool_items WHERE drop_pool_id = $1 AND item_id = $2", pool.ID, item.ID)
			},
			expectError: true,
			errorCode:   xerrors.CodeInvalidParams,
		},
		{
			name: "min_level大于max_level应该失败",
			request: &dto.AddDropPoolItemRequest{
				ItemID:      item.ID,
				MinQuantity: 1,
				MaxQuantity: 1,
				MinLevel:    int16Ptr(20),
				MaxLevel:    int16Ptr(10),
			},
			setup: func() {
				_, _ = db.Exec("DELETE FROM game_config.drop_pool_items WHERE drop_pool_id = $1 AND item_id = $2", pool.ID, item.ID)
			},
			expectError: true,
			errorCode:   xerrors.CodeInvalidParams,
		},
		{
			name: "无效的品质权重JSON应该失败",
			request: &dto.AddDropPoolItemRequest{
				ItemID:         item.ID,
				MinQuantity:    1,
				MaxQuantity:    1,
				QualityWeights: json.RawMessage(`{invalid json`),
			},
			setup: func() {
				_, _ = db.Exec("DELETE FROM game_config.drop_pool_items WHERE drop_pool_id = $1 AND item_id = $2", pool.ID, item.ID)
			},
			expectError: true,
			errorCode:   xerrors.CodeInvalidParams,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup()
			}

			resp, err := service.AddDropPoolItem(ctx, pool.ID, tc.request)

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

func TestDropPoolService_GetDropPoolItems(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewDropPoolService(db)
	ctx := context.Background()

	// 创建测试掉落池
	poolReq := &dto.CreateDropPoolRequest{
		PoolCode:        "TEST_ITEM_LIST_POOL",
		PoolName:        "物品列表测试池",
		PoolType:        "dungeon",
		MinDrops:        1,
		MaxDrops:        5,
		GuaranteedDrops: 1,
	}
	pool, err := service.CreateDropPool(ctx, poolReq)
	require.NoError(t, err)
	defer cleanupDropPools(t, db, []string{"TEST_ITEM_LIST_POOL"})

	// 创建多个测试物品
	items := make([]string, 3)
	for i := 0; i < 3; i++ {
		itemCode := "TEST_LIST_ITEM_" + string(rune('A'+i))
		item := createTestItem(t, db, itemCode)
		items[i] = itemCode

		// 添加到掉落池
		addReq := &dto.AddDropPoolItemRequest{
			ItemID:      item.ID,
			MinQuantity: 1,
			MaxQuantity: int16(i + 1),
			DropWeight:  int64Ptr(100 - i*10),
			MinLevel:    int16Ptr(int16(i * 10)),
			MaxLevel:    int16Ptr(int16((i + 1) * 10)),
		}
		_, err := service.AddDropPoolItem(ctx, pool.ID, addReq)
		require.NoError(t, err)
	}
	defer cleanupItems(t, db, items)

	t.Run("查询所有掉落池物品", func(t *testing.T) {
		params := interfaces.ListDropPoolItemParams{
			DropPoolID: pool.ID,
			Page:       1,
			PageSize:   10,
		}
		resp, err := service.GetDropPoolItems(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, 3, len(resp.Items))
		assert.Equal(t, int64(3), resp.Total)
	})

	t.Run("按等级范围筛选", func(t *testing.T) {
		minLevel := int16(5)
		maxLevel := int16(15)
		params := interfaces.ListDropPoolItemParams{
			DropPoolID: pool.ID,
			MinLevel:   &minLevel,
			MaxLevel:   &maxLevel,
			Page:       1,
			PageSize:   10,
		}
		resp, err := service.GetDropPoolItems(ctx, params)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(resp.Items), 1)
	})

	t.Run("分页测试", func(t *testing.T) {
		params := interfaces.ListDropPoolItemParams{
			DropPoolID: pool.ID,
			Page:       1,
			PageSize:   2,
		}
		resp, err := service.GetDropPoolItems(ctx, params)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(resp.Items), 2)
	})
}

func TestDropPoolService_UpdateDropPoolItem(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewDropPoolService(db)
	ctx := context.Background()

	// 创建测试数据
	poolReq := &dto.CreateDropPoolRequest{
		PoolCode:        "TEST_UPDATE_ITEM_POOL",
		PoolName:        "更新物品测试池",
		PoolType:        "boss",
		MinDrops:        1,
		MaxDrops:        3,
		GuaranteedDrops: 1,
	}
	pool, err := service.CreateDropPool(ctx, poolReq)
	require.NoError(t, err)
	defer cleanupDropPools(t, db, []string{"TEST_UPDATE_ITEM_POOL"})

	item := createTestItem(t, db, "TEST_UPDATE_ITEM")
	defer cleanupItems(t, db, []string{"TEST_UPDATE_ITEM"})

	addReq := &dto.AddDropPoolItemRequest{
		ItemID:      item.ID,
		MinQuantity: 1,
		MaxQuantity: 3,
		DropWeight:  int64Ptr(100),
	}
	_, err = service.AddDropPoolItem(ctx, pool.ID, addReq)
	require.NoError(t, err)

	t.Run("成功更新掉落物品", func(t *testing.T) {
		updateReq := &dto.UpdateDropPoolItemRequest{
			DropWeight:  int64Ptr(200),
			MinQuantity: int16Ptr(2),
			MaxQuantity: int16Ptr(5),
		}
		resp, err := service.UpdateDropPoolItem(ctx, pool.ID, item.ID, updateReq)
		require.NoError(t, err)
		assert.Equal(t, 200, *resp.DropWeight)
		assert.Equal(t, int16(2), resp.MinQuantity)
		assert.Equal(t, int16(5), resp.MaxQuantity)
	})
}

// 辅助函数
func int64Ptr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}
