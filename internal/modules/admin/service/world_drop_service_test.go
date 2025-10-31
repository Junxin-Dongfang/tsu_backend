package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tsu-self/internal/modules/admin/dto"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/interfaces"
)

// cleanupWorldDrops 清理世界掉落测试数据
func cleanupWorldDrops(t *testing.T, db *sql.DB, itemIDs []string) {
	for _, id := range itemIDs {
		_, err := db.Exec("DELETE FROM game_config.world_drop_configs WHERE item_id = $1", id)
		if err != nil {
			t.Logf("Warning: failed to cleanup world drop for item %s: %v", id, err)
		}
	}
}

func TestWorldDropService_CreateWorldDrop(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewWorldDropService(db)
	ctx := context.Background()

	// 创建测试物品
	item := createTestItem(t, db, "TEST_WORLD_DROP_ITEM_001")
	defer cleanupItems(t, db, []string{"TEST_WORLD_DROP_ITEM_001"})
	defer cleanupWorldDrops(t, db, []string{item.ID})

	testCases := []struct {
		name        string
		request     *dto.CreateWorldDropRequest
		setup       func()
		expectError bool
		errorCode   xerrors.ErrorCode
		validate    func(t *testing.T, resp *dto.WorldDropResponse)
	}{
		{
			name: "成功创建世界掉落配置",
			request: &dto.CreateWorldDropRequest{
				ItemID:       item.ID,
				BaseDropRate: 0.05,
			},
			expectError: false,
			validate: func(t *testing.T, resp *dto.WorldDropResponse) {
				assert.Equal(t, item.ID, resp.ItemID)
				assert.Equal(t, item.ItemCode, resp.ItemCode)
				assert.InDelta(t, 0.05, resp.BaseDropRate, 0.001)
				assert.True(t, resp.IsActive)
			},
		},
		{
			name: "带完整配置的世界掉落",
			request: &dto.CreateWorldDropRequest{
				ItemID:            item.ID,
				TotalDropLimit:    intPtr(1000),
				DailyDropLimit:    intPtr(100),
				HourlyDropLimit:   intPtr(10),
				MinDropInterval:   intPtr(300),
				MaxDropInterval:   intPtr(600),
				BaseDropRate:      0.1,
				TriggerConditions: json.RawMessage(`{"type": "level_range", "min_level": 10, "max_level": 50}`),
				DropRateModifiers: json.RawMessage(`{"vip_bonus": 0.05, "event_bonus": 0.1}`),
			},
			setup: func() {
				// 清理之前的配置
				cleanupWorldDrops(t, db, []string{item.ID})
			},
			expectError: false,
			validate: func(t *testing.T, resp *dto.WorldDropResponse) {
				assert.NotNil(t, resp.TotalDropLimit)
				assert.Equal(t, 1000, *resp.TotalDropLimit)
				assert.NotNil(t, resp.DailyDropLimit)
				assert.Equal(t, 100, *resp.DailyDropLimit)
				assert.NotNil(t, resp.HourlyDropLimit)
				assert.Equal(t, 10, *resp.HourlyDropLimit)
				assert.NotNil(t, resp.MinDropInterval)
				assert.Equal(t, 300, *resp.MinDropInterval)
				assert.NotNil(t, resp.MaxDropInterval)
				assert.Equal(t, 600, *resp.MaxDropInterval)
				assert.InDelta(t, 0.1, resp.BaseDropRate, 0.001)

				// 验证JSON字段
				assert.NotNil(t, resp.TriggerConditions)
				var conditions map[string]interface{}
				err := json.Unmarshal(resp.TriggerConditions, &conditions)
				require.NoError(t, err)
				assert.Equal(t, "level_range", conditions["type"])

				assert.NotNil(t, resp.DropRateModifiers)
				var modifiers map[string]interface{}
				err = json.Unmarshal(resp.DropRateModifiers, &modifiers)
				require.NoError(t, err)
				assert.Equal(t, 0.05, modifiers["vip_bonus"])
			},
		},
		{
			name: "物品不存在应该失败",
			request: &dto.CreateWorldDropRequest{
				ItemID:       uuid.New().String(),
				BaseDropRate: 0.05,
			},
			expectError: true,
			errorCode:   xerrors.CodeResourceNotFound,
		},
		{
			name: "概率超出范围应该失败",
			request: &dto.CreateWorldDropRequest{
				ItemID:       item.ID,
				BaseDropRate: 1.5,
			},
			setup: func() {
				cleanupWorldDrops(t, db, []string{item.ID})
			},
			expectError: true,
			errorCode:   xerrors.CodeInvalidParams,
		},
		{
			name: "概率为0应该失败",
			request: &dto.CreateWorldDropRequest{
				ItemID:       item.ID,
				BaseDropRate: 0,
			},
			setup: func() {
				cleanupWorldDrops(t, db, []string{item.ID})
			},
			expectError: true,
			errorCode:   xerrors.CodeInvalidParams,
		},
		{
			name: "min_drop_interval大于max_drop_interval应该失败",
			request: &dto.CreateWorldDropRequest{
				ItemID:          item.ID,
				BaseDropRate:    0.05,
				MinDropInterval: intPtr(600),
				MaxDropInterval: intPtr(300),
			},
			setup: func() {
				cleanupWorldDrops(t, db, []string{item.ID})
			},
			expectError: true,
			errorCode:   xerrors.CodeInvalidParams,
		},
		{
			name: "无效的触发条件JSON应该失败",
			request: &dto.CreateWorldDropRequest{
				ItemID:            item.ID,
				BaseDropRate:      0.05,
				TriggerConditions: json.RawMessage(`{invalid json`),
			},
			setup: func() {
				cleanupWorldDrops(t, db, []string{item.ID})
			},
			expectError: true,
			errorCode:   xerrors.CodeInvalidParams,
		},
		{
			name: "无效的概率修正因子JSON应该失败",
			request: &dto.CreateWorldDropRequest{
				ItemID:            item.ID,
				BaseDropRate:      0.05,
				DropRateModifiers: json.RawMessage(`{invalid json`),
			},
			setup: func() {
				cleanupWorldDrops(t, db, []string{item.ID})
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

			resp, err := service.CreateWorldDrop(ctx, tc.request)

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

func TestWorldDropService_GetWorldDropList(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewWorldDropService(db)
	ctx := context.Background()

	// 创建多个测试物品和世界掉落配置
	itemIDs := make([]string, 3)
	for i := 0; i < 3; i++ {
		itemCode := "TEST_WORLD_LIST_" + string(rune('A'+i))
		item := createTestItem(t, db, itemCode)
		itemIDs[i] = item.ID

		req := &dto.CreateWorldDropRequest{
			ItemID:       item.ID,
			BaseDropRate: 0.01 * float64(i+1),
		}
		_, err := service.CreateWorldDrop(ctx, req)
		require.NoError(t, err)
	}
	defer func() {
		cleanupWorldDrops(t, db, itemIDs)
		for i := 0; i < 3; i++ {
			itemCode := "TEST_WORLD_LIST_" + string(rune('A'+i))
			cleanupItems(t, db, []string{itemCode})
		}
	}()

	t.Run("查询所有世界掉落配置", func(t *testing.T) {
		params := interfaces.ListWorldDropConfigParams{
			Page:     1,
			PageSize: 10,
		}
		resp, err := service.GetWorldDropList(ctx, params)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(resp.Items), 3)
		assert.GreaterOrEqual(t, resp.Total, int64(3))
	})

	t.Run("按物品ID筛选", func(t *testing.T) {
		params := interfaces.ListWorldDropConfigParams{
			ItemID:   &itemIDs[0],
			Page:     1,
			PageSize: 10,
		}
		resp, err := service.GetWorldDropList(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, 1, len(resp.Items))
		assert.Equal(t, itemIDs[0], resp.Items[0].ItemID)
	})

	t.Run("分页测试", func(t *testing.T) {
		params := interfaces.ListWorldDropConfigParams{
			Page:     1,
			PageSize: 2,
		}
		resp, err := service.GetWorldDropList(ctx, params)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(resp.Items), 2)
		assert.Equal(t, 1, resp.Page)
		assert.Equal(t, 2, resp.PageSize)
	})
}

func TestWorldDropService_UpdateWorldDrop(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewWorldDropService(db)
	ctx := context.Background()

	// 创建测试数据
	item := createTestItem(t, db, "TEST_WORLD_UPDATE")
	defer cleanupItems(t, db, []string{"TEST_WORLD_UPDATE"})
	defer cleanupWorldDrops(t, db, []string{item.ID})

	createReq := &dto.CreateWorldDropRequest{
		ItemID:       item.ID,
		BaseDropRate: 0.05,
	}
	created, err := service.CreateWorldDrop(ctx, createReq)
	require.NoError(t, err)

	t.Run("成功更新世界掉落配置", func(t *testing.T) {
		updateReq := &dto.UpdateWorldDropRequest{
			BaseDropRate:    float64Ptr(0.1),
			TotalDropLimit:  intPtr(500),
			DailyDropLimit:  intPtr(50),
			HourlyDropLimit: intPtr(5),
		}
		resp, err := service.UpdateWorldDrop(ctx, created.ID, updateReq)
		require.NoError(t, err)
		assert.InDelta(t, 0.1, resp.BaseDropRate, 0.001)
		assert.NotNil(t, resp.TotalDropLimit)
		assert.Equal(t, 500, *resp.TotalDropLimit)
		assert.NotNil(t, resp.DailyDropLimit)
		assert.Equal(t, 50, *resp.DailyDropLimit)
	})

	t.Run("更新为无效概率应该失败", func(t *testing.T) {
		updateReq := &dto.UpdateWorldDropRequest{
			BaseDropRate: float64Ptr(1.5),
		}
		_, err := service.UpdateWorldDrop(ctx, created.ID, updateReq)
		require.Error(t, err)
	})

	t.Run("更新为无效间隔范围应该失败", func(t *testing.T) {
		updateReq := &dto.UpdateWorldDropRequest{
			MinDropInterval: intPtr(600),
			MaxDropInterval: intPtr(300),
		}
		_, err := service.UpdateWorldDrop(ctx, created.ID, updateReq)
		require.Error(t, err)
	})
}

func TestWorldDropService_DeleteWorldDrop(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewWorldDropService(db)
	ctx := context.Background()

	// 创建测试数据
	item := createTestItem(t, db, "TEST_WORLD_DELETE")
	defer cleanupItems(t, db, []string{"TEST_WORLD_DELETE"})

	createReq := &dto.CreateWorldDropRequest{
		ItemID:       item.ID,
		BaseDropRate: 0.05,
	}
	created, err := service.CreateWorldDrop(ctx, createReq)
	require.NoError(t, err)

	t.Run("成功删除世界掉落配置", func(t *testing.T) {
		err := service.DeleteWorldDrop(ctx, created.ID)
		require.NoError(t, err)

		// 验证已删除
		_, err = service.GetWorldDropByID(ctx, created.ID)
		require.Error(t, err)
	})

	t.Run("删除不存在的配置应该失败", func(t *testing.T) {
		err := service.DeleteWorldDrop(ctx, uuid.New().String())
		require.Error(t, err)
	})
}

// 辅助函数
func intPtr(i int) *int {
	return &i
}

