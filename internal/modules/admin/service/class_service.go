package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	apiReqAdmin "tsu-self/internal/api/model/request/admin"
	apiRespAdmin "tsu-self/internal/api/model/response/admin"
	"tsu-self/internal/converter/admin"
	"tsu-self/internal/entity"
	"tsu-self/internal/repository/interfaces"
	"tsu-self/internal/repository/query"
)

type ClassService struct {
	classRepo interfaces.ClassRepository
}

func NewClassService(classRepo interfaces.ClassRepository) *ClassService {
	return &ClassService{
		classRepo: classRepo,
	}
}

// CreateClass 创建职业
func (s *ClassService) CreateClass(ctx context.Context, req *apiReqAdmin.CreateClassRequest) (*apiRespAdmin.Class, error) {
	// 检查职业代码是否已存在
	existing, err := s.classRepo.GetByCode(ctx, req.Code)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("职业代码 %s 已存在", req.Code)
	}

	// 转换为实体并创建
	class := admin.ConvertCreateRequestToClass(req)
	if err := s.classRepo.Create(ctx, class); err != nil {
		return nil, fmt.Errorf("创建职业失败: %w", err)
	}

	return admin.ConvertClassToResponse(class), nil
}

// GetClass 获取职业详情
func (s *ClassService) GetClass(ctx context.Context, id uuid.UUID) (*apiRespAdmin.Class, error) {
	class, err := s.classRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取职业失败: %w", err)
	}

	return admin.ConvertClassToResponse(class), nil
}

// GetClassWithStats 获取带统计信息的职业详情
func (s *ClassService) GetClassWithStats(ctx context.Context, id uuid.UUID) (*apiRespAdmin.ClassWithStats, error) {
	class, err := s.classRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取职业失败: %w", err)
	}

	stats, err := s.classRepo.GetHeroStats(ctx, id)
	if err != nil {
		// 统计信息获取失败不影响主要数据返回
		stats = &query.ClassHeroStats{ClassID: id}
	}

	return admin.ConvertToClassWithStats(class, stats), nil
}

// UpdateClass 更新职业
func (s *ClassService) UpdateClass(ctx context.Context, id uuid.UUID, req *apiReqAdmin.UpdateClassRequest) (*apiRespAdmin.Class, error) {
	class, err := s.classRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取职业失败: %w", err)
	}

	// 如果更新职业代码，检查是否重复
	if req.Code != nil && *req.Code != class.ClassCode {
		existing, err := s.classRepo.GetByCode(ctx, *req.Code)
		if err == nil && existing != nil {
			return nil, fmt.Errorf("职业代码 %s 已存在", *req.Code)
		}
	}

	admin.UpdateClassFromRequest(class, req)
	if err := s.classRepo.Update(ctx, class); err != nil {
		return nil, fmt.Errorf("更新职业失败: %w", err)
	}

	return admin.ConvertClassToResponse(class), nil
}

// DeleteClass 删除职业
func (s *ClassService) DeleteClass(ctx context.Context, id uuid.UUID) error {
	if err := s.classRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("删除职业失败: %w", err)
	}
	return nil
}

// ListClasses 获取职业列表
func (s *ClassService) ListClasses(ctx context.Context, req *apiReqAdmin.ClassListRequest) (*apiRespAdmin.ClassListResponse, error) {
	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.SortBy == nil || *req.SortBy == "" {
		sortBy := "display_order"
		req.SortBy = &sortBy
	}
	if req.SortOrder == nil || *req.SortOrder == "" {
		sortOrder := "asc"
		req.SortOrder = &sortOrder
	}

	params := &query.ClassListParams{
		Tier:      req.Tier,
		IsActive:  req.IsActive,
		IsHidden:  req.IsHidden,
		Search:    req.Search,
		SortBy:    *req.SortBy,
		SortOrder: *req.SortOrder,
		Page:      req.Page,
		PageSize:  req.PageSize,
	}

	classes, total, err := s.classRepo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("查询职业列表失败: %w", err)
	}

	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}

	return &apiRespAdmin.ClassListResponse{
		Data: admin.ConvertClassesToResponse(classes),
		Pagination: apiRespAdmin.PaginationResponse{
			Page:       req.Page,
			PageSize:   req.PageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// GetClassHeroStats 获取职业英雄统计
func (s *ClassService) GetClassHeroStats(ctx context.Context, id uuid.UUID) (*apiRespAdmin.ClassHeroStats, error) {
	stats, err := s.classRepo.GetHeroStats(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取职业统计失败: %w", err)
	}

	return admin.ConvertToClassHeroStatsResponse(stats), nil
}

// CreateClassAttributeBonus 创建职业属性加成
func (s *ClassService) CreateClassAttributeBonus(ctx context.Context, classID uuid.UUID, req *apiReqAdmin.CreateClassAttributeBonusRequest) (*apiRespAdmin.ClassAttributeBonus, error) {
	// 检查职业是否存在
	if _, err := s.classRepo.GetByID(ctx, classID); err != nil {
		return nil, fmt.Errorf("职业不存在: %w", err)
	}

	// 检查是否已存在相同属性的加成
	existing, err := s.classRepo.GetAttributeBonus(ctx, classID, req.AttributeID)
	if err == nil && existing != nil {
		return nil, errors.New("该职业已存在此属性的加成配置")
	}

	bonus := admin.ConvertToAttributeBonusEntity(classID.String(), req)
	if err := s.classRepo.CreateAttributeBonus(ctx, bonus); err != nil {
		return nil, fmt.Errorf("创建属性加成失败: %w", err)
	}

	// 获取详细信息返回
	bonuses, err := s.classRepo.GetAttributeBonuses(ctx, classID)
	if err != nil {
		return nil, fmt.Errorf("获取属性加成详情失败: %w", err)
	}

	for _, b := range bonuses {
		if b.AttributeID == req.AttributeID {
			return admin.ConvertAttributeBonusToResponse(b), nil
		}
	}

	return nil, errors.New("创建成功但无法获取详情")
}

// GetClassAttributeBonuses 获取职业属性加成列表
func (s *ClassService) GetClassAttributeBonuses(ctx context.Context, classID uuid.UUID) ([]*apiRespAdmin.ClassAttributeBonus, error) {
	bonuses, err := s.classRepo.GetAttributeBonuses(ctx, classID)
	if err != nil {
		return nil, fmt.Errorf("获取属性加成列表失败: %w", err)
	}

	return admin.ConvertAttributeBonusesToResponse(bonuses), nil
}

// UpdateClassAttributeBonus 更新职业属性加成
func (s *ClassService) UpdateClassAttributeBonus(ctx context.Context, classID, attributeID uuid.UUID, req *apiReqAdmin.UpdateClassAttributeBonusRequest) (*apiRespAdmin.ClassAttributeBonus, error) {
	bonus, err := s.classRepo.GetAttributeBonus(ctx, classID, attributeID)
	if err != nil {
		return nil, fmt.Errorf("属性加成不存在: %w", err)
	}

	admin.UpdateAttributeBonusFromRequest(bonus, req)
	if err := s.classRepo.UpdateAttributeBonus(ctx, bonus); err != nil {
		return nil, fmt.Errorf("更新属性加成失败: %w", err)
	}

	// 获取详细信息返回
	bonuses, err := s.classRepo.GetAttributeBonuses(ctx, classID)
	if err != nil {
		return nil, fmt.Errorf("获取属性加成详情失败: %w", err)
	}

	for _, b := range bonuses {
		if b.AttributeID == attributeID {
			return admin.ConvertAttributeBonusToResponse(b), nil
		}
	}

	return nil, errors.New("更新成功但无法获取详情")
}

// DeleteClassAttributeBonus 删除职业属性加成
func (s *ClassService) DeleteClassAttributeBonus(ctx context.Context, classID, attributeID uuid.UUID) error {
	bonus, err := s.classRepo.GetAttributeBonus(ctx, classID, attributeID)
	if err != nil {
		return fmt.Errorf("属性加成不存在: %w", err)
	}

	bonusUUID, _ := uuid.Parse(bonus.ID)
	if err := s.classRepo.DeleteAttributeBonus(ctx, bonusUUID); err != nil {
		return fmt.Errorf("删除属性加成失败: %w", err)
	}

	return nil
}

// BatchCreateClassAttributeBonuses 批量创建职业属性加成
func (s *ClassService) BatchCreateClassAttributeBonuses(ctx context.Context, classID uuid.UUID, req *apiReqAdmin.BatchCreateClassAttributeBonusRequest) ([]*apiRespAdmin.ClassAttributeBonus, error) {
	// 检查职业是否存在
	if _, err := s.classRepo.GetByID(ctx, classID); err != nil {
		return nil, fmt.Errorf("职业不存在: %w", err)
	}

	// 转换为实体
	bonuses := make([]*entity.ClassAttributeBonuse, len(req.Bonuses))
	for i, bonusReq := range req.Bonuses {
		bonuses[i] = admin.ConvertToAttributeBonusEntity(classID.String(), &bonusReq)
	}

	// 批量创建
	if err := s.classRepo.BatchCreateAttributeBonuses(ctx, bonuses); err != nil {
		return nil, fmt.Errorf("批量创建属性加成失败: %w", err)
	}

	// 返回完整列表
	return s.GetClassAttributeBonuses(ctx, classID)
}

// CreateClassAdvancementRequirement 创建职业进阶要求
func (s *ClassService) CreateClassAdvancementRequirement(ctx context.Context, req *apiReqAdmin.CreateClassAdvancementRequest) (*apiRespAdmin.ClassAdvancementRequirement, error) {
	// 检查源职业和目标职业是否存在
	if _, err := s.classRepo.GetByID(ctx, req.FromClassID); err != nil {
		return nil, fmt.Errorf("源职业不存在: %w", err)
	}
	if _, err := s.classRepo.GetByID(ctx, req.ToClassID); err != nil {
		return nil, fmt.Errorf("目标职业不存在: %w", err)
	}

	// 检查是否已存在相同的进阶要求
	existing, err := s.classRepo.GetAdvancementRequirement(ctx, req.FromClassID, req.ToClassID)
	if err == nil && existing != nil {
		return nil, errors.New("该进阶路径已存在")
	}

	requirement := admin.ConvertToAdvancementRequirementEntity(req)
	if err := s.classRepo.CreateAdvancementRequirement(ctx, requirement); err != nil {
		return nil, fmt.Errorf("创建进阶要求失败: %w", err)
	}

	// 获取详细信息返回
	requirements, err := s.classRepo.GetAdvancementRequirements(ctx, req.FromClassID)
	if err != nil {
		return nil, fmt.Errorf("获取进阶要求详情失败: %w", err)
	}

	for _, r := range requirements {
		if r.FromClassID == req.FromClassID && r.ToClassID == req.ToClassID {
			return admin.ConvertAdvancementRequirementToResponse(r), nil
		}
	}

	return nil, errors.New("创建成功但无法获取详情")
}

// GetClassAdvancementRequirements 获取职业进阶要求
func (s *ClassService) GetClassAdvancementRequirements(ctx context.Context, classID uuid.UUID) ([]*apiRespAdmin.ClassAdvancementRequirement, error) {
	requirements, err := s.classRepo.GetAdvancementRequirements(ctx, classID)
	if err != nil {
		return nil, fmt.Errorf("获取进阶要求失败: %w", err)
	}

	return admin.ConvertAdvancementRequirementsToResponse(requirements), nil
}

// GetClassAdvancementPaths 获取职业进阶路径
func (s *ClassService) GetClassAdvancementPaths(ctx context.Context, fromClassID uuid.UUID) ([]*apiRespAdmin.ClassAdvancementRequirement, error) {
	paths, err := s.classRepo.GetAdvancementPaths(ctx, fromClassID)
	if err != nil {
		return nil, fmt.Errorf("获取进阶路径失败: %w", err)
	}

	return admin.ConvertAdvancementRequirementsToResponse(paths), nil
}

// GetClassAdvancementSources 获取职业进阶来源
func (s *ClassService) GetClassAdvancementSources(ctx context.Context, toClassID uuid.UUID) ([]*apiRespAdmin.ClassAdvancementRequirement, error) {
	sources, err := s.classRepo.GetAdvancementSources(ctx, toClassID)
	if err != nil {
		return nil, fmt.Errorf("获取进阶来源失败: %w", err)
	}

	return admin.ConvertAdvancementRequirementsToResponse(sources), nil
}

// UpdateClassAdvancementRequirement 更新职业进阶要求
func (s *ClassService) UpdateClassAdvancementRequirement(ctx context.Context, fromClassID, toClassID uuid.UUID, req *apiReqAdmin.UpdateClassAdvancementRequest) (*apiRespAdmin.ClassAdvancementRequirement, error) {
	requirement, err := s.classRepo.GetAdvancementRequirement(ctx, fromClassID, toClassID)
	if err != nil {
		return nil, fmt.Errorf("进阶要求不存在: %w", err)
	}

	admin.UpdateAdvancementRequirementFromRequest(requirement, req)
	if err := s.classRepo.UpdateAdvancementRequirement(ctx, requirement); err != nil {
		return nil, fmt.Errorf("更新进阶要求失败: %w", err)
	}

	// 获取详细信息返回
	requirements, err := s.classRepo.GetAdvancementRequirements(ctx, fromClassID)
	if err != nil {
		return nil, fmt.Errorf("获取进阶要求详情失败: %w", err)
	}

	for _, r := range requirements {
		if r.FromClassID == fromClassID && r.ToClassID == toClassID {
			return admin.ConvertAdvancementRequirementToResponse(r), nil
		}
	}

	return nil, errors.New("更新成功但无法获取详情")
}

// DeleteClassAdvancementRequirement 删除职业进阶要求
func (s *ClassService) DeleteClassAdvancementRequirement(ctx context.Context, fromClassID, toClassID uuid.UUID) error {
	requirement, err := s.classRepo.GetAdvancementRequirement(ctx, fromClassID, toClassID)
	if err != nil {
		return fmt.Errorf("进阶要求不存在: %w", err)
	}

	reqUUID, _ := uuid.Parse(requirement.ID)
	if err := s.classRepo.DeleteAdvancementRequirement(ctx, reqUUID); err != nil {
		return fmt.Errorf("删除进阶要求失败: %w", err)
	}

	return nil
}

// GetClassTags 获取职业标签
func (s *ClassService) GetClassTags(ctx context.Context, classID uuid.UUID) ([]*apiRespAdmin.ClassTag, error) {
	tags, err := s.classRepo.GetClassTags(ctx, classID)
	if err != nil {
		return nil, fmt.Errorf("获取职业标签失败: %w", err)
	}

	return admin.ConvertClassTagsToResponse(tags), nil
}

// AddClassTag 添加职业标签
func (s *ClassService) AddClassTag(ctx context.Context, classID uuid.UUID, req *apiReqAdmin.AddClassTagRequest) error {
	// 检查职业是否存在
	if _, err := s.classRepo.GetByID(ctx, classID); err != nil {
		return fmt.Errorf("职业不存在: %w", err)
	}

	if err := s.classRepo.AddClassTag(ctx, classID, req.TagID); err != nil {
		return fmt.Errorf("添加职业标签失败: %w", err)
	}

	return nil
}

// RemoveClassTag 移除职业标签
func (s *ClassService) RemoveClassTag(ctx context.Context, classID, tagID uuid.UUID) error {
	if err := s.classRepo.RemoveClassTag(ctx, classID, tagID); err != nil {
		return fmt.Errorf("移除职业标签失败: %w", err)
	}

	return nil
}

// GetAllTags 获取所有标签
func (s *ClassService) GetAllTags(ctx context.Context) ([]*apiRespAdmin.ClassTag, error) {
	tagType := "class" // 只获取职业类型的标签
	tags, err := s.classRepo.GetAllTags(ctx, &tagType)
	if err != nil {
		return nil, fmt.Errorf("获取标签列表失败: %w", err)
	}

	return admin.ConvertTagsToClassTags(tags), nil
}