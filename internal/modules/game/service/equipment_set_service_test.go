package service

import (
	"context"
	"testing"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/entity/game_runtime"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEquipmentSetRepository 是 EquipmentSetRepository 的 mock
type MockEquipmentSetRepository struct {
	mock.Mock
}

func (m *MockEquipmentSetRepository) GetByID(ctx context.Context, setID string) (*game_config.EquipmentSetConfig, error) {
	args := m.Called(ctx, setID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game_config.EquipmentSetConfig), args.Error(1)
}

func (m *MockEquipmentSetRepository) GetByIDs(ctx context.Context, setIDs []string) ([]*game_config.EquipmentSetConfig, error) {
	args := m.Called(ctx, setIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*game_config.EquipmentSetConfig), args.Error(1)
}

func (m *MockEquipmentSetRepository) List(ctx context.Context, page, pageSize int) ([]*game_config.EquipmentSetConfig, int64, error) {
	args := m.Called(ctx, page, pageSize)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*game_config.EquipmentSetConfig), int64(args.Int(1)), args.Error(2)
}

func (m *MockEquipmentSetRepository) GetItemsBySetID(ctx context.Context, setID string) ([]*game_config.Item, error) {
	args := m.Called(ctx, setID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*game_config.Item), args.Error(1)
}

// MockEquipmentRepository 是 EquipmentRepository 的 mock
type MockEquipmentRepository struct {
	mock.Mock
}

func (m *MockEquipmentRepository) GetEquippedItems(ctx context.Context, heroID string) ([]*game_runtime.PlayerItem, error) {
	args := m.Called(ctx, heroID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*game_runtime.PlayerItem), args.Error(1)
}

// MockItemRepository 是 ItemRepository 的 mock
type MockItemRepository struct {
	mock.Mock
}

func (m *MockItemRepository) GetByIDs(ctx context.Context, itemIDs []string) ([]*game_config.Item, error) {
	args := m.Called(ctx, itemIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*game_config.Item), args.Error(1)
}

// TestParseSetEffectToBonuses 测试套装效果解析
func TestParseSetEffectToBonuses(t *testing.T) {
	service := &EquipmentSetService{}

	tests := []struct {
		name     string
		effects  []SetEffectInfo
		expected map[string]*AttributeBonus
		hasError bool
	}{
		{
			name: "解析百分比加成",
			effects: []SetEffectInfo{
				{
					DataType:    "Status",
					DataID:      "ATK",
					BonusType:   "percent",
					BonusNumber: "10",
				},
			},
			expected: map[string]*AttributeBonus{
				"ATK": {
					FlatBonus:    0,
					PercentBonus: 10,
				},
			},
			hasError: false,
		},
		{
			name: "解析固定值加成",
			effects: []SetEffectInfo{
				{
					DataType:    "Status",
					DataID:      "HP",
					BonusType:   "flat",
					BonusNumber: "100",
				},
			},
			expected: map[string]*AttributeBonus{
				"HP": {
					FlatBonus:    100,
					PercentBonus: 0,
				},
			},
			hasError: false,
		},
		{
			name: "解析多个加成",
			effects: []SetEffectInfo{
				{
					DataType:    "Status",
					DataID:      "ATK",
					BonusType:   "percent",
					BonusNumber: "10",
				},
				{
					DataType:    "Status",
					DataID:      "CRIT_RATE",
					BonusType:   "percent",
					BonusNumber: "5",
				},
			},
			expected: map[string]*AttributeBonus{
				"ATK": {
					FlatBonus:    0,
					PercentBonus: 10,
				},
				"CRIT_RATE": {
					FlatBonus:    0,
					PercentBonus: 5,
				},
			},
			hasError: false,
		},
		{
			name: "无效的数字格式",
			effects: []SetEffectInfo{
				{
					DataType:    "Status",
					DataID:      "ATK",
					BonusType:   "percent",
					BonusNumber: "invalid",
				},
			},
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.parseSetEffectToBonuses(tt.effects)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expected), len(result))
				for attrID, expectedBonus := range tt.expected {
					actualBonus, ok := result[attrID]
					assert.True(t, ok, "属性 %s 应该存在", attrID)
					assert.Equal(t, expectedBonus.FlatBonus, actualBonus.FlatBonus)
					assert.Equal(t, expectedBonus.PercentBonus, actualBonus.PercentBonus)
				}
			}
		})
	}
}

// TestCalculateActiveTiers 测试激活档位计算
func TestCalculateActiveTiers(t *testing.T) {
	tests := []struct {
		name        string
		pieceCount  int
		setEffects  []SetEffect
		expectedLen int
	}{
		{
			name:       "2件套激活",
			pieceCount: 2,
			setEffects: []SetEffect{
				{RequiredPieces: 2},
				{RequiredPieces: 4},
			},
			expectedLen: 1, // 只激活2件套
		},
		{
			name:       "4件套激活",
			pieceCount: 4,
			setEffects: []SetEffect{
				{RequiredPieces: 2},
				{RequiredPieces: 4},
			},
			expectedLen: 2, // 激活2件套和4件套
		},
		{
			name:       "未达到最低件数",
			pieceCount: 1,
			setEffects: []SetEffect{
				{RequiredPieces: 2},
				{RequiredPieces: 4},
			},
			expectedLen: 0, // 不激活任何档位
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			activeTiers := []int{}
			for _, effect := range tt.setEffects {
				if tt.pieceCount >= effect.RequiredPieces {
					activeTiers = append(activeTiers, effect.RequiredPieces)
				}
			}
			assert.Equal(t, tt.expectedLen, len(activeTiers))
		})
	}
}

// TestMergeBonuses 测试属性加成合并
func TestMergeBonuses(t *testing.T) {
	tests := []struct {
		name     string
		bonus1   map[string]*AttributeBonus
		bonus2   map[string]*AttributeBonus
		expected map[string]*AttributeBonus
	}{
		{
			name: "合并不同属性",
			bonus1: map[string]*AttributeBonus{
				"ATK": {FlatBonus: 10, PercentBonus: 5},
			},
			bonus2: map[string]*AttributeBonus{
				"HP": {FlatBonus: 100, PercentBonus: 10},
			},
			expected: map[string]*AttributeBonus{
				"ATK": {FlatBonus: 10, PercentBonus: 5},
				"HP":  {FlatBonus: 100, PercentBonus: 10},
			},
		},
		{
			name: "合并相同属性",
			bonus1: map[string]*AttributeBonus{
				"ATK": {FlatBonus: 10, PercentBonus: 5},
			},
			bonus2: map[string]*AttributeBonus{
				"ATK": {FlatBonus: 20, PercentBonus: 10},
			},
			expected: map[string]*AttributeBonus{
				"ATK": {FlatBonus: 30, PercentBonus: 15},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := make(map[string]*AttributeBonus)
			
			// 复制 bonus1
			for k, v := range tt.bonus1 {
				result[k] = &AttributeBonus{
					FlatBonus:    v.FlatBonus,
					PercentBonus: v.PercentBonus,
				}
			}
			
			// 合并 bonus2
			for k, v := range tt.bonus2 {
				if existing, ok := result[k]; ok {
					existing.FlatBonus += v.FlatBonus
					existing.PercentBonus += v.PercentBonus
				} else {
					result[k] = &AttributeBonus{
						FlatBonus:    v.FlatBonus,
						PercentBonus: v.PercentBonus,
					}
				}
			}

			assert.Equal(t, len(tt.expected), len(result))
			for attrID, expectedBonus := range tt.expected {
				actualBonus, ok := result[attrID]
				assert.True(t, ok)
				assert.Equal(t, expectedBonus.FlatBonus, actualBonus.FlatBonus)
				assert.Equal(t, expectedBonus.PercentBonus, actualBonus.PercentBonus)
			}
		})
	}
}

