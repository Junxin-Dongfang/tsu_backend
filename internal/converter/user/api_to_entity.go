package user

import (
	apiUser "tsu-self/internal/api/model/request/user"
	apiUserResp "tsu-self/internal/api/model/response/user"
	"tsu-self/internal/entity"
	userDomain "tsu-self/internal/domain/user"
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
	if req.PhoneNumber != "" {
		updates["phone_number"] = req.PhoneNumber
	}

	return updates
}

// EntityToProfileResponse 把实体转换为api响应
func EntityToProfileResponse(user *userDomain.UserAggregate) *apiUserResp.Profile {
	if user == nil {
		return nil
	}
	// 处理可能为NULL的字段
	nickname := ""
	if user.Nickname.Valid {
		nickname = user.Nickname.String
	}

	phone := ""
	if user.PhoneNumber.Valid {
		phone = user.PhoneNumber.String
	}

	return &apiUserResp.Profile{
		ID:           user.ID,
		Username:     user.Username,
		Nickname:     nickname,
		Phone:        phone,
		Email:        user.Email,
		IsPremium:    user.IsPremium(),
		DiamondCount: user.GetDiamondCount(),
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

// BasicUserToProfileResponse 从基础User实体转换为API响应
func BasicUserToProfileResponse(user *entity.User) *apiUserResp.Profile {
	if user == nil {
		return nil
	}

	// 处理可能为NULL的字段
	nickname := ""
	if user.Nickname.Valid {
		nickname = user.Nickname.String
	}

	phone := ""
	if user.PhoneNumber.Valid {
		phone = user.PhoneNumber.String
	}

	return &apiUserResp.Profile{
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  nickname,
		Phone:     phone,
		Email:     user.Email,
		IsPremium: false, // 基础实体无财务信息，默认false
		DiamondCount: 0,  // 基础实体无财务信息，默认0
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
