package common

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"tsu-self/internal/api/model/response/user"
	userDomain "tsu-self/internal/domain/user"
	"tsu-self/internal/entity"
	"tsu-self/internal/rpc/generated/common"
)

// UserInfoFromRPC 从 RPC UserInfo 转换为 API UserProfile
func UserInfoFromRPC(rpcUser *common.UserInfo) *user.Profile {
	if rpcUser == nil {
		return nil
	}

	profile := &user.Profile{
		ID:           rpcUser.Id,
		Username:     rpcUser.Username,
		Email:        rpcUser.Email,
		IsPremium:    rpcUser.IsPremium,
		DiamondCount: int(rpcUser.DiamondCount),
	}

	// 转换时间戳
	if rpcUser.CreatedAt != nil {
		profile.CreatedAt = rpcUser.CreatedAt.AsTime()
	}
	if rpcUser.UpdatedAt != nil {
		profile.UpdatedAt = rpcUser.UpdatedAt.AsTime()
	}

	return profile
}

// UserInfoToRPC 从 API UserProfile 转换为 RPC UserInfo
func UserInfoToRPC(profile *user.Profile) *common.UserInfo {
	if profile == nil {
		return nil
	}

	return &common.UserInfo{
		Id:           profile.ID,
		Username:     profile.Username,
		Email:        profile.Email,
		IsPremium:    profile.IsPremium,
		DiamondCount: int32(profile.DiamondCount),
		CreatedAt:    timestamppb.New(profile.CreatedAt),
		UpdatedAt:    timestamppb.New(profile.UpdatedAt),
		Traits:       make(map[string]string), // 如果需要可以添加
	}
}

// UserInfoFromEntity 从数据库聚合实体转换为 API UserProfile
func UserInfoFromEntity(userEntity *userDomain.UserAggregate) *user.Profile {
	if userEntity == nil {
		return nil
	}

	return &user.Profile{
		ID:           userEntity.ID,
		Username:     userEntity.Username,
		Email:        userEntity.Email,
		IsPremium:    userEntity.IsPremium(),
		DiamondCount: userEntity.GetDiamondCount(),
		CreatedAt:    userEntity.CreatedAt,
		UpdatedAt:    userEntity.UpdatedAt,
	}
}

// UserInfoFromBasicModel 从基础User模型转换为 API UserProfile
func UserInfoFromBasicModel(userModel *entity.User) *user.Profile {
	if userModel == nil {
		return nil
	}

	return &user.Profile{
		ID:           userModel.ID,
		Username:     userModel.Username,
		Email:        userModel.Email,
		IsPremium:    false, // 基础模型没有财务信息，默认false
		DiamondCount: 0,     // 基础模型没有财务信息，默认0
		CreatedAt:    userModel.CreatedAt,
		UpdatedAt:    userModel.UpdatedAt,
	}
}

// UserInfoToEntity 从 API UserProfile 转换为数据库实体（部分字段）
// UserInfoToEntity 从 API UserProfile 转换为数据库模型（仅用户基础信息）
func UserInfoToEntity(profile *user.Profile) *entity.User {
	if profile == nil {
		return nil
	}

	return &entity.User{
		ID:        profile.ID,
		Username:  profile.Username,
		Email:     profile.Email,
		CreatedAt: profile.CreatedAt,
		UpdatedAt: profile.UpdatedAt,
		// 其他字段需要单独设置
	}
}

// EntityToRPCUserInfo 从数据库聚合实体转换为 RPC UserInfo
func EntityToRPCUserInfo(userEntity *userDomain.UserAggregate) *common.UserInfo {
	if userEntity == nil {
		return nil
	}

	return &common.UserInfo{
		Id:           userEntity.ID,
		Username:     userEntity.Username,
		Email:        userEntity.Email,
		IsPremium:    userEntity.IsPremium(),
		DiamondCount: int32(userEntity.GetDiamondCount()),
		CreatedAt:    timestamppb.New(userEntity.CreatedAt),
		UpdatedAt:    timestamppb.New(userEntity.UpdatedAt),
		Traits:       make(map[string]string),
	}
}
