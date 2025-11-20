package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/liangdas/mqant/module"
	mqrpc "github.com/liangdas/mqant/rpc"
	"google.golang.org/protobuf/proto"

	authpb "tsu-self/internal/pb/auth"
	"tsu-self/internal/pkg/metrics"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/validator"
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
	Email    string `json:"email" validate:"required,email" example:"player@example.com"` // 邮箱地址（必填，用于登录和密码找回）
	Username string `json:"username" validate:"required,min=3" example:"warrior123"`      // 用户名（必填，3-20字符，用于登录）
	Password string `json:"password" validate:"required,min=6" example:"password123"`     // 密码（必填，最少6字符）
}

// RegisterResponse HTTP registration response
type RegisterResponse struct {
	UserID     string `json:"user_id" example:"user-123"`         // 用户ID
	KratosID   string `json:"kratos_id" example:"kratos-456"`     // Kratos系统ID（内部使用）
	Email      string `json:"email" example:"player@example.com"` // 邮箱地址
	Username   string `json:"username" example:"warrior123"`      // 用户名
	NeedVerify bool   `json:"need_verify" example:"false"`        // 是否需要邮箱验证（当前版本为false）
}

// GetUserResponse HTTP get user response
type GetUserResponse struct {
	UserID      string  `json:"user_id" example:"user-123"`                                    // 用户ID
	Email       string  `json:"email" example:"player@example.com"`                            // 邮箱地址
	Username    string  `json:"username" example:"warrior123"`                                 // 用户名
	Nickname    *string `json:"nickname,omitempty" example:"勇敢的战士"`                            // 昵称（可选）
	AvatarURL   *string `json:"avatar_url,omitempty" example:"https://example.com/avatar.jpg"` // 头像URL（可选）
	IsBanned    bool    `json:"is_banned" example:"false"`                                     // 是否被封禁
	LoginCount  int     `json:"login_count" example:"10"`                                      // 登录次数
	LastLoginAt *string `json:"last_login_at,omitempty" example:"2025-10-17 10:30:00"`         // 最后登录时间
}

// ==================== HTTP Handlers ====================

// Register handles user registration
// @Summary 用户注册
// @Description 注册新的用户账号。注册后可以使用邮箱或用户名登录
// @Description
// @Description **填写说明**：
// @Description - `email`: 有效的邮箱地址，用于登录和密码找回
// @Description - `username`: 用户名，3-20字符，支持字母、数字、下划线，用于登录
// @Description - `password`: 密码，最少6字符，建议包含字母和数字
// @Description
// @Description **注册成功后**：
// @Description - 系统自动创建用户账号
// @Description - 可以直接登录，无需邮箱验证（当前版本）
// @Description - 返回用户ID，用于后续创建英雄
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册请求"
// @Success 200 {object} response.Response{data=RegisterResponse} "注册成功"
// @Failure 400 {object} response.Response "请求参数错误（如邮箱格式不正确、用户名已存在）"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/auth/register [post]
func (h *AuthHandler) Register(c echo.Context) error {
	// 1. 绑定和验证 HTTP 请求
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}

	if err := c.Validate(&req); err != nil {
		// 翻译验证错误为用户友好的中文消息
		friendlyMsg := validator.TranslateValidationError(err)
		return response.EchoBadRequest(c, h.respWriter, friendlyMsg)
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
		UserID:     rpcResp.UserId,
		KratosID:   rpcResp.KratosId,
		Email:      rpcResp.Email,
		Username:   rpcResp.Username,
		NeedVerify: rpcResp.NeedVerify,
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
// @Router /game/auth/users/{user_id} [get]
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
// @Router /game/auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	start := time.Now()
	outcome := "success"
	defer func() {
		metrics.DefaultLoginMetrics.ObserveDuration("game", outcome, time.Since(start))
	}()

	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		outcome = "error"
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}

	if err := c.Validate(&req); err != nil {
		// 翻译验证错误为用户友好的中文消息
		friendlyMsg := validator.TranslateValidationError(err)
		outcome = "error"
		return response.EchoBadRequest(c, h.respWriter, friendlyMsg)
	}

	// 构造 Protobuf RPC 请求
	sessionToken := readSessionToken(c)
	rpcReq := &authpb.LoginRequest{
		Identifier:    req.Identifier,
		Password:      req.Password,
		SessionToken:  sessionToken,
		ClientService: "game",
	}

	// 序列化 Protobuf 请求
	rpcReqBytes, err := proto.Marshal(rpcReq)
	if err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "序列化RPC请求失败")
		outcome = "error"
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
			outcome = "error"
			return response.EchoError(c, h.respWriter, appErr)
		}
		// 登录失败
		appErr := xerrors.NewAuthError("用户名或密码错误")
		outcome = "error"
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 反序列化 Protobuf 响应
	resultBytes, ok := result.([]byte)
	if !ok {
		appErr := xerrors.New(xerrors.CodeInternalError, "RPC响应类型错误")
		outcome = "error"
		return response.EchoError(c, h.respWriter, appErr)
	}

	rpcResp := &authpb.LoginResponse{}
	if err := proto.Unmarshal(resultBytes, rpcResp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		outcome = "error"
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
// @Router /game/auth/logout [post]
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
		SessionToken:  sessionToken,
		ClientService: "game",
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

func readSessionToken(c echo.Context) string {
	if token := c.Request().Header.Get("X-Session-Token"); token != "" {
		return token
	}
	if authHeader := c.Request().Header.Get(echo.HeaderAuthorization); authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
			return strings.TrimSpace(parts[1])
		}
	}
	if cookie, err := c.Cookie("ory_kratos_session"); err == nil && cookie != nil {
		return cookie.Value
	}
	return ""
}
