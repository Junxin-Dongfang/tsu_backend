// File: internal/app/login-shim/controller/auth_controller.go (完整更新版)
package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"regexp"

	"tsu-self/internal/app/login-shim/adapter"
	"tsu-self/internal/app/login-shim/domain"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/xerrors"
)

// AuthController 认证控制器
type AuthController struct {
	kratosAdapter *adapter.KratosAdapter
}

// NewAuthController 创建认证控制器实例
func NewAuthController(ka *adapter.KratosAdapter) *AuthController {
	return &AuthController{
		kratosAdapter: ka,
	}
}

// Login 处理登录请求
func (ac *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.InfoWithCtx(ctx, "收到登录请求")

	// 解析请求
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.ErrorWithCtx(ctx, "解析登录请求失败", err)
		appErr := xerrors.NewValidationError("请求格式错误")
		ac.addTraceToError(r, appErr)
		response.Error[response.EmptyData](w, r, appErr)
		return
	}

	// 参数验证
	if appErr := ac.validateLoginRequest(&req); appErr != nil {
		ac.addTraceToError(r, appErr)
		response.Error[response.EmptyData](w, r, appErr)
		return
	}

	// 执行登录
	sessionCookie, kratosErr, appErr := ac.kratosAdapter.Login(ctx, &req)
	if appErr != nil {
		log.WarnWithCtx(ctx, "登录失败", "error_code", appErr.Code, "error_msg", appErr.Message)
		ac.addTraceToError(r, appErr)
		response.Error[response.EmptyData](w, r, appErr)
		return
	}

	if kratosErr != nil {
		// 转换 Kratos 错误为业务错误
		appErr := ac.convertKratosError(kratosErr)
		ac.addTraceToError(r, appErr)
		log.WarnWithCtx(ctx, "Kratos 登录验证失败",
			"kratos_messages_count", len(kratosErr.UI.Messages),
			"converted_code", appErr.Code,
			"converted_msg", appErr.Message)
		response.Error[response.EmptyData](w, r, appErr)
		return
	}

	// 登录成功
	w.Header().Set("Set-Cookie", sessionCookie)
	response.OK(w, r, &response.EmptyData{})
	log.InfoWithCtx(ctx, "登录成功")
}

// Register 处理注册请求
func (ac *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.InfoWithCtx(ctx, "收到注册请求")

	// 解析请求
	var req domain.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.ErrorWithCtx(ctx, "解析注册请求失败", err)
		appErr := xerrors.NewValidationError("请求格式错误")
		ac.addTraceToError(r, appErr)
		response.Error[response.EmptyData](w, r, appErr)
		return
	}

	// 参数验证
	if appErr := ac.validateRegisterRequest(&req); appErr != nil {
		ac.addTraceToError(r, appErr)
		response.Error[response.EmptyData](w, r, appErr)
		return
	}

	// 执行注册
	sessionCookie, kratosErr, appErr := ac.kratosAdapter.Register(ctx, &req)
	if appErr != nil {
		log.WarnWithCtx(ctx, "注册失败", "error_code", appErr.Code, "error_msg", appErr.Message)
		ac.addTraceToError(r, appErr)
		response.Error[response.EmptyData](w, r, appErr)
		return
	}

	if kratosErr != nil {
		// 转换 Kratos 错误为业务错误
		appErr := ac.convertKratosError(kratosErr)
		ac.addTraceToError(r, appErr)
		log.WarnWithCtx(ctx, "Kratos 注册验证失败",
			"kratos_messages_count", len(kratosErr.UI.Messages),
			"converted_code", appErr.Code,
			"converted_msg", appErr.Message)
		response.Error[response.EmptyData](w, r, appErr)
		return
	}

	// 注册成功
	w.Header().Set("Set-Cookie", sessionCookie)
	response.OK(w, r, &response.EmptyData{})
	log.InfoWithCtx(ctx, "注册成功")
}

// --- 验证方法 ---

// validateLoginRequest 验证登录请求参数
func (ac *AuthController) validateLoginRequest(req *domain.LoginRequest) *xerrors.AppError {
	if req.Identifier == "" {
		return xerrors.NewValidationError("用户名或邮箱不能为空")
	}
	if req.Password == "" {
		return xerrors.NewValidationError("密码不能为空")
	}
	if len(req.Password) < 6 { // 降低前端验证要求，让 Kratos 处理具体策略
		return xerrors.New(xerrors.CodePasswordTooShort, "密码太短").
			WithUserMessage("密码长度不够")
	}
	if len(req.Password) > 128 {
		return xerrors.New(xerrors.CodePasswordTooLong, "密码太长").
			WithUserMessage("密码长度过长")
	}

	return nil
}

// convertKratosError 将 Kratos 错误转换为业务错误
func (ac *AuthController) convertKratosError(kratosErr *adapter.KratosErrorPayload) *xerrors.AppError {
	if kratosErr == nil || len(kratosErr.UI.Messages) == 0 {
		return xerrors.NewValidationError("请求验证失败")
	}

	// 收集所有可能的错误转换结果
	type ErrorCandidate struct {
		Code     int
		Message  string
		Priority int
		Source   string // "id" 或 "text"
	}

	var candidates []ErrorCandidate

	// 遍历所有错误消息
	for i, msg := range kratosErr.UI.Messages {
		// 1. 尝试通过 ID 转换
		if msg.ID != 0 {
			code, message := xerrors.TranslateKratosError(msg.ID)
			priority := xerrors.GetKratosErrorPriority(code)
			candidates = append(candidates, ErrorCandidate{
				Code:     code,
				Message:  message,
				Priority: priority,
				Source:   "id",
			})
		}

		// 2. 尝试通过文本转换
		if msg.Text != "" {
			code, message := xerrors.TranslateKratosErrorText(msg.Text)
			priority := xerrors.GetKratosErrorPriority(code)
			candidates = append(candidates, ErrorCandidate{
				Code:     code,
				Message:  message,
				Priority: priority,
				Source:   "text",
			})
		}

		// 记录原始 Kratos 错误信息用于调试
		log.DebugWithCtx(nil, "Kratos 错误详情",
			"index", i,
			"id", msg.ID,
			"text", msg.Text,
			"type", msg.Type)
	}

	// 选择优先级最高（数字最小）的错误
	if len(candidates) == 0 {
		return xerrors.NewValidationError("请求验证失败")
	}

	bestCandidate := candidates[0]
	for _, candidate := range candidates[1:] {
		if candidate.Priority < bestCandidate.Priority {
			bestCandidate = candidate
		}
	}

	// 创建 AppError，包含完整上下文信息
	appErr := xerrors.New(bestCandidate.Code, bestCandidate.Message).
		WithUserMessage(bestCandidate.Message).
		WithContext(&xerrors.ErrorContext{
			Service: "auth-controller",
			Method:  "convertKratosError",
			Metadata: map[string]string{
				"source":         bestCandidate.Source,
				"messages_count": fmt.Sprintf("%d", len(kratosErr.UI.Messages)),
			},
		})

	// 设置重试标识
	if xerrors.IsRetryableKratosError(kratosErr.UI.Messages[0].ID) {
		appErr.WithRetryable(true)
	}

	return appErr
}

// addTraceToError 为错误添加追踪信息
func (ac *AuthController) addTraceToError(r *http.Request, appErr *xerrors.AppError) {
	if appErr.Context == nil {
		appErr.WithContext(&xerrors.ErrorContext{})
	}

	// 添加 TraceID
	if traceID, ok := r.Context().Value("trace_id").(string); ok {
		appErr.Context.TraceID = traceID
	}

	// 添加 RequestID（如果有的话）
	if requestID, ok := r.Context().Value("request_id").(string); ok {
		appErr.Context.RequestID = requestID
	}

	// 添加用户信息（如果有的话）
	if userID, ok := r.Context().Value("user_id").(string); ok {
		appErr.Context.UserID = userID
	}
}

// --- 健康检查和辅助端点 ---

// HealthCheck 健康检查端点
func (ac *AuthController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	type healthCheckResponse struct {
		Status  string `json:"status"`
		Service string `json:"service"`
	}
	response.OK(w, r, &healthCheckResponse{
		Status:  "healthy",
		Service: "login-shim",
	})
}

// validateRegisterRequest 验证注册请求参数
func (ac *AuthController) validateRegisterRequest(req *domain.RegisterRequest) *xerrors.AppError {
	// 邮箱验证
	if req.Email == "" {
		return xerrors.NewValidationError("邮箱不能为空")
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return xerrors.New(xerrors.CodeInvalidParams, "邮箱格式错误").
			WithUserMessage("请输入有效的邮箱地址")
	}

	// 用户名验证
	if req.UserName == "" {
		return xerrors.NewValidationError("用户名不能为空")
	}
	if len(req.UserName) < 3 || len(req.UserName) > 30 {
		return xerrors.NewValidationError("用户名长度必须在3-30个字符之间")
	}
	if matched, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", req.UserName); !matched {
		return xerrors.NewValidationError("用户名只能包含字母、数字和下划线")
	}

	// 密码验证（基础验证，详细策略由 Kratos 处理）
	if req.Password == "" {
		return xerrors.NewValidationError("密码不能为空")
	}
	if len(req.Password) < 6 {
		return xerrors.New(xerrors.CodePasswordTooShort, "密码太短").
			WithUserMessage("密码长度不够")
	}
	if len(req.Password) > 128 {
		return xerrors.New(xerrors.CodePasswordTooLong, "密码太长").WithUserMessage("密码长度过长")
	}
	return nil
}
