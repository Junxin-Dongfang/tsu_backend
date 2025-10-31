package service

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"github.com/aarondl/null/v8"

	"tsu-self/internal/entity/game_runtime"
)

// EquipmentEffectService 装备效果解析服务
type EquipmentEffectService struct{}

// NewEquipmentEffectService 创建装备效果解析服务
func NewEquipmentEffectService() *EquipmentEffectService {
	return &EquipmentEffectService{}
}

// EquipmentEffectConfig 装备效果配置(来自配置表)
type EquipmentEffectConfig struct {
	DataType    string `json:"Data_type"`    // "Status"
	DataID      string `json:"Data_ID"`      // "MAX_HP"
	BonusType   string `json:"Bouns_type"`   // "bonus" / "percent"
	BonusNumber string `json:"Bouns_Number"` // "5"
}

// EquipmentEffect 装备效果(系统内部格式)
type EquipmentEffect struct {
	TargetAttribute string  // 目标属性(MAX_HP, STR, DEX等)
	BonusType       string  // 加成类型(flat/percent)
	BonusValue      float64 // 加成值
}

// AttributeBonus 属性加成
type AttributeBonus struct {
	AttributeCode string  // 属性代码
	FlatBonus     float64 // 固定加值
	PercentBonus  float64 // 百分比加成
}

// ParseOutOfCombatEffects 解析局外效果
func (s *EquipmentEffectService) ParseOutOfCombatEffects(effectsJSON null.JSON) ([]EquipmentEffect, error) {
	if !effectsJSON.Valid {
		return []EquipmentEffect{}, nil
	}

	var configs []EquipmentEffectConfig
	if err := json.Unmarshal(effectsJSON.JSON, &configs); err != nil {
		return nil, fmt.Errorf("解析装备效果JSON失败: %w", err)
	}

	effects := make([]EquipmentEffect, 0, len(configs))
	for _, config := range configs {
		effect, err := s.convertConfigToEffect(config)
		if err != nil {
			return nil, fmt.Errorf("转换装备效果失败: %w", err)
		}
		effects = append(effects, effect)
	}

	return effects, nil
}

// convertConfigToEffect 将配置格式转换为系统格式
func (s *EquipmentEffectService) convertConfigToEffect(config EquipmentEffectConfig) (EquipmentEffect, error) {
	// 解析加成值
	bonusValue, err := strconv.ParseFloat(config.BonusNumber, 64)
	if err != nil {
		return EquipmentEffect{}, fmt.Errorf("解析加成值失败: %w", err)
	}

	// 转换加成类型
	bonusType := "flat"
	if config.BonusType == "percent" {
		bonusType = "percent"
		// 百分比转换为小数(10% -> 0.1)
		bonusValue = bonusValue / 100.0
	}

	return EquipmentEffect{
		TargetAttribute: config.DataID,
		BonusType:       bonusType,
		BonusValue:      bonusValue,
	}, nil
}

// CalculateEffectiveEnhancementLevel 计算实际强化等级(考虑耐久度)
// 公式: 实际强化等级 = round(强化等级 * 当前耐久 / 最大耐久)
func (s *EquipmentEffectService) CalculateEffectiveEnhancementLevel(item *game_runtime.PlayerItem, maxDurability int) int {
	if !item.EnhancementLevel.Valid || item.EnhancementLevel.Int16 == 0 {
		return 0
	}

	if !item.CurrentDurability.Valid || maxDurability == 0 {
		return int(item.EnhancementLevel.Int16)
	}

	// 计算耐久度比例
	ratio := float64(item.CurrentDurability.Int) / float64(maxDurability)

	// 计算实际强化等级
	effectiveLevel := float64(item.EnhancementLevel.Int16) * ratio

	// 四舍五入
	return int(math.Round(effectiveLevel))
}

// ApplyEnhancementBonus 应用强化加成
// 每级强化提升5%的局外属性
func (s *EquipmentEffectService) ApplyEnhancementBonus(effect EquipmentEffect, enhancementLevel int) EquipmentEffect {
	if enhancementLevel <= 0 {
		return effect
	}

	// 强化加成: 每级+5%
	enhancementBonus := 1.0 + (float64(enhancementLevel) * 0.05)

	// 应用强化加成到效果值
	effect.BonusValue = effect.BonusValue * enhancementBonus

	return effect
}

// CalculateAttributeBonuses 计算装备的属性加成
func (s *EquipmentEffectService) CalculateAttributeBonuses(
	effectsJSON null.JSON,
	enhancementLevel int,
	currentDurability int,
	maxDurability int,
) (map[string]*AttributeBonus, error) {
	// 解析装备效果
	effects, err := s.ParseOutOfCombatEffects(effectsJSON)
	if err != nil {
		return nil, err
	}

	// 计算实际强化等级(考虑耐久度)
	effectiveEnhancementLevel := enhancementLevel
	if maxDurability > 0 && currentDurability >= 0 {
		ratio := float64(currentDurability) / float64(maxDurability)
		effectiveEnhancementLevel = int(math.Round(float64(enhancementLevel) * ratio))
	}

	// 初始化属性加成映射
	bonuses := make(map[string]*AttributeBonus)

	// 遍历所有效果
	for _, effect := range effects {
		// 应用强化加成
		enhancedEffect := s.ApplyEnhancementBonus(effect, effectiveEnhancementLevel)

		// 获取或创建属性加成
		bonus, exists := bonuses[enhancedEffect.TargetAttribute]
		if !exists {
			bonus = &AttributeBonus{
				AttributeCode: enhancedEffect.TargetAttribute,
				FlatBonus:     0,
				PercentBonus:  0,
			}
			bonuses[enhancedEffect.TargetAttribute] = bonus
		}

		// 累加加成值
		if enhancedEffect.BonusType == "flat" {
			bonus.FlatBonus += enhancedEffect.BonusValue
		} else if enhancedEffect.BonusType == "percent" {
			bonus.PercentBonus += enhancedEffect.BonusValue
		}
	}

	return bonuses, nil
}

// MergeAttributeBonuses 合并多个装备的属性加成
func (s *EquipmentEffectService) MergeAttributeBonuses(bonusesList []map[string]*AttributeBonus) map[string]*AttributeBonus {
	merged := make(map[string]*AttributeBonus)

	for _, bonuses := range bonusesList {
		for attrCode, bonus := range bonuses {
			// 获取或创建属性加成
			mergedBonus, exists := merged[attrCode]
			if !exists {
				mergedBonus = &AttributeBonus{
					AttributeCode: attrCode,
					FlatBonus:     0,
					PercentBonus:  0,
				}
				merged[attrCode] = mergedBonus
			}

			// 累加加成值
			mergedBonus.FlatBonus += bonus.FlatBonus
			mergedBonus.PercentBonus += bonus.PercentBonus
		}
	}

	return merged
}

// ValidateEffectJSON 验证效果JSON格式
func (s *EquipmentEffectService) ValidateEffectJSON(effectsJSON null.JSON) error {
	if !effectsJSON.Valid {
		return nil
	}

	var configs []EquipmentEffectConfig
	if err := json.Unmarshal(effectsJSON.JSON, &configs); err != nil {
		return fmt.Errorf("效果JSON格式错误: %w", err)
	}

	for i, config := range configs {
		// 验证必填字段
		if config.DataType == "" {
			return fmt.Errorf("效果[%d]: Data_type不能为空", i)
		}
		if config.DataID == "" {
			return fmt.Errorf("效果[%d]: Data_ID不能为空", i)
		}
		if config.BonusType == "" {
			return fmt.Errorf("效果[%d]: Bouns_type不能为空", i)
		}
		if config.BonusNumber == "" {
			return fmt.Errorf("效果[%d]: Bouns_Number不能为空", i)
		}

		// 验证加成类型
		if config.BonusType != "bonus" && config.BonusType != "percent" {
			return fmt.Errorf("效果[%d]: Bouns_type必须是'bonus'或'percent'", i)
		}

		// 验证加成值
		if _, err := strconv.ParseFloat(config.BonusNumber, 64); err != nil {
			return fmt.Errorf("效果[%d]: Bouns_Number必须是数字", i)
		}
	}

	return nil
}

