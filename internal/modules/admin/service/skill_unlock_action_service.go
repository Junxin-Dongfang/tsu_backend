package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// SkillUnlockActionService 技能解锁动作服务
type SkillUnlockActionService struct {
	unlockActionRepo interfaces.SkillUnlockActionRepository
	skillRepo        interfaces.SkillRepository
	actionRepo       interfaces.ActionRepository
	dictRepo         interfaces.MetadataDictionaryRepository
}

// NewSkillUnlockActionService 创建技能解锁动作服务
func NewSkillUnlockActionService(db *sql.DB) *SkillUnlockActionService {
	return &SkillUnlockActionService{
		unlockActionRepo: impl.NewSkillUnlockActionRepository(db),
		skillRepo:        impl.NewSkillRepository(db),
		actionRepo:       impl.NewActionRepository(db),
		dictRepo:         impl.NewMetadataDictionaryRepository(db),
	}
}

// GetSkillUnlockActionByID 根据ID获取技能解锁动作
func (s *SkillUnlockActionService) GetSkillUnlockActionByID(ctx context.Context, id string) (*game_config.SkillUnlockAction, error) {
	return s.unlockActionRepo.GetByID(ctx, id)
}

// GetSkillUnlockActions 获取技能的所有解锁动作
func (s *SkillUnlockActionService) GetSkillUnlockActions(ctx context.Context, skillID string) ([]*game_config.SkillUnlockAction, error) {
	// 验证技能存在
	if _, err := s.skillRepo.GetByID(ctx, skillID); err != nil {
		return nil, err
	}

	return s.unlockActionRepo.GetBySkillID(ctx, skillID)
}

// AddUnlockActionToSkill 为技能添加解锁动作
func (s *SkillUnlockActionService) AddUnlockActionToSkill(ctx context.Context, unlockAction *game_config.SkillUnlockAction) error {
	// 验证技能存在
	if _, err := s.skillRepo.GetByID(ctx, unlockAction.SkillID); err != nil {
		return err
	}

	// 验证动作存在
	action, err := s.actionRepo.GetByID(ctx, unlockAction.ActionID)
	if err != nil {
		return err
	}

	// 验证 level_scaling_config 配置
	if !unlockAction.LevelScalingConfig.IsZero() {
		var scalingConfig map[string]interface{}
		if err := json.Unmarshal(unlockAction.LevelScalingConfig.JSON, &scalingConfig); err != nil {
			return fmt.Errorf("level_scaling_config 格式错误: %w", err)
		}

		if err := s.validateLevelScalingConfig(ctx, action, scalingConfig); err != nil {
			return err
		}
	}

	return s.unlockActionRepo.Create(ctx, unlockAction)
}

// RemoveUnlockActionFromSkill 从技能移除解锁动作
func (s *SkillUnlockActionService) RemoveUnlockActionFromSkill(ctx context.Context, unlockActionID string) error {
	return s.unlockActionRepo.Delete(ctx, unlockActionID)
}

// UpdateUnlockAction 更新技能解锁动作
func (s *SkillUnlockActionService) UpdateUnlockAction(ctx context.Context, unlockAction *game_config.SkillUnlockAction) error {
	// 验证记录存在
	existing, err := s.unlockActionRepo.GetByID(ctx, unlockAction.ID)
	if err != nil {
		return err
	}

	// 验证动作存在并获取动作信息用于验证配置
	action, err := s.actionRepo.GetByID(ctx, existing.ActionID)
	if err != nil {
		return err
	}

	// 验证 level_scaling_config 配置
	if !unlockAction.LevelScalingConfig.IsZero() {
		var scalingConfig map[string]interface{}
		if err := json.Unmarshal(unlockAction.LevelScalingConfig.JSON, &scalingConfig); err != nil {
			return fmt.Errorf("level_scaling_config 格式错误: %w", err)
		}

		if err := s.validateLevelScalingConfig(ctx, action, scalingConfig); err != nil {
			return err
		}
	}

	return s.unlockActionRepo.Update(ctx, unlockAction)
}

// BatchSetSkillUnlockActions 批量设置技能解锁动作（先删后建）
func (s *SkillUnlockActionService) BatchSetSkillUnlockActions(ctx context.Context, skillID string, unlockActions []*game_config.SkillUnlockAction) error {
	// 验证技能存在
	if _, err := s.skillRepo.GetByID(ctx, skillID); err != nil {
		return err
	}

	// 验证所有动作存在并验证配置
	for _, unlockAction := range unlockActions {
		action, err := s.actionRepo.GetByID(ctx, unlockAction.ActionID)
		if err != nil {
			return fmt.Errorf("动作不存在: %s", unlockAction.ActionID)
		}

		// 验证 level_scaling_config 配置
		if !unlockAction.LevelScalingConfig.IsZero() {
			var scalingConfig map[string]interface{}
			if err := json.Unmarshal(unlockAction.LevelScalingConfig.JSON, &scalingConfig); err != nil {
				return fmt.Errorf("动作 %s 的 level_scaling_config 格式错误: %w", unlockAction.ActionID, err)
			}

			if err := s.validateLevelScalingConfig(ctx, action, scalingConfig); err != nil {
				return fmt.Errorf("动作 %s 的配置验证失败: %w", unlockAction.ActionID, err)
			}
		}
	}

	// 先删除旧关联
	if err := s.unlockActionRepo.DeleteAllBySkillID(ctx, skillID); err != nil {
		return err
	}

	// 批量创建新关联
	if len(unlockActions) > 0 {
		return s.unlockActionRepo.BatchCreate(ctx, unlockActions)
	}

	return nil
}

// validateLevelScalingConfig 验证等级成长配置
// 使用字典数据验证配置的属性名是否是动作的有效属性
func (s *SkillUnlockActionService) validateLevelScalingConfig(ctx context.Context, action *game_config.Action, config map[string]interface{}) error {
	if len(config) == 0 {
		return nil // 空配置是允许的
	}

	// 获取所有动作属性字典项
	dictEntries, err := s.dictRepo.GetActionAttributes(ctx)
	if err != nil {
		return fmt.Errorf("获取动作属性字典失败: %w", err)
	}

	// 构建字典属性映射（包含元数据）
	dictAttrs := make(map[string]*game_config.MetadataDictionary)
	for _, entry := range dictEntries {
		dictAttrs[entry.VariableCode] = entry
	}

	// 获取动作实际拥有的属性
	validAttrs := s.getValidActionAttributes(action)

	// 验证每个配置项
	for attrName, attrConfig := range config {
		// 检查属性是否在字典中定义
		dictEntry, existsInDict := dictAttrs[attrName]
		if !existsInDict {
			return fmt.Errorf("无效的属性名 '%s'，该属性未在字典中定义", attrName)
		}

		// 检查属性是否可配置
		if !dictEntry.Metadata.IsZero() {
			var metadata map[string]interface{}
			if err := json.Unmarshal(dictEntry.Metadata.JSON, &metadata); err == nil {
				if scalable, ok := metadata["scalable"].(bool); ok && !scalable {
					return fmt.Errorf("属性 '%s' 不支持等级成长配置", attrName)
				}
			}
		}

		// 检查动作是否实际拥有该属性
		if !validAttrs[attrName] {
			return fmt.Errorf("属性 '%s' 不存在于该动作的配置中", attrName)
		}

		// 验证配置格式
		if err := s.validateScalingConfigFormat(attrName, attrConfig); err != nil {
			return err
		}

		// 验证成长类型是否被允许
		configMap, ok := attrConfig.(map[string]interface{})
		if ok {
			if typeVal, hasType := configMap["type"]; hasType {
				if err := s.validateScalingType(dictEntry, typeVal.(string)); err != nil {
					return fmt.Errorf("属性 '%s' 的成长类型验证失败: %w", attrName, err)
				}
			}
		}
	}

	return nil
}

// ScalableAttribute 可配置的属性信息
type ScalableAttribute struct {
	Name        string `json:"name"`         // 属性名
	DisplayName string `json:"display_name"` // 显示名称
	Description string `json:"description"`  // 描述
	Category    string `json:"category"`     // 分类: basic(基础属性), hit_rate(命中率), effect(效果参数)
}

// GetActionScalableAttributes 获取动作的所有可配置属性（用于前端展示）
// 从字典表读取动作属性定义，并根据动作的实际字段筛选
func (s *SkillUnlockActionService) GetActionScalableAttributes(ctx context.Context, actionID string) ([]ScalableAttribute, error) {
	// 验证动作存在
	action, err := s.actionRepo.GetByID(ctx, actionID)
	if err != nil {
		return nil, err
	}

	// 从字典表获取所有动作属性定义
	dictEntries, err := s.dictRepo.GetActionAttributes(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取动作属性字典失败: %w", err)
	}

	// 获取动作的有效属性集合
	validAttrs := s.getValidActionAttributes(action)

	// 根据动作的实际字段筛选并转换为返回格式
	var attributes []ScalableAttribute
	for _, entry := range dictEntries {
		// 只返回动作实际拥有的属性
		if validAttrs[entry.VariableCode] {
			// 从 metadata 中获取 category
			category := "basic"
			if !entry.Metadata.IsZero() {
				var metadata map[string]interface{}
				if err := json.Unmarshal(entry.Metadata.JSON, &metadata); err == nil {
					if cat, ok := metadata["category"].(string); ok {
						category = cat
					}
				}
			}

			attr := ScalableAttribute{
				Name:        entry.VariableCode,
				DisplayName: entry.VariableName,
				Category:    category,
			}
			if entry.Description.Valid {
				attr.Description = entry.Description.String
			}
			attributes = append(attributes, attr)
		}
	}

	return attributes, nil
}

// getValidActionAttributes 获取动作的所有可配置属性
func (s *SkillUnlockActionService) getValidActionAttributes(action *game_config.Action) map[string]bool {
	validAttrs := make(map[string]bool)

	// 动作基础属性
	if action.ManaCost.Valid {
		validAttrs["mana_cost"] = true
		validAttrs["mp_cost"] = true // 别名
	}
	if action.CooldownTurns.Valid {
		validAttrs["cooldown"] = true
		validAttrs["cooldown_turns"] = true // 别名
	}
	if action.ActionPointCost.Valid {
		validAttrs["action_point_cost"] = true
	}

	// 命中率配置中的属性
	if !action.HitRateConfig.IsZero() {
		var hitRateConfig map[string]interface{}
		if err := json.Unmarshal(action.HitRateConfig.JSON, &hitRateConfig); err == nil {
			if _, ok := hitRateConfig["base_hit_rate"]; ok {
				validAttrs["base_hit_rate"] = true
			}
			if _, ok := hitRateConfig["accuracy_multiplier"]; ok {
				validAttrs["accuracy_multiplier"] = true
			}
			if _, ok := hitRateConfig["min_hit_rate"]; ok {
				validAttrs["min_hit_rate"] = true
			}
			if _, ok := hitRateConfig["max_hit_rate"]; ok {
				validAttrs["max_hit_rate"] = true
			}
		}
	}

	// 效果配置中的参数（从 LegacyEffectConfig 解析）
	if !action.LegacyEffectConfig.IsZero() {
		var effectConfig map[string]interface{}
		if err := json.Unmarshal(action.LegacyEffectConfig.JSON, &effectConfig); err == nil {
			// 解析 Effects 数组
			if effects, ok := effectConfig["Effects"].([]interface{}); ok {
				for _, eff := range effects {
					if effMap, ok := eff.(map[string]interface{}); ok {
						if params, ok := effMap["params"].(map[string]interface{}); ok {
							// 添加所有效果参数作为可配置属性
							for paramName := range params {
								validAttrs[paramName] = true
							}
						}
					}
				}
			}

			// 添加常见的效果参数（即使当前没有配置也允许）
			commonEffectParams := []string{}
			for _, param := range commonEffectParams {
				validAttrs[param] = true
			}
		}
	}

	return validAttrs
}

// validateScalingConfigFormat 验证单个成长配置的格式
func (s *SkillUnlockActionService) validateScalingConfigFormat(attrName string, config interface{}) error {
	configMap, ok := config.(map[string]interface{})
	if !ok {
		return fmt.Errorf("属性 '%s' 的配置格式错误，应为对象", attrName)
	}

	// 检查必需字段
	typeVal, hasType := configMap["type"]
	if !hasType {
		return fmt.Errorf("属性 '%s' 缺少 'type' 字段", attrName)
	}

	// 验证 type 值
	typeStr, ok := typeVal.(string)
	if !ok {
		return fmt.Errorf("属性 '%s' 的 'type' 必须是字符串", attrName)
	}

	validTypes := map[string]bool{
		"linear":     true,
		"percentage": true,
		"fixed":      true,
	}
	if !validTypes[typeStr] {
		return fmt.Errorf("属性 '%s' 的 'type' 值无效，必须是 linear/percentage/fixed 之一", attrName)
	}

	// 检查 base 字段
	if _, hasBase := configMap["base"]; !hasBase {
		return fmt.Errorf("属性 '%s' 缺少 'base' 字段", attrName)
	}

	// 检查 value 字段
	if _, hasValue := configMap["value"]; !hasValue {
		return fmt.Errorf("属性 '%s' 缺少 'value' 字段", attrName)
	}

	// 验证数值类型
	if err := s.validateNumericField(attrName, "base", configMap["base"]); err != nil {
		return err
	}
	if err := s.validateNumericField(attrName, "value", configMap["value"]); err != nil {
		return err
	}

	return nil
}

// validateNumericField 验证数值字段
func (s *SkillUnlockActionService) validateNumericField(attrName, fieldName string, value interface{}) error {
	switch value.(type) {
	case int, int32, int64, float32, float64:
		return nil
	default:
		return fmt.Errorf("属性 '%s' 的 '%s' 必须是数值类型", attrName, fieldName)
	}
}

// validateScalingType 验证成长类型是否被字典条目允许
func (s *SkillUnlockActionService) validateScalingType(dictEntry *game_config.MetadataDictionary, scalingType string) error {
	if dictEntry.Metadata.IsZero() {
		return nil // 没有元数据则不限制
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(dictEntry.Metadata.JSON, &metadata); err != nil {
		return nil // 解析失败则不限制
	}

	// 检查是否定义了允许的成长类型列表
	if scalingTypesRaw, ok := metadata["scaling_types"]; ok {
		if scalingTypesArr, ok := scalingTypesRaw.([]interface{}); ok {
			// 构建允许的类型集合
			allowedTypes := make(map[string]bool)
			for _, t := range scalingTypesArr {
				if typeStr, ok := t.(string); ok {
					allowedTypes[typeStr] = true
				}
			}

			// 检查当前类型是否允许
			if len(allowedTypes) > 0 && !allowedTypes[scalingType] {
				return fmt.Errorf("成长类型 '%s' 不被支持，允许的类型: %v", scalingType, scalingTypesArr)
			}
		}
	}

	return nil
}
