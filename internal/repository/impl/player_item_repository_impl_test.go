package impl

import (
	"context"
	"testing"

	"github.com/aarondl/null/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/repository/interfaces"
	"tsu-self/internal/test"
)

func TestPlayerItemRepository_Create(t *testing.T) {
	// 设置测试数据库
	db := test.SetupTestDB(t)
	defer test.TeardownTestDB(t, db)

	repo := NewPlayerItemRepository(db)
	ctx := context.Background()

	tests := []struct {
		name    string
		item    *game_runtime.PlayerItem
		wantErr bool
	}{
		{
			name: "创建装备物品",
			item: &game_runtime.PlayerItem{
				ItemID:            "test-item-001",
				OwnerID:           "test-user-001",
				SourceType:        "dungeon",
				SourceID:          null.StringFrom("dungeon-001"),
				ItemLocation:      "backpack",
				CurrentDurability: null.IntFrom(100),
				EnhancementLevel:  null.Int16From(0),
				StackCount:        null.IntFrom(1),
			},
			wantErr: false,
		},
		{
			name: "创建消耗品",
			item: &game_runtime.PlayerItem{
				ItemID:       "test-item-002",
				OwnerID:      "test-user-001",
				SourceType:   "shop",
				ItemLocation: "backpack",
				StackCount:   null.IntFrom(99),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(ctx, db, tt.item)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tt.item.ID)
				assert.NotZero(t, tt.item.CreatedAt)
			}
		})
	}
}

func TestPlayerItemRepository_GetByID(t *testing.T) {
	db := test.SetupTestDB(t)
	defer test.TeardownTestDB(t, db)

	repo := NewPlayerItemRepository(db)
	ctx := context.Background()

	// 创建测试数据
	item := &game_runtime.PlayerItem{
		ItemID:       "test-item-001",
		OwnerID:      "test-user-001",
		SourceType:   "dungeon",
		ItemLocation: "backpack",
		StackCount:   null.IntFrom(1),
	}
	err := repo.Create(ctx, db, item)
	require.NoError(t, err)

	tests := []struct {
		name    string
		itemID  string
		wantErr bool
	}{
		{
			name:    "查询存在的物品",
			itemID:  item.ID,
			wantErr: false,
		},
		{
			name:    "查询不存在的物品",
			itemID:  "non-existent-id",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByID(ctx, tt.itemID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.itemID, got.ID)
			}
		})
	}
}

func TestPlayerItemRepository_UpdateLocation(t *testing.T) {
	db := test.SetupTestDB(t)
	defer test.TeardownTestDB(t, db)

	repo := NewPlayerItemRepository(db)
	ctx := context.Background()

	// 创建测试数据
	item := &game_runtime.PlayerItem{
		ItemID:       "test-item-001",
		OwnerID:      "test-user-001",
		SourceType:   "dungeon",
		ItemLocation: "backpack",
		StackCount:   null.IntFrom(1),
	}
	err := repo.Create(ctx, db, item)
	require.NoError(t, err)

	tests := []struct {
		name        string
		itemID      string
		newLocation string
		wantErr     bool
	}{
		{
			name:        "移动到仓库",
			itemID:      item.ID,
			newLocation: "warehouse",
			wantErr:     false,
		},
		{
			name:        "移动到装备栏",
			itemID:      item.ID,
			newLocation: "equipped",
			wantErr:     false,
		},
		{
			name:        "移动不存在的物品",
			itemID:      "non-existent-id",
			newLocation: "warehouse",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.UpdateLocation(ctx, db, tt.itemID, tt.newLocation)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// 验证位置已更新
				updated, err := repo.GetByID(ctx, tt.itemID)
				require.NoError(t, err)
				assert.Equal(t, tt.newLocation, updated.ItemLocation)
			}
		})
	}
}

func TestPlayerItemRepository_UpdateDurability(t *testing.T) {
	db := test.SetupTestDB(t)
	defer test.TeardownTestDB(t, db)

	repo := NewPlayerItemRepository(db)
	ctx := context.Background()

	// 创建测试数据
	item := &game_runtime.PlayerItem{
		ItemID:            "test-item-001",
		OwnerID:           "test-user-001",
		SourceType:        "dungeon",
		ItemLocation:      "backpack",
		CurrentDurability: null.IntFrom(100),
		StackCount:        null.IntFrom(1),
	}
	err := repo.Create(ctx, db, item)
	require.NoError(t, err)

	tests := []struct {
		name           string
		itemID         string
		newDurability  int
		wantErr        bool
		expectedResult int
	}{
		{
			name:           "减少耐久度",
			itemID:         item.ID,
			newDurability:  80,
			wantErr:        false,
			expectedResult: 80,
		},
		{
			name:           "耐久度归零",
			itemID:         item.ID,
			newDurability:  0,
			wantErr:        false,
			expectedResult: 0,
		},
		{
			name:          "更新不存在的物品",
			itemID:        "non-existent-id",
			newDurability: 50,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.UpdateDurability(ctx, db, tt.itemID, tt.newDurability)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// 验证耐久度已更新
				updated, err := repo.GetByID(ctx, tt.itemID)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, updated.CurrentDurability.Int)
			}
		})
	}
}

func TestPlayerItemRepository_Delete(t *testing.T) {
	db := test.SetupTestDB(t)
	defer test.TeardownTestDB(t, db)

	repo := NewPlayerItemRepository(db)
	ctx := context.Background()

	// 创建测试数据
	item := &game_runtime.PlayerItem{
		ItemID:       "test-item-001",
		OwnerID:      "test-user-001",
		SourceType:   "dungeon",
		ItemLocation: "backpack",
		StackCount:   null.IntFrom(1),
	}
	err := repo.Create(ctx, db, item)
	require.NoError(t, err)

	// 删除物品
	err = repo.Delete(ctx, item.ID)
	assert.NoError(t, err)

	// 验证物品已被软删除
	deleted, err := repo.GetByID(ctx, item.ID)
	assert.Error(t, err)
	assert.Nil(t, deleted)
}

func TestPlayerItemRepository_GetByOwnerPaginated(t *testing.T) {
	db := test.SetupTestDB(t)
	defer test.TeardownTestDB(t, db)

	repo := NewPlayerItemRepository(db)
	ctx := context.Background()

	// 创建测试数据
	ownerID := "test-user-001"
	location := "backpack"

	for i := 0; i < 25; i++ {
		item := &game_runtime.PlayerItem{
			ItemID:       "test-item-001",
			OwnerID:      ownerID,
			SourceType:   "dungeon",
			ItemLocation: location,
			StackCount:   null.IntFrom(1),
		}
		err := repo.Create(ctx, db, item)
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		params        interfaces.GetPlayerItemsParams
		expectedCount int
		expectedTotal int64
	}{
		{
			name: "第一页_20条",
			params: interfaces.GetPlayerItemsParams{
				OwnerID:      ownerID,
				ItemLocation: &location,
				Page:         1,
				PageSize:     20,
			},
			expectedCount: 20,
			expectedTotal: 25,
		},
		{
			name: "第二页_5条",
			params: interfaces.GetPlayerItemsParams{
				OwnerID:      ownerID,
				ItemLocation: &location,
				Page:         2,
				PageSize:     20,
			},
			expectedCount: 5,
			expectedTotal: 25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, total, err := repo.GetByOwnerPaginated(ctx, tt.params)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCount, len(items))
			assert.Equal(t, tt.expectedTotal, total)
		})
	}
}
