package admin

import (
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/ericlagergren/decimal"
	"github.com/google/uuid"

	apiReqAdmin "tsu-self/internal/api/model/request/admin"
	apiResAdmin "tsu-self/internal/api/model/response/admin"
	"tsu-self/internal/entity"
)

// ConvertToAttributeTypeEntity 将创建请求转换为数据库实体
func ConvertToAttributeTypeEntity(req *apiReqAdmin.CreateAttributeTypeRequest) *entity.HeroAttributeType {
	attributeType := &entity.HeroAttributeType{
		ID:            uuid.New().String(),
		AttributeCode: req.AttributeCode,
		AttributeName: req.AttributeName,
		Category:      req.Category,
		DataType:      req.DataType,
		IsActive:      null.BoolFrom(true),
		IsVisible:     null.BoolFrom(true),
		DisplayOrder:  null.Int16From(1),
	}

	// 处理可选字段 - 使用正确的 NullDecimal 方法
	if req.MinValue != nil {
		// 使用 ericlagergren/decimal 创建 Big 值
		var big decimal.Big
		big.SetFloat64(*req.MinValue)
		attributeType.MinValue = types.NewNullDecimal(&big)
	}
	if req.MaxValue != nil {
		var big decimal.Big
		big.SetFloat64(*req.MaxValue)
		attributeType.MaxValue = types.NewNullDecimal(&big)
	}
	if req.DefaultValue != nil {
		var big decimal.Big
		big.SetFloat64(*req.DefaultValue)
		attributeType.DefaultValue = types.NewNullDecimal(&big)
	}
	if req.CalculationFormula != nil {
		attributeType.CalculationFormula = null.StringFrom(*req.CalculationFormula)
	}
	if req.DependencyAttributes != nil {
		attributeType.DependencyAttributes = null.StringFrom(*req.DependencyAttributes)
	}
	if req.Icon != nil {
		attributeType.Icon = null.StringFrom(*req.Icon)
	}
	if req.Color != nil {
		attributeType.Color = null.StringFrom(*req.Color)
	}
	if req.Unit != nil {
		attributeType.Unit = null.StringFrom(*req.Unit)
	}
	if req.DisplayOrder != nil {
		attributeType.DisplayOrder = null.Int16From(int16(*req.DisplayOrder))
	}
	if req.IsVisible != nil {
		attributeType.IsVisible = null.BoolFrom(*req.IsVisible)
	}
	if req.Description != nil {
		attributeType.Description = null.StringFrom(*req.Description)
	}

	return attributeType
}

// UpdateAttributeTypeEntity 将更新请求应用到数据库实体
func UpdateAttributeTypeEntity(entity *entity.HeroAttributeType, req *apiReqAdmin.UpdateAttributeTypeRequest) {
	if req.AttributeCode != nil {
		entity.AttributeCode = *req.AttributeCode
	}
	if req.AttributeName != nil {
		entity.AttributeName = *req.AttributeName
	}
	if req.Category != nil {
		entity.Category = *req.Category
	}
	if req.DataType != nil {
		entity.DataType = *req.DataType
	}
	// 处理 Decimal 字段更新
	if req.MinValue != nil {
		var big decimal.Big
		big.SetFloat64(*req.MinValue)
		entity.MinValue = types.NewNullDecimal(&big)
	}
	if req.MaxValue != nil {
		var big decimal.Big
		big.SetFloat64(*req.MaxValue)
		entity.MaxValue = types.NewNullDecimal(&big)
	}
	if req.DefaultValue != nil {
		var big decimal.Big
		big.SetFloat64(*req.DefaultValue)
		entity.DefaultValue = types.NewNullDecimal(&big)
	}
	if req.CalculationFormula != nil {
		entity.CalculationFormula = null.StringFrom(*req.CalculationFormula)
	}
	if req.DependencyAttributes != nil {
		entity.DependencyAttributes = null.StringFrom(*req.DependencyAttributes)
	}
	if req.Icon != nil {
		entity.Icon = null.StringFrom(*req.Icon)
	}
	if req.Color != nil {
		entity.Color = null.StringFrom(*req.Color)
	}
	if req.Unit != nil {
		entity.Unit = null.StringFrom(*req.Unit)
	}
	if req.DisplayOrder != nil {
		entity.DisplayOrder = null.Int16From(int16(*req.DisplayOrder))
	}
	if req.IsActive != nil {
		entity.IsActive = null.BoolFrom(*req.IsActive)
	}
	if req.IsVisible != nil {
		entity.IsVisible = null.BoolFrom(*req.IsVisible)
	}
	if req.Description != nil {
		entity.Description = null.StringFrom(*req.Description)
	}
}

// ConvertToAttributeTypeResponse 将数据库实体转换为响应模型
func ConvertToAttributeTypeResponse(entity *entity.HeroAttributeType) *apiResAdmin.AttributeType {
	response := &apiResAdmin.AttributeType{
		ID:            uuid.MustParse(entity.ID),
		AttributeCode: entity.AttributeCode,
		AttributeName: entity.AttributeName,
		Category:      entity.Category,
		DataType:      entity.DataType,
		DisplayOrder:  int(entity.DisplayOrder.Int16),
		IsActive:      entity.IsActive.Bool,
		IsVisible:     entity.IsVisible.Bool,
		CreatedAt:     entity.CreatedAt,
		UpdatedAt:     entity.UpdatedAt,
	}

	// 处理可选字段
	if !entity.MinValue.IsZero() {
		value, _ := entity.MinValue.Float64()
		response.MinValue = &value
	}
	if !entity.MaxValue.IsZero() {
		value, _ := entity.MaxValue.Float64()
		response.MaxValue = &value
	}
	if !entity.DefaultValue.IsZero() {
		value, _ := entity.DefaultValue.Float64()
		response.DefaultValue = &value
	}
	if entity.CalculationFormula.Valid {
		response.CalculationFormula = &entity.CalculationFormula.String
	}
	if entity.DependencyAttributes.Valid {
		response.DependencyAttributes = &entity.DependencyAttributes.String
	}
	if entity.Icon.Valid {
		response.Icon = &entity.Icon.String
	}
	if entity.Color.Valid {
		response.Color = &entity.Color.String
	}
	if entity.Unit.Valid {
		response.Unit = &entity.Unit.String
	}
	if entity.Description.Valid {
		response.Description = &entity.Description.String
	}

	return response
}

// ConvertToAttributeTypeOptionResponse 将数据库实体转换为选项响应模型
func ConvertToAttributeTypeOptionResponse(entity *entity.HeroAttributeType) *apiResAdmin.AttributeTypeOption {
	response := &apiResAdmin.AttributeTypeOption{
		ID:            uuid.MustParse(entity.ID),
		AttributeCode: entity.AttributeCode,
		AttributeName: entity.AttributeName,
		Category:      entity.Category,
		DataType:      entity.DataType,
	}

	// 处理可选字段
	if entity.Icon.Valid {
		response.Icon = &entity.Icon.String
	}
	if entity.Color.Valid {
		response.Color = &entity.Color.String
	}
	if entity.Unit.Valid {
		response.Unit = &entity.Unit.String
	}

	return response
}

// ConvertToAttributeTypeListResponse 将数据库实体列表转换为列表响应
func ConvertToAttributeTypeListResponse(entities []*entity.HeroAttributeType, total int64, page, pageSize int) *apiResAdmin.AttributeTypeList {
	data := make([]apiResAdmin.AttributeType, len(entities))
	for i, entity := range entities {
		data[i] = *ConvertToAttributeTypeResponse(entity)
	}

	totalPages := int(total+int64(pageSize)-1) / pageSize // 向上取整

	return &apiResAdmin.AttributeTypeList{
		Data: data,
		Pagination: apiResAdmin.PaginationResponse{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}

// ConvertToAttributeTypeOptionsResponse 将数据库实体列表转换为选项响应
func ConvertToAttributeTypeOptionsResponse(entities []*entity.HeroAttributeType) *apiResAdmin.AttributeTypeOptions {
	data := make([]apiResAdmin.AttributeTypeOption, len(entities))
	for i, entity := range entities {
		data[i] = *ConvertToAttributeTypeOptionResponse(entity)
	}

	return &apiResAdmin.AttributeTypeOptions{
		Data: data,
	}
}
