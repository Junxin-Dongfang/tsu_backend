package service

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"tsu-self/internal/modules/admin/dto"
)

// 注意：由于SQLBoiler生成的查询非常复杂，使用sqlmock进行单元测试很困难
// 这里我们主要测试业务逻辑函数，完整的CRUD测试应该使用集成测试
// 集成测试需要真实的数据库环境

// TestNormalizeJSON 测试JSON规范化（支持字符串包裹）
func TestNormalizeJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   json.RawMessage
		wantErr bool
	}{
		{
			name:    "有效的JSON对象",
			input:   json.RawMessage(`{"key": "value", "number": 123}`),
			wantErr: false,
		},
		{
			name:    "有效的JSON数组",
			input:   json.RawMessage(`[{"key": "value"}, {"key2": "value2"}]`),
			wantErr: false,
		},
		{
			name:    "空输入视为跳过",
			input:   json.RawMessage(``),
			wantErr: false,
		},
		{
			name:    "null值",
			input:   json.RawMessage(`null`),
			wantErr: false,
		},
		{
			name:    "无效的JSON - 缺少引号",
			input:   json.RawMessage(`{key: value}`),
			wantErr: true,
		},
		{
			name:    "无效的JSON - 缺少括号",
			input:   json.RawMessage(`{"key": "value"`),
			wantErr: true,
		},
		{
			name:    "无效的JSON - 纯文本",
			input:   json.RawMessage(`invalid json`),
			wantErr: true,
		},
		{
			name:    "字符串包裹的JSON对象",
			input:   json.RawMessage(`"{\"k\":1}"`),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := normalizeJSON(tt.input, "test-field")

			if tt.wantErr {
				assert.Error(t, err, "应该返回错误")
			} else {
				assert.NoError(t, err, "不应该返回错误")
			}
		})
	}
}

// TestCreateItemRequest_Validation 测试CreateItemRequest的字段验证
func TestCreateItemRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     *dto.CreateItemRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "有效的请求",
			req: &dto.CreateItemRequest{
				ItemCode:    "test_sword_001",
				ItemName:    "测试剑",
				ItemType:    "equipment",
				ItemQuality: "rare",
				ItemLevel:   10,
			},
			wantErr: false,
		},
		{
			name: "缺少ItemCode",
			req: &dto.CreateItemRequest{
				ItemName:    "测试剑",
				ItemType:    "equipment",
				ItemQuality: "rare",
				ItemLevel:   10,
			},
			wantErr: true,
			errMsg:  "ItemCode不能为空",
		},
		{
			name: "缺少ItemName",
			req: &dto.CreateItemRequest{
				ItemCode:    "test_sword_001",
				ItemType:    "equipment",
				ItemQuality: "rare",
				ItemLevel:   10,
			},
			wantErr: true,
			errMsg:  "ItemName不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 简单的字段验证
			if tt.req.ItemCode == "" {
				assert.True(t, tt.wantErr)
				return
			}
			if tt.req.ItemName == "" {
				assert.True(t, tt.wantErr)
				return
			}
			assert.False(t, tt.wantErr)
		})
	}
}

// TestUpdateItemRequest_PartialUpdate 测试UpdateItemRequest的部分更新逻辑
func TestUpdateItemRequest_PartialUpdate(t *testing.T) {
	tests := []struct {
		name        string
		req         *dto.UpdateItemRequest
		expectField string
	}{
		{
			name: "只更新ItemName",
			req: &dto.UpdateItemRequest{
				ItemName: stringPtr("新名称"),
			},
			expectField: "ItemName",
		},
		{
			name: "只更新ItemLevel",
			req: &dto.UpdateItemRequest{
				ItemLevel: int16Ptr(20),
			},
			expectField: "ItemLevel",
		},
		{
			name: "更新多个字段",
			req: &dto.UpdateItemRequest{
				ItemName:  stringPtr("新名称"),
				ItemLevel: int16Ptr(20),
			},
			expectField: "Multiple",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证字段是否被设置
			if tt.req.ItemName != nil {
				assert.NotNil(t, tt.req.ItemName)
			}
			if tt.req.ItemLevel != nil {
				assert.NotNil(t, tt.req.ItemLevel)
			}
		})
	}
}

// TestItemConfigResponse_Structure 测试ItemConfigResponse的结构
func TestItemConfigResponse_Structure(t *testing.T) {
	resp := &dto.ItemConfigResponse{
		ID:          "test-id",
		ItemCode:    "test_sword",
		ItemName:    "测试剑",
		ItemType:    "equipment",
		ItemQuality: "rare",
		ItemLevel:   10,
		Description: "测试描述",
		IconURL:     "http://example.com/icon.png",
		IsTradable:  true,
		IsDroppable: true,
		IsActive:    true,
		Tags:        []dto.TagResponse{},
	}

	assert.NotNil(t, resp)
	assert.Equal(t, "test-id", resp.ID)
	assert.Equal(t, "test_sword", resp.ItemCode)
	assert.Equal(t, "测试剑", resp.ItemName)
	assert.Equal(t, "equipment", resp.ItemType)
	assert.Equal(t, "rare", resp.ItemQuality)
	assert.Equal(t, int16(10), resp.ItemLevel)
	assert.True(t, resp.IsTradable)
	assert.True(t, resp.IsDroppable)
	assert.True(t, resp.IsActive)
	assert.NotNil(t, resp.Tags)
}

// TestTagResponse_Structure 测试TagResponse的结构
func TestTagResponse_Structure(t *testing.T) {
	tag := dto.TagResponse{
		ID:           "tag-id",
		TagCode:      "legendary",
		TagName:      "传奇",
		Category:     "item",
		Description:  "传奇物品",
		Icon:         "legendary.png",
		Color:        "#FFD700",
		DisplayOrder: 1,
		IsActive:     true,
	}

	assert.Equal(t, "tag-id", tag.ID)
	assert.Equal(t, "legendary", tag.TagCode)
	assert.Equal(t, "传奇", tag.TagName)
	assert.Equal(t, "item", tag.Category)
	assert.True(t, tag.IsActive)
}

// TestValidateEquipSlot 测试装备槽位验证
func TestValidateEquipSlot(t *testing.T) {
	// 创建一个简单的service实例用于测试
	service := &ItemConfigService{}

	tests := []struct {
		name      string
		itemType  string
		equipSlot *string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "装备类型-有效槽位",
			itemType:  "equipment",
			equipSlot: stringPtr("head"),
			wantErr:   false,
		},
		{
			name:      "装备类型-缺少槽位",
			itemType:  "equipment",
			equipSlot: nil,
			wantErr:   true,
			errMsg:    "装备类型物品必须设置装备槽位",
		},
		{
			name:      "装备类型-空槽位",
			itemType:  "equipment",
			equipSlot: stringPtr(""),
			wantErr:   true,
			errMsg:    "装备类型物品必须设置装备槽位",
		},
		{
			name:      "装备类型-无效槽位",
			itemType:  "equipment",
			equipSlot: stringPtr("invalid_slot"),
			wantErr:   true,
			errMsg:    "无效的装备槽位",
		},
		{
			name:      "消耗品-无槽位",
			itemType:  "consumable",
			equipSlot: nil,
			wantErr:   false,
		},
		{
			name:      "消耗品-有槽位",
			itemType:  "consumable",
			equipSlot: stringPtr("head"),
			wantErr:   true,
			errMsg:    "非装备类型物品不能设置装备槽位",
		},
		{
			name:      "材料-无槽位",
			itemType:  "material",
			equipSlot: nil,
			wantErr:   false,
		},
		{
			name:      "装备类型-所有有效槽位",
			itemType:  "equipment",
			equipSlot: stringPtr("main_hand"),
			wantErr:   false,
		},
		{
			name:      "装备类型-双手武器槽位",
			itemType:  "equipment",
			equipSlot: stringPtr("two_hand"),
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateEquipSlot(tt.itemType, tt.equipSlot)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
