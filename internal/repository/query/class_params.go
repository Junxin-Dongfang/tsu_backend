package query

import (
	"time"

	"github.com/google/uuid"
)

// ClassListParams 职业列表查询参数
type ClassListParams struct {
	// 筛选条件
	Tier     *int    `json:"tier,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
	IsHidden *bool   `json:"is_hidden,omitempty"`
	Search   *string `json:"search,omitempty"`

	// 排序
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"`

	// 分页
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// ClassHeroStats 职业英雄统计信息
type ClassHeroStats struct {
	ClassID      uuid.UUID `json:"class_id"`
	TotalHeroes  int       `json:"total_heroes"`
	ActiveHeroes int       `json:"active_heroes"`
	AverageLevel float64   `json:"average_level"`
	MaxLevel     int       `json:"max_level"`
}

// ClassAttributeBonusWithDetails 职业属性加成详细信息
type ClassAttributeBonusWithDetails struct {
	ID            uuid.UUID `json:"id"`
	ClassID       uuid.UUID `json:"class_id"`
	AttributeID   uuid.UUID `json:"attribute_id"`
	AttributeCode string    `json:"attribute_code"`
	AttributeName string    `json:"attribute_name"`
	BaseBonus     float64   `json:"base_bonus"`
	PerLevelBonus float64   `json:"per_level_bonus"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ClassAdvancementWithDetails 职业进阶要求详细信息
type ClassAdvancementWithDetails struct {
	ID                     uuid.UUID              `json:"id"`
	FromClassID            uuid.UUID              `json:"from_class_id"`
	FromClassName          string                 `json:"from_class_name"`
	ToClassID              uuid.UUID              `json:"to_class_id"`
	ToClassName            string                 `json:"to_class_name"`
	RequiredLevel          int                    `json:"required_level"`
	RequiredHonor          int                    `json:"required_honor"`
	RequiredJobChangeCount int                    `json:"required_job_change_count"`
	RequiredAttributes     map[string]interface{} `json:"required_attributes,omitempty"`
	RequiredSkills         map[string]interface{} `json:"required_skills,omitempty"`
	IsActive               bool                   `json:"is_active"`
	DisplayOrder           int                    `json:"display_order"`
	CreatedAt              time.Time              `json:"created_at"`
	UpdatedAt              time.Time              `json:"updated_at"`
}

// ClassTagWithDetails 职业标签详细信息
type ClassTagWithDetails struct {
	ID           uuid.UUID  `json:"id"`
	Code         string     `json:"code"`
	Name         string     `json:"name"`
	Description  *string    `json:"description,omitempty"`
	Color        *string    `json:"color,omitempty"`
	Icon         *string    `json:"icon,omitempty"`
	DisplayOrder int        `json:"display_order"`
	CreatedAt    time.Time  `json:"created_at"`
}