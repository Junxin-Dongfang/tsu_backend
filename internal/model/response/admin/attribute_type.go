package admin

import (
	"time"

	"github.com/google/uuid"
)

// AttributeType 属性类型响应
type AttributeType struct {
	ID                   uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000" description:"属性ID"`
	AttributeCode        string    `json:"attribute_code" example:"STRENGTH" description:"属性代码"`
	AttributeName        string    `json:"attribute_name" example:"力量" description:"属性名称"`
	Category             string    `json:"category" example:"basic" description:"属性分类"`
	DataType             string    `json:"data_type" example:"integer" description:"数据类型"`
	MinValue             *float64  `json:"min_value,omitempty" example:"0" description:"最小值"`
	MaxValue             *float64  `json:"max_value,omitempty" example:"999" description:"最大值"`
	DefaultValue         *float64  `json:"default_value,omitempty" example:"10" description:"默认值"`
	CalculationFormula   *string   `json:"calculation_formula,omitempty" example:"base + level * 2" description:"计算公式"`
	DependencyAttributes *string   `json:"dependency_attributes,omitempty" example:"strength,agility" description:"依赖属性列表"`
	Icon                 *string   `json:"icon,omitempty" example:"💪" description:"图标"`
	Color                *string   `json:"color,omitempty" example:"#FF0000" description:"颜色"`
	Unit                 *string   `json:"unit,omitempty" example:"点" description:"单位"`
	DisplayOrder         int       `json:"display_order" example:"1" description:"显示顺序"`
	IsActive             bool      `json:"is_active" example:"true" description:"是否启用"`
	IsVisible            bool      `json:"is_visible" example:"true" description:"是否可见"`
	Description          *string   `json:"description,omitempty" example:"影响物理攻击力" description:"属性描述"`
	CreatedAt            time.Time `json:"created_at" example:"2023-01-01T00:00:00Z" description:"创建时间"`
	UpdatedAt            time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z" description:"更新时间"`
}

// AttributeTypeList 属性类型列表响应
type AttributeTypeList struct {
	Data       []AttributeType `json:"data" description:"属性类型列表"`
	Pagination PaginationResponse `json:"pagination" description:"分页信息"`
}

// AttributeTypeOption 属性类型选项（用于下拉选择）
type AttributeTypeOption struct {
	ID            uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000" description:"属性ID"`
	AttributeCode string    `json:"attribute_code" example:"STRENGTH" description:"属性代码"`
	AttributeName string    `json:"attribute_name" example:"力量" description:"属性名称"`
	Category      string    `json:"category" example:"basic" description:"属性分类"`
	DataType      string    `json:"data_type" example:"integer" description:"数据类型"`
	Icon          *string   `json:"icon,omitempty" example:"💪" description:"图标"`
	Color         *string   `json:"color,omitempty" example:"#FF0000" description:"颜色"`
	Unit          *string   `json:"unit,omitempty" example:"点" description:"单位"`
}

// AttributeTypeOptions 属性类型选项列表
type AttributeTypeOptions struct {
	Data []AttributeTypeOption `json:"data" description:"属性类型选项列表"`
}