// +build integration

package service

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tsu-self/internal/modules/admin/dto"
	"tsu-self/internal/repository/interfaces"
)

// 集成测试需要真实的数据库环境
// 运行方式: go test -tags=integration ./internal/modules/admin/service/...

var testDB *sql.DB

func TestMain(m *testing.M) {
	// 设置测试数据库
	var err error
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		// 默认使用本地开发数据库
		dsn = "host=localhost port=5432 user=postgres password=postgres dbname=tsu_db sslmode=disable"
	}

	testDB, err = sql.Open("postgres", dsn)
	if err != nil {
		fmt.Printf("无法连接测试数据库: %v\n", err)
		os.Exit(1)
	}

	// 测试连接
	if err := testDB.Ping(); err != nil {
		fmt.Printf("无法ping测试数据库: %v\n", err)
		os.Exit(1)
	}

	// 运行测试
	code := m.Run()

	// 清理
	testDB.Close()
	os.Exit(code)
}

func setupTestService(t *testing.T) (*ItemConfigService, func()) {
	service := NewItemConfigService(testDB)

	cleanup := func() {
		// 清理测试数据
		ctx := context.Background()
		testDB.ExecContext(ctx, `DELETE FROM game_config.tags_relations WHERE entity_type = 'item'`)
		testDB.ExecContext(ctx, `DELETE FROM game_config.items WHERE item_code LIKE 'test_%'`)
		testDB.ExecContext(ctx, `DELETE FROM game_config.tags WHERE tag_code LIKE 'test_%'`)
	}

	return service, cleanup
}

func createTestTag(t *testing.T, ctx context.Context) string {
	tagCode := fmt.Sprintf("test_tag_%d", time.Now().UnixNano())
	var tagID string
	err := testDB.QueryRowContext(ctx, `
		INSERT INTO game_config.tags (tag_code, tag_name, category, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, tagCode, "测试标签", "item", true, time.Now(), time.Now()).Scan(&tagID)
	require.NoError(t, err)
	return tagID
}

func TestItemConfigService_CreateItem_Integration(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("成功创建装备物品", func(t *testing.T) {
		equipSlot := "head"
		req := &dto.CreateItemRequest{
			ItemCode:    fmt.Sprintf("test_sword_%d", time.Now().UnixNano()),
			ItemName:    "测试剑",
			ItemType:    "equipment",
			ItemQuality: "fine",
			ItemLevel:   10,
			Description: "一把测试用的剑",
			EquipSlot:   &equipSlot,
		}

		item, err := service.CreateItem(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, item)
		assert.Equal(t, req.ItemCode, item.ItemCode)
		assert.Equal(t, req.ItemName, item.ItemName)
		assert.Equal(t, req.ItemType, item.ItemType)
		assert.NotEmpty(t, item.ID)
	})

	t.Run("成功创建消耗品", func(t *testing.T) {
		req := &dto.CreateItemRequest{
			ItemCode:    fmt.Sprintf("test_potion_%d", time.Now().UnixNano()),
			ItemName:    "测试药水",
			ItemType:    "consumable",
			ItemQuality: "normal",
			ItemLevel:   1,
		}

		item, err := service.CreateItem(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, item)
		assert.Nil(t, item.EquipSlot)
	})

	t.Run("装备类型缺少槽位-应该失败", func(t *testing.T) {
		req := &dto.CreateItemRequest{
			ItemCode:    fmt.Sprintf("test_invalid_%d", time.Now().UnixNano()),
			ItemName:    "无效装备",
			ItemType:    "equipment",
			ItemQuality: "normal",
			ItemLevel:   1,
			// 缺少EquipSlot
		}

		_, err := service.CreateItem(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "装备类型物品必须设置装备槽位")
	})

	t.Run("重复的item_code-应该失败", func(t *testing.T) {
		itemCode := fmt.Sprintf("test_duplicate_%d", time.Now().UnixNano())
		equipSlot := "chest"

		req1 := &dto.CreateItemRequest{
			ItemCode:    itemCode,
			ItemName:    "第一个物品",
			ItemType:    "equipment",
			ItemQuality: "normal",
			ItemLevel:   1,
			EquipSlot:   &equipSlot,
		}

		_, err := service.CreateItem(ctx, req1)
		require.NoError(t, err)

		req2 := &dto.CreateItemRequest{
			ItemCode:    itemCode,
			ItemName:    "第二个物品",
			ItemType:    "equipment",
			ItemQuality: "normal",
			ItemLevel:   1,
			EquipSlot:   &equipSlot,
		}

		_, err = service.CreateItem(ctx, req2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "物品代码已存在")
	})

	t.Run("创建物品并关联标签", func(t *testing.T) {
		// 创建测试标签
		tagID := createTestTag(t, ctx)

		equipSlot := "mainhand"
		req := &dto.CreateItemRequest{
			ItemCode:    fmt.Sprintf("test_with_tag_%d", time.Now().UnixNano()),
			ItemName:    "带标签的物品",
			ItemType:    "equipment",
			ItemQuality: "epic",
			ItemLevel:   20,
			EquipSlot:   &equipSlot,
			TagIDs:      []string{tagID},
		}

		item, err := service.CreateItem(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, item)
		assert.Len(t, item.Tags, 1)
		assert.Equal(t, tagID, item.Tags[0].ID)
	})
}

func TestItemConfigService_GetItems_Integration(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// 创建测试数据
	equipSlot1 := "head"
	equipSlot2 := "chest"

	item1Code := fmt.Sprintf("test_list_1_%d", time.Now().UnixNano())
	item2Code := fmt.Sprintf("test_list_2_%d", time.Now().UnixNano())

	_, err := service.CreateItem(ctx, &dto.CreateItemRequest{
		ItemCode:    item1Code,
		ItemName:    "测试物品1",
		ItemType:    "equipment",
		ItemQuality: "fine",
		ItemLevel:   10,
		EquipSlot:   &equipSlot1,
	})
	require.NoError(t, err)

	_, err = service.CreateItem(ctx, &dto.CreateItemRequest{
		ItemCode:    item2Code,
		ItemName:    "测试物品2",
		ItemType:    "equipment",
		ItemQuality: "epic",
		ItemLevel:   20,
		EquipSlot:   &equipSlot2,
	})
	require.NoError(t, err)

	t.Run("查询所有测试物品", func(t *testing.T) {
		keyword := "test_list_"
		params := interfaces.ListItemParams{
			Page:     1,
			PageSize: 10,
			Keyword:  &keyword,
		}

		items, total, err := service.GetItems(ctx, params)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(items), 2)
		assert.GreaterOrEqual(t, total, int64(2))
	})

	t.Run("按品质筛选", func(t *testing.T) {
		quality := "fine"
		keyword := "test_list_"
		params := interfaces.ListItemParams{
			Page:        1,
			PageSize:    10,
			ItemQuality: &quality,
			Keyword:     &keyword,
		}

		items, _, err := service.GetItems(ctx, params)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(items), 1)
		for _, item := range items {
			assert.Equal(t, "fine", item.ItemQuality)
		}
	})
}

func TestItemConfigService_UpdateItem_Integration(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// 创建测试物品
	equipSlot := "legs"
	itemCode := fmt.Sprintf("test_update_%d", time.Now().UnixNano())
	item, err := service.CreateItem(ctx, &dto.CreateItemRequest{
		ItemCode:    itemCode,
		ItemName:    "原始名称",
		ItemType:    "equipment",
		ItemQuality: "normal",
		ItemLevel:   5,
		EquipSlot:   &equipSlot,
	})
	require.NoError(t, err)

	t.Run("更新物品名称", func(t *testing.T) {
		newName := "更新后的名称"
		req := &dto.UpdateItemRequest{
			ItemName: &newName,
		}

		updated, err := service.UpdateItem(ctx, item.ID, req)
		require.NoError(t, err)
		assert.Equal(t, newName, updated.ItemName)
		assert.Equal(t, item.ItemCode, updated.ItemCode) // 其他字段不变
	})

	t.Run("更新物品等级", func(t *testing.T) {
		newLevel := int16(15)
		req := &dto.UpdateItemRequest{
			ItemLevel: &newLevel,
		}

		updated, err := service.UpdateItem(ctx, item.ID, req)
		require.NoError(t, err)
		assert.Equal(t, newLevel, updated.ItemLevel)
	})
}

func TestItemConfigService_DeleteItem_Integration(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// 创建测试物品
	equipSlot := "feet"
	itemCode := fmt.Sprintf("test_delete_%d", time.Now().UnixNano())
	item, err := service.CreateItem(ctx, &dto.CreateItemRequest{
		ItemCode:    itemCode,
		ItemName:    "待删除物品",
		ItemType:    "equipment",
		ItemQuality: "normal",
		ItemLevel:   1,
		EquipSlot:   &equipSlot,
	})
	require.NoError(t, err)

	t.Run("删除物品", func(t *testing.T) {
		err := service.DeleteItem(ctx, item.ID)
		require.NoError(t, err)

		// 验证物品已被软删除
		_, err = service.GetItemByID(ctx, item.ID)
		assert.Error(t, err)
	})
}

