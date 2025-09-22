// internal/modules/auth/rpc_handle.go - 完整版本
package auth

import (
	"context"

	"tsu-self/internal/model/authmodel"
	"tsu-self/internal/modules/auth/service"
	"tsu-self/internal/pkg/log"
)

type AuthRPCHandler struct {
	kratosService       *service.KratosService
	ketoService         *service.KetoService
	sessionService      *service.SessionService
	syncService         *service.SyncService
	notificationService *service.NotificationService
	logger              log.Logger
}

func NewAuthRPCHandler(
	kratosService *service.KratosService,
	ketoService *service.KetoService,
	sessionService *service.SessionService,
	syncService *service.SyncService,
	notificationService *service.NotificationService,
	logger log.Logger,
) *AuthRPCHandler {
	return &AuthRPCHandler{
		kratosService:       kratosService,
		ketoService:         ketoService,
		sessionService:      sessionService,
		syncService:         syncService,
		notificationService: notificationService,
		logger:              logger,
	}
}

func (h *AuthRPCHandler) Login(req *authmodel.LoginRPCRequest) (*authmodel.LoginRPCResponse, error) {
	ctx := context.Background()

	h.logger.InfoContext(ctx, "RPC Login 请求", log.String("identifier", req.Identifier))

	// 调用 Kratos 登录
	kratosResp, err := h.kratosService.Login(ctx, req)
	if err != nil {
		h.logger.ErrorContext(ctx, "登录失败", log.Any("error", err))
		return nil, err
	}

	// 创建会话
	token, sessionErr := h.sessionService.CreateSession(ctx, kratosResp.UserInfo, req.ClientIP, req.UserAgent)
	if sessionErr != nil {
		h.logger.WarnContext(ctx, "创建会话失败，使用 Kratos token", log.Any("error", sessionErr))
		token = kratosResp.Token
	}

	// 更新响应中的 token
	kratosResp.Token = token

	h.logger.InfoContext(ctx, "登录成功", log.String("user_id", kratosResp.UserInfo.ID))
	return kratosResp, nil
}

func (h *AuthRPCHandler) Register(req *authmodel.RegisterRPCRequest) (*authmodel.RegisterRPCResponse, error) {
	ctx := context.Background()

	h.logger.InfoContext(ctx, "RPC Register 请求",
		log.String("email", req.Email),
		log.String("username", req.Username))

	// 调用 Kratos 注册
	kratosResp, err := h.kratosService.Register(ctx, req)
	if err != nil {
		h.logger.ErrorContext(ctx, "注册失败", log.Any("error", err))
		return nil, err
	}

	// 创建会话
	token, sessionErr := h.sessionService.CreateSession(ctx, kratosResp.UserInfo, req.ClientIP, req.UserAgent)
	if sessionErr != nil {
		h.logger.WarnContext(ctx, "创建会话失败，使用 Kratos token", log.Any("error", sessionErr))
		token = kratosResp.Token
	}

	// 更新响应中的 token
	kratosResp.Token = token

	h.logger.InfoContext(ctx, "注册成功", log.String("user_id", kratosResp.UserInfo.ID))
	return kratosResp, nil
}

func (h *AuthRPCHandler) ValidateToken(req *authmodel.ValidateTokenRequest) (*authmodel.ValidateTokenResponse, error) {
	ctx := context.Background()

	// 首先尝试 Session Service 验证
	sessionResp, err := h.sessionService.ValidateToken(ctx, req.Token)
	if err == nil && sessionResp.Valid {
		return sessionResp, nil
	}

	// Session 验证失败，尝试 Kratos 验证
	h.logger.DebugContext(ctx, "Session 验证失败，尝试 Kratos 验证")

	userInfo, kratosErr := h.kratosService.GetUserInfo(ctx, req.Token)
	if kratosErr != nil {
		return &authmodel.ValidateTokenResponse{Valid: false}, nil
	}

	return &authmodel.ValidateTokenResponse{
		Valid:    true,
		UserID:   userInfo.ID,
		UserInfo: userInfo,
	}, nil
}

func (h *AuthRPCHandler) Logout(req *authmodel.LogoutRequest) (*authmodel.LogoutResponse, error) {
	ctx := context.Background()

	h.logger.InfoContext(ctx, "RPC Logout 请求", log.String("user_id", req.UserID))

	// 清理会话
	if err := h.sessionService.InvalidateAllUserSessions(ctx, req.UserID); err != nil {
		h.logger.WarnContext(ctx, "清理会话失败", log.Any("error", err))
	}

	// 调用 Kratos 登出
	if logoutErr := h.kratosService.Logout(ctx, req.Token); logoutErr != nil {
		h.logger.WarnContext(ctx, "Kratos 登出失败", log.Any("error", logoutErr))
	}

	return &authmodel.LogoutResponse{Success: true}, nil
}

func (h *AuthRPCHandler) CheckPermission(req *authmodel.CheckPermissionRequest) (*authmodel.CheckPermissionResponse, error) {
	ctx := context.Background()

	h.logger.DebugContext(ctx, "RPC CheckPermission 请求",
		log.String("user_id", req.UserID),
		log.String("resource", req.Resource),
		log.String("action", req.Action))

	return h.ketoService.CheckPermission(ctx, req)
}

func (h *AuthRPCHandler) GetUserInfo(req *authmodel.GetUserInfoRequest) (*authmodel.GetUserInfoResponse, error) {
	ctx := context.Background()

	userInfo, err := h.syncService.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	return &authmodel.GetUserInfoResponse{
		UserInfo: userInfo,
	}, nil
}

func (h *AuthRPCHandler) UpdateUserTraits(req *authmodel.UpdateUserTraitsRequest) (*authmodel.UpdateUserTraitsResponse, error) {
	ctx := context.Background()

	return h.kratosService.UpdateUserTraits(ctx, req)
}

func (h *AuthRPCHandler) AssignRole(req *authmodel.AssignRoleRequest) (*authmodel.AssignRoleResponse, error) {
	ctx := context.Background()

	// 分配角色
	resp, err := h.ketoService.AssignRole(ctx, req)
	if err != nil {
		return nil, err
	}

	// 发布权限变更事件
	roles := []string{req.Role}
	metadata := map[string]interface{}{
		"action": "role_assigned",
		"role":   req.Role,
	}

	if notifyErr := h.notificationService.PublishPermissionChanged(ctx, req.UserID, "role_assigned", roles, metadata); notifyErr != nil {
		h.logger.WarnContext(ctx, "发布权限变更事件失败", log.Any("error", notifyErr))
	}

	return resp, nil
}

func (h *AuthRPCHandler) RevokeRole(req *authmodel.RevokeRoleRequest) (*authmodel.RevokeRoleResponse, error) {
	ctx := context.Background()

	// 撤销角色
	resp, err := h.ketoService.RevokeRole(ctx, req)
	if err != nil {
		return nil, err
	}

	// 发布权限变更事件
	roles := []string{req.Role}
	metadata := map[string]interface{}{
		"action": "role_revoked",
		"role":   req.Role,
	}

	if notifyErr := h.notificationService.PublishPermissionChanged(ctx, req.UserID, "role_revoked", roles, metadata); notifyErr != nil {
		h.logger.WarnContext(ctx, "发布权限变更事件失败", log.Any("error", notifyErr))
	}

	return resp, nil
}

func (h *AuthRPCHandler) CreateRole(req *authmodel.CreateRoleRequest) (*authmodel.CreateRoleResponse, error) {
	ctx := context.Background()

	return h.ketoService.CreateRole(ctx, req)
}
