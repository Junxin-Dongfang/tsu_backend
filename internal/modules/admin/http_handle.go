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
