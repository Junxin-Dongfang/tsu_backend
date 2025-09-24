// internal/modules/auth/rpc_handle_new.go - 使用 protobuf 的新版本
package auth

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"tsu-self/internal/modules/auth/service"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/rpc/generated/auth"
	"tsu-self/internal/rpc/generated/common"
	"tsu-self/internal/rpc/generated/user"
)

type AuthRPCHandler struct {
	kratosService       *service.KratosService
	ketoService         *service.KetoService
	sessionService      *service.SessionService
	notificationService *service.NotificationService
	logger              log.Logger
}

func NewAuthRPCHandler(
	kratosService *service.KratosService,
	ketoService *service.KetoService,
	sessionService *service.SessionService,
	notificationService *service.NotificationService,
	logger log.Logger,
) *AuthRPCHandler {
	return &AuthRPCHandler{
		kratosService:       kratosService,
		ketoService:         ketoService,
		sessionService:      sessionService,
		notificationService: notificationService,
		logger:              logger,
	}
}

// Login RPC 方法 - 使用 protobuf
func (h *AuthRPCHandler) Login(req *auth.LoginRequest) (*auth.LoginResponse, error) {
	ctx := context.Background()

	h.logger.InfoContext(ctx, "RPC Login 请求", log.String("identifier", req.Identifier))

	// 调用真正的 Kratos 登录
	resp, appErr := h.kratosService.Login(ctx, req)
	if appErr != nil {
		h.logger.ErrorContext(ctx, "Kratos 登录失败", log.Any("error", appErr))
		return &auth.LoginResponse{
			Success:      false,
			ErrorMessage: appErr.Error(),
		}, nil
	}

	return resp, nil
}

// Register RPC 方法 - 纯Kratos认证，不操作主数据库
func (h *AuthRPCHandler) Register(req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	ctx := context.Background()

	h.logger.InfoContext(ctx, "RPC Register 请求",
		log.String("email", req.Email),
		log.String("username", req.Username))

	// 调用 Kratos 注册（auth服务只负责身份认证）
	resp, appErr := h.kratosService.Register(ctx, req)
	if appErr != nil {
		h.logger.ErrorContext(ctx, "Kratos 注册失败", log.Any("error", appErr))
		return &auth.RegisterResponse{
			Success:      false,
			ErrorMessage: appErr.Error(),
		}, nil
	}

	// 如果注册成功，发布事件通知其他服务（可选）
	if resp.Success && resp.IdentityId != "" {
		h.logger.InfoContext(ctx, "Kratos 注册成功",
			log.String("identity_id", resp.IdentityId))

		// 发布用户注册事件，让admin服务监听并同步到主数据库
		if h.notificationService != nil {
			metadata := map[string]interface{}{
				"identity_id": resp.IdentityId,
				"email":       req.Email,
				"username":    req.Username,
				"phone":       req.Phone,
			}

			if notifyErr := h.notificationService.PublishUserRegistered(ctx, resp.IdentityId, metadata); notifyErr != nil {
				h.logger.WarnContext(ctx, "发布用户注册事件失败", log.Any("error", notifyErr))
			}
		}
	}

	return resp, nil
}

// ValidateToken RPC 方法 - 使用 protobuf
func (h *AuthRPCHandler) ValidateToken(req *auth.ValidateTokenRequest) (*auth.ValidateTokenResponse, error) {
	ctx := context.Background()

	// 首先尝试 Session Service 验证
	sessionResp, err := h.sessionService.ValidateToken(ctx, req.Token)
	if err == nil && sessionResp.Valid {
		// 转换为 protobuf 格式
		return &auth.ValidateTokenResponse{
			Valid:  true,
			UserId: sessionResp.UserId,
			UserInfo: &common.UserInfo{
				Id:       sessionResp.UserId,
				Email:    "user@example.com", // 需要从实际用户信息获取
				Username: "username",         // 需要从实际用户信息获取
				// 其他字段需要根据实际需要填充
			},
		}, nil
	}

	// Session 验证失败，返回无效
	return &auth.ValidateTokenResponse{Valid: false}, nil
}

// Logout RPC 方法 - 使用 protobuf
func (h *AuthRPCHandler) Logout(req *auth.LogoutRequest) (*auth.LogoutResponse, error) {
	ctx := context.Background()

	h.logger.InfoContext(ctx, "RPC Logout 请求", log.String("user_id", req.UserId))

	// 清理会话
	if err := h.sessionService.InvalidateAllUserSessions(ctx, req.UserId); err != nil {
		h.logger.WarnContext(ctx, "清理会话失败", log.Any("error", err))
	}

	return &auth.LogoutResponse{Success: true}, nil
}

// CheckPermission RPC 方法 - 使用 protobuf
func (h *AuthRPCHandler) CheckPermission(req *common.CheckPermissionRequest) (*common.CheckPermissionResponse, error) {
	ctx := context.Background()

	h.logger.DebugContext(ctx, "RPC CheckPermission 请求",
		log.String("user_id", req.UserId),
		log.String("resource", req.Resource),
		log.String("action", req.Action))

	// 调用 Keto 检查权限
	response, err := h.ketoService.CheckPermission(ctx, req)
	if err != nil {
		h.logger.WarnContext(ctx, "Keto 权限检查失败", log.Any("error", err))
		// 权限检查失败时默认拒绝访问
		return &common.CheckPermissionResponse{Allowed: false}, nil
	}

	return response, nil
}

// GetUserInfo RPC 方法 - 使用 protobuf
func (h *AuthRPCHandler) GetUserInfo(req *user.GetUserInfoRequest) (*user.GetUserInfoResponse, error) {
	ctx := context.Background()

	// 调用 Kratos 获取用户信息
	userInfo, err := h.kratosService.GetUserInfo(ctx, req.UserId)
	if err != nil {
		h.logger.ErrorContext(ctx, "Kratos 获取用户信息失败", log.Any("error", err))
		// 返回空的用户信息而不是错误，让调用方处理
		userInfo = &common.UserInfo{
			Id:        req.UserId,
			Email:     "",
			Username:  "",
			CreatedAt: timestamppb.New(time.Now()),
			UpdatedAt: timestamppb.New(time.Now()),
			Traits:    make(map[string]string),
		}
	}

	return &user.GetUserInfoResponse{
		UserInfo: userInfo,
	}, nil
}

// UpdateUserTraits RPC 方法 - 使用 protobuf
func (h *AuthRPCHandler) UpdateUserTraits(req *user.UpdateUserTraitsRequest) (*user.UpdateUserTraitsResponse, error) {
	// 调用 Kratos 更新用户特征
	// 实际实现中需要调用 Kratos API

	return &user.UpdateUserTraitsResponse{Success: true}, nil
}

// AssignRole RPC 方法 - 使用 protobuf
func (h *AuthRPCHandler) AssignRole(req *common.AssignRoleRequest) (*common.AssignRoleResponse, error) {
	ctx := context.Background()

	// 调用 Keto 分配角色
	response, err := h.ketoService.AssignRole(ctx, req)
	if err != nil {
		h.logger.ErrorContext(ctx, "Keto 角色分配失败", log.Any("error", err))
		return &common.AssignRoleResponse{Success: false}, nil
	}

	// 发布权限变更事件
	metadata := map[string]interface{}{
		"action": "role_assigned",
		"role":   req.Role,
	}

	if notifyErr := h.notificationService.PublishPermissionChanged(ctx, req.UserId, "role_assigned", []string{req.Role}, metadata); notifyErr != nil {
		h.logger.WarnContext(ctx, "发布权限变更事件失败", log.Any("error", notifyErr))
	}

	return response, nil
}

// RevokeRole RPC 方法 - 使用 protobuf
func (h *AuthRPCHandler) RevokeRole(req *common.RevokeRoleRequest) (*common.RevokeRoleResponse, error) {
	ctx := context.Background()

	// 调用 Keto 撤销角色
	response, err := h.ketoService.RevokeRole(ctx, req)
	if err != nil {
		h.logger.ErrorContext(ctx, "Keto 角色撤销失败", log.Any("error", err))
		return &common.RevokeRoleResponse{Success: false}, nil
	}

	// 发布权限变更事件
	metadata := map[string]interface{}{
		"action": "role_revoked",
		"role":   req.Role,
	}

	if notifyErr := h.notificationService.PublishPermissionChanged(ctx, req.UserId, "role_revoked", []string{req.Role}, metadata); notifyErr != nil {
		h.logger.WarnContext(ctx, "发布权限变更事件失败", log.Any("error", notifyErr))
	}

	return response, nil
}

// CreateRole RPC 方法 - 使用 protobuf
func (h *AuthRPCHandler) CreateRole(req *common.CreateRoleRequest) (*common.CreateRoleResponse, error) {
	ctx := context.Background()

	// 调用 Keto 创建角色
	response, err := h.ketoService.CreateRole(ctx, req)
	if err != nil {
		h.logger.ErrorContext(ctx, "Keto 角色创建失败", log.Any("error", err))
		return &common.CreateRoleResponse{Success: false}, nil
	}

	return response, nil
}
