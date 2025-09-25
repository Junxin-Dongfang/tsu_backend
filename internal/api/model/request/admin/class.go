package admin

import (
	"github.com/google/uuid"
)

// CreateClassRequest 创建职业请求
// @Description 创建新职业的请求参数
type CreateClassRequest struct {
	// 职业代码，必须唯一，用于系统内部识别
	Code string `json:"code" binding:"required,min=2,max=50" example:"WARRIOR"`
	// 职业显示名称
	Name string `json:"name" binding:"required,min=2,max=100" example:"战士"`
	// 职业描述信息
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000" example:"擅长近战与防御的强力职业"`
	// 职业背景故事文本
	LoreText *string `json:"lore_text,omitempty" binding:"omitempty,max=2000" example:"自古以来守护村庄的勇士"`
	// 职业层级，1为基础职业，2为进阶职业
	Tier int `json:"tier" binding:"required,min=1,max=10" example:"1"`
	// 转职时获得的属性加成百分比
	JobChangeBonus *int `json:"job_change_bonus,omitempty" binding:"omitempty,min=0" example:"5"`
	// 职业图标URL
	Icon *string `json:"icon,omitempty" example:"/assets/icons/warrior.png"`
	// 职业主题色彩（十六进制颜色代码）
	ColorTheme *string `json:"color_theme,omitempty" example:"#FFAA00"`
	// 是否在列表中隐藏
	IsHidden *bool `json:"is_hidden,omitempty" example:"false"`
	// 显示排序顺序
	DisplayOrder *int `json:"display_order,omitempty" binding:"omitempty,min=0" example:"10"`
} // @name CreateClassRequest

// UpdateClassRequest 更新职业请求
// @Description 更新职业信息的请求参数
type UpdateClassRequest struct {
	// 职业代码，必须唯一，用于系统内部识别
	Code *string `json:"code,omitempty" binding:"omitempty,min=2,max=50" example:"WARRIOR"`
	// 职业显示名称
	Name *string `json:"name,omitempty" binding:"omitempty,min=2,max=100" example:"战士"`
	// 职业描述信息
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000" example:"擅长近战与防御的强力职业"`
	// 职业背景故事文本
	LoreText *string `json:"lore_text,omitempty" binding:"omitempty,max=2000" example:"自古以来守护村庄的勇士"`
	// 职业层级，1为基础职业，2为进阶职业
	Tier *int `json:"tier,omitempty" binding:"omitempty,min=1,max=10" example:"1"`
	// 转职时获得的属性加成百分比
	JobChangeBonus *int `json:"job_change_bonus,omitempty" binding:"omitempty,min=0" example:"5"`
	// 职业图标URL
	Icon *string `json:"icon,omitempty" example:"/assets/icons/warrior.png"`
	// 职业主题色彩（十六进制颜色代码）
	ColorTheme *string `json:"color_theme,omitempty" example:"#FFAA00"`
	// 是否在列表中隐藏
	IsHidden *bool `json:"is_hidden,omitempty" example:"false"`
	// 显示排序顺序
	DisplayOrder *int `json:"display_order,omitempty" binding:"omitempty,min=0" example:"10"`
} // @name UpdateClassRequest

// ClassListRequest 职业列表查询请求
// @Description 获取职业列表的查询参数
type ClassListRequest struct {
	// 职业层级筛选
	Tier *int `form:"tier" binding:"omitempty,min=1,max=10" example:"1"`
	// 是否启用状态筛选
	IsActive *bool `form:"is_active" example:"true"`
	// 是否隐藏状态筛选
	IsHidden *bool `form:"is_hidden" example:"false"`
	// 关键词搜索（搜索名称、代码、描述）
	Search *string `form:"search" example:"战士"`
	// 排序字段
	SortBy *string `form:"sort_by" binding:"omitempty,oneof=name code tier display_order created_at" example:"display_order"`
	// 排序方向
	SortOrder *string `form:"sort_order" binding:"omitempty,oneof=asc desc" example:"asc"`
	// 页码，从1开始
	Page int `form:"page" binding:"omitempty,min=1" example:"1"`
	// 每页数量
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100" example:"20"`
} // @name ClassListRequest

// CreateClassAttributeBonusRequest 创建职业属性加成请求
// @Description 创建职业属性加成配置的请求参数
type CreateClassAttributeBonusRequest struct {
	// 属性ID
	AttributeID uuid.UUID `json:"attribute_id" binding:"required" example:"50000000-0000-0000-0000-000000000001"`
	// 基础加成值，职业获得的固定属性加成，不受等级影响
	BaseBonus float64 `json:"base_bonus" binding:"min=0" example:"5"`
	// 每级加成值，每升一级获得的属性加成
	PerLevelBonus float64 `json:"per_level_bonus" binding:"min=0" example:"1.2"`
} // @name CreateClassAttributeBonusRequest

// UpdateClassAttributeBonusRequest 更新职业属性加成请求
// @Description 更新职业属性加成配置的请求参数
type UpdateClassAttributeBonusRequest struct {
	// 基础加成值，职业获得的固定属性加成，不受等级影响
	BaseBonus *float64 `json:"base_bonus,omitempty" binding:"omitempty,min=0" example:"5"`
	// 每级加成值，每升一级获得的属性加成
	PerLevelBonus *float64 `json:"per_level_bonus,omitempty" binding:"omitempty,min=0" example:"1.2"`
} // @name UpdateClassAttributeBonusRequest

// BatchCreateClassAttributeBonusRequest 批量创建职业属性加成请求
// @Description 批量创建职业属性加成配置的请求参数
type BatchCreateClassAttributeBonusRequest struct {
	// 属性加成配置列表
	Bonuses []CreateClassAttributeBonusRequest `json:"bonuses" binding:"required,min=1,max=20"`
} // @name BatchCreateClassAttributeBonusRequest

// CreateClassAdvancementRequest 创建职业进阶要求请求
// @Description 创建职业进阶要求配置的请求参数
type CreateClassAdvancementRequest struct {
	// 源职业ID，进阶的起始职业
	FromClassID uuid.UUID `json:"from_class_id" binding:"required" example:"30000000-0000-0000-0000-000000000001"`
	// 目标职业ID，进阶的目标职业
	ToClassID uuid.UUID `json:"to_class_id" binding:"required" example:"30000000-0000-0000-0000-000000000010"`
	// 所需等级
	RequiredLevel int `json:"required_level" binding:"required,min=1" example:"20"`
	// 所需荣誉值
	RequiredHonor *int `json:"required_honor,omitempty" binding:"omitempty,min=0" example:"1000"`
	// 所需转职次数
	RequiredJobChangeCount *int `json:"required_job_change_count,omitempty" binding:"omitempty,min=0" example:"1"`
	// 所需属性要求（JSON格式）
	RequiredAttributes map[string]interface{} `json:"required_attributes,omitempty" example:"{\"strength\": 50, \"dexterity\": 30}"`
	// 所需技能要求（JSON格式）
	RequiredSkills map[string]interface{} `json:"required_skills,omitempty" example:"{\"skill_id_1\": 5, \"skill_id_2\": 3}"`
	// 显示排序顺序
	DisplayOrder *int `json:"display_order,omitempty" binding:"omitempty,min=0" example:"1"`
} // @name CreateClassAdvancementRequest

// UpdateClassAdvancementRequest 更新职业进阶要求请求
// @Description 更新职业进阶要求配置的请求参数
type UpdateClassAdvancementRequest struct {
	// 所需等级
	RequiredLevel *int `json:"required_level,omitempty" binding:"omitempty,min=1" example:"20"`
	// 所需荣誉值
	RequiredHonor *int `json:"required_honor,omitempty" binding:"omitempty,min=0" example:"1000"`
	// 所需转职次数
	RequiredJobChangeCount *int `json:"required_job_change_count,omitempty" binding:"omitempty,min=0" example:"1"`
	// 所需属性要求（JSON格式）
	RequiredAttributes map[string]interface{} `json:"required_attributes,omitempty" example:"{\"strength\": 50, \"dexterity\": 30}"`
	// 所需技能要求（JSON格式）
	RequiredSkills map[string]interface{} `json:"required_skills,omitempty" example:"{\"skill_id_1\": 5, \"skill_id_2\": 3}"`
	// 是否启用
	IsActive *bool `json:"is_active,omitempty" example:"true"`
	// 显示排序顺序
	DisplayOrder *int `json:"display_order,omitempty" binding:"omitempty,min=0" example:"1"`
} // @name UpdateClassAdvancementRequest

// AddClassTagRequest 添加职业标签请求
// @Description 为职业添加标签的请求参数
type AddClassTagRequest struct {
	// 标签ID
	TagID uuid.UUID `json:"tag_id" binding:"required" example:"60000000-0000-0000-0000-000000000001"`
} // @name AddClassTagRequest