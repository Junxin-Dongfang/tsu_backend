package client

import (
	"context"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/xerrors"

	ory "github.com/ory/kratos-client-go"
)

// KratosClient 封装 Ory Kratos Admin API 和 Public API 调用
type KratosClient struct {
	adminURL     string
	publicURL    string
	adminClient  *ory.APIClient
	publicClient *ory.APIClient
}

// NewKratosClient 创建 Kratos 客户端
func NewKratosClient(adminURL string) *KratosClient {
	adminConfig := ory.NewConfiguration()
	adminConfig.Servers = []ory.ServerConfiguration{
		{
			URL: adminURL,
		},
	}

	return &KratosClient{
		adminURL:     adminURL,
		adminClient:  ory.NewAPIClient(adminConfig),
		publicClient: nil, // Will be set by SetPublicURL
	}
}

// SetPublicURL 设置 Public API URL 并创建对应的客户端
func (c *KratosClient) SetPublicURL(publicURL string) {
	c.publicURL = publicURL
	publicConfig := ory.NewConfiguration()
	publicConfig.Servers = []ory.ServerConfiguration{
		{
			URL: publicURL,
		},
	}
	c.publicClient = ory.NewAPIClient(publicConfig)
}

// CreateIdentity 在 Kratos 中创建新的身份
// 参数：
//   - email: 用户邮箱
//   - username: 用户名
//   - password: 密码
//
// 返回：
//   - Identity: Kratos 身份信息
//   - error: 错误信息
func (c *KratosClient) CreateIdentity(ctx context.Context, email, username, password string) (*ory.Identity, error) {
	// 构建身份特征数据（traits）
	traits := map[string]interface{}{
		"email":    email,
		"username": username,
	}

	// 构建密码凭证
	credentials := ory.IdentityWithCredentials{
		Password: &ory.IdentityWithCredentialsPassword{
			Config: &ory.IdentityWithCredentialsPasswordConfig{
				Password: &password,
			},
		},
	}

	// 创建身份请求体
	createIdentityBody := ory.CreateIdentityBody{
		SchemaId:    "default",
		Traits:      traits,
		Credentials: &credentials,
	}

	// 调用 Kratos API
	identity, resp, err := c.adminClient.IdentityAPI.CreateIdentity(ctx).
		CreateIdentityBody(createIdentityBody).
		Execute()

	if err != nil {
		log.ErrorContext(ctx, "创建 Kratos identity 失败", err)
		return nil, xerrors.NewKratosError("CreateIdentity", err).
			WithService("kratos_client", "CreateIdentity").
			WithMetadata("email", email).
			WithMetadata("username", username)
	}

	if resp.StatusCode >= 400 {
		log.WarnContext(ctx, "Kratos API 返回错误状态码",
			"status_code", resp.StatusCode,
			"operation", "CreateIdentity")
		return nil, xerrors.NewKratosAPIError("CreateIdentity", resp.StatusCode).
			WithService("kratos_client", "CreateIdentity").
			WithMetadata("email", email)
	}

	log.InfoContext(ctx, "成功创建 Kratos identity",
		"identity_id", identity.Id,
		"email", email)

	return identity, nil
}

// GetIdentity 根据 ID 获取身份信息
func (c *KratosClient) GetIdentity(ctx context.Context, identityID string) (*ory.Identity, error) {
	identity, resp, err := c.adminClient.IdentityAPI.GetIdentity(ctx, identityID).Execute()

	if err != nil {
		log.ErrorContext(ctx, "获取 Kratos identity 失败", err)
		return nil, xerrors.NewKratosError("GetIdentity", err).
			WithService("kratos_client", "GetIdentity").
			WithMetadata("identity_id", identityID)
	}

	if resp.StatusCode >= 400 {
		log.WarnContext(ctx, "Kratos API 返回错误状态码",
			"status_code", resp.StatusCode,
			"operation", "GetIdentity")
		return nil, xerrors.NewKratosAPIError("GetIdentity", resp.StatusCode).
			WithService("kratos_client", "GetIdentity").
			WithMetadata("identity_id", identityID)
	}

	return identity, nil
}

// UpdateIdentity 更新身份信息
func (c *KratosClient) UpdateIdentity(ctx context.Context, identityID string, email, username string) (*ory.Identity, error) {
	// 构建更新后的 traits
	traits := map[string]interface{}{
		"email":    email,
		"username": username,
	}

	updateIdentityBody := ory.UpdateIdentityBody{
		SchemaId: "default",
		Traits:   traits,
	}

	identity, resp, err := c.adminClient.IdentityAPI.UpdateIdentity(ctx, identityID).
		UpdateIdentityBody(updateIdentityBody).
		Execute()

	if err != nil {
		log.ErrorContext(ctx, "更新 Kratos identity 失败", err)
		return nil, xerrors.NewKratosError("UpdateIdentity", err).
			WithService("kratos_client", "UpdateIdentity").
			WithMetadata("identity_id", identityID).
			WithMetadata("email", email)
	}

	if resp.StatusCode >= 400 {
		log.WarnContext(ctx, "Kratos API 返回错误状态码",
			"status_code", resp.StatusCode,
			"operation", "UpdateIdentity")
		return nil, xerrors.NewKratosAPIError("UpdateIdentity", resp.StatusCode).
			WithService("kratos_client", "UpdateIdentity").
			WithMetadata("identity_id", identityID)
	}

	log.InfoContext(ctx, "成功更新 Kratos identity",
		"identity_id", identityID,
		"email", email)

	return identity, nil
}

// DeleteIdentity 删除身份
func (c *KratosClient) DeleteIdentity(ctx context.Context, identityID string) error {
	resp, err := c.adminClient.IdentityAPI.DeleteIdentity(ctx, identityID).Execute()

	if err != nil {
		log.ErrorContext(ctx, "删除 Kratos identity 失败", err)
		return xerrors.NewKratosError("DeleteIdentity", err).
			WithService("kratos_client", "DeleteIdentity").
			WithMetadata("identity_id", identityID)
	}

	if resp.StatusCode >= 400 {
		log.WarnContext(ctx, "Kratos API 返回错误状态码",
			"status_code", resp.StatusCode,
			"operation", "DeleteIdentity")
		return xerrors.NewKratosAPIError("DeleteIdentity", resp.StatusCode).
			WithService("kratos_client", "DeleteIdentity").
			WithMetadata("identity_id", identityID)
	}

	log.InfoContext(ctx, "成功删除 Kratos identity", "identity_id", identityID)
	return nil
}

// ValidateSession 验证 Session
// 注意：这个方法通过 Public API 调用
func (c *KratosClient) ValidateSession(ctx context.Context, sessionToken string) (*ory.Session, error) {
	if c.publicClient == nil {
		return nil, xerrors.NewKratosClientNotInitializedError("public_api").
			WithService("kratos_client", "ValidateSession")
	}

	session, resp, err := c.publicClient.FrontendAPI.ToSession(ctx).
		XSessionToken(sessionToken).
		Execute()

	if err != nil {
		log.ErrorContext(ctx, "验证 Session 失败", err)
		return nil, xerrors.NewSessionInvalidError("session validation failed").
			WithService("kratos_client", "ValidateSession")
	}

	if resp.StatusCode >= 400 {
		log.WarnContext(ctx, "Kratos API 返回错误状态码",
			"status_code", resp.StatusCode,
			"operation", "ValidateSession")
		if resp.StatusCode == 401 {
			return nil, xerrors.NewSessionExpiredError().
				WithService("kratos_client", "ValidateSession")
		}
		return nil, xerrors.NewKratosAPIError("ValidateSession", resp.StatusCode).
			WithService("kratos_client", "ValidateSession")
	}

	return session, nil
}

// LoginWithPassword 使用密码登录
// 返回 Session Token 和 Identity ID
func (c *KratosClient) LoginWithPassword(ctx context.Context, identifier, password string) (sessionToken, identityID string, err error) {
	if c.publicClient == nil {
		return "", "", xerrors.NewKratosClientNotInitializedError("public_api").
			WithService("kratos_client", "LoginWithPassword")
	}

	// 1. 创建 Login Flow
	flow, _, err := c.publicClient.FrontendAPI.CreateNativeLoginFlow(ctx).Execute()
	if err != nil {
		log.ErrorContext(ctx, "创建登录流程失败", err)
		return "", "", xerrors.NewKratosError("CreateNativeLoginFlow", err).
			WithService("kratos_client", "LoginWithPassword").
			WithMetadata("identifier", identifier)
	}

	// 2. 提交登录凭证
	updateLoginBody := ory.UpdateLoginFlowBody{
		UpdateLoginFlowWithPasswordMethod: &ory.UpdateLoginFlowWithPasswordMethod{
			Method:     "password",
			Identifier: identifier,
			Password:   password,
		},
	}

	result, _, err := c.publicClient.FrontendAPI.UpdateLoginFlow(ctx).
		Flow(flow.Id).
		UpdateLoginFlowBody(updateLoginBody).
		Execute()

	if err != nil {
		log.ErrorContext(ctx, "登录失败", err)
		return "", "", xerrors.NewAuthError("登录失败").
			WithService("kratos_client", "LoginWithPassword").
			WithMetadata("identifier", identifier)
	}

	// 3. 提取 Session Token 和 Identity ID
	// SessionToken 优先使用 API 返回的 session_token (用于 API 流程)
	if result.SessionToken != nil {
		sessionToken = *result.SessionToken
	} else {
		// 如果没有 session_token,使用 session.id 作为 token
		sessionToken = result.Session.Id
	}

	// 提取 Identity ID
	if result.Session.Identity != nil {
		identityID = result.Session.Identity.Id
	} else {
		return "", "", xerrors.NewKratosDataIntegrityError("session.identity", "登录成功但未返回 Identity").
			WithService("kratos_client", "LoginWithPassword")
	}

	log.InfoContext(ctx, "用户登录成功",
		"identity_id", identityID,
		"identifier", identifier)

	return sessionToken, identityID, nil
}

// RevokeSession 撤销 Session (登出)
// ✅ 使用官方推荐的 Native Logout API
// 参考: https://www.ory.sh/docs/kratos/self-service/flows/user-logout
func (c *KratosClient) RevokeSession(ctx context.Context, sessionToken string) error {
	if c.publicClient == nil {
		return xerrors.NewKratosClientNotInitializedError("Public").WithContext(&xerrors.ErrorContext{
			Service:   "auth",
			Operation: "RevokeSession",
		})
	}

	// ✅ 使用 PerformNativeLogout (Native/API 应用的正确方式)
	// 不需要创建 logout flow，直接调用 API 端点
	logoutBody := ory.NewPerformNativeLogoutBody(sessionToken)

	_, err := c.publicClient.FrontendAPI.PerformNativeLogout(ctx).
		PerformNativeLogoutBody(*logoutBody).
		Execute()

	if err != nil {
		return xerrors.ParseKratosError("RevokeSession", err)
	}

	return nil
}

// GetIdentityByIdentifier 根据标识符(email/username/phone)查询 Identity
// 注意：Kratos Admin API 支持通过 credentials_identifier 查询
func (c *KratosClient) GetIdentityByIdentifier(ctx context.Context, identifier string) (*ory.Identity, error) {
	// 使用 ListIdentities API 并通过 credentials_identifier 过滤
	identities, _, err := c.adminClient.IdentityAPI.ListIdentities(ctx).
		CredentialsIdentifier(identifier).
		Execute()

	if err != nil {
		return nil, xerrors.ParseKratosError("GetIdentityByIdentifier", err)
	}

	if len(identities) == 0 {
		return nil, xerrors.New(xerrors.CodeResourceNotFound, "未找到匹配的用户").
			WithMetadata("identifier", identifier).
			WithContext(&xerrors.ErrorContext{
				Service:   "auth",
				Operation: "GetIdentityByIdentifier",
			})
	}

	return &identities[0], nil
}

// GetIdentityTraits 从 Identity 中提取 traits（类型安全的辅助方法）
func GetIdentityTraits(identity *ory.Identity) (email, username string, err error) {
	traits, ok := identity.Traits.(map[string]interface{})
	if !ok {
		return "", "", xerrors.NewKratosDataIntegrityError("identity traits 类型错误", "traits").
			WithMetadata("expected_type", "map[string]interface{}").
			WithContext(&xerrors.ErrorContext{
				Service:   "auth",
				Operation: "GetIdentityTraits",
			})
	}

	emailVal, ok := traits["email"]
	if ok {
		email, _ = emailVal.(string)
	}

	usernameVal, ok := traits["username"]
	if ok {
		username, _ = usernameVal.(string)
	}

	return email, username, nil
}

// ==================== Registration Flow (注册流程) ====================

// CreateRegistrationFlow 创建注册流程（API/Native App 模式）
func (c *KratosClient) CreateRegistrationFlow(ctx context.Context) (*ory.RegistrationFlow, error) {
	if c.publicClient == nil {
		return nil, xerrors.NewKratosClientNotInitializedError("Public").WithContext(&xerrors.ErrorContext{
			Service:   "auth",
			Operation: "CreateRegistrationFlow",
		})
	}

	flow, resp, err := c.publicClient.FrontendAPI.CreateNativeRegistrationFlow(ctx).
		Execute()

	if err != nil {
		return nil, xerrors.ParseKratosError("CreateRegistrationFlow", err)
	}

	if resp.StatusCode >= 400 {
		return nil, xerrors.HandleKratosAPIError("CreateRegistrationFlow", resp.StatusCode)
	}

	return flow, nil
}

// Register 完成注册（API/Native App 模式）
// 返回 session token 和 identity ID
// username: 用户名（必需，因为 schema 要求）
func (c *KratosClient) Register(ctx context.Context, flowID, email, username, password string) (sessionToken, identityID string, err error) {
	if c.publicClient == nil {
		return "", "", xerrors.NewKratosClientNotInitializedError("Public").WithContext(&xerrors.ErrorContext{
			Service:   "auth",
			Operation: "Register",
		})
	}

	// 构造 traits（根据 identity schema 的要求）
	traits := map[string]interface{}{
		"email":    email,
		"username": username, // 必需字段
	}

	// 构造注册请求
	updateBody := ory.UpdateRegistrationFlowBody{
		UpdateRegistrationFlowWithPasswordMethod: &ory.UpdateRegistrationFlowWithPasswordMethod{
			Method:   "password",
			Password: password,
			Traits:   traits,
		},
	}

	// 调用 SDK
	result, resp, err := c.publicClient.FrontendAPI.UpdateRegistrationFlow(ctx).
		Flow(flowID).
		UpdateRegistrationFlowBody(updateBody).
		Execute()

	if err != nil {
		return "", "", xerrors.ParseKratosError("Register", err)
	}

	if resp.StatusCode >= 400 {
		return "", "", xerrors.HandleKratosAPIError("Register", resp.StatusCode)
	}

	// 从结果中提取 session token 和 identity ID
	if result.Session != nil && result.Session.Id != "" {
		identityID = result.Session.Identity.Id

		// 方式1：从响应体获取 session token（如果 Kratos 配置返回）
		if result.SessionToken != nil && *result.SessionToken != "" {
			return *result.SessionToken, identityID, nil
		}

		// 方式2：从响应头获取
		sessionToken := resp.Header.Get("X-Session-Token")
		if sessionToken != "" {
			return sessionToken, identityID, nil
		}

		// 方式3：返回 session ID（可以用于后续查询）
		return result.Session.Id, identityID, nil
	}

	return "", "", xerrors.NewKratosDataIntegrityError("注册成功但未返回 session", "session").WithContext(&xerrors.ErrorContext{
		Service:   "auth",
		Operation: "Register",
	})
}

// ==================== 密码重置功能 ====================

// CreateRecoveryFlow 创建密码恢复流程 (用户端)
// 返回 Flow ID 用于后续操作
func (c *KratosClient) CreateRecoveryFlow(ctx context.Context) (*ory.RecoveryFlow, error) {
	if c.publicClient == nil {
		return nil, xerrors.NewKratosClientNotInitializedError("Public").WithContext(&xerrors.ErrorContext{
			Service:   "auth",
			Operation: "CreateRecoveryFlow",
		})
	}

	flow, resp, err := c.publicClient.FrontendAPI.CreateNativeRecoveryFlow(ctx).Execute()
	if err != nil {
		return nil, xerrors.ParseKratosError("CreateRecoveryFlow", err)
	}

	if resp.StatusCode >= 400 {
		return nil, xerrors.HandleKratosAPIError("CreateRecoveryFlow", resp.StatusCode)
	}

	return flow, nil
}

// UpdateRecoveryFlowWithCode 提交邮箱请求验证码（使用 SDK）
func (c *KratosClient) UpdateRecoveryFlowWithCode(ctx context.Context, flowID, email string) (*ory.RecoveryFlow, error) {
	if c.publicClient == nil {
		return nil, xerrors.NewKratosClientNotInitializedError("Public").WithContext(&xerrors.ErrorContext{
			Service:   "auth",
			Operation: "UpdateRecoveryFlowWithCode",
		})
	}

	// 构造请求体
	updateBody := ory.UpdateRecoveryFlowBody{
		UpdateRecoveryFlowWithCodeMethod: &ory.UpdateRecoveryFlowWithCodeMethod{
			Method: "code",
			Email:  &email,
		},
	}

	// 调用 SDK
	flow, resp, err := c.publicClient.FrontendAPI.UpdateRecoveryFlow(ctx).
		Flow(flowID).
		UpdateRecoveryFlowBody(updateBody).
		Execute()

	if err != nil {
		return nil, xerrors.ParseKratosError("UpdateRecoveryFlowWithCode", err)
	}

	if resp.StatusCode >= 400 {
		return nil, xerrors.HandleKratosAPIError("UpdateRecoveryFlowWithCode", resp.StatusCode)
	}

	return flow, nil
}

// VerifyRecoveryCodeAndGetSessionToken 提交验证码并提取特权信息（使用 SDK）
// 返回：(sessionToken, settingsFlowID, error)
// 注意：启用 use_continue_with_transitions 后，Kratos 会在 ContinueWith 中返回 session token 和 settings flow
func (c *KratosClient) VerifyRecoveryCodeAndGetSessionToken(ctx context.Context, flowID, code string) (sessionToken, settingsFlowID string, err error) {
	if c.publicClient == nil {
		return "", "", xerrors.NewKratosClientNotInitializedError("Public").WithContext(&xerrors.ErrorContext{
			Service:   "auth",
			Operation: "VerifyRecoveryCodeAndGetSessionToken",
		})
	}

	// 构造请求体
	updateBody := ory.UpdateRecoveryFlowBody{
		UpdateRecoveryFlowWithCodeMethod: &ory.UpdateRecoveryFlowWithCodeMethod{
			Method: "code",
			Code:   &code,
		},
	}

	// 调用 SDK
	flow, _, err := c.publicClient.FrontendAPI.UpdateRecoveryFlow(ctx).
		Flow(flowID).
		UpdateRecoveryFlowBody(updateBody).
		Execute()

	if err != nil {
		return "", "", xerrors.ParseKratosError("VerifyRecoveryCodeAndGetSessionToken", err)
	}

	// ✅ 正确的方式：从 ContinueWith 中提取信息
	// 当启用 use_continue_with_transitions 后，成功的响应会包含：
	// 1. ContinueWithSetOrySessionToken - 包含 session token
	// 2. ContinueWithSettingsUi - 包含 settings flow ID
	if len(flow.ContinueWith) > 0 {
		var foundSessionToken string
		var foundSettingsFlowID string

		for _, item := range flow.ContinueWith {
			// 提取 session token
			if sessionTokenItem := item.ContinueWithSetOrySessionToken; sessionTokenItem != nil {
				foundSessionToken = sessionTokenItem.OrySessionToken
				log.DebugContext(ctx, "从 ContinueWith 提取到 session token",
					"token_prefix", foundSessionToken[:30])
			}

			// 提取 settings flow ID
			if settingsUiItem := item.ContinueWithSettingsUi; settingsUiItem != nil {
				if settingsUiItem.Flow.Id != "" {
					foundSettingsFlowID = settingsUiItem.Flow.Id
					log.DebugContext(ctx, "从 ContinueWith 提取到 settings flow ID",
						"flow_id", foundSettingsFlowID)
				}
			}
		}

		// 两者都必须存在
		if foundSessionToken != "" && foundSettingsFlowID != "" {
			return foundSessionToken, foundSettingsFlowID, nil
		}

		// 如果只有一个，返回详细错误
		if foundSessionToken == "" {
			return "", "", xerrors.NewKratosDataIntegrityError("ContinueWith 中缺少 session token", "session_token").WithContext(&xerrors.ErrorContext{
				Service:   "auth",
				Operation: "VerifyRecoveryCodeAndGetSessionToken",
			})
		}
		if foundSettingsFlowID == "" {
			return "", "", xerrors.NewKratosDataIntegrityError("ContinueWith 中缺少 settings flow ID", "settings_flow_id").WithContext(&xerrors.ErrorContext{
				Service:   "auth",
				Operation: "VerifyRecoveryCodeAndGetSessionToken",
			})
		}
	}

	return "", "", xerrors.New(xerrors.CodeInternalError, "验证码验证成功但响应中缺少 ContinueWith 信息").
		WithMetadata("hint", "请检查 Kratos 配置中的 feature_flags.use_continue_with_transitions").
		WithContext(&xerrors.ErrorContext{
			Service:   "auth",
			Operation: "VerifyRecoveryCodeAndGetSessionToken",
		})
}

// CreateSettingsFlow 创建设置流程 (用于修改密码)
func (c *KratosClient) CreateSettingsFlow(ctx context.Context, sessionToken string) (*ory.SettingsFlow, error) {
	if c.publicClient == nil {
		return nil, xerrors.NewKratosClientNotInitializedError("Public").WithContext(&xerrors.ErrorContext{
			Service:   "auth",
			Operation: "CreateSettingsFlow",
		})
	}

	flow, resp, err := c.publicClient.FrontendAPI.CreateNativeSettingsFlow(ctx).
		XSessionToken(sessionToken).
		Execute()

	if err != nil {
		return nil, xerrors.ParseKratosError("CreateSettingsFlow", err)
	}

	if resp.StatusCode >= 400 {
		return nil, xerrors.HandleKratosAPIError("CreateSettingsFlow", resp.StatusCode)
	}

	return flow, nil
}

// UpdatePasswordInSettingsFlow 在设置流程中修改密码
func (c *KratosClient) UpdatePasswordInSettingsFlow(ctx context.Context, flowID, sessionToken, newPassword string) (*ory.SettingsFlow, error) {
	if c.publicClient == nil {
		return nil, xerrors.NewKratosClientNotInitializedError("Public").WithContext(&xerrors.ErrorContext{
			Service:   "auth",
			Operation: "UpdatePasswordInSettingsFlow",
		})
	}

	// 构造请求体
	updateBody := ory.UpdateSettingsFlowBody{
		UpdateSettingsFlowWithPasswordMethod: &ory.UpdateSettingsFlowWithPasswordMethod{
			Method:   "password",
			Password: newPassword,
		},
	}

	flow, resp, err := c.publicClient.FrontendAPI.UpdateSettingsFlow(ctx).
		Flow(flowID).
		XSessionToken(sessionToken).
		UpdateSettingsFlowBody(updateBody).
		Execute()

	if err != nil {
		return nil, xerrors.ParseKratosError("UpdatePasswordInSettingsFlow", err)
	}

	if resp.StatusCode >= 400 {
		return nil, xerrors.HandleKratosAPIError("UpdatePasswordInSettingsFlow", resp.StatusCode)
	}

	return flow, nil
}

// GetIdentityFromSettingsFlow 从 settings flow 中获取 identity 信息
// 注意：settings flow ID 实际上是从 recovery 422 响应中获取的
// 我们需要用特殊方式获取其中的 identity 信息
func (c *KratosClient) GetIdentityFromSettingsFlow(ctx context.Context, settingsFlowID string) (*ory.Identity, error) {
	// 由于 settings flow 需要 session 才能访问，而我们在 API 模式下没有 session
	// 我们需要换一个思路：直接解析 settingsFlowID 或使用其他方式
	//
	// 最简单的方案：让调用方提供 email，我们用 email 查找 identity
	// 但这需要修改 API 设计
	//
	// 临时方案：返回错误，提示需要其他方式
	return nil, xerrors.New(xerrors.CodeOperationNotAllowed, "无法从 settings flow 获取 identity").
		WithMetadata("hint", "需要使用 email 查找").
		WithMetadata("settings_flow_id", settingsFlowID).
		WithContext(&xerrors.ErrorContext{
			Service:   "auth",
			Operation: "GetIdentityFromSettingsFlow",
		})
}

// GetIdentityByEmail 通过邮箱查找 identity
func (c *KratosClient) GetIdentityByEmail(ctx context.Context, email string) (*ory.Identity, error) {
	if c.adminClient == nil {
		return nil, xerrors.NewKratosClientNotInitializedError("Admin").WithContext(&xerrors.ErrorContext{
			Service:   "auth",
			Operation: "GetIdentityByEmail",
		})
	}

	// 使用 Admin API 列出用户，通过 credentials_identifier 过滤（即邮箱）
	identities, resp, err := c.adminClient.IdentityAPI.ListIdentities(ctx).
		CredentialsIdentifier(email).
		Execute()

	if err != nil {
		return nil, xerrors.ParseKratosError("GetIdentityByEmail", err)
	}

	if resp.StatusCode >= 400 {
		return nil, xerrors.HandleKratosAPIError("GetIdentityByEmail", resp.StatusCode)
	}

	if len(identities) == 0 {
		return nil, xerrors.New(xerrors.CodeResourceNotFound, "未找到匹配的用户").
			WithMetadata("email", email).
			WithContext(&xerrors.ErrorContext{
				Service:   "auth",
				Operation: "GetIdentityByEmail",
			})
	}

	// 返回第一个匹配的用户
	return &identities[0], nil
}

// UpdatePasswordWithPrivilegedFlow 使用特权 settings flow 更新密码
// privilegedFlowID: 从 recovery flow 返回的特权 settings flow ID
// sessionToken: 从 recovery flow 返回的特权 session token
// newPassword: 新密码
func (c *KratosClient) UpdatePasswordWithPrivilegedFlow(ctx context.Context, privilegedFlowID, sessionToken, newPassword string) error {
	if c.publicClient == nil {
		return xerrors.NewKratosClientNotInitializedError("Public").WithContext(&xerrors.ErrorContext{
			Service:   "auth",
			Operation: "UpdatePasswordWithPrivilegedFlow",
		})
	}

	if sessionToken == "" {
		return xerrors.New(xerrors.CodeInvalidParams, "验证流程异常：未获取到有效的会话凭证").
			WithMetadata("hint", "请重新发起密码恢复").
			WithContext(&xerrors.ErrorContext{
				Service:   "auth",
				Operation: "UpdatePasswordWithPrivilegedFlow",
			})
	}

	// 构造密码更新请求
	updateBody := ory.UpdateSettingsFlowBody{
		UpdateSettingsFlowWithPasswordMethod: &ory.UpdateSettingsFlowWithPasswordMethod{
			Method:   "password",
			Password: newPassword,
		},
	}

	// 调试输出
	if len(sessionToken) > 30 {
		log.DebugContext(ctx, "使用 Session Token",
			"token_prefix", sessionToken[:30])
	} else {
		log.DebugContext(ctx, "使用 Session Token",
			"token", sessionToken)
	}

	// ✅ 正确的方式：使用 X-Session-Token header (API/Native App 模式)
	// 参考 Kratos 官方文档：https://www.ory.sh/docs/kratos/self-service/flows/user-settings
	_, resp, err := c.publicClient.FrontendAPI.UpdateSettingsFlow(ctx).
		Flow(privilegedFlowID).
		XSessionToken(sessionToken). // 使用 X-Session-Token header
		UpdateSettingsFlowBody(updateBody).
		Execute()

	if err != nil {
		return xerrors.ParseKratosError("UpdatePasswordWithPrivilegedFlow", err)
	}

	if resp != nil && resp.StatusCode >= 400 {
		return xerrors.HandleKratosAPIError("UpdatePasswordWithPrivilegedFlow", resp.StatusCode)
	}

	log.InfoContext(ctx, "密码更新成功", "status_code", resp.StatusCode)
	return nil
}

// AdminUpdateIdentityPassword 使用 Admin API 更新用户密码
// ⚠️ 注意：这不是推荐的密码重置方式，仅用于管理员操作
func (c *KratosClient) AdminUpdateIdentityPassword(ctx context.Context, identityID, newPassword string) error {
	if c.adminClient == nil {
		return xerrors.NewKratosClientNotInitializedError("Admin").WithContext(&xerrors.ErrorContext{
			Service:   "auth",
			Operation: "AdminUpdateIdentityPassword",
		})
	}

	// 先获取当前 identity
	identity, resp, err := c.adminClient.IdentityAPI.GetIdentity(ctx, identityID).Execute()
	if err != nil {
		return xerrors.ParseKratosError("AdminUpdateIdentityPassword.GetIdentity", err)
	}

	if resp.StatusCode >= 400 {
		return xerrors.HandleKratosAPIError("AdminUpdateIdentityPassword.GetIdentity", resp.StatusCode)
	}

	// 构建更新请求，保留现有的 traits，只更新密码
	credentials := ory.IdentityWithCredentials{
		Password: &ory.IdentityWithCredentialsPassword{
			Config: &ory.IdentityWithCredentialsPasswordConfig{
				Password: &newPassword,
			},
		},
	}

	// Traits 需要类型断言为 map[string]interface{}
	traits, ok := identity.Traits.(map[string]interface{})
	if !ok {
		return xerrors.NewKratosDataIntegrityError("identity traits 类型错误", "traits").
			WithMetadata("identity_id", identityID).
			WithContext(&xerrors.ErrorContext{
				Service:   "auth",
				Operation: "AdminUpdateIdentityPassword",
			})
	}

	updateBody := ory.UpdateIdentityBody{
		SchemaId:    identity.SchemaId,
		Traits:      traits,
		Credentials: &credentials,
		State:       *identity.State,
	}

	// 更新 identity
	_, resp, err = c.adminClient.IdentityAPI.UpdateIdentity(ctx, identityID).
		UpdateIdentityBody(updateBody).
		Execute()

	if err != nil {
		return xerrors.ParseKratosError("AdminUpdateIdentityPassword.UpdateIdentity", err)
	}

	if resp.StatusCode >= 400 {
		return xerrors.HandleKratosAPIError("AdminUpdateIdentityPassword.UpdateIdentity", resp.StatusCode)
	}

	return nil
}

// AdminCreateRecoveryLink 管理员为用户创建恢复链接 (备用方案)
func (c *KratosClient) AdminCreateRecoveryLink(ctx context.Context, identityID string, expiresIn string) (*ory.RecoveryLinkForIdentity, error) {
	// 构造请求体
	createBody := ory.CreateRecoveryLinkForIdentityBody{
		IdentityId: identityID,
		ExpiresIn:  &expiresIn,
	}

	result, resp, err := c.adminClient.IdentityAPI.CreateRecoveryLinkForIdentity(ctx).
		CreateRecoveryLinkForIdentityBody(createBody).
		Execute()

	if err != nil {
		return nil, xerrors.ParseKratosError("AdminCreateRecoveryLink", err)
	}

	if resp.StatusCode >= 400 {
		return nil, xerrors.HandleKratosAPIError("AdminCreateRecoveryLink", resp.StatusCode)
	}

	return result, nil
}
