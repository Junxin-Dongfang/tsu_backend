package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/liangdas/mqant/module"
	mqrpc "github.com/liangdas/mqant/rpc"
	"google.golang.org/protobuf/proto"

	authpb "tsu-self/internal/pb/auth"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/xerrors"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	rpcCaller  module.RPCModule
	respWriter response.Writer
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(rpcCaller module.RPCModule, respWriter response.Writer) *AuthHandler {
	return &AuthHandler{
		rpcCaller:  rpcCaller,
		respWriter: respWriter,
	}
}

// ==================== HTTP Request/Response Models ====================
// 这些是 HTTP API 专用的结构,用于前端交互

// RegisterRequest HTTP registration request
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=3"`
	Password string `json:"password" validate:"required,min=6"`
}

// RegisterResponse HTTP registration response
type RegisterResponse struct {
	UserID       string `json:"user_id"`
	KratosID     string `json:"kratos_id"`
	Email        string `json:"email"`
	Username     string `json:"username"`
	SessionToken string `json:"session_token"` // Registration Flow 返回的 session token
	NeedVerify   bool   `json:"need_verify"`
}

// GetUserResponse HTTP get user response
type GetUserResponse struct {
	UserID      string  `json:"user_id"`
	Email       string  `json:"email"`
	Username    string  `json:"username"`
	Nickname    *string `json:"nickname,omitempty"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
	IsBanned    bool    `json:"is_banned"`
	LoginCount  int     `json:"login_count"`
	LastLoginAt *string `json:"last_login_at,omitempty"`
}

// ==================== HTTP Handlers ====================

// Register handles user registration
// @Summary 用户注册
// @Description 注册新的用户账号
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册请求"
// @Success 200 {object} response.Response{data=RegisterResponse} "注册成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/auth/register [post]
func (h *AuthHandler) Register(c echo.Context) error {
	// 1. 绑定和验证 HTTP 请求
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoValidationError(c, h.respWriter, err)
	}

	// 2. 构造 Protobuf RPC 请求
	rpcReq := &authpb.RegisterRequest{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	}

	// 3. 序列化 Protobuf 请求
	rpcReqBytes, err := proto.Marshal(rpcReq)
	if err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "序列化RPC请求失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 4. 调用 Auth RPC (使用 Call 方法，支持超时控制)
	ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
	defer cancel()

	result, errStr := h.rpcCaller.Call(
		ctx,
		"auth",
		"Register",
		mqrpc.Param(rpcReqBytes),
	)

	if errStr != "" {
		// 检查超时
		if ctx.Err() == context.DeadlineExceeded {
			appErr := xerrors.New(xerrors.CodeExternalServiceError, "Auth服务超时")
			return response.EchoError(c, h.respWriter, appErr)
		}
		// RPC 错误,可能是业务错误
		appErr := xerrors.New(xerrors.CodeExternalServiceError, errStr)
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 5. 反序列化 Protobuf 响应
	resultBytes, ok := result.([]byte)
	if !ok {
		appErr := xerrors.New(xerrors.CodeInternalError, "RPC响应类型错误")
		return response.EchoError(c, h.respWriter, appErr)
	}

	rpcResp := &authpb.RegisterResponse{}
	if err := proto.Unmarshal(resultBytes, rpcResp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 6. 转换为 HTTP Response Model
	resp := RegisterResponse{
		UserID:       rpcResp.UserId,
		KratosID:     rpcResp.KratosId,
		Email:        rpcResp.Email,
		Username:     rpcResp.Username,
		SessionToken: rpcResp.SessionToken,
		NeedVerify:   rpcResp.NeedVerify,
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// GetUser handles getting user info
// @Summary 获取用户信息
// @Description 通过用户ID获取用户信息
// @Tags 认证
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Success 200 {object} response.Response{data=GetUserResponse} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "用户不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/auth/users/{user_id} [get]
// @Security BearerAuth
func (h *AuthHandler) GetUser(c echo.Context) error {
	userID := c.Param("user_id")
	if userID == "" {
		return response.EchoBadRequest(c, h.respWriter, "用户ID不能为空")
	}

	// 构造 Protobuf RPC 请求
	rpcReq := &authpb.GetUserRequest{
		UserId: userID,
	}

	// 序列化 Protobuf 请求
	rpcReqBytes, err := proto.Marshal(rpcReq)
	if err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "序列化RPC请求失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 调用 Auth RPC (使用 Call 方法)
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	result, errStr := h.rpcCaller.Call(
		ctx,
		"auth",
		"GetUser",
		mqrpc.Param(rpcReqBytes),
	)
	if errStr != "" {
		if ctx.Err() == context.DeadlineExceeded {
			appErr := xerrors.New(xerrors.CodeExternalServiceError, "Auth服务超时")
			return response.EchoError(c, h.respWriter, appErr)
		}
		// 用户不存在
		appErr := xerrors.NewUserNotFoundError(userID)
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 反序列化 Protobuf 响应
	resultBytes, ok := result.([]byte)
	if !ok {
		appErr := xerrors.New(xerrors.CodeInternalError, "RPC响应类型错误")
		return response.EchoError(c, h.respWriter, appErr)
	}

	rpcResp := &authpb.GetUserResponse{}
	if err := proto.Unmarshal(resultBytes, rpcResp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// Protobuf → HTTP Response Model
	user := rpcResp.User
	resp := GetUserResponse{
		UserID:     user.UserId,
		Email:      user.Email,
		Username:   user.Username,
		IsBanned:   user.IsBanned,
		LoginCount: int(user.LoginCount),
	}

	// 处理可选字段
	if user.Nickname != "" {
		resp.Nickname = &user.Nickname
	}
	if user.AvatarUrl != "" {
		resp.AvatarURL = &user.AvatarUrl
	}
	if user.LastLoginAt != "" {
		resp.LastLoginAt = &user.LastLoginAt
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// LoginRequest HTTP login request
type LoginRequest struct {
	Identifier string `json:"identifier" validate:"required"` // email or username
	Password   string `json:"password" validate:"required"`
}

// LoginResponse HTTP login response
type LoginResponse struct {
	SessionToken string `json:"session_token"`
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
	Username     string `json:"username"`
}

// Login handles user login
// @Summary 用户登录
// @Description 使用邮箱或用户名和密码登录,返回会话信息
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录请求"
// @Success 200 {object} response.Response{data=LoginResponse} "登录成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "认证失败"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoValidationError(c, h.respWriter, err)
	}

	// 构造 Protobuf RPC 请求
	rpcReq := &authpb.LoginRequest{
		Identifier: req.Identifier,
		Password:   req.Password,
	}

	// 序列化 Protobuf 请求
	rpcReqBytes, err := proto.Marshal(rpcReq)
	if err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "序列化RPC请求失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 调用 Auth RPC (使用 Call 方法)
	ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
	defer cancel()

	result, errStr := h.rpcCaller.Call(
		ctx,
		"auth",
		"Login",
		mqrpc.Param(rpcReqBytes),
	)

	if errStr != "" {
		if ctx.Err() == context.DeadlineExceeded {
			appErr := xerrors.New(xerrors.CodeExternalServiceError, "Auth服务超时")
			return response.EchoError(c, h.respWriter, appErr)
		}
		// 登录失败
		appErr := xerrors.NewAuthError("用户名或密码错误")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 反序列化 Protobuf 响应
	resultBytes, ok := result.([]byte)
	if !ok {
		appErr := xerrors.New(xerrors.CodeInternalError, "RPC响应类型错误")
		return response.EchoError(c, h.respWriter, appErr)
	}

	rpcResp := &authpb.LoginResponse{}
	if err := proto.Unmarshal(resultBytes, rpcResp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 设置 Session Cookie
	c.SetCookie(&http.Cookie{
		Name:     "ory_kratos_session",
		Value:    rpcResp.SessionToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400, // 24 hours
	})

	// 将 Session Token 作为 Bearer Token 写入响应头,方便前端直接复用
	c.Response().Header().Set("Authorization", "Bearer "+rpcResp.SessionToken)

	// 转换为 HTTP Response
	resp := LoginResponse{
		SessionToken: rpcResp.SessionToken,
		UserID:       rpcResp.UserId,
		Email:        rpcResp.Email,
		Username:     rpcResp.Username,
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// Logout handles user logout
// @Summary 用户登出
// @Description 登出并使会话失效
// @Tags 认证
// @Accept json
// @Produce json
// @Success 200 {object} response.Response "登出成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/auth/logout [post]
// @Security BearerAuth
func (h *AuthHandler) Logout(c echo.Context) error {
	// 从 Cookie 中获取 Session Token
	cookie, err := c.Cookie("ory_kratos_session")
	var sessionToken string

	if err == nil && cookie != nil {
		sessionToken = cookie.Value
	} else {
		// 尝试从 Header 中获取
		sessionToken = c.Request().Header.Get("X-Session-Token")
	}

	if sessionToken == "" {
		return response.EchoBadRequest(c, h.respWriter, "未找到会话令牌")
	}

	// 构造 Protobuf RPC 请求
	rpcReq := &authpb.LogoutRequest{
		SessionToken: sessionToken,
	}

	// 序列化 Protobuf 请求
	rpcReqBytes, err := proto.Marshal(rpcReq)
	if err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "序列化RPC请求失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 调用 Auth RPC (使用 Call 方法)
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	_, errStr := h.rpcCaller.Call(
		ctx,
		"auth",
		"Logout",
		mqrpc.Param(rpcReqBytes),
	)

	if errStr != "" {
		if ctx.Err() == context.DeadlineExceeded {
			appErr := xerrors.New(xerrors.CodeExternalServiceError, "Auth服务超时")
			return response.EchoError(c, h.respWriter, appErr)
		}
		appErr := xerrors.New(xerrors.CodeExternalServiceError, "登出失败: "+errStr)
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 清除 Cookie
	c.SetCookie(&http.Cookie{
		Name:     "ory_kratos_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1, // Delete cookie
	})

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": "登出成功",
	})
}

// DeleteUser 删除用户（管理员操作）
// @Summary 删除用户
// @Description 删除指定用户（软删除业务数据 + 删除 Kratos identity）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "用户不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/users/{user_id} [delete]
// @Security BearerAuth
func (h *AuthHandler) DeleteUser(c echo.Context) error {
	userID := c.Param("user_id")
	if userID == "" {
		return response.EchoBadRequest(c, h.respWriter, "用户ID不能为空")
	}

	// 构造 Protobuf RPC 请求
	rpcReq := &authpb.DeleteUserRequest{
		UserId: userID,
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
		"DeleteUser",
		mqrpc.Param(rpcReqBytes),
	)

	if errStr != "" {
		if ctx.Err() == context.DeadlineExceeded {
			appErr := xerrors.New(xerrors.CodeExternalServiceError, "Auth服务超时")
			return response.EchoError(c, h.respWriter, appErr)
		}
		// 用户不存在或删除失败
		appErr := xerrors.New(xerrors.CodeExternalServiceError, "删除用户失败: "+errStr)
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 反序列化 Protobuf 响应
	resultBytes, ok := result.([]byte)
	if !ok {
		appErr := xerrors.New(xerrors.CodeInternalError, "RPC响应类型错误")
		return response.EchoError(c, h.respWriter, appErr)
	}

	rpcResp := &authpb.DeleteUserResponse{}
	if err := proto.Unmarshal(resultBytes, rpcResp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": rpcResp.Message,
		"success": rpcResp.Status.Success,
	})
}
