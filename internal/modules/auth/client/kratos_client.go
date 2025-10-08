package client

import (
	"context"
	"fmt"

	ory "github.com/ory/kratos-client-go"
)

// KratosClient 封装 Ory Kratos Admin API 和 Public API 调用
type KratosClient struct {
	adminURL  string
	publicURL string
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
		return nil, fmt.Errorf("创建 Kratos identity 失败: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Kratos API 返回错误状态码: %d", resp.StatusCode)
	}

	return identity, nil
}

// GetIdentity 根据 ID 获取身份信息
func (c *KratosClient) GetIdentity(ctx context.Context, identityID string) (*ory.Identity, error) {
	identity, resp, err := c.adminClient.IdentityAPI.GetIdentity(ctx, identityID).Execute()

	if err != nil {
		return nil, fmt.Errorf("获取 Kratos identity 失败: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Kratos API 返回错误状态码: %d", resp.StatusCode)
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
		return nil, fmt.Errorf("更新 Kratos identity 失败: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Kratos API 返回错误状态码: %d", resp.StatusCode)
	}

	return identity, nil
}

// DeleteIdentity 删除身份
func (c *KratosClient) DeleteIdentity(ctx context.Context, identityID string) error {
	resp, err := c.adminClient.IdentityAPI.DeleteIdentity(ctx, identityID).Execute()

	if err != nil {
		return fmt.Errorf("删除 Kratos identity 失败: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Kratos API 返回错误状态码: %d", resp.StatusCode)
	}

	return nil
}

// ValidateSession 验证 Session
// 注意：这个方法通过 Public API 调用
func (c *KratosClient) ValidateSession(ctx context.Context, sessionToken string) (*ory.Session, error) {
	if c.publicClient == nil {
		return nil, fmt.Errorf("Public API client 未初始化,请先调用 SetPublicURL")
	}

	session, resp, err := c.publicClient.FrontendAPI.ToSession(ctx).
		XSessionToken(sessionToken).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("验证 Session 失败: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Kratos API 返回错误状态码: %d", resp.StatusCode)
	}

	return session, nil
}

// LoginWithPassword 使用密码登录
// 返回 Session Token 和 Identity ID
func (c *KratosClient) LoginWithPassword(ctx context.Context, identifier, password string) (sessionToken, identityID string, err error) {
	if c.publicClient == nil {
		return "", "", fmt.Errorf("Public API client 未初始化,请先调用 SetPublicURL")
	}

	// 1. 创建 Login Flow
	flow, _, err := c.publicClient.FrontendAPI.CreateNativeLoginFlow(ctx).Execute()
	if err != nil {
		return "", "", fmt.Errorf("创建登录流程失败: %w", err)
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
		return "", "", fmt.Errorf("登录失败: %w", err)
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
		return "", "", fmt.Errorf("登录成功但未返回 Identity")
	}

	return sessionToken, identityID, nil
}

// RevokeSession 撤销 Session (登出)
func (c *KratosClient) RevokeSession(ctx context.Context, sessionToken string) error {
	if c.publicClient == nil {
		return fmt.Errorf("Public API client 未初始化,请先调用 SetPublicURL")
	}

	// 方法1: 通过 Admin API 删除 Session (需要先获取 Session ID)
	// 由于我们只有 sessionToken,需要先验证获取 Session 对象
	session, err := c.ValidateSession(ctx, sessionToken)
	if err != nil {
		return fmt.Errorf("获取 Session 失败: %w", err)
	}

	// 使用 Admin API 删除 Session
	_, err = c.adminClient.IdentityAPI.DisableSession(ctx, session.Id).Execute()
	if err != nil {
		return fmt.Errorf("撤销 Session 失败: %w", err)
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
		return nil, fmt.Errorf("查询 Identity 失败: %w", err)
	}

	if len(identities) == 0 {
		return nil, fmt.Errorf("未找到匹配的用户: %s", identifier)
	}

	return &identities[0], nil
}

// GetIdentityTraits 从 Identity 中提取 traits（类型安全的辅助方法）
func GetIdentityTraits(identity *ory.Identity) (email, username string, err error) {
	traits, ok := identity.Traits.(map[string]interface{})
	if !ok {
		return "", "", fmt.Errorf("identity traits 类型错误")
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
