package handler

import (
	"context"
	"database/sql"

	"google.golang.org/protobuf/proto"

	"tsu-self/internal/modules/auth/service"
	authpb "tsu-self/internal/pb/auth"
	commonpb "tsu-self/internal/pb/common"
	"tsu-self/internal/repository/interfaces"
)

// UserRPCHandler 用户 RPC 处理器
type UserRPCHandler struct {
	db          *sql.DB
	userService *service.UserService
}

// NewUserRPCHandler 创建用户 RPC 处理器
func NewUserRPCHandler(db *sql.DB, userService *service.UserService) *UserRPCHandler {
	return &UserRPCHandler{
		db:          db,
		userService: userService,
	}
}

// GetUsers 获取用户列表 RPC
func (h *UserRPCHandler) GetUsers(reqBytes []byte) ([]byte, error) {
	ctx := context.Background()

	// 1. 解析请求
	req := &authpb.GetUsersRequest{}
	if err := proto.Unmarshal(reqBytes, req); err != nil {
		return nil, err
	}

	// 2. 构建查询参数
	params := interfaces.UserQueryParams{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
		Keyword:  req.Keyword,
		SortBy:   req.SortBy,
		SortDir:  req.SortDir,
	}

	if req.IsBanned != nil {
		params.IsBanned = req.IsBanned
	}

	// 3. 查询用户列表
	users, total, err := h.userService.GetUsers(ctx, params)
	if err != nil {
		return nil, err
	}

	// 4. 转换为 Protobuf 响应
	pbUsers := make([]*commonpb.UserInfo, 0, len(users))
	for _, u := range users {
		pbUser := &commonpb.UserInfo{
			UserId:     u.ID,
			Username:   u.Username,
			Email:      u.Email,
			IsBanned:   u.IsBanned,
			LoginCount: int32(u.LoginCount),
			CreatedAt:  u.CreatedAt.Unix(),
			UpdatedAt:  u.UpdatedAt.Unix(),
		}

		if u.Nickname.Valid {
			pbUser.Nickname = u.Nickname.String
		}
		if u.PhoneNumber.Valid {
			pbUser.PhoneNumber = u.PhoneNumber.String
		}
		if u.AvatarURL.Valid {
			pbUser.AvatarUrl = u.AvatarURL.String
		}
		if u.Bio.Valid {
			pbUser.Bio = u.Bio.String
		}
		if u.BirthDate.Valid {
			pbUser.BirthDate = u.BirthDate.Time.Format("2006-01-02T15:04:05Z07:00")
		}
		if u.Gender.Valid {
			pbUser.Gender = u.Gender.String
		}
		if u.Timezone.Valid {
			pbUser.Timezone = u.Timezone.String
		}
		if u.Language.Valid {
			pbUser.Language = u.Language.String
		}
		if u.BanUntil.Valid {
			pbUser.BanUntil = u.BanUntil.Time.Format("2006-01-02T15:04:05Z07:00")
		}
		if u.BanReason.Valid {
			pbUser.BanReason = u.BanReason.String
		}
		if u.LastLoginAt.Valid {
			pbUser.LastLoginAt = u.LastLoginAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}
		if u.LastLoginIP.Valid {
			pbUser.LastLoginIp = u.LastLoginIP.String
		}

		pbUsers = append(pbUsers, pbUser)
	}

	// 计算总页数
	totalPages := int32(total) / req.PageSize
	if int32(total)%req.PageSize > 0 {
		totalPages++
	}

	resp := &authpb.GetUsersResponse{
		Users:      pbUsers,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}

	return proto.Marshal(resp)
}

// UpdateUser 更新用户 RPC
func (h *UserRPCHandler) UpdateUser(reqBytes []byte) ([]byte, error) {
	ctx := context.Background()

	// 1. 解析请求
	req := &authpb.UpdateUserRequest{}
	if err := proto.Unmarshal(reqBytes, req); err != nil {
		return nil, err
	}

	// 2. 构建更新字段
	updates := make(map[string]interface{})

	if req.Username != nil {
		updates["username"] = *req.Username
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Nickname != nil {
		updates["nickname"] = *req.Nickname
	}
	if req.PhoneNumber != nil {
		updates["phone_number"] = *req.PhoneNumber
	}
	if req.AvatarUrl != nil {
		updates["avatar_url"] = *req.AvatarUrl
	}
	if req.Bio != nil {
		updates["bio"] = *req.Bio
	}
	if req.BirthDate != nil {
		updates["birth_date"] = *req.BirthDate
	}
	if req.Gender != nil {
		updates["gender"] = *req.Gender
	}
	if req.Timezone != nil {
		updates["timezone"] = *req.Timezone
	}
	if req.Language != nil {
		updates["language"] = *req.Language
	}

	// 3. 更新用户
	if err := h.userService.UpdateUser(ctx, req.UserId, updates); err != nil {
		return nil, err
	}

	// 4. 返回响应
	resp := &authpb.UpdateUserResponse{
		Status: &commonpb.Status{
			Success: true,
			Message: "用户信息更新成功",
		},
	}

	return proto.Marshal(resp)
}

// BanUser 封禁用户 RPC
func (h *UserRPCHandler) BanUser(reqBytes []byte) ([]byte, error) {
	ctx := context.Background()

	// 1. 解析请求
	req := &authpb.BanUserRequest{}
	if err := proto.Unmarshal(reqBytes, req); err != nil {
		return nil, err
	}

	// 2. 封禁用户
	if err := h.userService.BanUser(ctx, req.UserId, req.BanUntil, req.BanReason); err != nil {
		return nil, err
	}

	// 3. 返回响应
	resp := &authpb.BanUserResponse{
		Status: &commonpb.Status{
			Success: true,
			Message: "用户封禁成功",
		},
	}

	return proto.Marshal(resp)
}

// UnbanUser 解禁用户 RPC
func (h *UserRPCHandler) UnbanUser(reqBytes []byte) ([]byte, error) {
	ctx := context.Background()

	// 1. 解析请求
	req := &authpb.UnbanUserRequest{}
	if err := proto.Unmarshal(reqBytes, req); err != nil {
		return nil, err
	}

	// 2. 解禁用户
	if err := h.userService.UnbanUser(ctx, req.UserId); err != nil {
		return nil, err
	}

	// 3. 返回响应
	resp := &authpb.UnbanUserResponse{
		Status: &commonpb.Status{
			Success: true,
			Message: "用户解禁成功",
		},
	}

	return proto.Marshal(resp)
}
