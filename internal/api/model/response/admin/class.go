package admin

import (
	"time"

	"github.com/google/uuid"
)

// Class 职业响应模型
// @Description 职业的完整信息
type Class struct {
	// 职业唯一标识符
	ID uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	// 职业代码，用于系统内部识别，如WARRIOR、MAGE等
	Code string `json:"code" example:"WARRIOR"`
	// 职业显示名称
	Name string `json:"name" example:"战士"`
	// 职业描述信息
	Description *string `json:"description,omitempty" example:"擅长近战与防御的职业"`
	// 职业背景故事文本
	LoreText *string `json:"lore_text,omitempty" example:"上古部族守护者"`
	// 职业层级，1为基础职业，2为进阶职业，以此类推
	Tier int `json:"tier" example:"1"`
	// 转职时获得的属性加成百分比
	JobChangeBonus int `json:"job_change_bonus" example:"5"`
	// 职业图标URL
	Icon *string `json:"icon,omitempty" example:"/assets/icons/warrior.png"`
	// 职业主题色彩（十六进制颜色代码）
	ColorTheme *string `json:"color_theme,omitempty" example:"#FFAA00"`
	// 职业是否启用状态
	IsActive bool `json:"is_active" example:"true"`
	// 是否在列表中隐藏
	IsHidden bool `json:"is_hidden" example:"false"`
	// 显示排序顺序
	DisplayOrder int `json:"display_order" example:"10"`
	// 创建时间
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	// 最后更新时间
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-02T00:00:00Z"`
} // @name Class

// ClassWithStats 包含统计信息的职业响应模型
// @Description 职业信息及其英雄统计数据
type ClassWithStats struct {
	Class
	// 该职业总英雄数量
	TotalHeroes int `json:"total_heroes" example:"150"`
	// 该职业活跃英雄数量
	ActiveHeroes int `json:"active_heroes" example:"120"`
	// 该职业英雄平均等级
	AverageLevel float64 `json:"average_level" example:"45.8"`
	// 该职业英雄最高等级
	MaxLevel int `json:"max_level" example:"85"`
} // @name ClassWithStats

// ClassListResponse 职业列表响应
// @Description 职业列表查询的分页响应
type ClassListResponse struct {
	// 职业列表
	Data []Class `json:"data"`
	// 分页信息
	Pagination PaginationResponse `json:"pagination"`
} // @name ClassListResponse

// ClassAttributeBonus 职业属性加成响应模型
// @Description 职业属性加成配置信息
type ClassAttributeBonus struct {
	// 加成配置唯一标识符
	ID uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	// 所属职业ID
	ClassID uuid.UUID `json:"class_id" example:"30000000-0000-0000-0000-000000000010"`
	// 属性ID
	AttributeID uuid.UUID `json:"attribute_id" example:"50000000-0000-0000-0000-000000000001"`
	// 属性代码（从关联查询获得）
	AttributeCode string `json:"attribute_code" example:"strength"`
	// 属性名称（从关联查询获得）
	AttributeName string `json:"attribute_name" example:"力量"`
	// 基础加成值，职业获得的固定加成
	BaseBonus float64 `json:"base_bonus" example:"5"`
	// 每级加成值，每升一级获得的加成
	PerLevelBonus float64 `json:"per_level_bonus" example:"1.2"`
	// 创建时间
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	// 最后更新时间
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-02T00:00:00Z"`
} // @name ClassAttributeBonus

// ClassAdvancementRequirement 职业进阶要求响应模型
// @Description 职业进阶要求配置信息
type ClassAdvancementRequirement struct {
	// 进阶要求唯一标识符
	ID uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	// 源职业ID，进阶的起始职业
	FromClassID uuid.UUID `json:"from_class_id" example:"30000000-0000-0000-0000-000000000001"`
	// 源职业名称，从关联查询获得
	FromClassName string `json:"from_class_name" example:"新手"`
	// 目标职业ID，进阶的目标职业
	ToClassID uuid.UUID `json:"to_class_id" example:"30000000-0000-0000-0000-000000000010"`
	// 目标职业名称，从关联查询获得
	ToClassName string `json:"to_class_name" example:"剑士"`
	// 所需等级
	RequiredLevel int `json:"required_level" example:"20"`
	// 所需荣誉值
	RequiredHonor int `json:"required_honor" example:"1000"`
	// 所需转职次数
	RequiredJobChangeCount int `json:"required_job_change_count" example:"1"`
	// 所需属性要求（JSON格式）
	RequiredAttributes map[string]interface{} `json:"required_attributes,omitempty" swaggertype:"object" example:"{\"strength\": 50, \"dexterity\": 30}"`
	// 所需技能要求（JSON格式）
	RequiredSkills map[string]interface{} `json:"required_skills,omitempty" swaggertype:"object" example:"{\"skill_id_1\": 5, \"skill_id_2\": 3}"`
	// 是否启用
	IsActive bool `json:"is_active" example:"true"`
	// 显示排序顺序
	DisplayOrder int `json:"display_order" example:"1"`
	// 创建时间
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	// 最后更新时间
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-02T00:00:00Z"`
} // @name ClassAdvancementRequirement

// ClassTag 职业标签响应模型
// @Description 职业标签信息
type ClassTag struct {
	// 标签唯一标识符
	ID uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	// 标签代码
	Code string `json:"code" example:"MELEE"`
	// 标签名称
	Name string `json:"name" example:"近战"`
	// 标签描述信息
	Description *string `json:"description,omitempty" example:"近战类职业标签"`
	// 标签颜色值
	Color *string `json:"color,omitempty" example:"#FF0000"`
	// 标签图标URL
	Icon *string `json:"icon,omitempty" example:"/assets/tags/warrior.png"`
	// 显示排序顺序
	DisplayOrder int `json:"display_order" example:"1"`
} // @name ClassTag

// ClassHeroStats 职业英雄统计响应模型
// @Description 职业英雄数量统计信息
type ClassHeroStats struct {
	// 职业ID
	ClassID uuid.UUID `json:"class_id" example:"30000000-0000-0000-0000-000000000010"`
	// 该职业总英雄数量
	TotalHeroes int `json:"total_heroes" example:"150"`
	// 该职业活跃英雄数量
	ActiveHeroes int `json:"active_heroes" example:"120"`
	// 该职业英雄平均等级
	AverageLevel float64 `json:"average_level" example:"45.8"`
	// 该职业英雄最高等级
	MaxLevel int `json:"max_level" example:"85"`
} // @name ClassHeroStats

// PaginationResponse 分页响应信息
// @Description 分页查询的元数据信息
type PaginationResponse struct {
	// 当前页码
	Page int `json:"page" example:"1"`
	// 每页数量
	PageSize int `json:"page_size" example:"20"`
	// 总记录数
	Total int64 `json:"total" example:"100"`
	// 总页数
	TotalPages int `json:"total_pages" example:"5"`
} // @name PaginationResponse