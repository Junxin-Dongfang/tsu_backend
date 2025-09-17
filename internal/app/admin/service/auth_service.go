// File: internal/app/admin/service/auth_service.go
package service

import (
	"context"
	"net/http"

	client "github.com/ory/client-go"

	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/xerrors"
)

// AuthService 认证服务
type AuthService struct {
	publicClient *client.APIClient
	adminClient  *client.APIClient
	logger       log.Logger
}

// NewAuthService 创建认证服务
func NewAuthService(publicURL, adminURL string, logger log.Logger) (*AuthService, error) {
	// 创建公共客户端配置
	publicConfig := client.NewConfiguration()
	publicConfig.Servers = []client.ServerConfiguration{
		{URL: publicURL},
	}

	// 创建管理员客户端配置
	adminConfig := client.NewConfiguration()
	adminConfig.Servers = []client.ServerConfiguration{
		{URL: adminURL},
	}

	return &AuthService{
		publicClient: client.NewAPIClient(publicConfig),
		adminClient:  client.NewAPIClient(adminConfig),
		logger:       logger,
	}, nil
}

// LoginResult 登录结果
type LoginResult struct {
	Success       bool
	SessionToken  string
	SessionCookie string
}

// RegisterResult 注册结果
type RegisterResult struct {
	Success       bool
	IdentityID    string
	SessionToken  string
	SessionCookie string
}

// LoginRequest 登录请求
type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,min=3,max=30"`
	Password string `json:"password" binding:"required,min=8"`
}

// RecoveryRequest 恢复请求
type RecoveryRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// Login 执行登录
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResult, *xerrors.AppError) {
	s.logger.InfoContext(ctx, "开始登录流程",
		log.String("identifier", req.Identifier),
	)

	// 1. 创建登录流程
	flow, resp, err := s.publicClient.FrontendAPI.CreateNativeLoginFlow(ctx).Execute()
	if err != nil {
		return nil, s.handleKratosError(ctx, "create_login_flow", err, resp)
	}

	s.logger.DebugContext(ctx, "创建登录流程成功",
		log.String("flow_id", flow.Id),
	)

	// 2. 提交登录信息
	submitReq := client.UpdateLoginFlowBody{
		UpdateLoginFlowWithPasswordMethod: &client.UpdateLoginFlowWithPasswordMethod{
			Method:     "password",
			Identifier: req.Identifier,
			Password:   req.Password,
		},
	}

	loginResult, resp, err := s.publicClient.FrontendAPI.UpdateLoginFlow(ctx).
		Flow(flow.Id).
		UpdateLoginFlowBody(submitReq).
		Execute()

	if err != nil {
		return nil, s.handleKratosError(ctx, "submit_login", err, resp)
	}

	// 3. 处理登录结果
	if loginResult.Session.Identity != nil {
		s.logger.InfoContext(ctx, "登录成功",
			log.String("identity_id", loginResult.Session.Identity.Id),
		)

		sessionToken := ""
		if loginResult.SessionToken != nil {
			sessionToken = *loginResult.SessionToken
		}

		return &LoginResult{
			Success:       true,
			SessionToken:  sessionToken,
			SessionCookie: extractSessionCookie(resp),
		}, nil
	}

	return nil, xerrors.FromCode(xerrors.CodeInternalError).
		WithService("auth-service", "login").
		WithMetadata("unexpected_state", "no_session_in_successful_login")
}

// Register 执行注册
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*RegisterResult, *xerrors.AppError) {
	s.logger.InfoContext(ctx, "开始注册流程",
		log.String("email", req.Email),
		log.String("username", req.Username),
	)

	// 1. 创建注册流程
	flow, resp, err := s.publicClient.FrontendAPI.CreateNativeRegistrationFlow(ctx).Execute()
	if err != nil {
		return nil, s.handleKratosError(ctx, "create_registration_flow", err, resp)
	}

	s.logger.DebugContext(ctx, "创建注册流程成功",
		log.String("flow_id", flow.Id),
	)

	// 2. 提交注册信息
	traits := map[string]interface{}{
		"email":    req.Email,
		"username": req.Username,
	}

	submitReq := client.UpdateRegistrationFlowBody{
		UpdateRegistrationFlowWithPasswordMethod: &client.UpdateRegistrationFlowWithPasswordMethod{
			Method:   "password",
			Password: req.Password,
			Traits:   traits,
		},
	}

	registrationResult, resp, err := s.publicClient.FrontendAPI.UpdateRegistrationFlow(ctx).
		Flow(flow.Id).
		UpdateRegistrationFlowBody(submitReq).
		Execute()

	if err != nil {
		return nil, s.handleKratosError(ctx, "submit_registration", err, resp)
	}

	// 3. 处理注册结果
	if registrationResult.Session != nil {
		s.logger.InfoContext(ctx, "注册成功",
			log.String("identity_id", registrationResult.Session.Identity.Id),
			log.String("email", req.Email),
		)

		sessionToken := ""
		if registrationResult.SessionToken != nil {
			sessionToken = *registrationResult.SessionToken
		}

		return &RegisterResult{
			Success:       true,
			IdentityID:    registrationResult.Session.Identity.Id,
			SessionToken:  sessionToken,
			SessionCookie: extractSessionCookie(resp),
		}, nil
	}

	return nil, xerrors.FromCode(xerrors.CodeInternalError).
		WithService("auth-service", "register").
		WithMetadata("unexpected_state", "no_session_in_successful_registration")
}

// Logout 执行登出
func (s *AuthService) Logout(ctx context.Context, sessionToken string) *xerrors.AppError {
	s.logger.InfoContext(ctx, "开始登出流程")

	_, err := s.publicClient.FrontendAPI.PerformNativeLogout(ctx).
		PerformNativeLogoutBody(*client.NewPerformNativeLogoutBody(sessionToken)).
		Execute()
	if err != nil {
		return s.handleKratosError(ctx, "logout", err, nil)
	}
	s.logger.InfoContext(ctx, "登出成功")
	return nil
}

// GetSession 获取会话信息
func (s *AuthService) GetSession(ctx context.Context, sessionToken string) (*client.Session, *xerrors.AppError) {
	s.logger.DebugContext(ctx, "获取会话信息")

	session, resp, err := s.publicClient.FrontendAPI.ToSession(ctx).
		XSessionToken(sessionToken).
		Execute()

	if err != nil {
		return nil, s.handleKratosError(ctx, "get_session", err, resp)
	}

	return session, nil
}

// InitRecovery 初始化账户恢复
func (s *AuthService) InitRecovery(ctx context.Context, req *RecoveryRequest) *xerrors.AppError {
	s.logger.InfoContext(ctx, "开始账户恢复流程",
		log.String("email", req.Email),
	)

	// 1. 创建恢复流程
	flow, resp, err := s.publicClient.FrontendAPI.CreateNativeRecoveryFlow(ctx).Execute()
	if err != nil {
		return s.handleKratosError(ctx, "create_recovery_flow", err, resp)
	}

	// 2. 提交恢复请求
	submitReq := client.UpdateRecoveryFlowBody{
		UpdateRecoveryFlowWithLinkMethod: &client.UpdateRecoveryFlowWithLinkMethod{
			Method: "link",
			Email:  req.Email,
		},
	}

	_, resp, err = s.publicClient.FrontendAPI.UpdateRecoveryFlow(ctx).
		Flow(flow.Id).
		UpdateRecoveryFlowBody(submitReq).
		Execute()

	if err != nil {
		return s.handleKratosError(ctx, "submit_recovery", err, resp)
	}

	s.logger.InfoContext(ctx, "恢复邮件发送成功",
		log.String("email", req.Email),
	)

	return nil
}

// handleKratosError 处理 Kratos 错误
func (s *AuthService) handleKratosError(ctx context.Context, operation string, err error, resp *http.Response) *xerrors.AppError {
	s.logger.ErrorContext(ctx, "Kratos API 调用失败",
		log.String("operation", operation),
		log.Any("error", err),
	)

	// 尝试解析 Kratos 错误响应
	if resp != nil {
		s.logger.DebugContext(ctx, "Kratos 响应详情",
			log.Int("status_code", resp.StatusCode),
			log.String("status", resp.Status),
		)

		// 根据状态码转换错误
		switch resp.StatusCode {
		case http.StatusBadRequest:
			// 尝试解析具体的验证错误
			if kratosErr := parseKratosValidationError(resp); kratosErr != nil {
				return kratosErr
			}
			return xerrors.FromCode(xerrors.CodeInvalidParams).
				WithService("auth-service", operation).
				WithMetadata("kratos_status", "400")

		case http.StatusUnauthorized:
			return xerrors.FromCode(xerrors.CodeAuthenticationFailed).
				WithService("auth-service", operation).
				WithMetadata("kratos_status", "401")

		case http.StatusForbidden:
			return xerrors.FromCode(xerrors.CodePermissionDenied).
				WithService("auth-service", operation).
				WithMetadata("kratos_status", "403")

		case http.StatusNotFound:
			return xerrors.FromCode(xerrors.CodeResourceNotFound).
				WithService("auth-service", operation).
				WithMetadata("kratos_status", "404")

		case http.StatusConflict:
			return xerrors.FromCode(xerrors.CodeDuplicateResource).
				WithService("auth-service", operation).
				WithMetadata("kratos_status", "409")

		case http.StatusTooManyRequests:
			return xerrors.FromCode(xerrors.CodeRateLimitExceeded).
				WithService("auth-service", operation).
				WithMetadata("kratos_status", "429")

		default:
			if resp.StatusCode >= 500 {
				return xerrors.NewExternalServiceError("kratos", err).
					WithService("auth-service", operation).
					WithMetadata("kratos_status", string(rune(resp.StatusCode)))
			}
		}
	}

	// 默认系统错误
	return xerrors.NewExternalServiceError("kratos", err).
		WithService("auth-service", operation)
}

// parseKratosValidationError 解析 Kratos 验证错误
func parseKratosValidationError(resp *http.Response) *xerrors.AppError {
	// 这里可以解析 Kratos 的具体错误信息
	// 由于使用了官方客户端，错误信息会更结构化
	// 具体实现取决于 Kratos 返回的错误格式
	return nil
}

// extractSessionCookie 从响应中提取会话 Cookie
func extractSessionCookie(resp *http.Response) string {
	if resp == nil {
		return ""
	}

	cookies := resp.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "ory_kratos_session" {
			return cookie.String()
		}
	}

	return ""
}

// ValidateLoginRequest 验证登录请求
func (s *AuthService) ValidateLoginRequest(req *LoginRequest) *xerrors.AppError {
	if req.Identifier == "" {
		return xerrors.NewValidationError("identifier", "用户名或邮箱不能为空")
	}

	if req.Password == "" {
		return xerrors.NewValidationError("password", "密码不能为空")
	}

	if len(req.Password) < 6 {
		return xerrors.FromCode(xerrors.CodePasswordTooShort)
	}

	if len(req.Password) > 128 {
		return xerrors.FromCode(xerrors.CodePasswordTooLong)
	}

	return nil
}

// ValidateRegisterRequest 验证注册请求
func (s *AuthService) ValidateRegisterRequest(req *RegisterRequest) *xerrors.AppError {
	if req.Email == "" {
		return xerrors.NewValidationError("email", "邮箱不能为空")
	}

	if req.Username == "" {
		return xerrors.NewValidationError("username", "用户名不能为空")
	}

	if len(req.Username) < 3 || len(req.Username) > 30 {
		return xerrors.NewValidationError("username", "用户名长度必须在3-30个字符之间")
	}

	if req.Password == "" {
		return xerrors.NewValidationError("password", "密码不能为空")
	}

	if len(req.Password) < 8 {
		return xerrors.FromCode(xerrors.CodePasswordTooShort)
	}

	if len(req.Password) > 128 {
		return xerrors.FromCode(xerrors.CodePasswordTooLong)
	}

	return nil
}
