package usermodel

// CreateIdentityRequest 创建身份请求
type CreateIdentityRequest struct {
	Email    string                 `json:"email" binding:"required,email"`
	Username string                 `json:"username" binding:"required,min=3,max=30"`
	Password string                 `json:"password,omitempty" binding:"omitempty,min=8"`
	Traits   map[string]interface{} `json:"traits,omitempty"`
}

// UpdateIdentityRequest 更新身份请求
type UpdateIdentityRequest struct {
	Email    string                 `json:"email,omitempty" binding:"omitempty,email"`
	Username string                 `json:"username,omitempty" binding:"omitempty,min=3,max=30"`
	Traits   map[string]interface{} `json:"traits,omitempty"`
}

// ListIdentitiesQuery 列表查询参数
type ListIdentitiesQuery struct {
	Page    int64  `query:"page" validate:"min=1"`
	PerPage int64  `query:"per_page" validate:"min=1,max=1000"`
	Ids     string `query:"ids"`
}
