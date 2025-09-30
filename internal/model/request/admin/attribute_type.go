package admin

import (
	businessValidator "tsu-self/internal/pkg/validator"
)

// CreateAttributeTypeRequest 创建属性类型请求
type CreateAttributeTypeRequest struct {
	AttributeCode        string   `json:"attribute_code" binding:"required,attribute_code" example:"STRENGTH" description:"属性代码"`
	AttributeName        string   `json:"attribute_name" binding:"required,chinese_name" example:"力量" description:"属性名称"`
	Category             string   `json:"category" binding:"required,oneof=base derived resistance" example:"base" description:"属性分类: base-基础, derived-衍生, resistance-抗性"`
	DataType             string   `json:"data_type" binding:"required,oneof=integer percentage" example:"integer" description:"数据类型: integer-整数, percentage-百分比"`
	MinValue             *float64 `json:"min_value,omitempty" binding:"omitempty,min=0" example:"0" description:"最小值"`
	MaxValue             *float64 `json:"max_value,omitempty" binding:"omitempty,min=1" example:"999" description:"最大值"`
	DefaultValue         *float64 `json:"default_value,omitempty" binding:"omitempty,min=0" example:"10" description:"默认值"`
	CalculationFormula   *string  `json:"calculation_formula,omitempty" binding:"omitempty,safe_description" example:"base + level * 2" description:"计算公式"`
	DependencyAttributes *string  `json:"dependency_attributes,omitempty" binding:"omitempty,safe_description" example:"strength,agility" description:"依赖属性列表"`
	Icon                 *string  `json:"icon,omitempty" example:"💪" description:"图标"`
	Color                *string  `json:"color,omitempty" binding:"omitempty,color_hex" example:"#FF0000" description:"颜色"`
	Unit                 *string  `json:"unit,omitempty" binding:"omitempty,chinese_name" example:"点" description:"单位"`
	DisplayOrder         *int     `json:"display_order,omitempty" binding:"omitempty,display_order" example:"1" description:"显示顺序"`
	IsVisible            *bool    `json:"is_visible,omitempty" example:"true" description:"是否可见"`
	Description          *string  `json:"description,omitempty" binding:"omitempty,safe_description" example:"影响物理攻击力" description:"属性描述"`
}

// UpdateAttributeTypeRequest 更新属性类型请求
type UpdateAttributeTypeRequest struct {
	AttributeCode        *string  `json:"attribute_code,omitempty" binding:"omitempty,min=2,max=50" example:"STRENGTH" description:"属性代码"`
	AttributeName        *string  `json:"attribute_name,omitempty" binding:"omitempty,min=1,max=100" example:"力量" description:"属性名称"`
	Category             *string  `json:"category,omitempty" binding:"omitempty,oneof=basic combat special" example:"basic" description:"属性分类"`
	DataType             *string  `json:"data_type,omitempty" binding:"omitempty,oneof=integer decimal percentage boolean" example:"integer" description:"数据类型"`
	MinValue             *float64 `json:"min_value,omitempty" example:"0" description:"最小值"`
	MaxValue             *float64 `json:"max_value,omitempty" example:"999" description:"最大值"`
	DefaultValue         *float64 `json:"default_value,omitempty" example:"10" description:"默认值"`
	CalculationFormula   *string  `json:"calculation_formula,omitempty" example:"base + level * 2" description:"计算公式"`
	DependencyAttributes *string  `json:"dependency_attributes,omitempty" example:"strength,agility" description:"依赖属性列表"`
	Icon                 *string  `json:"icon,omitempty" example:"💪" description:"图标"`
	Color                *string  `json:"color,omitempty" example:"#FF0000" description:"颜色"`
	Unit                 *string  `json:"unit,omitempty" example:"点" description:"单位"`
	DisplayOrder         *int     `json:"display_order,omitempty" example:"1" description:"显示顺序"`
	IsActive             *bool    `json:"is_active,omitempty" example:"true" description:"是否启用"`
	IsVisible            *bool    `json:"is_visible,omitempty" example:"true" description:"是否可见"`
	Description          *string  `json:"description,omitempty" example:"影响物理攻击力" description:"属性描述"`
}

// GetAttributeTypesRequest 获取属性类型列表请求
type GetAttributeTypesRequest struct {
	Page      int    `form:"page" binding:"omitempty,min=1" example:"1" description:"页码"`
	PageSize  int    `form:"page_size" binding:"omitempty,min=1,max=100" example:"20" description:"每页数量"`
	Category  string `form:"category" binding:"omitempty,oneof=basic combat special" example:"basic" description:"属性分类过滤"`
	IsActive  *bool  `form:"is_active" example:"true" description:"是否启用过滤"`
	IsVisible *bool  `form:"is_visible" example:"true" description:"是否可见过滤"`
	Keyword   string `form:"keyword" binding:"omitempty,max=100" example:"力量" description:"关键词搜索(代码或名称)"`
	SortBy    string `form:"sort_by" binding:"omitempty,oneof=created_at updated_at display_order attribute_name" example:"display_order" description:"排序字段"`
	SortOrder string `form:"sort_order" binding:"omitempty,oneof=asc desc" example:"asc" description:"排序方向"`
}

// Validate 验证请求参数
func (req *CreateAttributeTypeRequest) Validate() error {
	businessVal := businessValidator.NewBusinessValidator()
	if err := businessVal.Validate(req); err != nil {
		return err
	}

	// 验证数值范围的业务逻辑
	return businessValidator.ValidateValueRange(req.MinValue, req.MaxValue, req.DefaultValue)
}

func (req *UpdateAttributeTypeRequest) Validate() error {
	businessVal := businessValidator.NewBusinessValidator()
	if err := businessVal.Validate(req); err != nil {
		return err
	}

	// 验证数值范围的业务逻辑
	return businessValidator.ValidateValueRange(req.MinValue, req.MaxValue, req.DefaultValue)
}

func (req *GetAttributeTypesRequest) Validate() error {
	businessVal := businessValidator.NewBusinessValidator()
	return businessVal.Validate(req)
}
