package controller

import (
	"encoding/json"
	"net/http"
	"net/mail"
	"regexp"
	"tsu-self/internal/app/login-shim/adapter"
	"tsu-self/internal/app/login-shim/domain"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/response" // 引入项目级 response 包
	"tsu-self/internal/pkg/xerrors"  // 引入项目级 errors 包
)

// AuthController 持有其所依赖的服务，这里是 KratosAdapter
type AuthController struct {
	kratosAdapter *adapter.KratosAdapter
}

// NewAuthController 是 AuthController 的构造函数，用于依赖注入
func NewAuthController(ka *adapter.KratosAdapter) *AuthController {
	return &AuthController{
		kratosAdapter: ka,
	}
}

// Login 是处理登录请求的 HTTP Handler
func (ac *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.InfoWithCtx(ctx, "AuthController: 接收到登录请求")

	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.ErrorWithCtx(ctx, "解析登录请求体失败", err)
		resp := response.Error[any](xerrors.Validation, "无效的请求体", err.Error())
		response.JSON(w, http.StatusBadRequest, resp)
		return
	}

	if req.Identifier == "" || req.Password == "" {
		resp := response.Error[any](xerrors.Validation, "无效的请求体", "identifier/password 不能为空")
		response.JSON(w, http.StatusBadRequest, resp)
		return
	}

	// 调用 adapter 层来执行核心的登录逻辑
	sessionCookie, kratosErr, appErr := ac.kratosAdapter.Login(ctx, &req)
	if appErr != nil {
		log.WarnWithCtx(ctx, "登录流程失败", "error_code", appErr.Code, "error_message", appErr.Message)
		statusCode := ac.getHTTPStatusFromAppError(appErr.Code)
		resp := response.Error[any](appErr.Code, appErr.Message, appErr.Error())
		response.JSON(w, statusCode, resp)
		return
	}

	if kratosErr != nil {
		firstErrorID := kratosErr.UI.Messages[0].ID
		appCode, appMessage := ac.pickBestKratosError(kratosErr)
		log.WarnWithCtx(ctx, "Kratos 登录验证失败", "kratos_id", firstErrorID, "app_code", appCode, "app_message", appMessage)
		statusCode := ac.getHTTPStatusFromAppError(appCode)
		resp := response.Error[any](appCode, appMessage, kratosErr.UI.Messages[0].Text)
		response.JSON(w, statusCode, resp)
		return
	}

	w.Header().Set("Set-Cookie", sessionCookie)
	resp := response.Success(&response.EmptyData{})
	response.JSON(w, http.StatusOK, resp)
	log.InfoWithCtx(ctx, "AuthController: 登录请求处理成功")
}

// Register 是处理注册请求的 HTTP Handler
func (ac *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.InfoWithCtx(ctx, "AuthController: 接收到注册请求")

	var req domain.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.ErrorWithCtx(ctx, "解析注册请求体失败", err)
		resp := response.Error[any](xerrors.Validation, "无效的请求体", err.Error())
		response.JSON(w, http.StatusBadRequest, resp)
		return
	}

	// 输入预校验：邮箱、密码、用户名
	if _, err := mail.ParseAddress(req.Email); err != nil {
		resp := response.Error[any](xerrors.EmailFormatError, "电子邮箱地址格式不正确", err.Error())
		response.JSON(w, http.StatusBadRequest, resp)
		return
	}
	if len(req.Password) < 8 {
		resp := response.Error[any](xerrors.PasswordTooShort, "密码长度不足，请设置更长的密码", "password 长度需≥8")
		response.JSON(w, http.StatusBadRequest, resp)
		return
	}
	if req.UserName == "" {
		resp := response.Error[any](xerrors.Validation, "无效的请求体", "username 不能为空")
		response.JSON(w, http.StatusBadRequest, resp)
		return
	}
	if ok, _ := regexp.MatchString("^[a-zA-Z0-9_]{3,30}$", req.UserName); !ok {
		resp := response.Error[any](xerrors.Validation, "用户名格式不正确", "username 需匹配 ^[a-zA-Z0-9_]{3,30}$")
		response.JSON(w, http.StatusBadRequest, resp)
		return
	}

	sessionCookie, kratosErr, appErr := ac.kratosAdapter.Register(ctx, &req)
	if appErr != nil {
		log.WarnWithCtx(ctx, "注册流程失败", "error_code", appErr.Code, "error_message", appErr.Message)
		statusCode := ac.getHTTPStatusFromAppError(appErr.Code)
		resp := response.Error[any](appErr.Code, appErr.Message, appErr.Error())
		response.JSON(w, statusCode, resp)
		return
	}

	if kratosErr != nil {
		firstErrorID := kratosErr.UI.Messages[0].ID
		appCode, appMessage := ac.pickBestKratosError(kratosErr)
		log.WarnWithCtx(ctx, "Kratos 注册验证失败", "kratos_id", firstErrorID, "app_code", appCode, "app_message", appMessage)
		statusCode := ac.getHTTPStatusFromAppError(appCode)
		resp := response.Error[any](appCode, appMessage, kratosErr.UI.Messages[0].Text)
		response.JSON(w, statusCode, resp)
		return
	}

	w.Header().Set("Set-Cookie", sessionCookie)
	resp := response.Success(&response.EmptyData{})
	response.JSON(w, http.StatusOK, resp)
	log.InfoWithCtx(ctx, "AuthController: 注册请求处理成功")
}

// getHTTPStatusFromAppError 根据应用错误码返回合适的HTTP状态码
func (ac *AuthController) getHTTPStatusFromAppError(appCode int) int {
	switch appCode {
	case xerrors.InvalidCredentials:
		return http.StatusUnauthorized
	case xerrors.IdentityAlreadyExists:
		return http.StatusConflict
	case xerrors.PasswordPolicyError, xerrors.PasswordTooShort:
		return http.StatusBadRequest
	case xerrors.EmailFormatError:
		return http.StatusBadRequest
	case xerrors.Validation:
		return http.StatusBadRequest
	case xerrors.SessionAlreadyAvailable:
		return http.StatusConflict
	case xerrors.SessionNotFound:
		return http.StatusUnauthorized
	case xerrors.NotFound:
		return http.StatusNotFound
	case xerrors.PermissionDenied:
		return http.StatusForbidden
	case xerrors.RateLimitExceeded:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

// pickBestKratosError 根据 Kratos 返回的全部 UI 消息挑选最合适的业务错误码与文案
func (ac *AuthController) pickBestKratosError(kratosErr *adapter.KratosErrorPayload) (int, string) {
	if kratosErr == nil || len(kratosErr.UI.Messages) == 0 {
		return xerrors.Validation, "提交的信息有误，请检查后重试"
	}

	priority := map[int]int{
		xerrors.IdentityAlreadyExists: 1,
		xerrors.InvalidCredentials:    2,
		xerrors.PasswordTooShort:      3,
		xerrors.PasswordPolicyError:   4,
		xerrors.EmailFormatError:      5,
		xerrors.Validation:            9,
	}

	bestCode := xerrors.Validation
	bestMsg := "提交的信息有误，请检查后重试"
	bestRank := 999

	for _, m := range kratosErr.UI.Messages {
		// 1) 先用 ID 试译
		codeID, msgID := xerrors.TranslateKratosError(m.ID)
		rankID, ok := priority[codeID]
		if !ok {
			rankID = 50
		}
		if rankID < bestRank {
			bestRank = rankID
			bestCode = codeID
			bestMsg = msgID
		}
		// 2) 再用文本兜底试译
		codeTXT, msgTXT := xerrors.TranslateKratosErrorText(m.Text)
		rankTXT, ok := priority[codeTXT]
		if !ok {
			rankTXT = 50
		}
		if rankTXT < bestRank {
			bestRank = rankTXT
			bestCode = codeTXT
			bestMsg = msgTXT
		}
	}

	return bestCode, bestMsg
}
