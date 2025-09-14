// cmd/login-shim/main.go
package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"fmt"
)

// Kratos 的 Public API 地址
const kratosPublicURL = "http://kratos:4433"

func main() {
	http.HandleFunc("/auth/login", handleLogin)
	log.Println("Login API Shim is listening on :8090")
	http.ListenAndServe(":8090", nil)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	log.Println("1")
	// 步骤 0: 检查请求方法并解析前端发送的凭据
	if r.Method != http.MethodPost {
		http.Error(w, "只允许 POST 方法", http.StatusMethodNotAllowed)
		return
	}

	var creds struct {
		Identifier string `json:"identifier"`
		Password   string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		log.Printf("ERROR: 步骤 0 - 解析请求体失败: %v", err)
		http.Error(w, "无效的请求体", http.StatusBadRequest)
		return
	}

	// 步骤 1: 创建一个特殊的 HTTP Client
	// - 使用 Cookie Jar 自动管理 Kratos 的 CSRF Cookie
	// - 禁用自动重定向，以便我们能手动处理 303 响应
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// --- 在后端完整模拟 Kratos 的浏览器登录流程 ---

	// 步骤 2: 初始化登录流程，获取 Flow ID
	log.Println("DEBUG: 步骤 2 - 开始初始化 Kratos 登录流程...")
	initResp, err := client.Get(kratosPublicURL + "/self-service/login/browser")
	if err != nil {
		log.Printf("ERROR: 步骤 2 - 初始化 Kratos 登录流程失败: %v", err)
		http.Error(w, "初始化 Kratos 登录流程失败", http.StatusInternalServerError)
		return
	}
	defer initResp.Body.Close()

	location, err := initResp.Location()
	if err != nil {
		log.Printf("ERROR: 步骤 2 - Kratos 未返回 location 头: %v", err)
		http.Error(w, "Kratos 未返回 location 头", http.StatusInternalServerError)
		return
	}
	flowID := location.Query().Get("flow")
	if flowID == "" {
		log.Println("ERROR: 步骤 2 - 未能从 Kratos 获取 flow ID")
		http.Error(w, "未能从 Kratos 获取 flow ID", http.StatusInternalServerError)
		return
	}
	log.Printf("DEBUG: 步骤 2 - 成功获取 Flow ID: %s", flowID)

	// 步骤 3: 使用 Flow ID 获取登录表单的详细信息 (包括 Action URL 和 CSRF Token)
	log.Println("DEBUG: 步骤 3 - 开始获取 Flow 详细信息...")
	flowResp, err := client.Get(kratosPublicURL + "/self-service/login/flows?id=" + flowID)
	if err != nil {
		log.Printf("ERROR: 步骤 3 - 从 Kratos 获取 Flow 详细信息失败: %v", err)
		http.Error(w, "从 Kratos 获取 Flow 详细信息失败", http.StatusInternalServerError)
		return
	}
	defer flowResp.Body.Close()

	if flowResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(flowResp.Body)
		log.Printf("ERROR: 步骤 3 - Kratos 获取 Flow 详情返回非 200 状态: %s", string(body))
		http.Error(w, "Kratos 获取 Flow 详情返回非 200 状态", http.StatusInternalServerError)
		return
	}

	var flowData map[string]interface{}
	if err := json.NewDecoder(flowResp.Body).Decode(&flowData); err != nil {
		log.Printf("ERROR: 步骤 3 - 解析 Kratos Flow JSON 失败: %v", err)
		http.Error(w, "解析 Kratos Flow JSON 失败", http.StatusInternalServerError)
		return
	}

	// [安全解析] 提取 Action URL
	uiData, ok := flowData["ui"].(map[string]interface{})
	if !ok {
		log.Printf("ERROR: 步骤 3 - Kratos 响应中缺少 'ui' 字段")
		http.Error(w, "Kratos 响应中缺少 'ui' 字段", http.StatusInternalServerError)
		return
	}
	actionURL, ok := uiData["action"].(string)
	if !ok || actionURL == "" {
		log.Printf("ERROR: 步骤 3 - Kratos 响应中缺少 'action' 字段")
		http.Error(w, "Kratos 响应中缺少 'action' 字段", http.StatusInternalServerError)
		return
	}

	// [安全解析] 提取 CSRF Token
	var csrfToken string
	nodes, ok := uiData["nodes"].([]interface{})
	if ok {
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
		log.Println("ERROR: 步骤 3 - 未能在 Kratos 响应中找到 csrf_token")
		http.Error(w, "未能在 Kratos 响应中找到 csrf_token", http.StatusInternalServerError)
		return
	}
	log.Println("DEBUG: 步骤 3 - 成功提取 Action URL 和 CSRF Token")

	// 步骤 4: [关键修正] 替换 Action URL 的主机名，以适应 Docker 内部网络
	parsedActionURL, err := url.Parse(actionURL)
	if err != nil {
		log.Printf("ERROR: 步骤 4 - 解析 Action URL '%s' 失败: %v", actionURL, err)
		http.Error(w, "解析 Action URL 失败", http.StatusInternalServerError)
		return
	}
	parsedActionURL.Host = "kratos:4433" // 强制使用 Docker 服务名
	finalActionURL := parsedActionURL.String()
	log.Printf("DEBUG: 步骤 4 - 最终提交表单的 URL: %s", finalActionURL)

	// 步骤 5: 构造并提交登录表单给 Kratos
	log.Println("DEBUG: 步骤 5 - 开始向 Kratos 提交登录表单...")
	formData := url.Values{
		"method":     {"password"},
		"identifier": {creds.Identifier},
		"password":   {creds.Password},
		"csrf_token": {csrfToken},
	}
	loginResp, err := client.PostForm(finalActionURL, formData)
	if err != nil {
		log.Printf("ERROR: 步骤 5 - 提交登录表单到 Kratos 失败: %v", err)
		http.Error(w, "提交登录表单到 Kratos 失败", http.StatusInternalServerError)
		return
	}
	defer loginResp.Body.Close()

	// 步骤 6: 处理 Kratos 的最终响应
	log.Printf("DEBUG: 步骤 6 - Kratos 返回最终响应，状态码: %d", loginResp.StatusCode)
	// 如果登录失败 (例如密码错误), Kratos 会返回 >= 400 的状态码和错误详情
	if loginResp.StatusCode >= 400 {
		log.Println("INFO: 步骤 6 - 登录失败，将 Kratos 错误转发给前端")
		w.WriteHeader(loginResp.StatusCode)
		io.Copy(w, loginResp.Body) // 将 Kratos 的错误信息直接透传给前端
		return
	}

	// 如果登录成功, 从 Kratos 的响应中提取 Set-Cookie 头并转发给前端浏览器
	sessionCookie := loginResp.Header.Get("Set-Cookie")
	if sessionCookie == "" {
		log.Println("ERROR: 步骤 6 - 登录成功但 Kratos 未返回 session cookie")
		http.Error(w, "Kratos 未返回 session cookie", http.StatusInternalServerError)
		return
	}

	log.Println("INFO: 步骤 6 - 登录成功！转发 session cookie 给前端")
	w.Header().Set("Set-Cookie", sessionCookie)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, `{"status": "login successful"}`)
}