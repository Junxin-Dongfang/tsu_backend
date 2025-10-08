package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/types"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// ActionService 动作服务
type ActionService struct {
	repo interfaces.ActionRepository
}

// NewActionService 创建动作服务
func NewActionService(db *sql.DB) *ActionService {
	return &ActionService{
		repo: impl.NewActionRepository(db),
	}
}

// GetActions 获取动作列表
func (s *ActionService) GetActions(ctx context.Context, params interfaces.ActionQueryParams) ([]*game_config.Action, int64, error) {
	return s.repo.List(ctx, params)
}

// GetActionByID 根据ID获取动作
func (s *ActionService) GetActionByID(ctx context.Context, actionID string) (*game_config.Action, error) {
	return s.repo.GetByID(ctx, actionID)
}

// CreateAction 创建动作
func (s *ActionService) CreateAction(ctx context.Context, action *game_config.Action) error {
	// 业务验证：检查动作代码是否已存在
	exists, err := s.repo.Exists(ctx, action.ActionCode)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("动作代码已存在: %s", action.ActionCode)
	}

	return s.repo.Create(ctx, action)
}

// UpdateAction 更新动作信息
func (s *ActionService) UpdateAction(ctx context.Context, actionID string, updates map[string]interface{}) error {
	action, err := s.repo.GetByID(ctx, actionID)
	if err != nil {
		return err
	}

	// 更新字段
	if actionCode, ok := updates["action_code"].(string); ok && actionCode != "" {
		existing, err := s.repo.GetByCode(ctx, actionCode)
		if err == nil && existing.ID != actionID {
			return fmt.Errorf("动作代码已被使用: %s", actionCode)
		}
		action.ActionCode = actionCode
	}

	if actionName, ok := updates["action_name"].(string); ok && actionName != "" {
		action.ActionName = actionName
	}

	if actionType, ok := updates["action_type"].(string); ok && actionType != "" {
		action.ActionType = actionType
	}

	if categoryID, ok := updates["action_category_id"].(string); ok {
		if categoryID != "" {
			action.ActionCategoryID.SetValid(categoryID)
		} else {
			action.ActionCategoryID.Valid = false
		}
	}

	if relatedSkillID, ok := updates["related_skill_id"].(string); ok {
		if relatedSkillID != "" {
			action.RelatedSkillID.SetValid(relatedSkillID)
		} else {
			action.RelatedSkillID.Valid = false
		}
	}

	if rangeConfig, ok := updates["range_config"].(types.JSON); ok {
		action.RangeConfig = rangeConfig
	}

	if targetConfig, ok := updates["target_config"].(null.JSON); ok {
		action.TargetConfig = targetConfig
	}

	if areaConfig, ok := updates["area_config"].(null.JSON); ok {
		action.AreaConfig = areaConfig
	}

	if actionPointCost, ok := updates["action_point_cost"].(int); ok {
		action.ActionPointCost.SetValid(actionPointCost)
	}

	if manaCost, ok := updates["mana_cost"].(int); ok {
		action.ManaCost.SetValid(manaCost)
	}

	if manaCostFormula, ok := updates["mana_cost_formula"].(string); ok {
		if manaCostFormula != "" {
			action.ManaCostFormula.SetValid(manaCostFormula)
		} else {
			action.ManaCostFormula.Valid = false
		}
	}

	if cooldownTurns, ok := updates["cooldown_turns"].(int); ok {
		action.CooldownTurns.SetValid(cooldownTurns)
	}

	if usesPerBattle, ok := updates["uses_per_battle"].(int); ok {
		action.UsesPerBattle.SetValid(usesPerBattle)
	}

	if hitRateConfig, ok := updates["hit_rate_config"].(null.JSON); ok {
		action.HitRateConfig = hitRateConfig
	}

	if requirements, ok := updates["requirements"].(null.JSON); ok {
		action.Requirements = requirements
	}

	if animationConfig, ok := updates["animation_config"].(null.JSON); ok {
		action.AnimationConfig = animationConfig
	}

	if visualEffects, ok := updates["visual_effects"].(null.JSON); ok {
		action.VisualEffects = visualEffects
	}

	if soundEffects, ok := updates["sound_effects"].(null.JSON); ok {
		action.SoundEffects = soundEffects
	}

	if featureTags, ok := updates["feature_tags"].(types.StringArray); ok {
		action.FeatureTags = featureTags
	}

	if startFlags, ok := updates["start_flags"].(types.StringArray); ok {
		action.StartFlags = startFlags
	}

	if description, ok := updates["description"].(string); ok {
		if description != "" {
			action.Description.SetValid(description)
		} else {
			action.Description.Valid = false
		}
	}

	if isActive, ok := updates["is_active"].(bool); ok {
		action.IsActive.SetValid(isActive)
	}

	return s.repo.Update(ctx, action)
}

// DeleteAction 删除动作（软删除）
func (s *ActionService) DeleteAction(ctx context.Context, actionID string) error {
	return s.repo.Delete(ctx, actionID)
}
