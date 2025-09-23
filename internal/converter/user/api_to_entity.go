package user

import (
	apiUser "tsu-self/internal/api/request/user"
	apiUserResp "tsu-self/internal/api/response/user"
	"tsu-self/internal/repository/entity"
)

// UpdateProfileRequestToEntity 把api请求转换为实体
func UpdateProfileRequestToEntity(req *apiUser.UpdateProfileRequest) map[string]interface{} {
	updates := make(map[string]interface{})

	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}

	return updates
}

// EntityToProfileResponse 把实体转换为api响应
func EntityToProfileResponse(user *entity.User) *apiUserResp.Profile {
	if user == nil {
		return nil
	}
	return &apiUserResp.Profile{
		ID:           user.ID,
		Username:     user.Username,
		Nickname:     user.Nickname,
		Phone:        user.PhoneNumber.String,
		Email:        user.Email,
		IsPremium:    user.IsPremium,
		DiamondCount: user.DiamondCount,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}
