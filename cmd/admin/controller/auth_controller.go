// File: internal/app/admin/controller/auth_controller.go
package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/app/admin/service"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/xerrors"
)

// AuthController 认证控制器
type AuthController struct {
	authService *service.AuthService
	respWriter  response.Writer
	logger      log.Logger
}

// NewAuthController 创建认证控制器
func NewAuthController(authService *service.AuthService, respWriter response.Writer, logger log.Logger) *AuthController {
	return &AuthController{
		authService: authService,
		respWriter:  respWriter,
		logger:      logger,
	}
}

// Login 处理登录
func (ac *AuthController) Login(c echo.Context) error {
	ctx := c.Request().Context()

	var req service.LoginRequest
	if err := c.Bind(&req); err != nil {
		appErr := xerrors.NewValidationError("request", "请求格式错误")
		return ac.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	// 验证请求
	if appErr := ac.authService.ValidateLoginRequest(&req); appErr != nil {
		return ac.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	// 执行登录
	result, appErr := ac.authService.Login(ctx, &req)
	if appErr != nil {
		return ac.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	// 设置会话 Cookie
	if result.SessionCookie != "" {
		c.Response().Header().Set("Set-Cookie", result.SessionCookie)
	}

	// 返回成功响应
	respData := map[string]interface{}{
		"success":       result.Success,
		"session_token": result.SessionToken,
		"message":       "登录成功",
	}

	return ac.respWriter.WriteSuccess(ctx, c.Response().Writer, respData)
}

// Register 处理注册
func (ac *AuthController) Register(c echo.Context) error {
	ctx := c.Request().Context()

	var req service.RegisterRequest
	if err := c.Bind(&req); err != nil {
		appErr := xerrors.NewValidationError("request", "请求格式错误")
		return ac.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	// 验证请求
	if appErr := ac.authService.ValidateRegisterRequest(&req); appErr != nil {
		return ac.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	// 执行注册
	result, appErr := ac.authService.Register(ctx, &req)
	if appErr != nil {
		return ac.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	// 设置会话 Cookie
	if result.SessionCookie != "" {
		c.Response().Header().Set("Set-Cookie", result.SessionCookie)
	}

	// 返回成功响应
	respData := map[string]interface{}{
		"success":       result.Success,
		"identity_id":   result.IdentityID,
		"session_token": result.SessionToken,
		"message":       "注册成功",
	}

	return ac.respWriter.WriteSuccess(ctx, c.Response().Writer, respData)
}

// Logout 处理登出
func (ac *AuthController) Logout(c echo.Context) error {
	ctx := c.Request().Context()

	sessionToken := c.Request().Header.Get("X-Session-Token")
	if sessionToken == "" {
		appErr := xerrors.NewValidationError("session_token", "会话令牌缺失")
		return ac.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	if appErr := ac.authService.Logout(ctx, sessionToken); appErr != nil {
		return ac.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	respData := map[string]string{
		"message": "登出成功",
	}

	return ac.respWriter.WriteSuccess(ctx, c.Response().Writer, respData)
}

// GetSession 获取会话信息
func (ac *AuthController) GetSession(c echo.Context) error {
	ctx := c.Request().Context()

	sessionToken := c.Request().Header.Get("X-Session-Token")
	if sessionToken == "" {
		appErr := xerrors.NewValidationError("session_token", "会话令牌缺失")
		return ac.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	session, appErr := ac.authService.GetSession(ctx, sessionToken)
	if appErr != nil {
		return ac.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	return ac.respWriter.WriteSuccess(ctx, c.Response().Writer, session)
}

// InitRecovery 初始化账户恢复
func (ac *AuthController) InitRecovery(c echo.Context) error {
	ctx := c.Request().Context()

	var req service.RecoveryRequest
	if err := c.Bind(&req); err != nil {
		appErr := xerrors.NewValidationError("request", "请求格式错误")
		return ac.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	if appErr := ac.authService.InitRecovery(ctx, &req); appErr != nil {
		return ac.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	respData := map[string]string{
		"message": "恢复邮件已发送",
	}

	return ac.respWriter.WriteSuccess(ctx, c.Response().Writer, respData)
}

// SubmitRecovery 提交恢复请求
func (ac *AuthController) SubmitRecovery(c echo.Context) error {
	// 这里需要根据具体的恢复流程实现
	// 通常需要处理恢复令牌和新密码
	return c.JSON(http.StatusNotImplemented, map[string]string{
		"message": "功能开发中",
	})
}
