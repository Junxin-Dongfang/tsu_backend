package service

import (
	"testing"

	"github.com/aarondl/null/v8"
	"github.com/stretchr/testify/assert"

	"tsu-self/internal/entity/game_runtime"
)

func TestEquipmentEffectService_ParseOutOfCombatEffects(t *testing.T) {
	svc := NewEquipmentEffectService()

	tests := []struct {
		name        string
		effectsJSON null.JSON
		want        int // 期望的效果数量
		wantErr     bool
	}{
		{
			name: "解析单个效果",
			effectsJSON: null.JSONFrom([]byte(`[
				{
					"Data_type": "Status",
					"Data_ID": "STR",
					"Bouns_type": "bonus",
					"Bouns_Number": "10"
				}
			]`)),
			want:    1,
			wantErr: false,
		},
		{
			name: "解析多个效果",
			effectsJSON: null.JSONFrom([]byte(`[
				{
					"Data_type": "Status",
					"Data_ID": "STR",
					"Bouns_type": "bonus",
					"Bouns_Number": "10"
				},
				{
					"Data_type": "Status",
					"Data_ID": "ATK",
					"Bouns_type": "percent",
					"Bouns_Number": "15"
				}
			]`)),
			want:    2,
			wantErr: false,
		},
		{
			name:        "空JSON",
			effectsJSON: null.JSON{},
			want:        0,
			wantErr:     false,
		},
		{
			name:        "无效JSON",
			effectsJSON: null.JSONFrom([]byte(`invalid json`)),
			want:        0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.ParseOutOfCombatEffects(tt.effectsJSON)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, len(got))
			}
		})
	}
}

func TestEquipmentEffectService_CalculateEffectiveEnhancementLevel(t *testing.T) {
	svc := NewEquipmentEffectService()

	tests := []struct {
		name           string
		item           *game_runtime.PlayerItem
		maxDurability  int
		expectedLevel  int
	}{
		{
			name: "满耐久度",
			item: &game_runtime.PlayerItem{
				EnhancementLevel:  null.Int16From(5),
				CurrentDurability: null.IntFrom(100),
			},
			maxDurability: 100,
			expectedLevel: 5,
		},
		{
			name: "半耐久度",
			item: &game_runtime.PlayerItem{
				EnhancementLevel:  null.Int16From(10),
				CurrentDurability: null.IntFrom(50),
			},
			maxDurability: 100,
			expectedLevel: 5, // round(10 * 50/100) = 5
		},
		{
			name: "低耐久度",
			item: &game_runtime.PlayerItem{
				EnhancementLevel:  null.Int16From(10),
				CurrentDurability: null.IntFrom(10),
			},
			maxDurability: 100,
			expectedLevel: 1, // round(10 * 10/100) = 1
		},
		{
			name: "零耐久度",
			item: &game_runtime.PlayerItem{
				EnhancementLevel:  null.Int16From(10),
				CurrentDurability: null.IntFrom(0),
			},
			maxDurability: 100,
			expectedLevel: 0,
		},
		{
			name: "未强化",
			item: &game_runtime.PlayerItem{
				EnhancementLevel:  null.Int16From(0),
				CurrentDurability: null.IntFrom(100),
			},
			maxDurability: 100,
			expectedLevel: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.CalculateEffectiveEnhancementLevel(tt.item, tt.maxDurability)
			assert.Equal(t, tt.expectedLevel, got)
		})
	}
}

func TestEquipmentEffectService_ApplyEnhancementBonus(t *testing.T) {
	svc := NewEquipmentEffectService()

	tests := []struct {
		name             string
		effect           EquipmentEffect
		enhancementLevel int
		expectedValue    float64
	}{
		{
			name: "无强化",
			effect: EquipmentEffect{
				TargetAttribute: "STR",
				BonusType:       "flat",
				BonusValue:      10.0,
			},
			enhancementLevel: 0,
			expectedValue:    10.0,
		},
		{
			name: "强化+5",
			effect: EquipmentEffect{
				TargetAttribute: "STR",
				BonusType:       "flat",
				BonusValue:      10.0,
			},
			enhancementLevel: 5,
			expectedValue:    12.5, // 10 * (1 + 5*0.05) = 12.5
		},
		{
			name: "强化+10",
			effect: EquipmentEffect{
				TargetAttribute: "ATK",
				BonusType:       "percent",
				BonusValue:      0.15,
			},
			enhancementLevel: 10,
			expectedValue:    0.225, // 0.15 * (1 + 10*0.05) = 0.225
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.ApplyEnhancementBonus(tt.effect, tt.enhancementLevel)
			assert.InDelta(t, tt.expectedValue, got.BonusValue, 0.001)
		})
	}
}

func TestEquipmentEffectService_CalculateAttributeBonuses(t *testing.T) {
	svc := NewEquipmentEffectService()

	effectsJSON := null.JSONFrom([]byte(`[
		{
			"Data_type": "Status",
			"Data_ID": "STR",
			"Bouns_type": "bonus",
			"Bouns_Number": "10"
		},
		{
			"Data_type": "Status",
			"Data_ID": "ATK",
			"Bouns_type": "percent",
			"Bouns_Number": "15"
		}
	]`))

	tests := []struct {
		name              string
		effectsJSON       null.JSON
		enhancementLevel  int
		currentDurability int
		maxDurability     int
		wantSTR           float64
		wantATK           float64
		wantErr           bool
	}{
		{
			name:              "无强化_满耐久",
			effectsJSON:       effectsJSON,
			enhancementLevel:  0,
			currentDurability: 100,
			maxDurability:     100,
			wantSTR:           10.0,
			wantATK:           0.15,
			wantErr:           false,
		},
		{
			name:              "强化+5_满耐久",
			effectsJSON:       effectsJSON,
			enhancementLevel:  5,
			currentDurability: 100,
			maxDurability:     100,
			wantSTR:           12.5,  // 10 * 1.25
			wantATK:           0.1875, // 0.15 * 1.25
			wantErr:           false,
		},
		{
			name:              "强化+10_半耐久",
			effectsJSON:       effectsJSON,
			enhancementLevel:  10,
			currentDurability: 50,
			maxDurability:     100,
			wantSTR:           12.5,  // 10 * (1 + 5*0.05) = 12.5
			wantATK:           0.1875, // 0.15 * (1 + 5*0.05) = 0.1875
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bonuses, err := svc.CalculateAttributeBonuses(
				tt.effectsJSON,
				tt.enhancementLevel,
				tt.currentDurability,
				tt.maxDurability,
			)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, bonuses)

				if strBonus, ok := bonuses["STR"]; ok {
					assert.InDelta(t, tt.wantSTR, strBonus.FlatBonus, 0.001)
				}

				if atkBonus, ok := bonuses["ATK"]; ok {
					assert.InDelta(t, tt.wantATK, atkBonus.PercentBonus, 0.001)
				}
			}
		})
	}
}

func TestEquipmentEffectService_MergeAttributeBonuses(t *testing.T) {
	svc := NewEquipmentEffectService()

	bonuses1 := map[string]*AttributeBonus{
		"STR": {
			AttributeCode: "STR",
			FlatBonus:     10.0,
			PercentBonus:  0.05,
		},
		"ATK": {
			AttributeCode: "ATK",
			FlatBonus:     20.0,
			PercentBonus:  0.0,
		},
	}

	bonuses2 := map[string]*AttributeBonus{
		"STR": {
			AttributeCode: "STR",
			FlatBonus:     5.0,
			PercentBonus:  0.03,
		},
		"DEF": {
			AttributeCode: "DEF",
			FlatBonus:     15.0,
			PercentBonus:  0.1,
		},
	}

	merged := svc.MergeAttributeBonuses([]map[string]*AttributeBonus{bonuses1, bonuses2})

	// 验证STR合并
	assert.Equal(t, 15.0, merged["STR"].FlatBonus)    // 10 + 5
	assert.Equal(t, 0.08, merged["STR"].PercentBonus) // 0.05 + 0.03

	// 验证ATK保留
	assert.Equal(t, 20.0, merged["ATK"].FlatBonus)
	assert.Equal(t, 0.0, merged["ATK"].PercentBonus)

	// 验证DEF添加
	assert.Equal(t, 15.0, merged["DEF"].FlatBonus)
	assert.Equal(t, 0.1, merged["DEF"].PercentBonus)
}

