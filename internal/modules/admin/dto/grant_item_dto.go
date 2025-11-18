package dto

// GrantItemRequest 管理端发放物品请求
type GrantItemRequest struct {
	TargetType string `json:"target_type" validate:"required,oneof=user team_warehouse"` // user=玩家背包
	TargetID   string `json:"target_id" validate:"required"`                             // userID 或 teamID
	ItemID     string `json:"item_id" validate:"required,uuid4"`
	Quantity   int    `json:"quantity" validate:"required,min=1,max=999"`
}

// GrantItemResponse 发放结果
type GrantItemResponse struct {
	Granted int `json:"granted"`
}
