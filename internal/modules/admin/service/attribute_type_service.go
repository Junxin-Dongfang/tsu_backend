package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	apiReqAdmin "tsu-self/internal/api/model/request/admin"
	apiResAdmin "tsu-self/internal/api/model/response/admin"
	"tsu-self/internal/converter/admin"
	"tsu-self/internal/pkg/validator"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/interfaces"
)

// AttributeTypeService 属性类型服务
type AttributeTypeService struct {
	attributeTypeRepo interfaces.AttributeTypeRepository
	validator         *validator.BusinessValidator
}

// NewAttributeTypeService 创建属性类型服务
func NewAttributeTypeService(attributeTypeRepo interfaces.AttributeTypeRepository) *AttributeTypeService {
	return &AttributeTypeService{
		attributeTypeRepo: attributeTypeRepo,
		validator:         validator.NewBusinessValidator(),
	}
}

// CreateAttributeType 创建属性类型
func (s *AttributeTypeService) CreateAttributeType(ctx context.Context, req *apiReqAdmin.CreateAttributeTypeRequest) (*apiResAdmin.AttributeType, error) {
	// 1. 业务规则验证
	if err := s.validator.Validate(req); err != nil {
		return nil, xerrors.NewValidationError("validation", validator.GetValidationErrorMessage(err))
	}

	// 2. 验证数值范围的业务逻辑
	if err := validator.ValidateValueRange(req.MinValue, req.MaxValue, req.DefaultValue); err != nil {
		return nil, xerrors.NewValidationError("value_range", err.Error())
	}

	// 3. 验证属性代码是否已存在
	exists, err := s.attributeTypeRepo.ExistsByCode(ctx, req.AttributeCode)
	if err != nil {
		return nil, fmt.Errorf("failed to check attribute code existence: %w", err)
	}
	if exists {
		return nil, xerrors.NewValidationError("attribute_code", "属性代码已存在")
	}

	// 4. 复杂业务规则验证
	if err := validator.ValidateBusinessRules("attribute_type", req); err != nil {
		return nil, xerrors.NewValidationError("business_rules", err.Error())
	}

	// 5. 转换为实体
	entity := admin.ConvertToAttributeTypeEntity(req)

	// 6. 保存到数据库
	if err := s.attributeTypeRepo.Create(ctx, entity); err != nil {
		return nil, fmt.Errorf("failed to create attribute type: %w", err)
	}

	// 7. 转换为响应模型
	return admin.ConvertToAttributeTypeResponse(entity), nil
}

// GetAttributeType 获取属性类型详情
func (s *AttributeTypeService) GetAttributeType(ctx context.Context, id string) (*apiResAdmin.AttributeType, error) {
	// 验证UUID格式
	if _, err := uuid.Parse(id); err != nil {
		return nil, xerrors.NewValidationError("id", "无效的ID格式")
	}

	// 从数据库获取
	entity, err := s.attributeTypeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get attribute type: %w", err)
	}
	if entity == nil {
		return nil, xerrors.NewNotFoundError("id", "属性类型不存在")
	}

	// 转换为响应模型
	return admin.ConvertToAttributeTypeResponse(entity), nil
}

// GetAttributeTypes 获取属性类型列表
func (s *AttributeTypeService) GetAttributeTypes(ctx context.Context, req *apiReqAdmin.GetAttributeTypesRequest) (*apiResAdmin.AttributeTypeList, error) {
	// 从数据库获取列表
	entities, total, err := s.attributeTypeRepo.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get attribute types: %w", err)
	}

	// 转换为响应模型
	return admin.ConvertToAttributeTypeListResponse(entities, total, req.Page, req.PageSize), nil
}

// UpdateAttributeType 更新属性类型
func (s *AttributeTypeService) UpdateAttributeType(ctx context.Context, id string, req *apiReqAdmin.UpdateAttributeTypeRequest) (*apiResAdmin.AttributeType, error) {
	// 验证UUID格式
	if _, err := uuid.Parse(id); err != nil {
		return nil, xerrors.NewValidationError("id", "无效的ID格式")
	}

	// 获取现有实体
	entity, err := s.attributeTypeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get attribute type: %w", err)
	}
	if entity == nil {
		return nil, xerrors.NewNotFoundError("id", "属性类型不存在")
	}

	// 检查属性代码是否重复（如果更新了代码）
	if req.AttributeCode != nil && *req.AttributeCode != entity.AttributeCode {
		exists, err := s.attributeTypeRepo.ExistsByCode(ctx, *req.AttributeCode, id)
		if err != nil {
			return nil, fmt.Errorf("failed to check attribute code existence: %w", err)
		}
		if exists {
			return nil, xerrors.NewValidationError("attribute_code", "属性代码已存在")
		}
	}

	// 应用更新
	admin.UpdateAttributeTypeEntity(entity, req)

	// 保存更新
	if err := s.attributeTypeRepo.Update(ctx, entity); err != nil {
		return nil, fmt.Errorf("failed to update attribute type: %w", err)
	}

	// 转换为响应模型
	return admin.ConvertToAttributeTypeResponse(entity), nil
}

// DeleteAttributeType 删除属性类型
func (s *AttributeTypeService) DeleteAttributeType(ctx context.Context, id string) error {
	// 验证UUID格式
	if _, err := uuid.Parse(id); err != nil {
		return xerrors.NewValidationError("id", "无效的ID格式")
	}

	// 检查是否存在
	entity, err := s.attributeTypeRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get attribute type: %w", err)
	}
	if entity == nil {
		return xerrors.NewNotFoundError("id", "属性类型不存在")
	}

	// 软删除
	if err := s.attributeTypeRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete attribute type: %w", err)
	}

	return nil
}

// GetAttributeTypeOptions 获取属性类型选项列表
func (s *AttributeTypeService) GetAttributeTypeOptions(ctx context.Context, category string) (*apiResAdmin.AttributeTypeOptions, error) {
	// 获取启用的属性类型列表
	entities, err := s.attributeTypeRepo.GetActiveList(ctx, category)
	if err != nil {
		return nil, fmt.Errorf("failed to get attribute type options: %w", err)
	}

	// 转换为选项响应模型
	return admin.ConvertToAttributeTypeOptionsResponse(entities), nil
}
