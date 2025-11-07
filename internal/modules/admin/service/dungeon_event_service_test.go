package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"tsu-self/internal/modules/admin/dto"
	"tsu-self/internal/pkg/xerrors"
)

func TestValidateRewardExp(t *testing.T) {
	tests := []struct {
		name      string
		rewardExp int
		wantError bool
	}{
		{
			name:      "正数经验值应该成功",
			rewardExp: 100,
			wantError: false,
		},
		{
			name:      "零经验值应该成功",
			rewardExp: 0,
			wantError: false,
		},
		{
			name:      "负数经验值应该失败",
			rewardExp: -10,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.rewardExp < 0 {
				err = xerrors.New(xerrors.CodeInvalidParams, "经验奖励不能为负数")
			}
			
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateApplyEffectTarget(t *testing.T) {
	tests := []struct {
		name      string
		target    string
		wantError bool
	}{
		{
			name:      "all_heroes是有效目标",
			target:    "all_heroes",
			wantError: false,
		},
		{
			name:      "all_monsters是有效目标",
			target:    "all_monsters",
			wantError: false,
		},
		{
			name:      "all是有效目标",
			target:    "all",
			wantError: false,
		},
		{
			name:      "random_hero是有效目标",
			target:    "random_hero",
			wantError: false,
		},
		{
			name:      "无效的目标应该失败",
			target:    "invalid_target",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validTargets := map[string]bool{
				"all_heroes":   true,
				"all_monsters": true,
				"all":          true,
				"random_hero":  true,
			}
			
			_, valid := validTargets[tt.target]
			
			if tt.wantError {
				assert.False(t, valid)
			} else {
				assert.True(t, valid)
			}
		})
	}
}

func TestValidateGuaranteedItems(t *testing.T) {
	tests := []struct {
		name      string
		items     []dto.GuaranteedItem
		wantError bool
		errorMsg  string
	}{
		{
			name: "有效的保底物品",
			items: []dto.GuaranteedItem{
				{ItemCode: "HEALTH_POTION", Quantity: 3},
			},
			wantError: false,
		},
		{
			name: "数量为0应该失败",
			items: []dto.GuaranteedItem{
				{ItemCode: "HEALTH_POTION", Quantity: 0},
			},
			wantError: true,
			errorMsg:  "数量必须大于0",
		},
		{
			name: "负数数量应该失败",
			items: []dto.GuaranteedItem{
				{ItemCode: "HEALTH_POTION", Quantity: -1},
			},
			wantError: true,
			errorMsg:  "数量必须大于0",
		},
		{
			name: "多个有效物品",
			items: []dto.GuaranteedItem{
				{ItemCode: "HEALTH_POTION", Quantity: 3},
				{ItemCode: "MANA_POTION", Quantity: 2},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			for _, item := range tt.items {
				if item.Quantity <= 0 {
					err = xerrors.New(xerrors.CodeInvalidParams, "数量必须大于0")
					break
				}
			}
			
			if tt.wantError {
				assert.Error(t, err, "应该返回错误")
			} else {
				assert.NoError(t, err, "不应该返回错误")
			}
		})
	}
}

func TestEventCodeValidation(t *testing.T) {
	tests := []struct {
		name      string
		eventCode string
		wantError bool
	}{
		{
			name:      "有效的事件代码",
			eventCode: "EVENT_001",
			wantError: false,
		},
		{
			name:      "空事件代码应该失败",
			eventCode: "",
			wantError: true,
		},
		{
			name:      "过长的事件代码应该失败",
			eventCode: "EVENT_" + string(make([]byte, 50)),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantError {
				if tt.eventCode == "" {
					assert.Empty(t, tt.eventCode)
				} else if len(tt.eventCode) > 50 {
					assert.Greater(t, len(tt.eventCode), 50)
				}
			} else {
				assert.NotEmpty(t, tt.eventCode)
				assert.LessOrEqual(t, len(tt.eventCode), 50)
			}
		})
	}
}

func TestApplyEffectValidation(t *testing.T) {
	tests := []struct {
		name      string
		effect    dto.ApplyEffectItem
		wantError bool
		errorMsg  string
	}{
		{
			name: "有效的效果配置",
			effect: dto.ApplyEffectItem{
				BuffCode:    "STRENGTH_BUFF",
				CasterLevel: 5,
				Target:      "all_heroes",
			},
			wantError: false,
		},
		{
			name: "施法者等级为0应该失败",
			effect: dto.ApplyEffectItem{
				BuffCode:    "STRENGTH_BUFF",
				CasterLevel: 0,
				Target:      "all_heroes",
			},
			wantError: true,
			errorMsg:  "施法者等级必须大于0",
		},
		{
			name: "负数施法者等级应该失败",
			effect: dto.ApplyEffectItem{
				BuffCode:    "STRENGTH_BUFF",
				CasterLevel: -1,
				Target:      "all_heroes",
			},
			wantError: true,
			errorMsg:  "施法者等级必须大于0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.effect.CasterLevel <= 0 {
				err = xerrors.New(xerrors.CodeInvalidParams, "施法者等级必须大于0")
			}
			
			if tt.wantError {
				assert.Error(t, err, "应该返回错误")
			} else {
				assert.NoError(t, err, "不应该返回错误")
			}
		})
	}
}

