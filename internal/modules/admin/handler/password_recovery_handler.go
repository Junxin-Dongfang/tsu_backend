package handler

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/liangdas/mqant/module"
	mqrpc "github.com/liangdas/mqant/rpc"
	"google.golang.org/protobuf/proto"

	authpb "tsu-self/internal/pb/auth"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/xerrors"
)

// PasswordRecoveryHandler 密码恢复 HTTP Handler
type PasswordRecoveryHandler struct {
	rpcCaller  module.RPCModule
	respWriter response.Writer
}

// NewPasswordRecoveryHandler 创建密码恢复 Handler
func NewPasswordRecoveryHandler(rpcCaller module.RPCModule, respWriter response.Writer) *PasswordRecoveryHandler {
	return &PasswordRecoveryHandler{
		rpcCaller:  rpcCaller,
		respWriter: respWriter,
	}
}

// ==================== HTTP Request/Response Models ====================

// InitiateRecoveryRequest 发起密码恢复请求
type InitiateRecoveryRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// InitiateRecoveryResponse 发起密码恢复响应
type InitiateRecoveryResponse struct {
	CodeSent bool   `json:"code_sent"`
	Message  string `json:"message"`
	// flow_id 已移除：后端通过 Redis 管理，前端不需要知道
}

// VerifyRecoveryCodeRequest 验证恢复码请求
type VerifyRecoveryCodeRequest struct {
	Email string `json:"email" validate:"required,email"` // 用户邮箱（用于从 Redis 获取 flow_id）
	Code  string `json:"code" validate:"required,len=6"`  // 验证码
	// flow_id 已移除：后端通过 Redis 管理
}

// VerifyRecoveryCodeResponse 验证恢复码响应
type VerifyRecoveryCodeResponse struct {
	Verified     bool   `json:"verified"`
	Message      string `json:"message"`
	SessionToken string `json:"session_token"` // Kratos 特权 session token
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	SessionToken string `json:"session_token" validate:"required"`      // Kratos session token
	Email        string `json:"email" validate:"required,email"`        // 用户邮箱
	NewPassword  string `json:"new_password" validate:"required,min=6"` // 新密码
}

// ResetPasswordResponse 重置密码响应
type ResetPasswordResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// AdminCreateRecoveryCodeRequest 管理员创建恢复码请求
type AdminCreateRecoveryCodeRequest struct {
	UserID    string `json:"user_id" validate:"required"`
	ExpiresIn string `json:"expires_in,omitempty"` // 例如: "12h", "1h"
}

// AdminCreateRecoveryCodeResponse 管理员创建恢复码响应
type AdminCreateRecoveryCodeResponse struct {
	RecoveryCode string `json:"recovery_code"`
	RecoveryLink string `json:"recovery_link"`
	ExpiresAt    string `json:"expires_at"`
}

// ==================== HTTP Handlers ====================

// InitiateRecovery 用户发起密码恢复
// @Summary 发起密码恢复
// @Description 请求发送密码恢复验证码到邮箱
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body InitiateRecoveryRequest true "恢复请求"
// @Success 200 {object} response.Response{data=InitiateRecoveryResponse} "验证码发送成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /auth/recovery/initiate [post]
func (h *PasswordRecoveryHandler) InitiateRecovery(c echo.Context) error {
	var req InitiateRecoveryRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 构造 Protobuf RPC 请求
	rpcReq := &authpb.InitiateRecoveryRequest{
		Email: req.Email,
	}

	// 序列化 Protobuf 请求
	rpcReqBytes, err := proto.Marshal(rpcReq)
	if err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "序列化RPC请求失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 调用 Auth RPC
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	result, errStr := h.rpcCaller.Call(
		ctx,
		"auth",
		"InitiateRecovery",
		mqrpc.Param(rpcReqBytes),
	)

	if errStr != "" {
		if ctx.Err() == context.DeadlineExceeded {
			appErr := xerrors.New(xerrors.CodeExternalServiceError, "Auth服务超时")
			return response.EchoError(c, h.respWriter, appErr)
		}
		appErr := xerrors.New(xerrors.CodeExternalServiceError, errStr)
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 反序列化 Protobuf 响应
	resultBytes, ok := result.([]byte)
	if !ok {
		appErr := xerrors.New(xerrors.CodeInternalError, "RPC响应类型错误")
		return response.EchoError(c, h.respWriter, appErr)
	}

	rpcResp := &authpb.InitiateRecoveryResponse{}
	if err := proto.Unmarshal(resultBytes, rpcResp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 转换为 HTTP Response
	resp := InitiateRecoveryResponse{
		CodeSent: rpcResp.CodeSent,
		Message:  rpcResp.Message,
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// VerifyRecoveryCode 验证恢复验证码
// @Summary 验证恢复码
// @Description 验证用户输入的恢复验证码
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body VerifyRecoveryCodeRequest true "验证请求"
// @Success 200 {object} response.Response{data=VerifyRecoveryCodeResponse} "验证成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "验证码错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /auth/recovery/verify [post]
func (h *PasswordRecoveryHandler) VerifyRecoveryCode(c echo.Context) error {
	var req VerifyRecoveryCodeRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 构造 Protobuf RPC 请求
	rpcReq := &authpb.VerifyRecoveryCodeRequest{
		Email: req.Email,
		Code:  req.Code,
	}

	// 序列化 Protobuf 请求
	rpcReqBytes, err := proto.Marshal(rpcReq)
	if err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "序列化RPC请求失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 调用 Auth RPC
	ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
	defer cancel()

	result, errStr := h.rpcCaller.Call(
		ctx,
		"auth",
		"VerifyRecoveryCode",
		mqrpc.Param(rpcReqBytes),
	)

	if errStr != "" {
		if ctx.Err() == context.DeadlineExceeded {
			appErr := xerrors.New(xerrors.CodeExternalServiceError, "Auth服务超时")
			return response.EchoError(c, h.respWriter, appErr)
		}
		appErr := xerrors.NewAuthError("验证码错误或已过期")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 反序列化 Protobuf 响应
	resultBytes, ok := result.([]byte)
	if !ok {
		appErr := xerrors.New(xerrors.CodeInternalError, "RPC响应类型错误")
		return response.EchoError(c, h.respWriter, appErr)
	}

	rpcResp := &authpb.VerifyRecoveryCodeResponse{}
	if err := proto.Unmarshal(resultBytes, rpcResp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 转换为 HTTP Response
	resp := VerifyRecoveryCodeResponse{
		Verified:     rpcResp.Verified,
		Message:      rpcResp.Message,
		SessionToken: rpcResp.SessionToken,
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// ResetPassword 重置密码
// @Summary 重置密码
// @Description 使用特权 Session Token 重置用户密码
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "重置请求"
// @Success 200 {object} response.Response{data=ResetPasswordResponse} "重置成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "Session Token 无效"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /auth/password/reset [post]
func (h *PasswordRecoveryHandler) ResetPassword(c echo.Context) error {
	var req ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 构造 Protobuf RPC 请求
	rpcReq := &authpb.ResetPasswordRequest{
		SessionToken: req.SessionToken,
		Email:        req.Email,
		NewPassword:  req.NewPassword,
	}

	// 序列化 Protobuf 请求
	rpcReqBytes, err := proto.Marshal(rpcReq)
	if err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "序列化RPC请求失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 调用 Auth RPC
	ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
	defer cancel()

	result, errStr := h.rpcCaller.Call(
		ctx,
		"auth",
		"ResetPassword",
		mqrpc.Param(rpcReqBytes),
	)

	if errStr != "" {
		if ctx.Err() == context.DeadlineExceeded {
			appErr := xerrors.New(xerrors.CodeExternalServiceError, "Auth服务超时")
			return response.EchoError(c, h.respWriter, appErr)
		}
		appErr := xerrors.New(xerrors.CodeExternalServiceError, "重置密码失败: "+errStr)
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 反序列化 Protobuf 响应
	resultBytes, ok := result.([]byte)
	if !ok {
		appErr := xerrors.New(xerrors.CodeInternalError, "RPC响应类型错误")
		return response.EchoError(c, h.respWriter, appErr)
	}

	rpcResp := &authpb.ResetPasswordResponse{}
	if err := proto.Unmarshal(resultBytes, rpcResp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 转换为 HTTP Response
	resp := ResetPasswordResponse{
		Success: rpcResp.Status.Success,
		Message: rpcResp.Message,
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// AdminCreateRecoveryCode 管理员为用户创建恢复码
// @Summary 创建用户恢复码 (管理员)
// @Description 管理员为指定用户生成密码恢复码和链接
// @Tags 管理员-用户管理
// @Accept json
// @Produce json
// @Param request body AdminCreateRecoveryCodeRequest true "创建恢复码请求"
// @Success 200 {object} response.Response{data=AdminCreateRecoveryCodeResponse} "创建成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "用户不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/users/recovery-code [post]
// @Security BearerAuth
func (h *PasswordRecoveryHandler) AdminCreateRecoveryCode(c echo.Context) error {
	var req AdminCreateRecoveryCodeRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 设置默认过期时间
	if req.ExpiresIn == "" {
		req.ExpiresIn = "12h"
	}

	// 构造 Protobuf RPC 请求
	rpcReq := &authpb.AdminCreateRecoveryCodeRequest{
		UserId:    req.UserID,
		ExpiresIn: req.ExpiresIn,
	}

	// 序列化 Protobuf 请求
	rpcReqBytes, err := proto.Marshal(rpcReq)
	if err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "序列化RPC请求失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 调用 Auth RPC
	ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
	defer cancel()

	result, errStr := h.rpcCaller.Call(
		ctx,
		"auth",
		"AdminCreateRecoveryCode",
		mqrpc.Param(rpcReqBytes),
	)

	if errStr != "" {
		if ctx.Err() == context.DeadlineExceeded {
			appErr := xerrors.New(xerrors.CodeExternalServiceError, "Auth服务超时")
			return response.EchoError(c, h.respWriter, appErr)
		}
		appErr := xerrors.NewUserNotFoundError(req.UserID)
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 反序列化 Protobuf 响应
	resultBytes, ok := result.([]byte)
	if !ok {
		appErr := xerrors.New(xerrors.CodeInternalError, "RPC响应类型错误")
		return response.EchoError(c, h.respWriter, appErr)
	}

	rpcResp := &authpb.AdminCreateRecoveryCodeResponse{}
	if err := proto.Unmarshal(resultBytes, rpcResp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 转换为 HTTP Response
	resp := AdminCreateRecoveryCodeResponse{
		RecoveryCode: rpcResp.RecoveryCode,
		RecoveryLink: rpcResp.RecoveryLink,
		ExpiresAt:    rpcResp.ExpiresAt,
	}

	return response.EchoOK(c, h.respWriter, resp)
}
