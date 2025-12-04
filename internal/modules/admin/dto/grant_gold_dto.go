package dto

// GrantGoldRequest 仓库加金币（测试/运营工具）
type GrantGoldRequest struct {
	TeamID string `json:"team_id" validate:"required,uuid4"`
	Amount int64  `json:"amount" validate:"required,min=1"`
}

type GrantGoldResponse struct {
	Added int64 `json:"added"`
}
