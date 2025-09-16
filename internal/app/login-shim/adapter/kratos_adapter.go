// File: cmd/login-shim/internal/adapter/kratos_adapter.go (改进版)
package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"tsu-self/internal/app/login-shim/domain"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/xerrors"
)

// KratosAdapter 封装了与 ORY Kratos 交互的所有逻辑
type KratosAdapter struct {
	kratosPublicURL string
	timeout         time.Duration
}

// KratosErrorPayload 定义了 Kratos 错误响应结构
type KratosErrorPayload struct {
	UI struct {
		Messages []struct {
			ID   int    `json:"id"`
			Text string `json:"text"`
			Type string `json:"type"`
		} `json:"messages"`
	} `json:"ui"`
	Error *struct {
		ID int64 `json:"id"`
	} `json:"error,omitempty"`
}

// NewKratosAdapter 创建新的 KratosAdapter 实例
func NewKratosAdapter(kratosPublicURL string) *KratosAdapter {
	return &KratosAdapter{
		kratosPublicURL: kratosPublicURL,
		timeout:         30 * time.Second, // 设置合理的超时时间
	}
}

// Login 执行登录流程
func (ka *KratosAdapter) Login(ctx context.Context, req *domain.LoginRequest) (string, *KratosErrorPayload, *xerrors.AppError) {
	log.InfoWithCtx(ctx, "KratosAdapter: 开始登录流程", "identifier", req.Identifier)

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, ka.timeout)
	defer cancel()

	client := ka.createClient()

	// 1. 初始化登录流程
	flowID, err := ka.initializeFlow(ctx, client, "/self-service/login/browser")
	if err != nil {
		log.ErrorWithCtx(ctx, "初始化登录流程失败", err)
		return "", nil, xerrors.NewSystemError("初始化登录流程失败").
			WithContext(&xerrors.ErrorContext{
				Service: "kratos-adapter",
				Method:  "Login",
			})
	}

	// 2. 获取流程详情
	actionURL, csrfToken, err := ka.getFlowDetails(ctx, client, "/self-service/login/flows", flowID)
	if err != nil {
		log.ErrorWithCtx(ctx, "获取登录流程详情失败", err)
		return "", nil, xerrors.NewSystemError("获取登录流程详情失败").
			WithContext(&xerrors.ErrorContext{
				Service:  "kratos-adapter",
				Method:   "Login",
				Metadata: map[string]string{"flow_id": flowID},
			})
	}

	// 3. 修正 Action URL
	finalActionURL, err := ka.rewriteActionURL(actionURL)
	if err != nil {
		log.ErrorWithCtx(ctx, "修正 Action URL 失败", err)
		return "", nil, xerrors.NewSystemError("修正 Action URL 失败").
			WithContext(&xerrors.ErrorContext{
				Service:  "kratos-adapter",
				Method:   "Login",
				Metadata: map[string]string{"original_url": actionURL},
			})
	}

	// 4. 提交登录表单
	formData := url.Values{
		"method":     {"password"},
		"identifier": {req.Identifier},
		"password":   {req.Password},
		"csrf_token": {csrfToken},
	}

	loginResp, err := client.PostForm(finalActionURL, formData)
	if err != nil {
		log.ErrorWithCtx(ctx, "提交登录表单失败", err)
		return "", nil, xerrors.NewSystemError("提交登录表单失败").
			WithRetryable(true).
			WithContext(&xerrors.ErrorContext{
				Service: "kratos-adapter",
				Method:  "Login",
			})
	}
	defer loginResp.Body.Close()

	// 5. 处理响应
	return ka.processAuthResponse(ctx, client, loginResp, "/self-service/login/flows", flowID)
}

// Register 执行注册流程
func (ka *KratosAdapter) Register(ctx context.Context, req *domain.RegisterRequest) (string, *KratosErrorPayload, *xerrors.AppError) {
	log.InfoWithCtx(ctx, "KratosAdapter: 开始注册流程", "email", req.Email, "username", req.UserName)

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, ka.timeout)
	defer cancel()

	client := ka.createClient()

	// 1. 初始化注册流程
	flowID, err := ka.initializeFlow(ctx, client, "/self-service/registration/browser")
	if err != nil {
		log.ErrorWithCtx(ctx, "初始化注册流程失败", err)
		return "", nil, xerrors.NewSystemError("初始化注册流程失败").
			WithContext(&xerrors.ErrorContext{
				Service: "kratos-adapter",
				Method:  "Register",
			})
	}

	// 2. 获取流程详情
	actionURL, csrfToken, err := ka.getFlowDetails(ctx, client, "/self-service/registration/flows", flowID)
	if err != nil {
		log.ErrorWithCtx(ctx, "获取注册流程详情失败", err)
		return "", nil, xerrors.NewSystemError("获取注册流程详情失败").
			WithContext(&xerrors.ErrorContext{
				Service:  "kratos-adapter",
				Method:   "Register",
				Metadata: map[string]string{"flow_id": flowID},
			})
	}

	// 3. 修正 Action URL
	finalActionURL, err := ka.rewriteActionURL(actionURL)
	if err != nil {
		log.ErrorWithCtx(ctx, "修正 Action URL 失败", err)
		return "", nil, xerrors.NewSystemError("修正 Action URL 失败").
			WithContext(&xerrors.ErrorContext{
				Service:  "kratos-adapter",
				Method:   "Register",
				Metadata: map[string]string{"original_url": actionURL},
			})
	}

	// 4. 提交注册表单
	formData := url.Values{
		"method":          {"password"},
		"traits.email":    {req.Email},
		"traits.username": {req.UserName},
		"password":        {req.Password},
		"csrf_token":      {csrfToken},
	}

	registerResp, err := client.PostForm(finalActionURL, formData)
	if err != nil {
		log.ErrorWithCtx(ctx, "提交注册表单失败", err)
		return "", nil, xerrors.NewSystemError("提交注册表单失败").
			WithRetryable(true).
			WithContext(&xerrors.ErrorContext{
				Service: "kratos-adapter",
				Method:  "Register",
			})
	}
	defer registerResp.Body.Close()

	// 5. 处理响应
	return ka.processAuthResponse(ctx, client, registerResp, "/self-service/registration/flows", flowID)
}

// processAuthResponse 统一处理认证响应（登录/注册）
func (ka *KratosAdapter) processAuthResponse(
	ctx context.Context,
	client *http.Client,
	resp *http.Response,
	flowPath string,
	flowID string,
) (string, *KratosErrorPayload, *xerrors.AppError) {

	// 检查是否有 session cookie
	if sessionCookie := resp.Header.Get("Set-Cookie"); sessionCookie != "" {
		log.InfoWithCtx(ctx, "认证成功，获得 session cookie")
		return sessionCookie, nil, nil
	}

	// 认证失败，解析错误
	body, _ := io.ReadAll(resp.Body)

	// 尝试从响应体解析错误
	var kratosErr KratosErrorPayload
	if json.Unmarshal(body, &kratosErr) == nil && len(kratosErr.UI.Messages) > 0 {
		log.WarnWithCtx(ctx, "从响应体解析到 Kratos 错误", "messages_count", len(kratosErr.UI.Messages))
		return "", &kratosErr, nil
	}

	// 如果响应体没有错误信息，从流程中获取
	if resp.StatusCode == 302 || resp.StatusCode == 303 {
		if errPayload := ka.fetchFlowError(ctx, client, flowPath, flowID); errPayload != nil {
			return "", errPayload, nil
		}
		// 兜底错误
		return "", nil, xerrors.New(xerrors.CodeInvalidCredentials, "认证失败").
			WithUserMessage("用户名或密码错误")
	}

	// 其他状态码返回系统错误
	return "", nil, xerrors.NewSystemError(fmt.Sprintf("Kratos 返回异常状态码: %d", resp.StatusCode)).
		WithContext(&xerrors.ErrorContext{
			Service: "kratos-adapter",
			Metadata: map[string]string{
				"status_code":   fmt.Sprintf("%d", resp.StatusCode),
				"response_body": string(body),
			},
		})
}

// --- 辅助方法 ---

// createClient 创建配置好的 HTTP 客户端
func (ka *KratosAdapter) createClient() *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{
		Jar:     jar,
		Timeout: ka.timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // 禁止自动重定向
		},
	}
}

// initializeFlow 初始化认证流程
func (ka *KratosAdapter) initializeFlow(ctx context.Context, client *http.Client, path string) (string, error) {
	initURL := ka.kratosPublicURL + path

	req, err := http.NewRequestWithContext(ctx, "GET", initURL, nil)
	if err != nil {
		return "", err
	}

	initResp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer initResp.Body.Close()

	location, err := initResp.Location()
	if err != nil {
		return "", fmt.Errorf("kratos 未返回 location 头: %w", err)
	}

	flowID := location.Query().Get("flow")
	if flowID == "" {
		return "", fmt.Errorf("未能从重定向 URL 中获取 flow ID")
	}

	return flowID, nil
}

// getFlowDetails 获取流程详细信息
func (ka *KratosAdapter) getFlowDetails(ctx context.Context, client *http.Client, path, flowID string) (string, string, error) {
	flowURL := fmt.Sprintf("%s%s?id=%s", ka.kratosPublicURL, path, flowID)

	req, err := http.NewRequestWithContext(ctx, "GET", flowURL, nil)
	if err != nil {
		return "", "", err
	}

	flowResp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer flowResp.Body.Close()

	if flowResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(flowResp.Body)
		return "", "", fmt.Errorf("获取流程详情失败，状态码: %d, 响应: %s", flowResp.StatusCode, string(body))
	}

	var flowData map[string]interface{}
	if err := json.NewDecoder(flowResp.Body).Decode(&flowData); err != nil {
		return "", "", fmt.Errorf("解析流程 JSON 失败: %w", err)
	}

	// 提取 Action URL
	uiData, ok := flowData["ui"].(map[string]interface{})
	if !ok {
		return "", "", fmt.Errorf("响应中缺少 'ui' 字段")
	}

	actionURL, ok := uiData["action"].(string)
	if !ok || actionURL == "" {
		return "", "", fmt.Errorf("响应中缺少或空的 'action' 字段")
	}

	// 提取 CSRF Token
	var csrfToken string
	if nodes, ok := uiData["nodes"].([]interface{}); ok {
		for _, node := range nodes {
			if nodeMap, ok := node.(map[string]interface{}); ok {
				if attributes, ok := nodeMap["attributes"].(map[string]interface{}); ok {
					if name, ok := attributes["name"].(string); ok && name == "csrf_token" {
						if value, ok := attributes["value"].(string); ok {
							csrfToken = value
							break
						}
					}
				}
			}
		}
	}

	if csrfToken == "" {
		return "", "", fmt.Errorf("未找到 csrf_token")
	}

	return actionURL, csrfToken, nil
}

// rewriteActionURL 重写 Action URL 以适应 Docker 网络
func (ka *KratosAdapter) rewriteActionURL(originalURL string) (string, error) {
	parsedURL, err := url.Parse(originalURL)
	if err != nil {
		return "", fmt.Errorf("解析 URL 失败: %w", err)
	}

	// 将主机替换为 Docker 内部服务名
	parsedURL.Host = "kratos:4433"
	return parsedURL.String(), nil
}

// fetchFlowError 获取流程中的错误信息
func (ka *KratosAdapter) fetchFlowError(ctx context.Context, client *http.Client, path, flowID string) *KratosErrorPayload {
	flowURL := fmt.Sprintf("%s%s?id=%s", ka.kratosPublicURL, path, flowID)

	req, err := http.NewRequestWithContext(ctx, "GET", flowURL, nil)
	if err != nil {
		return nil
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var payload KratosErrorPayload
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil
	}

	if len(payload.UI.Messages) > 0 {
		return &payload
	}

	return nil
}
