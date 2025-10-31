package dto

import "time"

// TagResponse 标签响应
type TagResponse struct {
	ID           string `json:"id"`
	TagCode      string `json:"tag_code"`
	TagName      string `json:"tag_name"`
	Category     string `json:"category"`
	Description  string `json:"description,omitempty"`
	Icon         string `json:"icon,omitempty"`
	Color        string `json:"color,omitempty"`
	DisplayOrder int    `json:"display_order"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateTagRequest 创建标签请求
type CreateTagRequest struct {
	TagCode      string `json:"tag_code" validate:"required,min=1,max=100"`
	TagName      string `json:"tag_name" validate:"required,min=1,max=200"`
	Category     string `json:"category" validate:"required,oneof=item skill hero quest achievement"`
	Description  string `json:"description,omitempty" validate:"max=500"`
	Icon         string `json:"icon,omitempty" validate:"max=200"`
	Color        string `json:"color,omitempty" validate:"omitempty,hexcolor"`
	DisplayOrder int    `json:"display_order" validate:"min=0"`
}

// UpdateTagRequest 更新标签请求
type UpdateTagRequest struct {
	TagCode      *string `json:"tag_code,omitempty" validate:"omitempty,min=1,max=100"`
	TagName      *string `json:"tag_name,omitempty" validate:"omitempty,min=1,max=200"`
	Category     *string `json:"category,omitempty" validate:"omitempty,oneof=item skill hero quest achievement"`
	Description  *string `json:"description,omitempty" validate:"omitempty,max=500"`
	Icon         *string `json:"icon,omitempty" validate:"omitempty,max=200"`
	Color        *string `json:"color,omitempty" validate:"omitempty,hexcolor"`
	DisplayOrder *int    `json:"display_order,omitempty" validate:"omitempty,min=0"`
	IsActive     *bool   `json:"is_active,omitempty"`
}

// TagListResponse 标签列表响应
type TagListResponse struct {
	Tags     []TagResponse `json:"tags"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

