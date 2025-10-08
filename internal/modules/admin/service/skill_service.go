package service

import (
	"context"
	"database/sql"
	"fmt"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// SkillService 技能服务
type SkillService struct {
	repo interfaces.SkillRepository
}

// NewSkillService 创建技能服务
func NewSkillService(db *sql.DB) *SkillService {
	return &SkillService{
		repo: impl.NewSkillRepository(db),
	}
}

// GetSkills 获取技能列表
func (s *SkillService) GetSkills(ctx context.Context, params interfaces.SkillQueryParams) ([]*game_config.Skill, int64, error) {
	return s.repo.List(ctx, params)
}

// GetSkillByID 根据ID获取技能
func (s *SkillService) GetSkillByID(ctx context.Context, skillID string) (*game_config.Skill, error) {
	return s.repo.GetByID(ctx, skillID)
}

// CreateSkill 创建技能
func (s *SkillService) CreateSkill(ctx context.Context, skill *game_config.Skill) error {
	// 业务验证：检查技能代码是否已存在
	exists, err := s.repo.Exists(ctx, skill.SkillCode)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("技能代码已存在: %s", skill.SkillCode)
	}

	return s.repo.Create(ctx, skill)
}

// UpdateSkill 更新技能信息
func (s *SkillService) UpdateSkill(ctx context.Context, skillID string, updates map[string]interface{}) error {
	// 获取技能
	skill, err := s.repo.GetByID(ctx, skillID)
	if err != nil {
		return err
	}

	// 更新字段
	if skillCode, ok := updates["skill_code"].(string); ok && skillCode != "" {
		// 检查技能代码是否已被使用
		existing, err := s.repo.GetByCode(ctx, skillCode)
		if err == nil && existing.ID != skillID {
			return fmt.Errorf("技能代码已被使用: %s", skillCode)
		}
		skill.SkillCode = skillCode
	}

	if skillName, ok := updates["skill_name"].(string); ok && skillName != "" {
		skill.SkillName = skillName
	}

	if skillType, ok := updates["skill_type"].(string); ok && skillType != "" {
		skill.SkillType = skillType
	}

	if categoryID, ok := updates["category_id"].(string); ok {
		if categoryID != "" {
			skill.CategoryID.SetValid(categoryID)
		} else {
			skill.CategoryID.Valid = false
		}
	}

	if maxLevel, ok := updates["max_level"].(int); ok {
		skill.MaxLevel.SetValid(maxLevel)
	}

	if description, ok := updates["description"].(string); ok {
		if description != "" {
			skill.Description.SetValid(description)
		} else {
			skill.Description.Valid = false
		}
	}

	if detailedDescription, ok := updates["detailed_description"].(string); ok {
		if detailedDescription != "" {
			skill.DetailedDescription.SetValid(detailedDescription)
		} else {
			skill.DetailedDescription.Valid = false
		}
	}

	if icon, ok := updates["icon"].(string); ok {
		if icon != "" {
			skill.Icon.SetValid(icon)
		} else {
			skill.Icon.Valid = false
		}
	}

	if isActive, ok := updates["is_active"].(bool); ok {
		skill.IsActive.SetValid(isActive)
	}

	// 保存更新
	return s.repo.Update(ctx, skill)
}

// DeleteSkill 删除技能
func (s *SkillService) DeleteSkill(ctx context.Context, skillID string) error {
	return s.repo.Delete(ctx, skillID)
}
