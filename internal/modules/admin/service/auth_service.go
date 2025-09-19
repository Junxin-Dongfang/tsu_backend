// File: internal/app/admin/service/auth_service.go
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	client "github.com/ory/client-go"

	"tsu-self/internal/model/authmodel"
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

// ValidateLoginRequest 验证登录请求
func (s *AuthService) ValidateLoginRequest(req *authmodel.LoginRequest) *xerrors.AppError {
	if req.Identifier == "" {
		return xerrors.NewValidationError("identifier", "用户标识不能为空")
	}
	if req.Password == "" {
		return xerrors.NewValidationError("password", "密码不能为空")
	}
	if len(req.Password) < 8 {
		return xerrors.NewValidationError("password", "密码长度不能少于8位")
	}
	return nil
}

// Login 执行登录
func (s *AuthService) Login(ctx context.Context, req *authmodel.LoginRequest) (*authmodel.LoginResult, *xerrors.AppError) {
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

		return &authmodel.LoginResult{
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
func (s *AuthService) Register(ctx context.Context, req *authmodel.RegisterRequest) (*authmodel.RegisterResult, *xerrors.AppError) {
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

		return &authmodel.RegisterResult{
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
func (s *AuthService) InitRecovery(ctx context.Context, req *authmodel.RecoveryRequest) *xerrors.AppError {
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

		// 尝试读取响应体内容
		if resp.Body != nil {
			body, readErr := io.ReadAll(resp.Body)
			if readErr == nil {
				s.logger.DebugContext(ctx, "Kratos 响应体内容",
					log.String("body", string(body)),
					log.Int("body_length", len(body)))
				// 重置响应体
				resp.Body = io.NopCloser(bytes.NewBuffer(body))

				// 尝试解析响应体中的错误信息
				if len(body) > 0 {
					// 优先尝试解析为详细的错误格式
					if kratosErr := s.parseDetailedKratosError(ctx, body, operation); kratosErr != nil {
						return kratosErr
					}
				}
			} else {
				s.logger.ErrorContext(ctx, "读取响应体失败", log.Any("error", readErr))
			}
		}

		// 尝试从 Ory 客户端错误中提取详细信息
		if kratosErr := s.extractErrorFromOryClient(ctx, err, operation); kratosErr != nil {
			return kratosErr
		}

		// 根据状态码转换错误
		switch resp.StatusCode {
		case http.StatusBadRequest:
			// 尝试解析具体的验证错误
			if kratosErr := parseKratosValidationError(resp); kratosErr != nil {
				s.logger.InfoContext(ctx, "解析到 Kratos 验证错误",
					log.Int("app_code", kratosErr.Code),
					log.String("app_message", kratosErr.Message),
				)
				return kratosErr
			}
			s.logger.WarnContext(ctx, "无法解析 Kratos 验证错误，使用默认错误")
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

// parseDetailedKratosError 解析详细的 Kratos 错误信息
func (s *AuthService) parseDetailedKratosError(ctx context.Context, body []byte, operation string) *xerrors.AppError {
	if len(body) == 0 {
		return nil
	}

	// 尝试多种解析方式
	parsers := []func([]byte) *xerrors.AppError{
		parseKratosUIError,
		parseRegistrationFlowError,
		parseKratosErrorFromBody,
	}

	for _, parser := range parsers {
		if kratosErr := parser(body); kratosErr != nil {
			s.logger.InfoContext(ctx, "成功解析 Kratos 错误",
				log.Int("app_code", kratosErr.Code),
				log.String("app_message", kratosErr.Message),
				log.String("operation", operation),
			)
			return kratosErr.WithService("auth-service", operation)
		}
	}

	// 如果所有解析器都失败，尝试从原始文本中提取错误信息
	if kratosErr := s.parseErrorFromRawText(ctx, string(body)); kratosErr != nil {
		return kratosErr.WithService("auth-service", operation)
	}

	return nil
}

// extractErrorFromOryClient 从 Ory 客户端错误中提取详细信息
func (s *AuthService) extractErrorFromOryClient(ctx context.Context, err error, operation string) *xerrors.AppError {
	s.logger.DebugContext(ctx, "分析 Ory 客户端错误",
		log.String("error_type", fmt.Sprintf("%T", err)),
		log.String("error_string", err.Error()),
	)

	// 方法1：检查是否有 Body() 方法
	if genericErr, ok := err.(interface{ Body() []byte }); ok {
		body := genericErr.Body()
		if len(body) > 0 {
			s.logger.DebugContext(ctx, "从 Body() 方法获取错误体", log.String("body", string(body)))
			if kratosErr := s.parseDetailedKratosError(ctx, body, operation); kratosErr != nil {
				return kratosErr
			}
		}
	}

	// 方法2：检查是否有 ResponseBody() 方法
	if genericErr, ok := err.(interface{ ResponseBody() []byte }); ok {
		body := genericErr.ResponseBody()
		if len(body) > 0 {
			s.logger.DebugContext(ctx, "从 ResponseBody() 方法获取错误体", log.String("body", string(body)))
			if kratosErr := s.parseDetailedKratosError(ctx, body, operation); kratosErr != nil {
				return kratosErr
			}
		}
	}

	// 方法3：尝试使用反射获取错误体
	if kratosErr := extractErrorFromReflection(err); kratosErr != nil {
		s.logger.InfoContext(ctx, "从反射解析到 Kratos 验证错误",
			log.Int("app_code", kratosErr.Code),
			log.String("app_message", kratosErr.Message),
		)
		return kratosErr.WithService("auth-service", operation)
	}

	// 方法4：从错误字符串中提取信息
	if kratosErr := s.parseErrorFromRawText(ctx, err.Error()); kratosErr != nil {
		return kratosErr.WithService("auth-service", operation)
	}

	return nil
}

// parseErrorFromRawText 从原始文本中解析错误信息
func (s *AuthService) parseErrorFromRawText(ctx context.Context, errorText string) *xerrors.AppError {
	if errorText == "" {
		return nil
	}

	s.logger.DebugContext(ctx, "尝试从原始文本解析错误", log.String("text", errorText))

	// 使用改进的文本匹配功能
	appCode, appMessage := xerrors.TranslateKratosErrorText(errorText)

	// 如果找到了具体的错误码，创建 AppError
	if appCode != xerrors.CodeInvalidParams || appMessage != "输入信息有误，请检查后重试" {
		appErr := xerrors.FromCode(appCode)
		if appMessage != "" {
			appErr.Message = appMessage
		}

		// 添加原始错误文本作为元数据
		appErr = appErr.WithMetadata("kratos_raw_error", errorText)

		s.logger.InfoContext(ctx, "从原始文本解析到具体错误",
			log.Int("app_code", appCode),
			log.String("app_message", appMessage),
		)

		return appErr
	}

	return nil
}

// parseKratosValidationError 解析 Kratos 验证错误
func parseKratosValidationError(resp *http.Response) *xerrors.AppError {
	if resp == nil || resp.Body == nil {
		return nil
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return xerrors.NewExternalServiceError("kratos", err).
			WithMetadata("parse_error", "failed to read response body")
	}

	// 重置响应体，以便后续可能的读取
	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	// 添加调试日志
	fmt.Printf("DEBUG: Kratos 响应体内容: %s\n", string(body))

	// 尝试解析为标准的 Kratos 错误响应
	var kratosError struct {
		Error struct {
			ID      int    `json:"id"`
			Code    int    `json:"code"`
			Status  string `json:"status"`
			Reason  string `json:"reason"`
			Message string `json:"message"`
			Details []struct {
				Property string `json:"property"`
				Messages []struct {
					ID      int                    `json:"id"`
					Text    string                 `json:"text"`
					Type    string                 `json:"type"`
					Context map[string]interface{} `json:"context"`
				} `json:"messages"`
			} `json:"details"`
		} `json:"error"`
		UIErrorMessages []struct {
			ID      int                    `json:"id"`
			Text    string                 `json:"text"`
			Type    string                 `json:"type"`
			Context map[string]interface{} `json:"context"`
		} `json:"ui_error_messages"`
	}

	if err := json.Unmarshal(body, &kratosError); err != nil {
		// 如果无法解析为标准格式，尝试文本匹配
		return parseKratosErrorFromText(string(body))
	}

	// 收集所有错误信息
	var allErrors []kratosErrorInfo

	// 处理主错误
	if kratosError.Error.ID != 0 {
		allErrors = append(allErrors, kratosErrorInfo{
			ID:       kratosError.Error.ID,
			Text:     kratosError.Error.Message,
			Property: "",
			Context:  nil,
		})
	}

	// 处理字段级详细错误
	for _, detail := range kratosError.Error.Details {
		for _, msg := range detail.Messages {
			allErrors = append(allErrors, kratosErrorInfo{
				ID:       msg.ID,
				Text:     msg.Text,
				Property: detail.Property,
				Context:  msg.Context,
			})
		}
	}

	// 处理 UI 错误消息
	for _, uiMsg := range kratosError.UIErrorMessages {
		allErrors = append(allErrors, kratosErrorInfo{
			ID:       uiMsg.ID,
			Text:     uiMsg.Text,
			Property: "",
			Context:  uiMsg.Context,
		})
	}

	// 如果没有找到任何错误，返回 nil
	if len(allErrors) == 0 {
		return nil
	}

	// 根据优先级选择最重要的错误
	selectedError := selectHighestPriorityError(allErrors)

	// 转换为业务错误码，考虑字段特定的错误
	appCode, appMessage := translateKratosErrorWithField(selectedError)

	// 创建 AppError
	appErr := xerrors.FromCode(appCode)
	if appMessage != "" {
		appErr.Message = appMessage
	}

	// 添加字段信息
	if selectedError.Property != "" {
		appErr = appErr.WithMetadata("field", selectedError.Property)
	}

	// 添加原始错误信息
	if selectedError.Text != "" && selectedError.Text != appMessage {
		appErr = appErr.WithMetadata("kratos_message", selectedError.Text)
	}

	// 添加 Kratos 错误 ID
	appErr = appErr.WithMetadata("kratos_error_id", fmt.Sprintf("%d", selectedError.ID))

	// 添加上下文信息
	if selectedError.Context != nil {
		for key, value := range selectedError.Context {
			if strValue, ok := value.(string); ok {
				appErr = appErr.WithMetadata(fmt.Sprintf("context_%s", key), strValue)
			}
		}
	}

	return appErr
}

// kratosErrorInfo 表示一个 Kratos 错误信息
type kratosErrorInfo struct {
	ID       int                    `json:"id"`
	Text     string                 `json:"text"`
	Property string                 `json:"property"`
	Context  map[string]interface{} `json:"context"`
}

// selectHighestPriorityError 根据优先级选择最重要的错误
func selectHighestPriorityError(errors []kratosErrorInfo) kratosErrorInfo {
	if len(errors) == 0 {
		return kratosErrorInfo{}
	}

	// 选择第一个错误作为默认值
	selectedError := errors[0]
	highestPriority := xerrors.GetKratosErrorPriority(getAppCodeFromKratosID(selectedError.ID))

	// 遍历所有错误，找到优先级最高的
	for _, err := range errors[1:] {
		appCode := getAppCodeFromKratosID(err.ID)
		priority := xerrors.GetKratosErrorPriority(appCode)

		if priority < highestPriority {
			selectedError = err
			highestPriority = priority
		}
	}

	return selectedError
}

// getAppCodeFromKratosID 从 Kratos 错误 ID 获取应用错误码
func getAppCodeFromKratosID(kratosID int) int {
	appCode, _ := xerrors.TranslateKratosError(kratosID)
	return appCode
}

// parseKratosErrorFromText 从文本解析 Kratos 错误（兜底方案）
func parseKratosErrorFromText(errorText string) *xerrors.AppError {
	if errorText == "" {
		return nil
	}

	// 使用现有的文本匹配功能
	appCode, appMessage := xerrors.TranslateKratosErrorText(errorText)

	appErr := xerrors.FromCode(appCode)
	if appMessage != "" {
		appErr.Message = appMessage
	}

	// 添加原始错误文本
	appErr = appErr.WithMetadata("kratos_raw_error", errorText)

	return appErr
}

// parseKratosErrorFromBody 从字节数组解析 Kratos 错误
func parseKratosErrorFromBody(body []byte) *xerrors.AppError {
	if len(body) == 0 {
		return nil
	}

	// 尝试解析为标准的 Kratos 错误响应
	var kratosError struct {
		Error struct {
			ID      int    `json:"id"`
			Code    int    `json:"code"`
			Status  string `json:"status"`
			Reason  string `json:"reason"`
			Message string `json:"message"`
			Details []struct {
				Property string `json:"property"`
				Messages []struct {
					ID      int                    `json:"id"`
					Text    string                 `json:"text"`
					Type    string                 `json:"type"`
					Context map[string]interface{} `json:"context"`
				} `json:"messages"`
			} `json:"details"`
		} `json:"error"`
		UIErrorMessages []struct {
			ID      int                    `json:"id"`
			Text    string                 `json:"text"`
			Type    string                 `json:"type"`
			Context map[string]interface{} `json:"context"`
		} `json:"ui_error_messages"`
	}

	if err := json.Unmarshal(body, &kratosError); err != nil {
		// 如果无法解析为标准格式，尝试文本匹配
		return parseKratosErrorFromText(string(body))
	}

	// 收集所有错误信息
	var allErrors []kratosErrorInfo

	// 处理主错误
	if kratosError.Error.ID != 0 {
		allErrors = append(allErrors, kratosErrorInfo{
			ID:       kratosError.Error.ID,
			Text:     kratosError.Error.Message,
			Property: "",
			Context:  nil,
		})
	}

	// 处理字段级详细错误
	for _, detail := range kratosError.Error.Details {
		for _, msg := range detail.Messages {
			allErrors = append(allErrors, kratosErrorInfo{
				ID:       msg.ID,
				Text:     msg.Text,
				Property: detail.Property,
				Context:  msg.Context,
			})
		}
	}

	// 处理 UI 错误消息
	for _, uiMsg := range kratosError.UIErrorMessages {
		allErrors = append(allErrors, kratosErrorInfo{
			ID:       uiMsg.ID,
			Text:     uiMsg.Text,
			Property: "",
			Context:  uiMsg.Context,
		})
	}

	// 如果没有找到任何错误，返回 nil
	if len(allErrors) == 0 {
		return nil
	}

	// 根据优先级选择最重要的错误
	selectedError := selectHighestPriorityError(allErrors)

	// 转换为业务错误码，考虑字段特定的错误
	appCode, appMessage := translateKratosErrorWithField(selectedError)

	// 创建 AppError
	appErr := xerrors.FromCode(appCode)
	if appMessage != "" {
		appErr.Message = appMessage
	}

	// 添加字段信息
	if selectedError.Property != "" {
		appErr = appErr.WithMetadata("field", selectedError.Property)
	}

	// 添加原始错误信息
	if selectedError.Text != "" && selectedError.Text != appMessage {
		appErr = appErr.WithMetadata("kratos_message", selectedError.Text)
	}

	// 添加 Kratos 错误 ID
	appErr = appErr.WithMetadata("kratos_error_id", fmt.Sprintf("%d", selectedError.ID))

	// 添加上下文信息
	if selectedError.Context != nil {
		for key, value := range selectedError.Context {
			if strValue, ok := value.(string); ok {
				appErr = appErr.WithMetadata(fmt.Sprintf("context_%s", key), strValue)
			}
		}
	}

	return appErr
}

// parseKratosUIError 解析 Kratos UI 格式的错误（用于注册/登录流程）
func parseKratosUIError(body []byte) *xerrors.AppError {
	if len(body) == 0 {
		return nil
	}

	// 解析 Kratos UI 响应格式
	var uiResponse struct {
		UI struct {
			Messages []struct {
				ID      int                    `json:"id"`
				Text    string                 `json:"text"`
				Type    string                 `json:"type"`
				Context map[string]interface{} `json:"context"`
			} `json:"messages"`
		} `json:"ui"`
	}

	if err := json.Unmarshal(body, &uiResponse); err != nil {
		return nil
	}

	// 收集错误消息
	var allErrors []kratosErrorInfo
	for _, uiMsg := range uiResponse.UI.Messages {
		if uiMsg.Type == "error" { // 只处理错误消息
			allErrors = append(allErrors, kratosErrorInfo{
				ID:       uiMsg.ID,
				Text:     uiMsg.Text,
				Property: "",
				Context:  uiMsg.Context,
			})
		}
	}

	// 如果没有找到任何错误，返回 nil
	if len(allErrors) == 0 {
		return nil
	}

	// 根据优先级选择最重要的错误
	selectedError := selectHighestPriorityError(allErrors)

	// 转换为业务错误码，考虑字段特定的错误
	appCode, appMessage := translateKratosErrorWithField(selectedError)

	// 创建 AppError
	appErr := xerrors.FromCode(appCode)
	if appMessage != "" {
		appErr.Message = appMessage
	}

	// 添加原始错误信息
	if selectedError.Text != "" && selectedError.Text != appMessage {
		appErr = appErr.WithMetadata("kratos_message", selectedError.Text)
	}

	// 添加 Kratos 错误 ID
	appErr = appErr.WithMetadata("kratos_error_id", fmt.Sprintf("%d", selectedError.ID))

	// 添加上下文信息
	if selectedError.Context != nil {
		for key, value := range selectedError.Context {
			if strValue, ok := value.(string); ok {
				appErr = appErr.WithMetadata(fmt.Sprintf("context_%s", key), strValue)
			}
		}
	}

	return appErr
}

// extractErrorFromReflection 使用反射从错误中提取详细信息
func extractErrorFromReflection(err error) *xerrors.AppError {
	if err == nil {
		return nil
	}

	// 使用反射检查错误结构
	v := reflect.ValueOf(err)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// 尝试找到包含响应体的字段
	var body []byte

	// 常见的字段名
	fieldNames := []string{"body", "Body", "responseBody", "ResponseBody", "content", "Content"}

	for _, fieldName := range fieldNames {
		if v.Kind() == reflect.Struct {
			field := v.FieldByName(fieldName)
			if field.IsValid() && field.CanInterface() {
				if bodyBytes, ok := field.Interface().([]byte); ok && len(bodyBytes) > 0 {
					body = bodyBytes
					break
				}
				if bodyStr, ok := field.Interface().(string); ok && bodyStr != "" {
					body = []byte(bodyStr)
					break
				}
			}
		}
	}

	if len(body) > 0 {
		// 尝试解析错误体
		if kratosErr := parseKratosUIError(body); kratosErr != nil {
			return kratosErr
		}
	}

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

// ValidateRegisterRequest 验证注册请求
func (s *AuthService) ValidateRegisterRequest(req *authmodel.RegisterRequest) *xerrors.AppError {
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

// parseRegistrationFlowError 解析Kratos Registration Flow错误响应
func parseRegistrationFlowError(body []byte) *xerrors.AppError {
	if len(body) == 0 {
		return nil
	}

	// 尝试解析为Registration Flow的错误格式
	var flowResponse struct {
		UI struct {
			Messages []struct {
				ID      int                    `json:"id"`
				Text    string                 `json:"text"`
				Type    string                 `json:"type"`
				Context map[string]interface{} `json:"context"`
			} `json:"messages"`
			Nodes []struct {
				Type       string `json:"type"`
				Group      string `json:"group"`
				Attributes struct {
					Name  string `json:"name"`
					Type  string `json:"type"`
					Value string `json:"value"`
				} `json:"attributes"`
				Messages []struct {
					ID      int                    `json:"id"`
					Text    string                 `json:"text"`
					Type    string                 `json:"type"`
					Context map[string]interface{} `json:"context"`
				} `json:"messages"`
			} `json:"nodes"`
		} `json:"ui"`
	}

	if err := json.Unmarshal(body, &flowResponse); err != nil {
		return nil
	}

	// 收集所有错误消息
	var allErrors []kratosErrorInfo

	// 检查全局UI消息
	for _, msg := range flowResponse.UI.Messages {
		allErrors = append(allErrors, kratosErrorInfo{
			ID:       msg.ID,
			Text:     msg.Text,
			Property: "",
			Context:  msg.Context,
		})
	}

	// 检查字段级消息
	for _, node := range flowResponse.UI.Nodes {
		for _, msg := range node.Messages {
			property := node.Attributes.Name
			if property == "" {
				property = node.Group
			}
			allErrors = append(allErrors, kratosErrorInfo{
				ID:       msg.ID,
				Text:     msg.Text,
				Property: property,
				Context:  msg.Context,
			})
		}
	}

	if len(allErrors) == 0 {
		return nil
	}

	// 根据优先级选择最重要的错误
	selectedError := selectHighestPriorityError(allErrors)

	// 转换为业务错误码，考虑字段特定的错误
	appCode, appMessage := translateKratosErrorWithField(selectedError)

	// 创建AppError
	appErr := xerrors.FromCode(appCode)
	if appMessage != "" {
		appErr.Message = appMessage
	}

	// 添加字段信息
	if selectedError.Property != "" {
		appErr = appErr.WithMetadata("field", selectedError.Property)
	}

	// 添加原始错误信息
	if selectedError.Text != "" && selectedError.Text != appMessage {
		appErr = appErr.WithMetadata("kratos_message", selectedError.Text)
	}

	// 添加Kratos错误ID
	appErr = appErr.WithMetadata("kratos_error_id", fmt.Sprintf("%d", selectedError.ID))

	return appErr
}

// translateKratosErrorWithField 根据字段信息翻译 Kratos 错误
func translateKratosErrorWithField(errorInfo kratosErrorInfo) (int, string) {
	// 如果有字段信息，优先尝试字段特定的错误处理
	if errorInfo.Property != "" && errorInfo.Text != "" {
		fieldSpecificCode, fieldSpecificMessage := getFieldSpecificError(errorInfo.Property, errorInfo.Text)
		if fieldSpecificCode != 0 {
			return fieldSpecificCode, fieldSpecificMessage
		}
	}

	// 尝试基于 Kratos 错误 ID 进行翻译
	appCode, appMessage := xerrors.TranslateKratosError(errorInfo.ID)

	// 如果 ID 映射失败，尝试文本匹配
	if appCode == xerrors.CodeInvalidParams && errorInfo.Text != "" {
		appCode, appMessage = xerrors.TranslateKratosErrorText(errorInfo.Text)
	}

	return appCode, appMessage
}

// getFieldSpecificError 根据字段和错误文本获取具体的错误码和消息
func getFieldSpecificError(field, errorText string) (int, string) {
	errorTextLower := strings.ToLower(errorText)

	switch field {
	case "traits.email", "email":
		if strings.Contains(errorTextLower, "already taken") || strings.Contains(errorTextLower, "already exists") {
			return xerrors.CodeEmailExists, "该邮箱已被注册"
		}
		if strings.Contains(errorTextLower, "invalid") || strings.Contains(errorTextLower, "format") {
			return xerrors.CodeInvalidParams, "邮箱格式不正确"
		}
		if strings.Contains(errorTextLower, "required") || strings.Contains(errorTextLower, "missing") {
			return xerrors.CodeInvalidParams, "请输入邮箱地址"
		}

	case "traits.username", "username":
		if strings.Contains(errorTextLower, "already taken") || strings.Contains(errorTextLower, "already exists") {
			return xerrors.CodeUsernameExists, "该用户名已被使用"
		}
		if strings.Contains(errorTextLower, "invalid") || strings.Contains(errorTextLower, "format") {
			return xerrors.CodeInvalidParams, "用户名格式不正确"
		}
		if strings.Contains(errorTextLower, "required") || strings.Contains(errorTextLower, "missing") {
			return xerrors.CodeInvalidParams, "请输入用户名"
		}
		if strings.Contains(errorTextLower, "too short") || strings.Contains(errorTextLower, "minimum") {
			return xerrors.CodeInvalidParams, "用户名长度不能少于3个字符"
		}
		if strings.Contains(errorTextLower, "too long") || strings.Contains(errorTextLower, "maximum") {
			return xerrors.CodeInvalidParams, "用户名长度不能超过30个字符"
		}

	case "password":
		if strings.Contains(errorTextLower, "too short") || strings.Contains(errorTextLower, "minimum") {
			return xerrors.CodePasswordTooShort, "密码长度不能少于8个字符"
		}
		if strings.Contains(errorTextLower, "too long") || strings.Contains(errorTextLower, "maximum") {
			return xerrors.CodePasswordTooLong, "密码长度不能超过128个字符"
		}
		if strings.Contains(errorTextLower, "policy") || strings.Contains(errorTextLower, "strength") || strings.Contains(errorTextLower, "breaches") {
			return xerrors.CodePasswordPolicyError, "密码强度不够，请使用更复杂的密码"
		}
		if strings.Contains(errorTextLower, "similar") {
			return xerrors.CodePasswordTooSimilar, "密码不能与用户信息太相似"
		}
		if strings.Contains(errorTextLower, "required") || strings.Contains(errorTextLower, "missing") {
			return xerrors.CodeInvalidParams, "请输入密码"
		}

	case "identifier":
		if strings.Contains(errorTextLower, "invalid") || strings.Contains(errorTextLower, "credentials are invalid") {
			return xerrors.CodeInvalidCredentials, "用户名或密码错误"
		}
		if strings.Contains(errorTextLower, "not found") {
			return xerrors.CodeAccountNotFound, "用户名或邮箱不存在"
		}
		if strings.Contains(errorTextLower, "required") || strings.Contains(errorTextLower, "missing") {
			return xerrors.CodeInvalidParams, "请输入用户名或邮箱"
		}
	}

	// 如果没有找到字段特定的错误，返回 0 表示未找到
	return 0, ""
}
