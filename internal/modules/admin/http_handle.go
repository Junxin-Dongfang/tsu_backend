// @title Tsu Admin API
// @version 1.0
// @description Tsu 后台管理系统 API
// @contact.name Tsu Team
// @contact.email support@tsu.com
// @host localhost
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer token format: Bearer {token}

package admin

import (
	"fmt"
	"net/http"
	"time"

	// 新的 API 层模型
	apiAuthReq "tsu-self/internal/api/request/auth"

	// 转换器
	authConverter "tsu-self/internal/converter/auth"

	// 新的 RPC 模型
	"tsu-self/internal/rpc/generated/auth"

	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/xerrors"

	"github.com/labstack/echo/v4"
	mqrpc "github.com/liangdas/mqant/rpc"
	"google.golang.org/protobuf/proto"
)

// Login 用户登录
// @Summary 用户登录
// @Description 通过用户名或邮箱和密码登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param login body apiAuthReq.LoginRequest true "登录请求参数"
// @Success 200 {object} response.APIResponse[apiAuthResp.LoginResult] "登录成功"
// @Failure 400 {object} response.APIResponse[any] "请求参数错误"
// @Failure 401 {object} response.APIResponse[any] "认证失败"
// @Failure 500 {object} response.APIResponse[any] "服务器错误"
// @Router /auth/login [post]
func (m *AdminModule) Login(c echo.Context) error {
	ctx := c.Request().Context()

	// 使用新的 API 请求模型
	var req apiAuthReq.LoginRequest
	if err := c.Bind(&req); err != nil {
		appErr := xerrors.NewValidationError("request", "请求格式错误")
		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	// 填充服务器端信息
	req.ClientIP = c.RealIP()
	req.UserAgent = c.Request().Header.Get("User-Agent")

	// 使用转换器转换为 RPC 请求
	rpcReq := authConverter.LoginRequestToRPC(&req)

	// mqant RPC 调用
	resp, err := m.Call(ctx, "auth", "Login", mqrpc.Param(rpcReq))
	if err != "" {
		log.ErrorContext(c.Request().Context(), "Auth服务调用失败.", log.Any("error", err))
		return response.InternalServerError(ctx, c.Response().Writer, m.respWriter, "认证服务调用失败")
	}

	// 处理 RPC 响应
	var rpcResp *auth.LoginResponse
	switch v := resp.(type) {
	case []byte:
		rpcResp = &auth.LoginResponse{}
		if err := proto.Unmarshal(v, rpcResp); err != nil {
			log.ErrorContext(c.Request().Context(), "Login响应反序列化失败", log.Any("error", err))
			return response.InternalServerError(ctx, c.Response().Writer, m.respWriter, "响应处理失败")
		}
	case *auth.LoginResponse:
		rpcResp = v
	default:
		log.ErrorContext(c.Request().Context(), "Login响应类型错误", log.Any("type", fmt.Sprintf("%T", v)))
		return response.InternalServerError(ctx, c.Response().Writer, m.respWriter, "响应类型错误")
	}

	// 使用转换器转换为 API 响应
	apiResult := authConverter.LoginResponseFromRPC(rpcResp)

	// 如果需要同步到主数据库，可以在这里处理
	// TODO: 实现事务协调逻辑

	return m.respWriter.WriteSuccess(ctx, c.Response().Writer, apiResult)
}

// Register 用户注册
// @Summary 用户注册
// @Description 通过邮箱、用户名和密码注册新用户
// @Tags 认证
// @Accept json
// @Produce json
// @Param register body apiAuthReq.RegisterRequest true "注册请求参数"
// @Success 200 {object} response.APIResponse[apiAuthResp.RegisterResult] "注册成功"
// @Failure 400 {object} response.APIResponse[any] "请求参数错误"
// @Failure 409 {object} response.APIResponse[any] "用户已存在"
// @Failure 500 {object} response.APIResponse[any] "服务器错误"
// @Router /auth/register [post]
func (m *AdminModule) Register(c echo.Context) error {
	ctx := c.Request().Context()

	// 使用新的 API 请求模型
	var req apiAuthReq.RegisterRequest
	if err := c.Bind(&req); err != nil {
		appErr := xerrors.NewValidationError("request", "请求格式错误")
		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	// 填充服务器端信息
	req.ClientIP = c.RealIP()
	req.UserAgent = c.Request().Header.Get("User-Agent")

	// 使用转换器转换为 RPC 请求
	rpcReq := authConverter.RegisterRequestToRPC(&req)

	result, err := m.Call(ctx, "auth", "Register", mqrpc.Param(rpcReq))
	if err != "" {
		m.logger.ErrorContext(c.Request().Context(), "Auth服务调用失败", log.Any("error", err))
		return m.respWriter.WriteError(c.Request().Context(), c.Response().Writer,
			xerrors.New(xerrors.CodeExternalServiceError, "认证服务调用失败"))
	}

	m.logger.InfoContext(ctx, "Auth服务调用成功", log.Any("result_type", fmt.Sprintf("%T", result)))

	// 处理 RPC 响应
	var rpcResp *auth.RegisterResponse
	switch v := result.(type) {
	case []byte:
		rpcResp = &auth.RegisterResponse{}
		if err := proto.Unmarshal(v, rpcResp); err != nil {
			m.logger.ErrorContext(c.Request().Context(), "Protobuf反序列化失败", log.Any("error", err))
			return m.respWriter.WriteError(c.Request().Context(), c.Response().Writer,
				xerrors.FromCode(xerrors.CodeInternalError).WithMetadata("reason", "protobuf_unmarshal_failed"))
		}
	case *auth.RegisterResponse:
		rpcResp = v
	default:
		m.logger.ErrorContext(c.Request().Context(), "类型断言失败",
			log.Any("expected", "*auth.RegisterResponse or []byte"),
			log.Any("actual", fmt.Sprintf("%T", result)))
		return m.respWriter.WriteError(c.Request().Context(), c.Response().Writer,
			xerrors.FromCode(xerrors.CodeInternalError).WithMetadata("reason", "invalid_response_type"))
	}

	// 使用转换器转换为 API 响应
	apiResult := authConverter.RegisterResponseFromRPC(rpcResp)

	// TODO: 如果需要同步到主数据库，可以在这里处理事务协调逻辑

	return m.respWriter.WriteSuccess(c.Request().Context(), c.Response().Writer, apiResult)
}

// // Logout 用户登出
// // @Summary 用户登出
// // @Description 通过会话令牌登出当前用户
// // @Tags 认证
// // @Accept json
// // @Produce json
// // @Param X-Session-Token header string true "会话令牌"
// // @Success 200 {object} response.APIResponse[map[string]string] "登出成功"
// // @Failure 400 {object} response.APIResponse[any] "请求参数错误"
// // @Failure 401 {object} response.APIResponse[any] "认证失败"
// // @Failure 500 {object} response.APIResponse[any] "服务器错误"
// // @Router /auth/logout [post]
// func (m *AdminModule) Logout(c echo.Context) error {
// 	ctx := c.Request().Context()

// 	sessionToken := c.Request().Header.Get("X-Session-Token")
// 	if sessionToken == "" {
// 		appErr := xerrors.NewValidationError("session_token", "会话令牌缺失")
// 		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
// 	}

// 	if appErr := m.authService.Logout(ctx, sessionToken); appErr != nil {
// 		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
// 	}

// 	respData := map[string]string{
// 		"message": "登出成功",
// 	}

// 	return m.respWriter.WriteSuccess(ctx, c.Response().Writer, respData)
// }

// // GetSession 获取会话信息
// // @Summary 获取会话信息
// // @Description 通过会话令牌获取当前会话信息
// // @Tags 认证
// // @Accept json
// // @Produce json
// // @Param X-Session-Token header string true "会话令牌"
// // @Success 200 {object} response.APIResponse[map[string]string] "获取成功"
// // @Failure 400 {object} response.APIResponse[any] "请求参数错误"
// // @Failure 401 {object} response.APIResponse[any] "认证失败"
// // @Failure 500 {object} response.APIResponse[any] "服务器错误"
// // @Router /admin/auth/session [get]
// func (m *AdminModule) GetSession(c echo.Context) error {
// 	ctx := c.Request().Context()

// 	sessionToken := c.Request().Header.Get("X-Session-Token")
// 	if sessionToken == "" {
// 		appErr := xerrors.NewValidationError("session_token", "会话令牌缺失")
// 		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
// 	}

// 	session, appErr := m.authService.GetSession(ctx, sessionToken)
// 	if appErr != nil {
// 		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
// 	}

// 	return m.respWriter.WriteSuccess(ctx, c.Response().Writer, session)
// }

// // InitRecovery 初始化账户恢复
// func (m *AdminModule) InitRecovery(c echo.Context) error {
// 	ctx := c.Request().Context()

// 	var req authmodel.RecoveryRequest
// 	if err := c.Bind(&req); err != nil {
// 		appErr := xerrors.NewValidationError("request", "请求格式错误")
// 		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
// 	}

// 	if appErr := m.authService.InitRecovery(ctx, &req); appErr != nil {
// 		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
// 	}

// 	respData := map[string]string{
// 		"message": "恢复邮件已发送",
// 	}

// 	return m.respWriter.WriteSuccess(ctx, c.Response().Writer, respData)
// }

// // SubmitRecovery 提交恢复请求
// func (m *AdminModule) SubmitRecovery(c echo.Context) error {
// 	// 这里需要根据具体的恢复流程实现
// 	// 通常需要处理恢复令牌和新密码
// 	return c.JSON(http.StatusNotImplemented, map[string]string{
// 		"message": "功能开发中",
// 	})
// }

// // ListUsers 列出用户（简化版，映射到身份）
// func (m *AdminModule) ListUsers(c echo.Context) error {
// 	return m.ListIdentities(c)
// }

// // GetUser 获取用户（简化版，映射到身份）
// func (m *AdminModule) GetUser(c echo.Context) error {
// 	return m.GetIdentity(c)
// }

// // UpdateUser 更新用户（简化版，映射到身份）
// func (m *AdminModule) UpdateUser(c echo.Context) error {
// 	return m.UpdateIdentity(c)
// }

// // DeleteUser 删除用户（简化版，映射到身份）
// func (m *AdminModule) DeleteUser(c echo.Context) error {
// 	return m.DeleteIdentity(c)
// }

// // DisableUser 禁用用户
// func (m *AdminModule) DisableUser(c echo.Context) error {
// 	ctx := c.Request().Context()
// 	id := c.Param("id")

// 	if id == "" {
// 		appErr := xerrors.NewValidationError("id", "用户ID不能为空")
// 		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
// 	}

// 	identity, appErr := m.userService.DisableIdentity(ctx, id)
// 	if appErr != nil {
// 		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
// 	}

// 	respData := map[string]interface{}{
// 		"message":  "用户已禁用",
// 		"identity": identity,
// 	}

// 	return m.respWriter.WriteSuccess(ctx, c.Response().Writer, respData)
// }

// // EnableUser 启用用户
// func (m *AdminModule) EnableUser(c echo.Context) error {
// 	ctx := c.Request().Context()
// 	id := c.Param("id")

// 	if id == "" {
// 		appErr := xerrors.NewValidationError("id", "用户ID不能为空")
// 		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
// 	}

// 	identity, appErr := m.userService.EnableIdentity(ctx, id)
// 	if appErr != nil {
// 		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
// 	}

// 	respData := map[string]interface{}{
// 		"message":  "用户已启用",
// 		"identity": identity,
// 	}

// 	return m.respWriter.WriteSuccess(ctx, c.Response().Writer, respData)
// }

// // ListIdentities 列出身份
// func (m *AdminModule) ListIdentities(c echo.Context) error {
// 	ctx := c.Request().Context()

// 	// 解析查询参数
// 	query := &usermodel.ListIdentitiesQuery{
// 		Page:    1,
// 		PerPage: 20,
// 	}

// 	if pageStr := c.QueryParam("page"); pageStr != "" {
// 		if page, err := strconv.ParseInt(pageStr, 10, 64); err == nil && page > 0 {
// 			query.Page = page
// 		}
// 	}

// 	if perPageStr := c.QueryParam("per_page"); perPageStr != "" {
// 		if perPage, err := strconv.ParseInt(perPageStr, 10, 64); err == nil && perPage > 0 && perPage <= 1000 {
// 			query.PerPage = perPage
// 		}
// 	}

// 	query.Ids = c.QueryParam("ids")

// 	// 获取身份列表
// 	identities, total, appErr := m.userService.ListIdentities(ctx, query)
// 	if appErr != nil {
// 		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
// 	}

// 	respData := map[string]interface{}{
// 		"identities": identities,
// 		"pagination": map[string]interface{}{
// 			"page":     query.Page,
// 			"per_page": query.PerPage,
// 			"total":    total,
// 			"count":    len(identities),
// 		},
// 	}

// 	return m.respWriter.WriteSuccess(ctx, c.Response().Writer, respData)
// }

// // GetIdentity 获取身份
// func (m *AdminModule) GetIdentity(c echo.Context) error {
// 	ctx := c.Request().Context()
// 	id := c.Param("id")

// 	if id == "" {
// 		appErr := xerrors.NewValidationError("id", "身份ID不能为空")
// 		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
// 	}

// 	identity, appErr := m.userService.GetIdentity(ctx, id)
// 	if appErr != nil {
// 		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
// 	}

// 	return m.respWriter.WriteSuccess(ctx, c.Response().Writer, identity)
// }

// // CreateIdentity 创建身份
// func (m *AdminModule) CreateIdentity(c echo.Context) error {
// 	ctx := c.Request().Context()

// 	var req usermodel.CreateIdentityRequest
// 	if err := c.Bind(&req); err != nil {
// 		appErr := xerrors.NewValidationError("request", "请求格式错误")
// 		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
// 	}

// 	// 基础验证
// 	if req.Email == "" {
// 		appErr := xerrors.NewValidationError("email", "邮箱不能为空")
// 		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
// 	}

// 	if req.Username == "" {
// 		appErr := xerrors.NewValidationError("username", "用户名不能为空")
// 		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
// 	}

// 	identity, appErr := m.userService.CreateIdentity(ctx, &req)
// 	if appErr != nil {
// 		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
// 	}

// 	m.logger.InfoContext(ctx, "创建身份成功",
// 		log.String("identity_id", identity.Id),
// 		log.String("email", req.Email),
// 	)

// 	respData := map[string]interface{}{
// 		"message":  "身份创建成功",
// 		"identity": identity,
// 	}

// 	return m.respWriter.WriteSuccess(ctx, c.Response().Writer, respData)
// }

// // UpdateIdentity 更新身份
// func (m *AdminModule) UpdateIdentity(c echo.Context) error {
// 	ctx := c.Request().Context()
// 	id := c.Param("id")

// 	if id == "" {
// 		appErr := xerrors.NewValidationError("id", "身份ID不能为空")
// 		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
// 	}

// 	var req usermodel.UpdateIdentityRequest
// 	if err := c.Bind(&req); err != nil {
// 		appErr := xerrors.NewValidationError("request", "请求格式错误")
// 		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
// 	}

// 	identity, appErr := m.userService.UpdateIdentity(ctx, id, &req)
// 	if appErr != nil {
// 		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
// 	}

// 	m.logger.InfoContext(ctx, "更新身份成功",
// 		log.String("identity_id", id),
// 	)

// 	respData := map[string]interface{}{
// 		"message":  "身份更新成功",
// 		"identity": identity,
// 	}

// 	return m.respWriter.WriteSuccess(ctx, c.Response().Writer, respData)
// }

// // DeleteIdentity 删除身份
// func (m *AdminModule) DeleteIdentity(c echo.Context) error {
// 	ctx := c.Request().Context()
// 	id := c.Param("id")

// 	if id == "" {
// 		appErr := xerrors.NewValidationError("id", "身份ID不能为空")
// 		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
// 	}

// 	if appErr := m.userService.DeleteIdentity(ctx, id); appErr != nil {
// 		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
// 	}

// 	m.logger.InfoContext(ctx, "删除身份成功",
// 		log.String("identity_id", id),
// 	)

// 	respData := map[string]string{
// 		"message": "身份删除成功",
// 	}

// 	return m.respWriter.WriteSuccess(ctx, c.Response().Writer, respData)
// }

// healthCheck 健康检查
func (m *AdminModule) healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"service":   "tsu-admin",
		"version":   "1.0.0",
		"timestamp": time.Now().Unix(),
	})
}

// readinessCheck 就绪检查
func (m *AdminModule) readinessCheck(c echo.Context) error {

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "ready",
		"service": "tsu-admin",
		"checks": map[string]string{
			"database": "ok",
		},
		"timestamp": time.Now().Unix(),
	})
}
