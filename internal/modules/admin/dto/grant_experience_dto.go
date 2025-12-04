package dto

// GrantExperienceRequest 英雄加经验（测试/运营工具）
type GrantExperienceRequest struct {
	HeroID string `json:"hero_id" validate:"required,uuid4"`
	Amount int64  `json:"amount" validate:"required,min=1"`
}

type GrantExperienceResponse struct {
	Added int64 `json:"added"`
}
