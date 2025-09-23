package common

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"tsu-self/internal/api/response/user"
	"tsu-self/internal/repository/entity"
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

// UserInfoFromEntity 从数据库实体转换为 API UserProfile
func UserInfoFromEntity(userEntity *entity.User) *user.Profile {
	if userEntity == nil {
		return nil
	}

	return &user.Profile{
		ID:           userEntity.ID,
		Username:     userEntity.Username,
		Email:        userEntity.Email,
		IsPremium:    userEntity.IsPremium,
		DiamondCount: userEntity.DiamondCount,
		CreatedAt:    userEntity.CreatedAt,
		UpdatedAt:    userEntity.UpdatedAt,
	}
}

// UserInfoToEntity 从 API UserProfile 转换为数据库实体（部分字段）
func UserInfoToEntity(profile *user.Profile) *entity.User {
	if profile == nil {
		return nil
	}

	return &entity.User{
		ID:           profile.ID,
		Username:     profile.Username,
		Email:        profile.Email,
		IsPremium:    profile.IsPremium,
		DiamondCount: profile.DiamondCount,
		CreatedAt:    profile.CreatedAt,
		UpdatedAt:    profile.UpdatedAt,
		// 其他字段需要单独设置
		Timezone: "UTC",
		Language: "zh-CN",
	}
}

// EntityToRPCUserInfo 从数据库实体转换为 RPC UserInfo
func EntityToRPCUserInfo(userEntity *entity.User) *common.UserInfo {
	if userEntity == nil {
		return nil
	}

	return &common.UserInfo{
		Id:           userEntity.ID,
		Username:     userEntity.Username,
		Email:        userEntity.Email,
		IsPremium:    userEntity.IsPremium,
		DiamondCount: int32(userEntity.DiamondCount),
		CreatedAt:    timestamppb.New(userEntity.CreatedAt),
		UpdatedAt:    timestamppb.New(userEntity.UpdatedAt),
		Traits:       make(map[string]string),
	}
}