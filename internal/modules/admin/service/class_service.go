package service

import (
	"context"
	"database/sql"
	"fmt"
	"tsu-self/internal/pkg/xerrors"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// ClassService 职业服务
type ClassService struct {
	classRepo          interfaces.ClassRepository
	attributeBonusRepo interfaces.AttributeBonusRepository
}

// NewClassService 创建职业服务
func NewClassService(db *sql.DB) *ClassService {
	return &ClassService{
		classRepo:          impl.NewClassRepository(db),
		attributeBonusRepo: impl.NewAttributeBonusRepository(db),
	}
}

// GetClasses 获取职业列表
func (s *ClassService) GetClasses(ctx context.Context, params interfaces.ClassQueryParams) ([]*game_config.Class, int64, error) {
	return s.classRepo.List(ctx, params)
}

// GetClassByID 根据ID获取职业
func (s *ClassService) GetClassByID(ctx context.Context, classID string) (*game_config.Class, error) {
	return s.classRepo.GetByID(ctx, classID)
}

// CreateClass 创建职业
func (s *ClassService) CreateClass(ctx context.Context, class *game_config.Class) error {
	// 业务验证：检查职业代码是否已存在
	existing, err := s.classRepo.GetByCode(ctx, class.ClassCode)
	if err == nil && existing != nil {
		return xerrors.New(xerrors.CodeDuplicateResource, fmt.Sprintf("职业代码已存在: %s", class.ClassCode))
	}

	return s.classRepo.Create(ctx, class)
}

// UpdateClass 更新职业信息
func (s *ClassService) UpdateClass(ctx context.Context, classID string, updates map[string]interface{}) error {
	// 获取职业
	class, err := s.classRepo.GetByID(ctx, classID)
	if err != nil {
		return err
	}

	// 更新字段
	if classCode, ok := updates["class_code"].(string); ok && classCode != "" {
		// 检查职业代码是否已被使用
		existing, err := s.classRepo.GetByCode(ctx, classCode)
		if err == nil && existing.ID != classID {
			return xerrors.New(xerrors.CodeDuplicateResource, fmt.Sprintf("职业代码已被使用: %s", classCode))
		}
		class.ClassCode = classCode
	}

	if className, ok := updates["class_name"].(string); ok && className != "" {
		class.ClassName = className
	}

	if description, ok := updates["description"].(string); ok {
		if description != "" {
			class.Description.SetValid(description)
		} else {
			class.Description.Valid = false
		}
	}

	if loreText, ok := updates["lore_text"].(string); ok {
		if loreText != "" {
			class.LoreText.SetValid(loreText)
		} else {
			class.LoreText.Valid = false
		}
	}

	if specialty, ok := updates["specialty"].(string); ok {
		if specialty != "" {
			class.Specialty.SetValid(specialty)
		} else {
			class.Specialty.Valid = false
		}
	}

	if playstyle, ok := updates["playstyle"].(string); ok {
		if playstyle != "" {
			class.Playstyle.SetValid(playstyle)
		} else {
			class.Playstyle.Valid = false
		}
	}

	if tier, ok := updates["tier"].(string); ok && tier != "" {
		class.Tier = tier
	}

	if promotionCount, ok := updates["promotion_count"].(int16); ok {
		class.PromotionCount.SetValid(promotionCount)
	}

	if icon, ok := updates["icon"].(string); ok {
		if icon != "" {
			class.Icon.SetValid(icon)
		} else {
			class.Icon.Valid = false
		}
	}

	if color, ok := updates["color"].(string); ok {
		if color != "" {
			class.Color.SetValid(color)
		} else {
			class.Color.Valid = false
		}
	}

	if isActive, ok := updates["is_active"].(bool); ok {
		class.IsActive.SetValid(isActive)
	}

	if isVisible, ok := updates["is_visible"].(bool); ok {
		class.IsVisible.SetValid(isVisible)
	}

	if displayOrder, ok := updates["display_order"].(int16); ok {
		class.DisplayOrder.SetValid(displayOrder)
	}

	// 保存更新
	return s.classRepo.Update(ctx, class)
}

// DeleteClass 删除职业
func (s *ClassService) DeleteClass(ctx context.Context, classID string) error {
	return s.classRepo.Delete(ctx, classID)
}

// ==================== 属性加成管理 ====================

// GetClassAttributeBonuses 获取职业的所有属性加成
func (s *ClassService) GetClassAttributeBonuses(ctx context.Context, classID string) ([]*game_config.ClassAttributeBonuse, error) {
	// 验证职业是否存在
	if _, err := s.classRepo.GetByID(ctx, classID); err != nil {
		return nil, err
	}

	return s.attributeBonusRepo.GetByClassID(ctx, classID)
}

// GetAttributeBonusByID 根据ID获取属性加成
func (s *ClassService) GetAttributeBonusByID(ctx context.Context, bonusID string) (*game_config.ClassAttributeBonuse, error) {
	return s.attributeBonusRepo.GetByID(ctx, bonusID)
}

// CreateAttributeBonus 创建属性加成
func (s *ClassService) CreateAttributeBonus(ctx context.Context, classID string, bonus *game_config.ClassAttributeBonuse) error {
	// 验证职业是否存在
	if _, err := s.classRepo.GetByID(ctx, classID); err != nil {
		return err
	}

	// 验证是否已存在相同的职业-属性组合
	exists, err := s.attributeBonusRepo.Exists(ctx, classID, bonus.AttributeID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("该职业已配置此属性的加成")
	}

	bonus.ClassID = classID
	return s.attributeBonusRepo.Create(ctx, bonus)
}

// UpdateAttributeBonus 更新属性加成
func (s *ClassService) UpdateAttributeBonus(ctx context.Context, bonusID string, updates map[string]interface{}) error {
	// 获取属性加成
	bonus, err := s.attributeBonusRepo.GetByID(ctx, bonusID)
	if err != nil {
		return err
	}

	// 更新字段
	if baseBonusValue, ok := updates["base_bonus_value"].(string); ok {
		if err := bonus.BaseBonusValue.UnmarshalText([]byte(baseBonusValue)); err != nil {
			return fmt.Errorf("base_bonus_value 格式错误: %w", err)
		}
	}

	if bonusPerLevel, ok := updates["bonus_per_level"].(bool); ok {
		bonus.BonusPerLevel = bonusPerLevel
	}

	if perLevelBonusValue, ok := updates["per_level_bonus_value"].(string); ok {
		if err := bonus.PerLevelBonusValue.UnmarshalText([]byte(perLevelBonusValue)); err != nil {
			return fmt.Errorf("per_level_bonus_value 格式错误: %w", err)
		}
	}

	return s.attributeBonusRepo.Update(ctx, bonus)
}

// DeleteAttributeBonus 删除属性加成
func (s *ClassService) DeleteAttributeBonus(ctx context.Context, bonusID string) error {
	return s.attributeBonusRepo.Delete(ctx, bonusID)
}

// BatchSetAttributeBonuses 批量设置职业的属性加成（先删除旧的，再批量创建新的）
func (s *ClassService) BatchSetAttributeBonuses(ctx context.Context, classID string, bonuses []*game_config.ClassAttributeBonuse) error {
	// 验证职业是否存在
	if _, err := s.classRepo.GetByID(ctx, classID); err != nil {
		return err
	}

	// 先删除旧的
	if err := s.attributeBonusRepo.DeleteByClassID(ctx, classID); err != nil {
		return err
	}

	// 设置 class_id
	for _, bonus := range bonuses {
		bonus.ClassID = classID
	}

	// 批量创建新的
	return s.attributeBonusRepo.BatchCreate(ctx, bonuses)
}
