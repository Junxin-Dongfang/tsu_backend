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

// ClassAdvancedRequirementService 职业进阶要求服务
type ClassAdvancedRequirementService struct {
	repo      interfaces.ClassAdvancedRequirementRepository
	classRepo interfaces.ClassRepository
}

// NewClassAdvancedRequirementService 创建职业进阶要求服务
func NewClassAdvancedRequirementService(db *sql.DB) *ClassAdvancedRequirementService {
	return &ClassAdvancedRequirementService{
		repo:      impl.NewClassAdvancedRequirementRepository(db),
		classRepo: impl.NewClassRepository(db),
	}
}

// GetByID 根据ID获取进阶要求
func (s *ClassAdvancedRequirementService) GetByID(ctx context.Context, id string) (*game_config.ClassAdvancedRequirement, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByFromClass 获取指定源职业的所有进阶路径
func (s *ClassAdvancedRequirementService) GetByFromClass(ctx context.Context, fromClassID string) ([]*game_config.ClassAdvancedRequirement, error) {
	// 验证职业是否存在
	if _, err := s.classRepo.GetByID(ctx, fromClassID); err != nil {
		return nil, xerrors.NewNotFoundError("Class", fromClassID)
	}
	return s.repo.GetByFromClass(ctx, fromClassID)
}

// GetByToClass 获取可以进阶到指定职业的所有路径
func (s *ClassAdvancedRequirementService) GetByToClass(ctx context.Context, toClassID string) ([]*game_config.ClassAdvancedRequirement, error) {
	// 验证职业是否存在
	if _, err := s.classRepo.GetByID(ctx, toClassID); err != nil {
		return nil, xerrors.NewNotFoundError("Class", toClassID)
	}
	return s.repo.GetByToClass(ctx, toClassID)
}

// List 获取进阶要求列表
func (s *ClassAdvancedRequirementService) List(ctx context.Context, params interfaces.ListAdvancedRequirementsParams) ([]*game_config.ClassAdvancedRequirement, int64, error) {
	return s.repo.List(ctx, params)
}

// CreateInput 创建进阶要求输入
type CreateAdvancedRequirementInput struct {
	FromClassID            string
	ToClassID              string
	RequiredLevel          int
	RequiredHonor          int
	RequiredJobChangeCount int
	RequiredAttributes     []byte // JSONB
	RequiredSkills         []byte // JSONB
	RequiredItems          []byte // JSONB
	IsActive               bool
	DisplayOrder           int16
}

// Create 创建进阶要求
func (s *ClassAdvancedRequirementService) Create(ctx context.Context, input CreateAdvancedRequirementInput) (*game_config.ClassAdvancedRequirement, error) {
	// 验证源职业和目标职业是否存在
	fromClass, err := s.classRepo.GetByID(ctx, input.FromClassID)
	if err != nil {
		return nil, xerrors.NewNotFoundError("FromClass", input.FromClassID)
	}
	toClass, err := s.classRepo.GetByID(ctx, input.ToClassID)
	if err != nil {
		return nil, xerrors.NewNotFoundError("ToClass", input.ToClassID)
	}

	// 业务验证
	if input.FromClassID == input.ToClassID {
		return nil, xerrors.NewValidationError("to_class_id", "进阶的源职业和目标职业不能相同")
	}

	// Tier验证：只能往更高级的Tier进阶
	tierOrder := map[string]int{
		"basic":     1,
		"advanced":  2,
		"elite":     3,
		"legendary": 4,
		"mythic":    5,
	}
	fromTierLevel := tierOrder[fromClass.Tier]
	toTierLevel := tierOrder[toClass.Tier]
	if toTierLevel <= fromTierLevel {
		return nil, xerrors.NewValidationError("to_class_id",
			fmt.Sprintf("目标职业等级(%s)必须高于源职业等级(%s)", toClass.Tier, fromClass.Tier))
	}

	// 检查是否已存在相同的进阶关系
	existing, _ := s.repo.GetByClassPair(ctx, input.FromClassID, input.ToClassID)
	if existing != nil {
		return nil, xerrors.NewValidationError("from_class_id", "该进阶关系已存在")
	}

	// 创建实体
	requirement := &game_config.ClassAdvancedRequirement{
		FromClassID:            input.FromClassID,
		ToClassID:              input.ToClassID,
		RequiredLevel:          input.RequiredLevel,
		RequiredHonor:          input.RequiredHonor,
		RequiredJobChangeCount: input.RequiredJobChangeCount,
	}
	requirement.IsActive.SetValid(input.IsActive)
	requirement.DisplayOrder.SetValid(input.DisplayOrder)

	// 设置JSONB字段
	if len(input.RequiredAttributes) > 0 {
		requirement.RequiredAttributes.UnmarshalJSON(input.RequiredAttributes)
	}
	if len(input.RequiredSkills) > 0 {
		requirement.RequiredSkills.UnmarshalJSON(input.RequiredSkills)
	}
	if len(input.RequiredItems) > 0 {
		requirement.RequiredItems.UnmarshalJSON(input.RequiredItems)
	}

	if err := s.repo.Create(ctx, requirement); err != nil {
		return nil, err
	}

	return requirement, nil
}

// Update 更新进阶要求
func (s *ClassAdvancedRequirementService) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	// 验证记录是否存在
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 如果更新了from_class_id或to_class_id，需要验证职业是否存在
	if fromClassID, ok := updates["from_class_id"].(string); ok {
		if _, err := s.classRepo.GetByID(ctx, fromClassID); err != nil {
			return xerrors.NewNotFoundError("FromClass", fromClassID)
		}
	}
	if toClassID, ok := updates["to_class_id"].(string); ok {
		if _, err := s.classRepo.GetByID(ctx, toClassID); err != nil {
			return xerrors.NewNotFoundError("ToClass", toClassID)
		}
	}

	return s.repo.Update(ctx, id, updates)
}

// Delete 删除进阶要求
func (s *ClassAdvancedRequirementService) Delete(ctx context.Context, id string) error {
	// 验证记录是否存在
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, id)
}

// BatchCreate 批量创建进阶要求
func (s *ClassAdvancedRequirementService) BatchCreate(ctx context.Context, inputs []CreateAdvancedRequirementInput) ([]*game_config.ClassAdvancedRequirement, error) {
	requirements := make([]*game_config.ClassAdvancedRequirement, 0, len(inputs))

	for _, input := range inputs {
		// 验证每个输入
		_, err := s.classRepo.GetByID(ctx, input.FromClassID)
		if err != nil {
			return nil, xerrors.NewNotFoundError("FromClass", input.FromClassID)
		}
		_, err = s.classRepo.GetByID(ctx, input.ToClassID)
		if err != nil {
			return nil, xerrors.NewNotFoundError("ToClass", input.ToClassID)
		}

		requirement := &game_config.ClassAdvancedRequirement{
			FromClassID:            input.FromClassID,
			ToClassID:              input.ToClassID,
			RequiredLevel:          input.RequiredLevel,
			RequiredHonor:          input.RequiredHonor,
			RequiredJobChangeCount: input.RequiredJobChangeCount,
		}
		requirement.IsActive.SetValid(input.IsActive)
		requirement.DisplayOrder.SetValid(input.DisplayOrder)

		if len(input.RequiredAttributes) > 0 {
			requirement.RequiredAttributes.UnmarshalJSON(input.RequiredAttributes)
		}
		if len(input.RequiredSkills) > 0 {
			requirement.RequiredSkills.UnmarshalJSON(input.RequiredSkills)
		}
		if len(input.RequiredItems) > 0 {
			requirement.RequiredItems.UnmarshalJSON(input.RequiredItems)
		}

		requirements = append(requirements, requirement)
	}

	if err := s.repo.BatchCreate(ctx, requirements); err != nil {
		return nil, err
	}

	return requirements, nil
}

// GetAdvancementPaths 获取完整进阶路径树
func (s *ClassAdvancedRequirementService) GetAdvancementPaths(ctx context.Context, fromClassID string, maxDepth int) ([][]*game_config.ClassAdvancedRequirement, error) {
	// 验证职业是否存在
	if _, err := s.classRepo.GetByID(ctx, fromClassID); err != nil {
		return nil, xerrors.NewNotFoundError("Class", fromClassID)
	}

	return s.repo.GetAdvancementPaths(ctx, fromClassID, maxDepth)
}
