package handler

import (
	"context"
	"time"

	"google.golang.org/protobuf/proto"

	"tsu-self/internal/modules/auth/service"
	authpb "tsu-self/internal/pb/auth"
	commonpb "tsu-self/internal/pb/common"
)

// RPCHandler handles RPC requests for auth module
type RPCHandler struct {
	authService *service.AuthService
}

// NewRPCHandler creates a new RPC handler
func NewRPCHandler(authService *service.AuthService) *RPCHandler {
	return &RPCHandler{
		authService: authService,
	}
}

// Register handles user registration via RPC
// mqant RPC 参数: protobuf 序列化的字节数组
// mqant RPC 返回: protobuf 序列化的字节数组或错误
func (h *RPCHandler) Register(reqBytes []byte) ([]byte, error) {
	ctx := context.Background()
	// 反序列化 Protobuf 请求
	req := &authpb.RegisterRequest{}
	if err := proto.Unmarshal(reqBytes, req); err != nil {
		return nil, err
	}

	// 调用 Service 层
	result, err := h.authService.Register(ctx, service.RegisterInput{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}

	// 构造 Protobuf 响应
	resp := &authpb.RegisterResponse{
		UserId:       result.UserID,
		KratosId:     result.KratosID,
		Email:        result.Email,
		Username:     result.Username,
		NeedVerify:   result.NeedVerify,
		SessionToken: result.SessionToken,
	}

	// 序列化 Protobuf 响应
	return proto.Marshal(resp)
}

// GetUser handles getting user info via RPC
func (h *RPCHandler) GetUser(reqBytes []byte) ([]byte, error) {
	ctx := context.Background()
	// 反序列化 Protobuf 请求
	req := &authpb.GetUserRequest{}
	if err := proto.Unmarshal(reqBytes, req); err != nil {
		return nil, err
	}

	// 调用 Service 层
	user, err := h.authService.GetUserByID(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	// Entity → Protobuf (使用共享的 UserInfo)
	userInfo := &commonpb.UserInfo{
		UserId:     user.ID,
		KratosId:   "", // TODO: 添加 kratos_id 字段到数据库
		Email:      user.Email,
		Username:   user.Username,
		IsBanned:   user.IsBanned,
		LoginCount: int32(user.LoginCount),
		CreatedAt:  user.CreatedAt.Unix(),
		UpdatedAt:  user.UpdatedAt.Unix(),
	}

	// 处理可选字段
	if user.Nickname.Valid {
		userInfo.Nickname = user.Nickname.String
	}
	if user.AvatarURL.Valid {
		userInfo.AvatarUrl = user.AvatarURL.String
	}
	if user.PhoneNumber.Valid {
		userInfo.PhoneNumber = user.PhoneNumber.String
	}
	if user.Bio.Valid {
		userInfo.Bio = user.Bio.String
	}
	if user.BirthDate.Valid {
		userInfo.BirthDate = user.BirthDate.Time.Format(time.RFC3339)
	}
	if user.Gender.Valid {
		userInfo.Gender = user.Gender.String
	}
	if user.Timezone.Valid {
		userInfo.Timezone = user.Timezone.String
	}
	if user.Language.Valid {
		userInfo.Language = user.Language.String
	}
	if user.BanUntil.Valid {
		userInfo.BanUntil = user.BanUntil.Time.Format(time.RFC3339)
	}
	if user.BanReason.Valid {
		userInfo.BanReason = user.BanReason.String
	}
	if user.LastLoginAt.Valid {
		userInfo.LastLoginAt = user.LastLoginAt.Time.Format(time.RFC3339)
	}
	if user.LastLoginIP.Valid {
		userInfo.LastLoginIp = user.LastLoginIP.String
	}

	// 构造响应
	resp := &authpb.GetUserResponse{
		User: userInfo,
	}

	// 序列化 Protobuf 响应
	return proto.Marshal(resp)
}

// UpdateLoginInfo handles updating login info via RPC
func (h *RPCHandler) UpdateLoginInfo(reqBytes []byte) ([]byte, error) {
	ctx := context.Background()
	// 反序列化 Protobuf 请求
	req := &authpb.UpdateLoginInfoRequest{}
	if err := proto.Unmarshal(reqBytes, req); err != nil {
		return nil, err
	}

	// 调用 Service 层
	err := h.authService.UpdateLoginInfo(ctx, req.UserId, req.LoginIp)

	// 构造响应
	resp := &authpb.UpdateLoginInfoResponse{
		Status: &commonpb.Status{
			Success: err == nil,
			Message: "",
		},
	}

	if err != nil {
		resp.Status.Message = err.Error()
	}

	// 序列化 Protobuf 响应
	return proto.Marshal(resp)
}

// SyncUserFromKratos handles syncing user from Kratos via RPC
func (h *RPCHandler) SyncUserFromKratos(reqBytes []byte) ([]byte, error) {
	ctx := context.Background()
	// 反序列化 Protobuf 请求
	req := &authpb.SyncUserFromKratosRequest{}
	if err := proto.Unmarshal(reqBytes, req); err != nil {
		return nil, err
	}

	// 调用 Service 层
	err := h.authService.SyncUserFromKratos(ctx, req.KratosId)

	// 构造响应
	resp := &authpb.SyncUserFromKratosResponse{
		Status: &commonpb.Status{
			Success: err == nil,
			Message: "",
		},
	}

	if err != nil {
		resp.Status.Message = err.Error()
	}
	// TODO: 同步成功后,可以返回用户信息

	// 序列化 Protobuf 响应
	return proto.Marshal(resp)
}

// Login handles user login via RPC
// 支持使用 email, username, 或 phone_number 登录
func (h *RPCHandler) Login(reqBytes []byte) ([]byte, error) {
	ctx := context.Background()

	// 反序列化 Protobuf 请求
	req := &authpb.LoginRequest{}
	if err := proto.Unmarshal(reqBytes, req); err != nil {
		return nil, err
	}

	// 调用 Service 层
	result, err := h.authService.Login(ctx, service.LoginInput{
		Identifier: req.Identifier,
		Password:   req.Password,
	})
	if err != nil {
		return nil, err
	}

	// 构造 Protobuf 响应
	resp := &authpb.LoginResponse{
		SessionToken: result.SessionToken,
		UserId:       result.UserID,
		Email:        result.Email,
		Username:     result.Username,
	}

	// 序列化 Protobuf 响应
	return proto.Marshal(resp)
}

// Logout handles user logout via RPC
func (h *RPCHandler) Logout(reqBytes []byte) ([]byte, error) {
	ctx := context.Background()

	// 反序列化 Protobuf 请求
	req := &authpb.LogoutRequest{}
	if err := proto.Unmarshal(reqBytes, req); err != nil {
		return nil, err
	}

	// 调用 Service 层
	err := h.authService.Logout(ctx, service.LogoutInput{
		SessionToken: req.SessionToken,
	})

	// 构造 Protobuf 响应
	resp := &authpb.LogoutResponse{
		Status: &commonpb.Status{
			Success: err == nil,
			Message: "",
		},
	}

	if err != nil {
		resp.Status.Message = err.Error()
	}

	// 序列化 Protobuf 响应
	return proto.Marshal(resp)
}

// ==================== 密码重置功能 ====================

// InitiateRecovery 用户发起密码恢复
func (h *RPCHandler) InitiateRecovery(reqBytes []byte) ([]byte, error) {
	ctx := context.Background()

	req := &authpb.InitiateRecoveryRequest{}
	if err := proto.Unmarshal(reqBytes, req); err != nil {
		return nil, err
	}

	err := h.authService.InitiatePasswordRecovery(ctx, req.Email)

	resp := &authpb.InitiateRecoveryResponse{
		CodeSent: err == nil,
		Message:  "验证码已发送到您的邮箱",
	}

	if err != nil {
		resp.Message = err.Error()
		resp.CodeSent = false
	}

	return proto.Marshal(resp)
}

// VerifyRecoveryCode 验证恢复验证码
func (h *RPCHandler) VerifyRecoveryCode(reqBytes []byte) ([]byte, error) {
	ctx := context.Background()

	req := &authpb.VerifyRecoveryCodeRequest{}
	if err := proto.Unmarshal(reqBytes, req); err != nil {
		return nil, err
	}

	flowToken, err := h.authService.VerifyRecoveryCode(ctx, req.Email, req.Code)

	resp := &authpb.VerifyRecoveryCodeResponse{
		Verified:     err == nil,
		Message:      "验证码验证成功,请设置新密码",
		SessionToken: flowToken, // 注意：这实际上是特权 settings flow ID
	}

	if err != nil {
		resp.Message = err.Error()
		resp.Verified = false
	}

	return proto.Marshal(resp)
}

// ResetPassword 重置密码
func (h *RPCHandler) ResetPassword(reqBytes []byte) ([]byte, error) {
	ctx := context.Background()

	req := &authpb.ResetPasswordRequest{}
	if err := proto.Unmarshal(reqBytes, req); err != nil {
		return nil, err
	}

	err := h.authService.ResetPassword(ctx, req.SessionToken, req.Email, req.NewPassword)

	resp := &authpb.ResetPasswordResponse{
		Status: &commonpb.Status{
			Success: err == nil,
			Message: "密码重置成功",
		},
		Message: "密码重置成功,请使用新密码登录",
	}

	if err != nil {
		resp.Status.Message = err.Error()
		resp.Status.Success = false
		resp.Message = err.Error()
	}

	return proto.Marshal(resp)
}

// ResetPasswordWithCode 验证码重置密码（验证码 + 新密码）
func (h *RPCHandler) ResetPasswordWithCode(reqBytes []byte) ([]byte, error) {
	ctx := context.Background()

	req := &authpb.ResetPasswordWithCodeRequest{}
	if err := proto.Unmarshal(reqBytes, req); err != nil {
		return nil, err
	}

	err := h.authService.ResetPasswordWithCode(ctx, req.Email, req.Code, req.NewPassword)

	resp := &authpb.ResetPasswordWithCodeResponse{
		Status: &commonpb.Status{
			Success: err == nil,
			Message: "密码重置成功",
		},
		Message: "密码重置成功,请使用新密码登录",
	}

	if err != nil {
		resp.Status.Message = err.Error()
		resp.Status.Success = false
		resp.Message = err.Error()
	}

	return proto.Marshal(resp)
}

// DeleteUser 删除用户
func (h *RPCHandler) DeleteUser(reqBytes []byte) ([]byte, error) {
	ctx := context.Background()

	req := &authpb.DeleteUserRequest{}
	if err := proto.Unmarshal(reqBytes, req); err != nil {
		return nil, err
	}

	err := h.authService.DeleteUser(ctx, req.UserId)

	resp := &authpb.DeleteUserResponse{
		Status: &commonpb.Status{
			Success: err == nil,
			Message: "用户删除成功",
		},
		Message: "用户删除成功",
	}

	if err != nil {
		resp.Status.Message = err.Error()
		resp.Status.Success = false
		resp.Message = err.Error()
	}

	return proto.Marshal(resp)
}
