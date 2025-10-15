package service

import (
	"context"
	"database/sql"
	"fmt"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// DamageTypeService 伤害类型服务
type DamageTypeService struct {
	repo interfaces.DamageTypeRepository
}

// NewDamageTypeService 创建伤害类型服务
func NewDamageTypeService(db *sql.DB) *DamageTypeService {
	return &DamageTypeService{
		repo: impl.NewDamageTypeRepository(db),
	}
}

// GetDamageTypes 获取伤害类型列表
func (s *DamageTypeService) GetDamageTypes(ctx context.Context, params interfaces.DamageTypeQueryParams) ([]*game_config.DamageType, int64, error) {
	return s.repo.List(ctx, params)
}

// GetDamageTypeByID 根据ID获取伤害类型
func (s *DamageTypeService) GetDamageTypeByID(ctx context.Context, damageTypeID string) (*game_config.DamageType, error) {
	return s.repo.GetByID(ctx, damageTypeID)
}

// CreateDamageType 创建伤害类型
func (s *DamageTypeService) CreateDamageType(ctx context.Context, damageType *game_config.DamageType) error {
	// 业务验证：检查代码是否已存在
	exists, err := s.repo.Exists(ctx, damageType.Code)
	if err != nil {
		return err
	}
	if exists {
		return xerrors.New(xerrors.CodeDuplicateResource, fmt.Sprintf("伤害类型代码已存在: %s", damageType.Code))
	}

	return s.repo.Create(ctx, damageType)
}

// UpdateDamageType 更新伤害类型信息
func (s *DamageTypeService) UpdateDamageType(ctx context.Context, damageTypeID string, updates map[string]interface{}) error {
	// 获取伤害类型
	damageType, err := s.repo.GetByID(ctx, damageTypeID)
	if err != nil {
		return err
	}

	// 更新字段
	if code, ok := updates["code"].(string); ok && code != "" {
		// 检查代码是否已被使用
		existing, err := s.repo.GetByCode(ctx, code)
		if err == nil && existing.ID != damageTypeID {
			return xerrors.New(xerrors.CodeDuplicateResource, fmt.Sprintf("伤害类型代码已被使用: %s", code))
		}
		damageType.Code = code
	}

	if name, ok := updates["name"].(string); ok && name != "" {
		damageType.Name = name
	}

	if category, ok := updates["category"].(string); ok {
		if category != "" {
			damageType.Category.SetValid(category)
		} else {
			damageType.Category.Valid = false
		}
	}

	if resistanceAttrCode, ok := updates["resistance_attribute_code"].(string); ok {
		if resistanceAttrCode != "" {
			damageType.ResistanceAttributeCode.SetValid(resistanceAttrCode)
		} else {
			damageType.ResistanceAttributeCode.Valid = false
		}
	}

	if damageReductionAttrCode, ok := updates["damage_reduction_attribute_code"].(string); ok {
		if damageReductionAttrCode != "" {
			damageType.DamageReductionAttributeCode.SetValid(damageReductionAttrCode)
		} else {
			damageType.DamageReductionAttributeCode.Valid = false
		}
	}

	if resistanceCap, ok := updates["resistance_cap"].(int); ok {
		damageType.ResistanceCap.SetValid(resistanceCap)
	}

	if color, ok := updates["color"].(string); ok {
		if color != "" {
			damageType.Color.SetValid(color)
		} else {
			damageType.Color.Valid = false
		}
	}

	if icon, ok := updates["icon"].(string); ok {
		if icon != "" {
			damageType.Icon.SetValid(icon)
		} else {
			damageType.Icon.Valid = false
		}
	}

	if description, ok := updates["description"].(string); ok {
		if description != "" {
			damageType.Description.SetValid(description)
		} else {
			damageType.Description.Valid = false
		}
	}

	if isActive, ok := updates["is_active"].(bool); ok {
		damageType.IsActive.SetValid(isActive)
	}

	// 保存更新
	return s.repo.Update(ctx, damageType)
}

// DeleteDamageType 删除伤害类型
func (s *DamageTypeService) DeleteDamageType(ctx context.Context, damageTypeID string) error {
	return s.repo.Delete(ctx, damageTypeID)
}
