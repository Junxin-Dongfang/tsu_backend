// File: cmd/login-shim/internal/adapter/kratos_adapter.go
package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"tsu-self/cmd/login-shim/internal/domain"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/xerrors"
)

// KratosAdapter 封装了与 ORY Kratos 交互的所有逻辑。
// 它是 login-api-shim 服务与 Kratos 之间的唯一“翻译官”。
type KratosAdapter struct {
	kratosPublicURL string
}

// KratosErrorPayload 定义了 Kratos 登录/注册失败时返回的 JSON 结构。
// 我们只关心我们需要翻译的字段。
type KratosErrorPayload struct {
	UI struct {
		Messages []struct {
			ID   int    `json:"id"`
			Text string `json:"text"`
		} `json:"messages"`
	} `json:"ui"`
}

// NewKratosAdapter 是 KratosAdapter 的构造函数。
func NewKratosAdapter(kratosPublicURL string) *KratosAdapter {
	return &KratosAdapter{
		kratosPublicURL: kratosPublicURL,
	}
}

// Login 方法在后端完整地模拟了 Kratos 的浏览器登录流程。
// 它接收前端传来的凭证，返回 Kratos 下发的 session cookie 或一个自定义的 AppError。
func (ka *KratosAdapter) Login(ctx context.Context, req *domain.LoginRequest) (string, *KratosErrorPayload, *xerrors.AppError) {
	log.InfoWithCtx(ctx, "KratosAdapter: 开始登录流程")

	// 1. 为本次流程创建一个全新的、独立的 HTTP Client
	client := ka.createClient()

	// 2. 初始化登录流程
	flowID, err := ka.initializeFlow(ctx, client, "/self-service/login/browser")
	if err != nil {
		// 注意：这里返回的是一个通用的内部错误，因为这是基础设施问题，不是用户凭证问题
		return "", nil, xerrors.E_INTERNAL(fmt.Errorf("初始化 Kratos 登录流程失败: %w", err))
	}
	log.DebugWithCtx(ctx, "KratosAdapter: 成功获取 Flow ID", "flow_id", flowID)

	// 3. 获取 Flow 详情，包括 CSRF Token 和 Action URL
	actionURL, csrfToken, err := ka.getFlowDetails(ctx, client, "/self-service/login/flows", flowID)
	if err != nil {
		return "", nil, xerrors.E_INTERNAL(fmt.Errorf("获取 Kratos 流程详情失败: %w", err))
	}
	log.DebugWithCtx(ctx, "KratosAdapter: 成功获取 Action URL 和 CSRF Token")

	// 4. 修正 Action URL 以适应 Docker 内部网络
	finalActionURL, err := ka.rewriteActionURL(actionURL)
	if err != nil {
		return "", nil, xerrors.E_INTERNAL(fmt.Errorf("修正 Kratos Action URL 失败: %w", err))
	}

	// 5. 构造并提交登录表单
	formData := url.Values{
		"method":     {"password"},
		"identifier": {req.Identifier},
		"password":   {req.Password},
		"csrf_token": {csrfToken},
	}
	log.DebugWithCtx(ctx, "KratosAdapter: 准备提交登录表单", "url", finalActionURL)
	loginResp, err := client.PostForm(finalActionURL, formData)
	if err != nil {
		log.ErrorWithCtx(ctx, "提交登录表单到 Kratos 时发生网络错误", err)
		return "", nil, xerrors.E_INTERNAL(fmt.Errorf("提交登录表单到 Kratos 失败: %w", err))
	}
	defer loginResp.Body.Close()

	// 6. 处理 Kratos 的最终响应
	// 成功的唯一标志是 Kratos 在响应头中下发了 session cookie
	if sessionCookie := loginResp.Header.Get("Set-Cookie"); sessionCookie != "" {
		log.InfoWithCtx(ctx, "KratosAdapter: 登录成功，已获取 session cookie")
		return sessionCookie, nil, nil
	}

	// 如果没有 session cookie，说明登录失败。我们需要解析错误原因。
	log.WarnWithCtx(ctx, "KratosAdapter: Kratos 未返回 session cookie，判断为登录失败", "status_code", loginResp.StatusCode)

	// 尝试解析 Kratos 返回的错误详情 JSON
	var kratosErr KratosErrorPayload
	body, _ := io.ReadAll(loginResp.Body)
	if json.Unmarshal(body, &kratosErr) == nil && len(kratosErr.UI.Messages) > 0 {
		return "", &kratosErr, nil
	}

	// 如果无法解析出具体的错误 ID，根据状态码返回合适的错误
	// 对于登录流程，302/303 重定向通常表示凭证验证失败
	if loginResp.StatusCode == 302 || loginResp.StatusCode == 303 {
		// 回到 flow 查询 ui.messages，以便拿到可映射的错误 ID
		if errPayload := ka.fetchFlowError(ctx, client, "/self-service/login/flows", flowID); errPayload != nil {
			return "", errPayload, nil
		}
		appErr := xerrors.E_INVALID_CREDENTIALS(fmt.Errorf("kratos returned status %d", loginResp.StatusCode))
		return "", nil, appErr
	}

	// 其他状态码返回通用内部错误
	appErr := xerrors.E_INTERNAL(fmt.Errorf("kratos returned unexpected status %d", loginResp.StatusCode))
	return "", nil, appErr
}

// Register 方法的占位符，您可以仿照 Login 的逻辑来实现
func (ka *KratosAdapter) Register(ctx context.Context, req *domain.RegisterRequest) (string, *KratosErrorPayload, *xerrors.AppError) {
	log.InfoWithCtx(ctx, "KratosAdapter: 开始注册流程")
	// 1. 创建 Client
	client := ka.createClient()

	// 2. 初始化注册流程: /self-service/registration/browser
	flowID, err := ka.initializeFlow(ctx, client, "/self-service/registration/browser")
	if err != nil {
		return "", nil, xerrors.E_INTERNAL(fmt.Errorf("初始化 Kratos 注册流程失败: %w", err))
	}
	log.DebugWithCtx(ctx, "KratosAdapter: 成功获取注册 Flow ID", "flow_id", flowID)
	// 3. 获取注册 Flow 详情: /self-service/registration/flows
	actionURL, csrfToken, err := ka.getFlowDetails(ctx, client, "/self-service/registration/flows", flowID)
	if err != nil {
		return "", nil, xerrors.E_INTERNAL(fmt.Errorf("获取 Kratos 注册流程详情失败: %w", err))
	}
	log.DebugWithCtx(ctx, "KratosAdapter: 成功获取注册 Action URL 和 CSRF Token")
	// 4. 提交注册表单
	finalActionURL, err := ka.rewriteActionURL(actionURL)
	if err != nil {
		return "", nil, xerrors.E_INTERNAL(fmt.Errorf("修正 Kratos 注册 Action URL 失败: %w", err))
	}

	formData := url.Values{
		"method":           {"password"},
		"traits.email":     {req.Email},
		"traits.username":  {req.UserName},
		"password":         {req.Password},
		"password_confirm": {req.Password}, // 通常需要确认密码
		"csrf_token":       {csrfToken},
	}
	log.DebugWithCtx(ctx, "KratosAdapter: 准备提交注册表单", "url", finalActionURL)
	registerResp, err := client.PostForm(finalActionURL, formData)
	if err != nil {
		log.ErrorWithCtx(ctx, "提交注册表单到 Kratos 时发生网络错误", err)
		return "", nil, xerrors.E_INTERNAL(fmt.Errorf("提交注册表单到 Kratos 失败: %w", err))
	}
	// 5. 处理最终响应
	defer registerResp.Body.Close()
	if sessionCookie := registerResp.Header.Get("Set-Cookie"); sessionCookie != "" {
		log.InfoWithCtx(ctx, "KratosAdapter: 注册成功，已获取 session cookie")
		return sessionCookie, nil, nil
	}
	log.WarnWithCtx(ctx, "KratosAdapter: 注册失败，未获取到 session cookie", "status_code", registerResp.StatusCode)

	// 尝试解析 Kratos 返回的错误详情 JSON
	var kratosErr KratosErrorPayload
	body, _ := io.ReadAll(registerResp.Body)
	if json.Unmarshal(body, &kratosErr) == nil && len(kratosErr.UI.Messages) > 0 {
		return "", &kratosErr, nil
	}

	// 如果无法解析出具体的错误 ID，根据状态码和场景返回合适的错误
	// 对于注册流程，需要区分不同的错误类型
	if registerResp.StatusCode == 302 || registerResp.StatusCode == 303 {
		// 回到 flow 查询 ui.messages，以便拿到可映射的错误 ID
		if errPayload := ka.fetchFlowError(ctx, client, "/self-service/registration/flows", flowID); errPayload != nil {
			return "", errPayload, nil
		}
		// 若仍无法解析，返回通用的校验错误
		appErr := xerrors.New(xerrors.Validation, "注册信息不符合要求，请检查后重试",
			fmt.Errorf("kratos returned status %d", registerResp.StatusCode))
		return "", nil, appErr
	}

	// 其他状态码返回通用内部错误
	appErr := xerrors.E_INTERNAL(fmt.Errorf("kratos returned unexpected status %d", registerResp.StatusCode))
	return "", nil, appErr
}

// --- 以下是内部辅助函数 ---

// createClient 创建一个适用于与 Kratos 流程交互的 HTTP Client
func (ka *KratosAdapter) createClient() *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // 核心：禁止自动重定向
		},
	}
}

// initializeFlow 向 Kratos 发起请求以开始一个新的流程 (登录/注册等)
func (ka *KratosAdapter) initializeFlow(ctx context.Context, client *http.Client, path string) (string, error) {
	initURL := ka.kratosPublicURL + path
	log.DebugWithCtx(ctx, "KratosAdapter: 初始化流程", "url", initURL)
	initResp, err := client.Get(initURL)
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
		return "", fmt.Errorf("未能从 Kratos 的重定向 URL 中获取 flow ID")
	}
	return flowID, nil
}

// getFlowDetails 使用 flow ID 获取流程的详细信息，包括 CSRF Token 和 Action URL
func (ka *KratosAdapter) getFlowDetails(ctx context.Context, client *http.Client, path, flowID string) (actionURL, csrfToken string, err error) {
	flowURL := fmt.Sprintf("%s%s?id=%s", ka.kratosPublicURL, path, flowID)
	log.DebugWithCtx(ctx, "KratosAdapter: 获取流程详情", "url", flowURL)
	flowResp, err := client.Get(flowURL)
	if err != nil {
		return "", "", err
	}
	defer flowResp.Body.Close()

	if flowResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(flowResp.Body)
		return "", "", fmt.Errorf("kratos 获取 Flow 详情返回非 200 状态: %s", string(body))
	}

	var flowData map[string]interface{}
	if err := json.NewDecoder(flowResp.Body).Decode(&flowData); err != nil {
		return "", "", fmt.Errorf("解析 Kratos Flow JSON 失败: %w", err)
	}

	uiData, ok := flowData["ui"].(map[string]interface{})
	if !ok {
		return "", "", fmt.Errorf("kratos 响应中缺少 'ui' 字段")
	}

	actionURL, ok = uiData["action"].(string)
	if !ok || actionURL == "" {
		return "", "", fmt.Errorf("kratos 响应中缺少或空的 'action' 字段")
	}

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
		return "", "", fmt.Errorf("未能在 Kratos 响应中找到 csrf_token")
	}

	return actionURL, csrfToken, nil
}

// rewriteActionURL 将 Kratos 返回的 URL (通常是 localhost 或 127.0.0.1) 替换为 Docker 内部网络可达的地址
func (ka *KratosAdapter) rewriteActionURL(originalURL string) (string, error) {
	parsedURL, err := url.Parse(originalURL)
	if err != nil {
		return "", fmt.Errorf("解析 Action URL '%s' 失败: %w", originalURL, err)
	}
	// 将 URL 的主机部分强制替换为 Docker 内部网络的服务名
	parsedURL.Host = "kratos:4433"
	return parsedURL.String(), nil
}

// fetchFlowError 回到指定 flow 查询 ui.messages，从而获得可翻译的 Kratos 错误 ID
func (ka *KratosAdapter) fetchFlowError(ctx context.Context, client *http.Client, path, flowID string) *KratosErrorPayload {
	flowURL := fmt.Sprintf("%s%s?id=%s", ka.kratosPublicURL, path, flowID)
	log.DebugWithCtx(ctx, "KratosAdapter: 回查 Flow 错误消息", "url", flowURL)
	resp, err := client.Get(flowURL)
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
