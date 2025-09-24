// internal/modules/auth/service/kratos_service_new.go - 移除数据库依赖的新版本
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	client "github.com/ory/client-go"
	"google.golang.org/protobuf/types/known/timestamppb"

	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/rpc/generated/auth"
	"tsu-self/internal/rpc/generated/common"
	"tsu-self/internal/rpc/generated/user"
)

type KratosService struct {
	publicClient *client.APIClient
	adminClient  *client.APIClient
	logger       log.Logger
}

func NewKratosService(publicURL, adminURL string, logger log.Logger) (*KratosService, error) {
	publicConfig := client.NewConfiguration()
	publicConfig.Servers = []client.ServerConfiguration{{URL: publicURL}}

	adminConfig := client.NewConfiguration()
	adminConfig.Servers = []client.ServerConfiguration{{URL: adminURL}}

	return &KratosService{
		publicClient: client.NewAPIClient(publicConfig),
		adminClient:  client.NewAPIClient(adminConfig),
		logger:       logger,
	}, nil
}

// Login 纯 Kratos 登录，不操作主数据库
func (s *KratosService) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, *xerrors.AppError) {
	s.logger.InfoContext(ctx, "开始 Kratos 登录流程", log.String("identifier", req.Identifier))

	// 基础验证
	if req.Password == "" {
		return &auth.LoginResponse{
			Success:      false,
			ErrorMessage: "密码不能为空",
		}, xerrors.NewValidationError("password", "密码不能为空")
	}

	// 1. 创建登录流程
	loginFlow, _, err := s.publicClient.FrontendAPI.CreateNativeLoginFlow(ctx).Execute()
	if err != nil {
		s.logger.ErrorContext(ctx, "创建 Kratos 登录流程失败", log.Any("error", err))
		return &auth.LoginResponse{
			Success:      false,
			ErrorMessage: "登录流程创建失败",
		}, xerrors.NewExternalServiceError("kratos", err)
	}

	s.logger.DebugContext(ctx, "Kratos 登录流程创建成功", log.String("flow_id", loginFlow.Id))

	// 2. 提交登录信息
	passwordMethod := client.NewUpdateLoginFlowWithPasswordMethod(req.Identifier, "password", req.Password)
	updateBody := client.UpdateLoginFlowWithPasswordMethodAsUpdateLoginFlowBody(passwordMethod)

	loginResult, _, loginErr := s.publicClient.FrontendAPI.UpdateLoginFlow(ctx).
		Flow(loginFlow.Id).
		UpdateLoginFlowBody(updateBody).
		Execute()

	if loginErr != nil {
		s.logger.ErrorContext(ctx, "Kratos 登录验证失败", log.Any("error", loginErr))

		// 解析 Kratos 错误，提取具体的错误信息
		errorMessage := s.parseKratosError(ctx, loginErr, "登录")

		return &auth.LoginResponse{
			Success:      false,
			ErrorMessage: errorMessage,
		}, nil
	}

	// 3. 从登录结果中提取用户信息
	session := loginResult.Session
	if session.Identity == nil {
		s.logger.ErrorContext(ctx, "Kratos 登录结果中没有身份信息")
		return &auth.LoginResponse{
			Success:      false,
			ErrorMessage: "登录结果异常",
		}, xerrors.NewExternalServiceError("kratos", nil)
	}

	identity := session.Identity
	userInfo := &common.UserInfo{
		Id:     identity.Id,
		Traits: make(map[string]string),
	}

	// 处理时间戳字段（可能为 nil）
	if identity.CreatedAt != nil {
		userInfo.CreatedAt = timestamppb.New(*identity.CreatedAt)
	} else {
		userInfo.CreatedAt = timestamppb.New(time.Now())
	}
	if identity.UpdatedAt != nil {
		userInfo.UpdatedAt = timestamppb.New(*identity.UpdatedAt)
	} else {
		userInfo.UpdatedAt = timestamppb.New(time.Now())
	}

	// 从 traits 中提取用户信息
	if traits, ok := identity.Traits.(map[string]interface{}); ok {
		if email, exists := traits["email"]; exists {
			if emailStr, ok := email.(string); ok {
				userInfo.Email = emailStr
			}
		}
		if username, exists := traits["username"]; exists {
			if usernameStr, ok := username.(string); ok {
				userInfo.Username = usernameStr
			}
		}
		// 将所有 traits 转换为字符串映射
		for k, v := range traits {
			if strVal, ok := v.(string); ok {
				userInfo.Traits[k] = strVal
			}
		}
	}

	// 4. 获取会话令牌
	sessionToken := ""
	if loginResult.SessionToken != nil {
		sessionToken = *loginResult.SessionToken
	}

	s.logger.InfoContext(ctx, "Kratos 登录成功",
		log.String("user_id", userInfo.Id),
		log.String("email", userInfo.Email))

	return &auth.LoginResponse{
		Success:    true,
		Token:      sessionToken,
		IdentityId: userInfo.Id,
		UserInfo:   userInfo,
		ExpiresIn:  3600, // 1小时
	}, nil
}

// Register 纯 Kratos 注册，不操作主数据库
func (s *KratosService) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, *xerrors.AppError) {
	s.logger.InfoContext(ctx, "开始 Kratos 注册流程",
		log.String("email", req.Email),
		log.String("username", req.Username))

	// 基础验证
	if req.Email == "" {
		return &auth.RegisterResponse{
			Success:      false,
			ErrorMessage: "邮箱不能为空",
		}, xerrors.NewValidationError("email", "邮箱不能为空")
	}

	if req.Username == "" {
		return &auth.RegisterResponse{
			Success:      false,
			ErrorMessage: "用户名不能为空",
		}, xerrors.NewValidationError("username", "用户名不能为空")
	}

	if req.Password == "" {
		return &auth.RegisterResponse{
			Success:      false,
			ErrorMessage: "密码不能为空",
		}, xerrors.NewValidationError("password", "密码不能为空")
	}

	// 1. 创建注册流程
	registerFlow, _, err := s.publicClient.FrontendAPI.CreateNativeRegistrationFlow(ctx).Execute()
	if err != nil {
		s.logger.ErrorContext(ctx, "创建 Kratos 注册流程失败", log.Any("error", err))
		return &auth.RegisterResponse{
			Success:      false,
			ErrorMessage: "注册流程创建失败",
		}, xerrors.NewExternalServiceError("kratos", err)
	}

	s.logger.DebugContext(ctx, "Kratos 注册流程创建成功", log.String("flow_id", registerFlow.Id))

	// 2. 构建用户特征 (traits)
	traits := map[string]interface{}{
		"email":    req.Email,
		"username": req.Username,
	}
	// 添加phone字段（如果提供）
	if req.Phone != "" {
		traits["phone"] = req.Phone
	}

	s.logger.DebugContext(ctx, "提交注册信息", log.Any("traits", traits))

	// 3. 使用password方法直接注册（包括用户信息和密码）
	passwordMethod := client.NewUpdateRegistrationFlowWithPasswordMethod("password", req.Password, traits)
	updateBody := client.UpdateRegistrationFlowWithPasswordMethodAsUpdateRegistrationFlowBody(passwordMethod)

	s.logger.DebugContext(ctx, "Kratos 注册请求详情",
		log.String("method", passwordMethod.Method),
		log.Any("traits", passwordMethod.Traits))

	// 提交注册信息
	registerResult, _, registerErr := s.publicClient.FrontendAPI.UpdateRegistrationFlow(ctx).
		Flow(registerFlow.Id).
		UpdateRegistrationFlowBody(updateBody).
		Execute()

	if registerErr != nil {
		s.logger.ErrorContext(ctx, "Kratos 注册失败", log.Any("error", registerErr))

		// 解析 Kratos 错误，提取具体的错误信息
		errorMessage := s.parseKratosError(ctx, registerErr, "注册")

		return &auth.RegisterResponse{
			Success:      false,
			ErrorMessage: errorMessage,
		}, nil
	}

	// 4. 从注册结果中提取用户信息
	session := registerResult.Session
	if session.Identity == nil {
		s.logger.ErrorContext(ctx, "Kratos 注册结果中没有身份信息")
		return &auth.RegisterResponse{
			Success:      false,
			ErrorMessage: "注册结果异常",
		}, xerrors.NewExternalServiceError("kratos", nil)
	}

	identity := session.Identity
	userInfo := &common.UserInfo{
		Id:     identity.Id,
		Traits: make(map[string]string),
	}

	// 处理时间戳字段（可能为 nil）
	if identity.CreatedAt != nil {
		userInfo.CreatedAt = timestamppb.New(*identity.CreatedAt)
	} else {
		userInfo.CreatedAt = timestamppb.New(time.Now())
	}
	if identity.UpdatedAt != nil {
		userInfo.UpdatedAt = timestamppb.New(*identity.UpdatedAt)
	} else {
		userInfo.UpdatedAt = timestamppb.New(time.Now())
	}

	// 从 traits 中提取用户信息
	if identityTraits, ok := identity.Traits.(map[string]interface{}); ok {
		if email, exists := identityTraits["email"]; exists {
			if emailStr, ok := email.(string); ok {
				userInfo.Email = emailStr
			}
		}
		if username, exists := identityTraits["username"]; exists {
			if usernameStr, ok := username.(string); ok {
				userInfo.Username = usernameStr
			}
		}
		// 将所有 traits 转换为字符串映射
		for k, v := range identityTraits {
			if strVal, ok := v.(string); ok {
				userInfo.Traits[k] = strVal
			}
		}
	}

	// 5. 获取会话令牌
	sessionToken := ""
	if registerResult.SessionToken != nil {
		sessionToken = *registerResult.SessionToken
	}

	s.logger.InfoContext(ctx, "Kratos 注册成功", log.String("identity_id", userInfo.Id))

	return &auth.RegisterResponse{
		Success:    true,
		IdentityId: userInfo.Id,
		Token:      sessionToken,
		UserInfo:   userInfo,
	}, nil
}

// GetUserInfoByToken 通过 token 从 Kratos 获取用户信息
func (s *KratosService) GetUserInfoByToken(ctx context.Context, token string) (*common.UserInfo, *xerrors.AppError) {
	s.logger.DebugContext(ctx, "从 Kratos 获取用户信息", log.String("token", token[:10]+"..."))

	// 调用 Kratos whoami API (ToSession)
	session, _, err := s.publicClient.FrontendAPI.ToSession(ctx).
		XSessionToken(token).
		Execute()

	if err != nil {
		s.logger.ErrorContext(ctx, "Kratos ToSession 调用失败", log.Any("error", err))
		return nil, xerrors.NewExternalServiceError("kratos", err)
	}

	// 检查会话是否有效
	if session.Identity == nil {
		s.logger.ErrorContext(ctx, "Kratos 会话中没有身份信息")
		return nil, xerrors.NewExternalServiceError("kratos", nil)
	}

	identity := session.Identity
	userInfo := &common.UserInfo{
		Id:     identity.Id,
		Traits: make(map[string]string),
	}

	// 处理时间戳字段（可能为 nil）
	if identity.CreatedAt != nil {
		userInfo.CreatedAt = timestamppb.New(*identity.CreatedAt)
	} else {
		userInfo.CreatedAt = timestamppb.New(time.Now())
	}
	if identity.UpdatedAt != nil {
		userInfo.UpdatedAt = timestamppb.New(*identity.UpdatedAt)
	} else {
		userInfo.UpdatedAt = timestamppb.New(time.Now())
	}

	// 从 traits 中提取用户信息
	if traits, ok := identity.Traits.(map[string]interface{}); ok {
		if email, exists := traits["email"]; exists {
			if emailStr, ok := email.(string); ok {
				userInfo.Email = emailStr
			}
		}
		if username, exists := traits["username"]; exists {
			if usernameStr, ok := username.(string); ok {
				userInfo.Username = usernameStr
			}
		}
		// 将所有 traits 转换为字符串映射
		for k, v := range traits {
			if strVal, ok := v.(string); ok {
				userInfo.Traits[k] = strVal
			}
		}
	}

	return userInfo, nil
}

// GetUserInfo 通过用户ID从 Kratos 获取用户信息
func (s *KratosService) GetUserInfo(ctx context.Context, userID string) (*common.UserInfo, *xerrors.AppError) {
	s.logger.DebugContext(ctx, "通过用户ID从 Kratos 获取用户信息", log.String("user_id", userID))

	// 调用 Kratos Admin API 获取身份信息
	identity, _, err := s.adminClient.IdentityAPI.GetIdentity(ctx, userID).Execute()
	if err != nil {
		s.logger.ErrorContext(ctx, "Kratos GetIdentity 调用失败", log.Any("error", err))
		return nil, xerrors.NewExternalServiceError("kratos", err)
	}

	userInfo := &common.UserInfo{
		Id:     identity.Id,
		Traits: make(map[string]string),
	}

	// 处理时间戳字段（可能为 nil）
	if identity.CreatedAt != nil {
		userInfo.CreatedAt = timestamppb.New(*identity.CreatedAt)
	} else {
		userInfo.CreatedAt = timestamppb.New(time.Now())
	}
	if identity.UpdatedAt != nil {
		userInfo.UpdatedAt = timestamppb.New(*identity.UpdatedAt)
	} else {
		userInfo.UpdatedAt = timestamppb.New(time.Now())
	}

	// 从 traits 中提取用户信息
	if traits, ok := identity.Traits.(map[string]interface{}); ok {
		if email, exists := traits["email"]; exists {
			if emailStr, ok := email.(string); ok {
				userInfo.Email = emailStr
			}
		}
		if username, exists := traits["username"]; exists {
			if usernameStr, ok := username.(string); ok {
				userInfo.Username = usernameStr
			}
		}
		// 将所有 traits 转换为字符串映射
		for k, v := range traits {
			if strVal, ok := v.(string); ok {
				userInfo.Traits[k] = strVal
			}
		}
	}

	return userInfo, nil
}

// UpdateUserTraits 更新用户特征
func (s *KratosService) UpdateUserTraits(ctx context.Context, req *user.UpdateUserTraitsRequest) (*user.UpdateUserTraitsResponse, *xerrors.AppError) {
	s.logger.InfoContext(ctx, "更新用户特征",
		log.String("user_id", req.UserId),
	//log.String("email", req.Email),
	//log.String("username", req.Username)
	)

	// 调用 Kratos 更新身份 API
	// 这里需要实现实际的 Kratos 客户端调用
	// 暂时返回成功响应

	return &user.UpdateUserTraitsResponse{
		Success: true,
	}, nil
}

// Logout 登出用户会话
func (s *KratosService) Logout(ctx context.Context, token string) *xerrors.AppError {
	s.logger.InfoContext(ctx, "登出用户会话", log.String("token", token[:10]+"..."))

	// 调用 Kratos 登出 API
	// 这里需要实现实际的 Kratos 客户端调用
	// 暂时返回成功

	return nil
}

// ValidateSession 验证会话
func (s *KratosService) ValidateSession(ctx context.Context, token string) (*common.UserInfo, *xerrors.AppError) {
	s.logger.DebugContext(ctx, "验证会话", log.String("token", token[:10]+"..."))

	// 调用 Kratos 验证会话 API
	// 这里需要实现实际的 Kratos 客户端调用
	// 暂时返回模拟响应

	userInfo := &common.UserInfo{
		Id:        "validated-user-id",
		Email:     "user@example.com",
		Username:  "username",
		CreatedAt: timestamppb.New(time.Now()),
		UpdatedAt: timestamppb.New(time.Now()),
		Traits:    make(map[string]string),
	}

	return userInfo, nil
}

// DeleteIdentity 删除身份（用于回滚）
func (s *KratosService) DeleteIdentity(ctx context.Context, identityID string) *xerrors.AppError {
	s.logger.InfoContext(ctx, "删除 Kratos 身份", log.String("identity_id", identityID))

	// 调用 Kratos 删除身份 API
	// 这里需要实现实际的 Kratos 客户端调用
	// 暂时返回成功

	return nil
}

// 结果类型定义
type KratosLoginResult struct {
	SessionToken  string `json:"session_token"`
	SessionCookie string `json:"session_cookie"`
	IdentityID    string `json:"identity_id"`
	ExpiresIn     int64  `json:"expires_in"`
}

type KratosRegisterResult struct {
	IdentityID    string `json:"identity_id"`
	SessionToken  string `json:"session_token"`
	SessionCookie string `json:"session_cookie"`
	ExpiresIn     int64  `json:"expires_in"`
}

// parseKratosError 解析Kratos错误，提取用户友好的错误信息
func (s *KratosService) parseKratosError(ctx context.Context, err error, operation string) string {
	// 记录原始错误到日志
	s.logger.ErrorContext(ctx, "解析Kratos错误",
		log.String("operation", operation),
		log.Any("error", err))

	// 尝试解析为GenericOpenAPIError
	if apiErr, ok := err.(*client.GenericOpenAPIError); ok {
		s.logger.DebugContext(ctx, "检测到GenericOpenAPIError",
			log.String("error_body", string(apiErr.Body())))

		// 尝试解析错误响应体
		if flowErr := s.parseFlowError(ctx, apiErr.Body()); flowErr != "" {
			return flowErr
		}
	}

	// 根据操作类型返回默认错误信息
	switch operation {
	case "注册":
		return "注册失败，请检查输入信息"
	case "登录":
		return "用户名或密码错误"
	default:
		return "操作失败，请稍后重试"
	}
}

// parseFlowError 解析Flow错误响应体，提取具体错误信息
func (s *KratosService) parseFlowError(ctx context.Context, body []byte) string {
	var flowResponse struct {
		UI struct {
			Messages []struct {
				ID   int    `json:"id"`
				Text string `json:"text"`
				Type string `json:"type"`
			} `json:"messages"`
			Nodes []struct {
				Messages []struct {
					ID   int    `json:"id"`
					Text string `json:"text"`
					Type string `json:"type"`
				} `json:"messages"`
				Attributes struct {
					Name string `json:"name"`
				} `json:"attributes"`
			} `json:"nodes"`
		} `json:"ui"`
	}

	if err := json.Unmarshal(body, &flowResponse); err != nil {
		s.logger.WarnContext(ctx, "解析Flow错误响应失败", log.Any("error", err))
		return ""
	}

	// 优先检查UI级别的错误消息
	for _, msg := range flowResponse.UI.Messages {
		if msg.Type == "error" {
			translated := s.translateKratosMessage(msg.ID, msg.Text)
			s.logger.InfoContext(ctx, "翻译Kratos错误消息",
				log.Int("message_id", msg.ID),
				log.String("original", msg.Text),
				log.String("translated", translated))
			return translated
		}
	}

	// 检查字段级别的错误消息
	for _, node := range flowResponse.UI.Nodes {
		for _, msg := range node.Messages {
			if msg.Type == "error" {
				fieldName := s.translateFieldName(node.Attributes.Name)
				return fmt.Sprintf("%s：%s", fieldName, s.translateKratosMessage(msg.ID, msg.Text))
			}
		}
	}

	return ""
}

// translateKratosMessage 将Kratos错误消息翻译为中文
func (s *KratosService) translateKratosMessage(messageID int, originalText string) string {
	// 根据Kratos的错误ID进行翻译
	switch messageID {
	case 4000007:
		return "该邮箱、用户名或手机号已被使用"
	case 4000006:
		return "用户名或密码错误"
	case 4000001:
		return "输入的信息格式不正确"
	case 4000002:
		return "必填字段不能为空"
	case 4000003:
		return "邮箱格式不正确"
	case 4000004:
		return "密码不符合要求"
	case 4000005:
		return "用户名格式不正确"
	default:
		// 如果没有对应的翻译，尝试根据英文内容进行简单翻译
		return s.translateEnglishMessage(originalText)
	}
}

// translateEnglishMessage 简单的英文错误信息翻译
func (s *KratosService) translateEnglishMessage(text string) string {
	text = strings.ToLower(text)

	if strings.Contains(text, "already exists") || strings.Contains(text, "same identifier") {
		return "该邮箱、用户名或手机号已被使用"
	}
	if strings.Contains(text, "invalid credentials") || strings.Contains(text, "password") {
		return "用户名或密码错误"
	}
	if strings.Contains(text, "email") && strings.Contains(text, "invalid") {
		return "邮箱格式不正确"
	}
	if strings.Contains(text, "username") && strings.Contains(text, "invalid") {
		return "用户名格式不正确"
	}
	if strings.Contains(text, "required") {
		return "必填字段不能为空"
	}

	// 如果无法翻译，返回原文
	return text
}

// translateFieldName 翻译字段名称
func (s *KratosService) translateFieldName(fieldName string) string {
	switch fieldName {
	case "traits.email":
		return "邮箱"
	case "traits.username":
		return "用户名"
	case "traits.phone":
		return "手机号"
	case "password":
		return "密码"
	default:
		return fieldName
	}
}
