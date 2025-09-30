package user

// UpdateProfileRequest HTTP API 更新用户信息请求
// @Description 更新用户基本信息请求参数
type UpdateProfileRequest struct {
	// 新昵称
	Nickname string `json:"nickname" binding:"omitempty,max=30" validate:"omitempty,max=30" example:"newnickname"`
	// 新邮箱
	Email string `json:"email" binding:"omitempty,email" validate:"omitempty,email" example:"newemail@example.com"`
	// 新手机号
	PhoneNumber string `json:"phone_number" binding:"omitempty,e164" validate:"omitempty,e164" example:"+11234567890"`
}
