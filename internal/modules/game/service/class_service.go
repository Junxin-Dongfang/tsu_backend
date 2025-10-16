package service

import (
	"context"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/interfaces"
)

// ClassService 职业服务
type ClassService struct {
	classRepo                interfaces.ClassRepository
	classAdvancedReqRepo     interfaces.ClassAdvancedRequirementRepository
}

// NewClassService 创建职业服务
func NewClassService(
	classRepo interfaces.ClassRepository,
	classAdvancedReqRepo interfaces.ClassAdvancedRequirementRepository,
) *ClassService {
	return &ClassService{
		classRepo:            classRepo,
		classAdvancedReqRepo: classAdvancedReqRepo,
	}
}

// GetClassByID 根据ID获取职业
func (s *ClassService) GetClassByID(ctx context.Context, classID string) (*game_config.Class, error) {
	return s.classRepo.GetByID(ctx, classID)
}

// GetClassByCode 根据职业代码获取职业
func (s *ClassService) GetClassByCode(ctx context.Context, classCode string) (*game_config.Class, error) {
	return s.classRepo.GetByCode(ctx, classCode)
}

// GetClassList 获取职业列表
func (s *ClassService) GetClassList(ctx context.Context, params interfaces.ClassQueryParams) ([]*game_config.Class, int64, error) {
	return s.classRepo.List(ctx, params)
}

// GetAdvancementOptions 获取职业的可进阶选项
func (s *ClassService) GetAdvancementOptions(ctx context.Context, fromClassID string) ([]*game_config.ClassAdvancedRequirement, error) {
	return s.classAdvancedReqRepo.GetAdvancementOptions(ctx, fromClassID)
}
