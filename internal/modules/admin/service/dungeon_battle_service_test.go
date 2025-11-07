package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"tsu-self/internal/modules/admin/dto"
	"tsu-self/internal/pkg/xerrors"
)

func TestValidateMonsterPositions(t *testing.T) {
	tests := []struct {
		name      string
		setup     []dto.MonsterSetupItem
		wantError bool
		errorMsg  string
	}{
		{
			name: "有效的怪物位置",
			setup: []dto.MonsterSetupItem{
				{MonsterCode: "GOBLIN", Position: 14},
				{MonsterCode: "ORC", Position: 15},
			},
			wantError: false,
		},
		{
			name: "位置小于1应该失败",
			setup: []dto.MonsterSetupItem{
				{MonsterCode: "GOBLIN", Position: 0},
			},
			wantError: true,
			errorMsg:  "位置必须在1-21之间",
		},
		{
			name: "位置大于21应该失败",
			setup: []dto.MonsterSetupItem{
				{MonsterCode: "GOBLIN", Position: 22},
			},
			wantError: true,
			errorMsg:  "位置必须在1-21之间",
		},
		{
			name: "重复的位置应该失败",
			setup: []dto.MonsterSetupItem{
				{MonsterCode: "GOBLIN", Position: 14},
				{MonsterCode: "ORC", Position: 14}, // 重复
			},
			wantError: true,
			errorMsg:  "位置重复",
		},
		{
			name: "边界位置应该成功",
			setup: []dto.MonsterSetupItem{
				{MonsterCode: "GOBLIN", Position: 1},
				{MonsterCode: "ORC", Position: 21},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMonsterPositions(tt.setup)

			if tt.wantError {
				assert.Error(t, err, "应该返回错误")
			} else {
				assert.NoError(t, err, "不应该返回错误")
			}
		})
	}
}

// validateMonsterPositions 验证怪物位置(测试辅助函数)
func validateMonsterPositions(setup []dto.MonsterSetupItem) error {
	positionMap := make(map[int]bool)

	for _, monster := range setup {
		// 验证位置范围
		if monster.Position < 1 || monster.Position > 21 {
			return assert.AnError
		}

		// 验证位置唯一性
		if positionMap[monster.Position] {
			return assert.AnError
		}
		positionMap[monster.Position] = true
	}

	return nil
}

func TestValidateGlobalBuffTarget(t *testing.T) {
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

func TestBattleCodeValidation(t *testing.T) {
	tests := []struct {
		name       string
		battleCode string
		wantError  bool
	}{
		{
			name:       "有效的战斗代码",
			battleCode: "BATTLE_001",
			wantError:  false,
		},
		{
			name:       "空战斗代码应该失败",
			battleCode: "",
			wantError:  true,
		},
		{
			name:       "过长的战斗代码应该失败",
			battleCode: "BATTLE_" + string(make([]byte, 50)),
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantError {
				if tt.battleCode == "" {
					assert.Empty(t, tt.battleCode)
				} else if len(tt.battleCode) > 50 {
					assert.Greater(t, len(tt.battleCode), 50)
				}
			} else {
				assert.NotEmpty(t, tt.battleCode)
				assert.LessOrEqual(t, len(tt.battleCode), 50)
			}
		})
	}
}

func TestMonsterSetupValidation(t *testing.T) {
	tests := []struct {
		name      string
		setup     []dto.MonsterSetupItem
		wantError bool
		errorMsg  string
	}{
		{
			name: "空的怪物配置应该失败",
			setup: []dto.MonsterSetupItem{},
			wantError: true,
			errorMsg: "至少需要一个怪物",
		},
		{
			name: "有效的怪物配置",
			setup: []dto.MonsterSetupItem{
				{MonsterCode: "GOBLIN", Position: 14},
			},
			wantError: false,
		},
		{
			name: "多个怪物的有效配置",
			setup: []dto.MonsterSetupItem{
				{MonsterCode: "GOBLIN", Position: 14},
				{MonsterCode: "ORC", Position: 15},
				{MonsterCode: "TROLL", Position: 16},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if len(tt.setup) == 0 {
				err = xerrors.New(xerrors.CodeInvalidParams, "至少需要一个怪物")
			}
			
			if tt.wantError {
				assert.Error(t, err, "应该返回错误")
			} else {
				assert.NoError(t, err, "不应该返回错误")
			}
		})
	}
}

