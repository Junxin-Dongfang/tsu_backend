package service

import (
	"context"
	"encoding/json"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/interfaces"
)

// SkillDetailService 技能详情查询服务
type SkillDetailService struct {
	skillRepo             interfaces.SkillRepository
	skillUnlockActionRepo interfaces.SkillUnlockActionRepository
	actionRepo            interfaces.ActionRepository
	actionEffectRepo      interfaces.ActionEffectRepository
	effectRepo            interfaces.EffectRepository
	skillCategoryRepo     interfaces.SkillCategoryRepository
}

// NewSkillDetailService 创建技能详情查询服务
func NewSkillDetailService(
	skillRepo interfaces.SkillRepository,
	skillUnlockActionRepo interfaces.SkillUnlockActionRepository,
	actionRepo interfaces.ActionRepository,
	actionEffectRepo interfaces.ActionEffectRepository,
	effectRepo interfaces.EffectRepository,
	skillCategoryRepo interfaces.SkillCategoryRepository,
) *SkillDetailService {
	return &SkillDetailService{
		skillRepo:             skillRepo,
		skillUnlockActionRepo: skillUnlockActionRepo,
		actionRepo:            actionRepo,
		actionEffectRepo:      actionEffectRepo,
		effectRepo:            effectRepo,
		skillCategoryRepo:     skillCategoryRepo,
	}
}

// SkillQueryParams 技能查询参数
type SkillQueryParams struct {
	SkillType  *string // 技能类型
	CategoryID *string // 分类ID
	IsActive   *bool   // 是否启用
	Limit      int     // 每页数量
	Offset     int     // 偏移量
}

// ==================== 简化版响应（仅基本信息） ====================

// SkillBasicResponse 技能基本信息响应
type SkillBasicResponse struct {
	ID           string  `json:"id"`
	SkillCode    string  `json:"skill_code"`
	SkillName    string  `json:"skill_name"`
	SkillType    string  `json:"skill_type"` // active/passive
	Description  *string `json:"description,omitempty"`
	MaxLevel     int     `json:"max_level"`
	CategoryID   string  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	IsActive     bool    `json:"is_active"`
}

// GetSkillBasic 获取技能基本信息（简化版）
func (s *SkillDetailService) GetSkillBasic(ctx context.Context, skillID string) (*SkillBasicResponse, error) {
	// 查询技能
	skill, err := s.skillRepo.GetByID(ctx, skillID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "技能不存在")
	}

	// 查询分类
	var categoryName string
	if !skill.CategoryID.IsZero() {
		category, err := s.skillCategoryRepo.GetByID(ctx, skill.CategoryID.String)
		if err == nil {
			categoryName = category.CategoryName
		}
	}

	// 构造响应
	resp := &SkillBasicResponse{
		ID:           skill.ID,
		SkillCode:    skill.SkillCode,
		SkillName:    skill.SkillName,
		SkillType:    skill.SkillType,
		MaxLevel:     int(skill.MaxLevel.Int),
		CategoryID:   skill.CategoryID.String,
		CategoryName: categoryName,
		IsActive:     skill.IsActive.Bool,
	}

	if !skill.Description.IsZero() {
		resp.Description = &skill.Description.String
	}

	return resp, nil
}

// ListSkillsBasic 获取技能列表（简化版）
func (s *SkillDetailService) ListSkillsBasic(ctx context.Context, params SkillQueryParams) ([]*SkillBasicResponse, int64, error) {
	// 转换参数
	queryParams := interfaces.SkillQueryParams{
		SkillType:  params.SkillType,
		CategoryID: params.CategoryID,
		IsActive:   params.IsActive,
		Limit:      params.Limit,
		Offset:     params.Offset,
	}

	// 查询技能列表
	skills, total, err := s.skillRepo.List(ctx, queryParams)
	if err != nil {
		return nil, 0, xerrors.Wrap(err, xerrors.CodeInternalError, "查询技能列表失败")
	}

	// 批量查询分类（去重）
	categoryMap := make(map[string]string) // categoryID -> categoryName
	categoryIDs := make([]string, 0)
	for _, skill := range skills {
		if !skill.CategoryID.IsZero() {
			categoryID := skill.CategoryID.String
			if _, exists := categoryMap[categoryID]; !exists {
				categoryIDs = append(categoryIDs, categoryID)
			}
		}
	}

	// 查询分类信息
	if len(categoryIDs) > 0 {
		for _, catID := range categoryIDs {
			cat, err := s.skillCategoryRepo.GetByID(ctx, catID)
			if err == nil {
				categoryMap[catID] = cat.CategoryName
			}
		}
	}

	// 构造响应列表
	respList := make([]*SkillBasicResponse, len(skills))
	for i, skill := range skills {
		resp := &SkillBasicResponse{
			ID:           skill.ID,
			SkillCode:    skill.SkillCode,
			SkillName:    skill.SkillName,
			SkillType:    skill.SkillType,
			MaxLevel:     int(skill.MaxLevel.Int),
			CategoryID:   skill.CategoryID.String,
			CategoryName: categoryMap[skill.CategoryID.String],
			IsActive:     skill.IsActive.Bool,
		}

		if !skill.Description.IsZero() {
			resp.Description = &skill.Description.String
		}

		respList[i] = resp
	}

	return respList, total, nil
}

// ==================== 标准版响应（含 Actions 基本信息） ====================

// SkillStandardResponse 技能标准响应
type SkillStandardResponse struct {
	SkillBasicResponse
	UnlockActions []*UnlockActionInfo `json:"unlock_actions"`
}

// UnlockActionInfo 解锁动作信息
type UnlockActionInfo struct {
	ActionID           string                 `json:"action_id"`
	ActionCode         string                 `json:"action_code"`
	ActionName         string                 `json:"action_name"`
	UnlockLevel        int                    `json:"unlock_level"`
	IsDefault          bool                   `json:"is_default"`
	LevelScalingConfig map[string]interface{} `json:"level_scaling_config,omitempty"`
}

// GetSkillStandard 获取技能标准信息（含 Actions）
func (s *SkillDetailService) GetSkillStandard(ctx context.Context, skillID string) (*SkillStandardResponse, error) {
	// 1. 获取基本信息
	basicInfo, err := s.GetSkillBasic(ctx, skillID)
	if err != nil {
		return nil, err
	}

	// 2. 查询解锁动作
	unlockActions, err := s.skillUnlockActionRepo.GetBySkillID(ctx, skillID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询技能解锁动作失败")
	}

	// 3. 批量查询 Actions
	actionIDs := make([]string, len(unlockActions))
	for i, ua := range unlockActions {
		actionIDs[i] = ua.ActionID
	}

	actions, err := s.actionRepo.GetByIDs(ctx, actionIDs)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "批量查询动作失败")
	}

	// 构建 actionID -> Action 的映射
	actionMap := make(map[string]*game_config.Action)
	for _, action := range actions {
		actionMap[action.ID] = action
	}

	// 4. 构造 UnlockActionInfo
	unlockActionInfos := make([]*UnlockActionInfo, len(unlockActions))
	for i, ua := range unlockActions {
		action := actionMap[ua.ActionID]
		if action == nil {
			continue
		}

		info := &UnlockActionInfo{
			ActionID:    ua.ActionID,
			ActionCode:  action.ActionCode,
			ActionName:  action.ActionName,
			UnlockLevel: int(ua.UnlockLevel),
			IsDefault:   ua.IsDefault.Bool,
		}

		// 解析 level_scaling_config
		if !ua.LevelScalingConfig.IsZero() {
			var scalingConfig map[string]interface{}
			if err := json.Unmarshal(ua.LevelScalingConfig.JSON, &scalingConfig); err == nil {
				info.LevelScalingConfig = scalingConfig
			}
		}

		unlockActionInfos[i] = info
	}

	return &SkillStandardResponse{
		SkillBasicResponse: *basicInfo,
		UnlockActions:      unlockActionInfos,
	}, nil
}

// ==================== 完整版响应（深度关联 Actions 和 Effects） ====================

// SkillFullResponse 技能完整响应
type SkillFullResponse struct {
	SkillBasicResponse
	UnlockActions []*UnlockActionFullInfo `json:"unlock_actions"`
}

// UnlockActionFullInfo 解锁动作完整信息
type UnlockActionFullInfo struct {
	ActionID           string                 `json:"action_id"`
	ActionCode         string                 `json:"action_code"`
	ActionName         string                 `json:"action_name"`
	UnlockLevel        int                    `json:"unlock_level"`
	IsDefault          bool                   `json:"is_default"`
	LevelScalingConfig map[string]interface{} `json:"level_scaling_config,omitempty"`
	ActionDetails      *ActionDetailInfo      `json:"action_details"`
}

// ActionDetailInfo 动作详情
type ActionDetailInfo struct {
	ActionType     string                 `json:"action_type"`     // main/minor/reaction
	ActionCategory string                 `json:"action_category"` // BASIC_ATTACK/BASIC_SAVE 等
	RangeConfig    map[string]interface{} `json:"range_config"`
	TargetConfig   map[string]interface{} `json:"target_config,omitempty"`
	HitRateConfig  map[string]interface{} `json:"hit_rate_config,omitempty"`
	Description    *string                `json:"description,omitempty"`
	Effects        []*EffectInfo          `json:"effects"`
}

// EffectInfo 效果信息
type EffectInfo struct {
	EffectID           string                 `json:"effect_id"`
	EffectCode         string                 `json:"effect_code"`
	EffectName         string                 `json:"effect_name"`
	EffectType         string                 `json:"effect_type"`
	ExecutionOrder     int                    `json:"execution_order"`
	Parameters         map[string]interface{} `json:"parameters"`
	ParameterOverrides map[string]interface{} `json:"parameter_overrides,omitempty"`
}

// GetSkillFull 获取技能完整信息（深度关联）
func (s *SkillDetailService) GetSkillFull(ctx context.Context, skillID string) (*SkillFullResponse, error) {
	// 1. 获取基本信息
	basicInfo, err := s.GetSkillBasic(ctx, skillID)
	if err != nil {
		return nil, err
	}

	// 2. 查询解锁动作
	unlockActions, err := s.skillUnlockActionRepo.GetBySkillID(ctx, skillID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询技能解锁动作失败")
	}

	// 3. 批量查询 Actions
	actionIDs := make([]string, len(unlockActions))
	for i, ua := range unlockActions {
		actionIDs[i] = ua.ActionID
	}

	actions, err := s.actionRepo.GetByIDs(ctx, actionIDs)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "批量查询动作失败")
	}

	// 构建 actionID -> Action 的映射
	actionMap := make(map[string]*game_config.Action)
	for _, action := range actions {
		actionMap[action.ID] = action
	}

	// 4. 批量查询 ActionEffects
	actionEffectsMap := make(map[string][]*game_config.ActionEffect) // actionID -> []ActionEffect
	allEffectIDs := make([]string, 0)
	for _, actionID := range actionIDs {
		actionEffects, err := s.actionEffectRepo.GetByActionID(ctx, actionID)
		if err != nil {
			continue
		}
		actionEffectsMap[actionID] = actionEffects

		// 收集 effectID
		for _, ae := range actionEffects {
			allEffectIDs = append(allEffectIDs, ae.EffectID)
		}
	}

	// 5. 批量查询 Effects
	effects, err := s.effectRepo.GetByIDs(ctx, allEffectIDs)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "批量查询效果失败")
	}

	effectMap := make(map[string]*game_config.Effect)
	for _, effect := range effects {
		effectMap[effect.ID] = effect
	}

	// 6. 构造完整响应
	unlockActionFullInfos := make([]*UnlockActionFullInfo, len(unlockActions))
	for i, ua := range unlockActions {
		action := actionMap[ua.ActionID]
		if action == nil {
			continue
		}

		// 构造 ActionDetailInfo
		actionDetail := &ActionDetailInfo{
			ActionType:     action.ActionType,
			ActionCategory: action.ActionCategoryID.String, // 修正：ActionCategory 是 ID
			Effects:        make([]*EffectInfo, 0),
		}

		// 解析 RangeConfig
		if len(action.RangeConfig) > 0 {
			var rangeConfig map[string]interface{}
			if err := json.Unmarshal(action.RangeConfig, &rangeConfig); err == nil {
				actionDetail.RangeConfig = rangeConfig
			}
		}

		// 解析 TargetConfig
		if !action.TargetConfig.IsZero() {
			var targetConfig map[string]interface{}
			if err := json.Unmarshal(action.TargetConfig.JSON, &targetConfig); err == nil {
				actionDetail.TargetConfig = targetConfig
			}
		}

		// 解析 HitRateConfig
		if !action.HitRateConfig.IsZero() {
			var hitRateConfig map[string]interface{}
			if err := json.Unmarshal(action.HitRateConfig.JSON, &hitRateConfig); err == nil {
				actionDetail.HitRateConfig = hitRateConfig
			}
		}

		// 描述
		if !action.Description.IsZero() {
			actionDetail.Description = &action.Description.String
		}

		// 构造 Effects
		actionEffects := actionEffectsMap[ua.ActionID]
		for _, ae := range actionEffects {
			effect := effectMap[ae.EffectID]
			if effect == nil {
				continue
			}

			effectInfo := &EffectInfo{
				EffectID:       effect.ID,
				EffectCode:     effect.EffectCode,
				EffectName:     effect.EffectName,
				EffectType:     effect.EffectType,
				ExecutionOrder: int(ae.ExecutionOrder.Int),
			}

			// 解析 Effect Parameters
			if len(effect.Parameters) > 0 {
				var params map[string]interface{}
				if err := json.Unmarshal(effect.Parameters, &params); err == nil {
					effectInfo.Parameters = params
				}
			}

			// 解析 ParameterOverrides
			if !ae.ParameterOverrides.IsZero() {
				var overrides map[string]interface{}
				if err := json.Unmarshal(ae.ParameterOverrides.JSON, &overrides); err == nil {
					effectInfo.ParameterOverrides = overrides
				}
			}

			actionDetail.Effects = append(actionDetail.Effects, effectInfo)
		}

		// 构造 UnlockActionFullInfo
		info := &UnlockActionFullInfo{
			ActionID:      ua.ActionID,
			ActionCode:    action.ActionCode,
			ActionName:    action.ActionName,
			UnlockLevel:   int(ua.UnlockLevel),
			IsDefault:     ua.IsDefault.Bool,
			ActionDetails: actionDetail,
		}

		// 解析 level_scaling_config
		if !ua.LevelScalingConfig.IsZero() {
			var scalingConfig map[string]interface{}
			if err := json.Unmarshal(ua.LevelScalingConfig.JSON, &scalingConfig); err == nil {
				info.LevelScalingConfig = scalingConfig
			}
		}

		unlockActionFullInfos[i] = info
	}

	return &SkillFullResponse{
		SkillBasicResponse: *basicInfo,
		UnlockActions:      unlockActionFullInfos,
	}, nil
}

// ListSkillsStandard 获取技能列表（标准版）
func (s *SkillDetailService) ListSkillsStandard(ctx context.Context, params SkillQueryParams) ([]*SkillStandardResponse, int64, error) {
	// 转换参数
	queryParams := interfaces.SkillQueryParams{
		SkillType:  params.SkillType,
		CategoryID: params.CategoryID,
		IsActive:   params.IsActive,
		Limit:      params.Limit,
		Offset:     params.Offset,
	}

	// 查询技能列表
	skills, total, err := s.skillRepo.List(ctx, queryParams)
	if err != nil {
		return nil, 0, xerrors.Wrap(err, xerrors.CodeInternalError, "查询技能列表失败")
	}

	// 批量获取标准信息（避免 N+1）
	respList := make([]*SkillStandardResponse, len(skills))
	for i, skill := range skills {
		standardInfo, err := s.GetSkillStandard(ctx, skill.ID)
		if err != nil {
			// 降级为基本信息
			basicInfo, _ := s.GetSkillBasic(ctx, skill.ID)
			if basicInfo != nil {
				respList[i] = &SkillStandardResponse{
					SkillBasicResponse: *basicInfo,
					UnlockActions:      make([]*UnlockActionInfo, 0),
				}
			}
			continue
		}
		respList[i] = standardInfo
	}

	return respList, total, nil
}
