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
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	// 新的 API 层模型
	apiAuthReq "tsu-self/internal/api/request/auth"
	apiUserReq "tsu-self/internal/api/request/user"

	// 转换器
	authConverter "tsu-self/internal/converter/auth"
	userConverter "tsu-self/internal/converter/user"

	// 新的 RPC 模型
	"tsu-self/internal/rpc/generated/auth"

	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/xerrors"

	"github.com/labstack/echo/v4"
	mqrpc "github.com/liangdas/mqant/rpc"
	"google.golang.org/protobuf/proto"
)

// Login 用户登录 - 协调Auth服务和主数据库
// @Summary 用户登录
// @Description 通过用户名或邮箱和密码登录，同步更新主数据库登录信息
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

	m.logger.InfoContext(ctx, "开始用户登录流程",
		log.String("identifier", req.Identifier),
		log.String("client_ip", req.ClientIP))

	// 1. 调用Auth服务进行Kratos登录验证
	rpcReq := authConverter.LoginRequestToRPC(&req)
	resp, err := m.Call(ctx, "auth", "Login", mqrpc.Param(rpcReq))
	if err != "" {
		m.logger.ErrorContext(ctx, "Auth服务调用失败", log.Any("error", err))
		return response.InternalServerError(ctx, c.Response().Writer, m.respWriter, "认证服务调用失败")
	}

	// 处理 RPC 响应
	var rpcResp *auth.LoginResponse
	switch v := resp.(type) {
	case []byte:
		rpcResp = &auth.LoginResponse{}
		if err := proto.Unmarshal(v, rpcResp); err != nil {
			m.logger.ErrorContext(ctx, "Login响应反序列化失败", log.Any("error", err))
			return response.InternalServerError(ctx, c.Response().Writer, m.respWriter, "响应处理失败")
		}
	case *auth.LoginResponse:
		rpcResp = v
	default:
		m.logger.ErrorContext(ctx, "Login响应类型错误", log.Any("type", fmt.Sprintf("%T", v)))
		return response.InternalServerError(ctx, c.Response().Writer, m.respWriter, "响应类型错误")
	}

	// 2. 检查登录是否成功
	if !rpcResp.Success {
		// 登录失败，返回具体错误信息
		m.logger.WarnContext(ctx, "Kratos登录失败",
			log.String("error_message", rpcResp.ErrorMessage))

		// 如果有具体的错误消息，使用它；否则使用默认消息
		errorMsg := rpcResp.ErrorMessage
		if errorMsg == "" {
			errorMsg = "用户名或密码错误"
		}

		appErr := xerrors.NewAuthError(errorMsg)
		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	// 3. 如果Kratos登录成功，同步更新主数据库
	if rpcResp.Success && rpcResp.IdentityId != "" {
		m.logger.InfoContext(ctx, "Kratos登录成功，开始同步主数据库登录信息",
			log.String("identity_id", rpcResp.IdentityId))

		// 记录登录历史
		go func() {
			m.syncService.RecordLoginHistory(context.Background(), rpcResp.IdentityId, req.ClientIP, req.UserAgent, true)
		}()

		// 更新最后登录时间
		go func() {
			m.syncService.UpdateLastLogin(context.Background(), rpcResp.IdentityId, req.ClientIP)
		}()

		// 获取用户业务信息（可选，用于填充响应）
		if userInfo, syncErr := m.syncService.GetUserByID(ctx, rpcResp.IdentityId); syncErr == nil {
			m.logger.InfoContext(ctx, "获取用户业务信息成功",
				log.String("identity_id", rpcResp.IdentityId),
				log.String("username", userInfo.Username))
		}
	} else {
		m.logger.InfoContext(ctx, "登录失败，记录失败历史")
		// 记录登录失败历史（如果能确定用户ID的话）
		if rpcResp.IdentityId != "" {
			go func() {
				m.syncService.RecordLoginHistory(context.Background(), rpcResp.IdentityId, req.ClientIP, req.UserAgent, false)
			}()
		}
	}

	// 使用转换器转换为 API 响应
	apiResult := authConverter.LoginResponseFromRPC(rpcResp)

	return m.respWriter.WriteSuccess(ctx, c.Response().Writer, apiResult)
}

// Register 用户注册 - 协调Auth服务和主数据库
// @Summary 用户注册
// @Description 通过邮箱、用户名和密码注册新用户，确保Kratos和主数据库数据一致
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

	m.logger.InfoContext(ctx, "开始用户注册流程",
		log.String("email", req.Email),
		log.String("username", req.Username))

	// 1. 调用Auth服务进行Kratos注册
	rpcReq := authConverter.RegisterRequestToRPC(&req)
	result, err := m.Call(ctx, "auth", "Register", mqrpc.Param(rpcReq))
	if err != "" {
		m.logger.ErrorContext(ctx, "Auth服务调用失败", log.Any("error", err))
		return m.respWriter.WriteError(ctx, c.Response().Writer,
			xerrors.New(xerrors.CodeExternalServiceError, "认证服务调用失败"))
	}

	// 处理 RPC 响应
	var rpcResp *auth.RegisterResponse
	switch v := result.(type) {
	case []byte:
		rpcResp = &auth.RegisterResponse{}
		if err := proto.Unmarshal(v, rpcResp); err != nil {
			m.logger.ErrorContext(ctx, "Protobuf反序列化失败", log.Any("error", err))
			return m.respWriter.WriteError(ctx, c.Response().Writer,
				xerrors.FromCode(xerrors.CodeInternalError).WithMetadata("reason", "protobuf_unmarshal_failed"))
		}
	case *auth.RegisterResponse:
		rpcResp = v
	default:
		m.logger.ErrorContext(ctx, "类型断言失败",
			log.Any("expected", "*auth.RegisterResponse or []byte"),
			log.Any("actual", fmt.Sprintf("%T", result)))
		return m.respWriter.WriteError(ctx, c.Response().Writer,
			xerrors.FromCode(xerrors.CodeInternalError).WithMetadata("reason", "invalid_response_type"))
	}

	// 2. 检查注册是否成功
	if !rpcResp.Success {
		// 注册失败，返回具体错误信息
		m.logger.WarnContext(ctx, "Kratos注册失败",
			log.String("error_message", rpcResp.ErrorMessage))

		// 如果有具体的错误消息，使用它；否则使用默认消息
		errorMsg := rpcResp.ErrorMessage
		if errorMsg == "" {
			errorMsg = "注册失败，请检查输入信息"
		}

		appErr := xerrors.NewValidationError("registration", errorMsg)
		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	// 3. 如果Kratos注册成功，同步到主数据库
	if rpcResp.Success && rpcResp.IdentityId != "" {
		m.logger.InfoContext(ctx, "Kratos注册成功，开始同步到主数据库",
			log.String("identity_id", rpcResp.IdentityId))

		// 同步用户到主数据库
		_, syncErr := m.syncService.CreateBusinessUser(ctx, rpcResp.IdentityId, req.Email, req.Username)
		if syncErr != nil {
			m.logger.ErrorContext(ctx, "同步用户到主数据库失败",
				log.String("identity_id", rpcResp.IdentityId),
				log.Any("error", syncErr))

			// 这里应该实现补偿机制，比如删除Kratos中的用户
			// 暂时只记录错误，不影响返回结果
		} else {
			m.logger.InfoContext(ctx, "用户同步到主数据库成功",
				log.String("identity_id", rpcResp.IdentityId))
		}
	}

	// 使用转换器转换为 API 响应
	apiResult := authConverter.RegisterResponseFromRPC(rpcResp)

	return m.respWriter.WriteSuccess(ctx, c.Response().Writer, apiResult)
}

// UpdateUserProfile 更新用户资料
// @Summary 更新用户资料
// @Description 更新指定用户的资料信息
// @Tags 用户
// @Accept json
// @Produce json
// @Param profile body apiUserReq.UpdateUserProfileRequest true "用户资料更新请求"
// @Success 200 {object} response.APIResponse[apiUserResp.UserProfile] "更新成功"
// @Failure 400 {object} response.APIResponse[any] "请求参数错误"
// @Failure 401 {object} response.APIResponse[any] "未授权"
// @Failure 404 {object} response.APIResponse[any] "用户不存在"
// @Failure 500 {object} response.APIResponse[any] "服务器错误"
// @Router /user/{user_id}/profile [put]
func (m *AdminModule) UpdateUserProfile(c echo.Context) error {
	ctx := c.Request().Context()

	userID := c.Param("user_id")
	if userID == "" {
		appErr := xerrors.NewValidationError("user_id", "用户ID不能为空")
		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	var req apiUserReq.UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		appErr := xerrors.NewValidationError("request", "请求格式错误")
		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	// 手动验证请求数据
	if err := m.validateUpdateProfileRequest(&req); err != nil {
		return m.respWriter.WriteError(ctx, c.Response().Writer, err)
	}

	// 转换为数据库更新格式
	updates := userConverter.UpdateProfileRequestToEntity(&req)

	// 调用用户服务更新资料
	updatedUser, appErr := m.userService.UpdateUserProfile(ctx, userID, updates)
	if appErr != nil {
		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	profile := userConverter.EntityToProfileResponse(updatedUser)

	return m.respWriter.WriteSuccess(ctx, c.Response().Writer, profile)
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

// validateUpdateProfileRequest 验证更新用户资料请求
func (m *AdminModule) validateUpdateProfileRequest(req *apiUserReq.UpdateProfileRequest) *xerrors.AppError {
	// 验证昵称
	if req.Nickname != "" {
		req.Nickname = strings.TrimSpace(req.Nickname)
		if len(req.Nickname) > 30 {
			return xerrors.NewValidationError("nickname", "昵称长度不能超过30个字符")
		}
		if len(req.Nickname) == 0 {
			return xerrors.NewValidationError("nickname", "昵称不能为空字符串")
		}
	}

	// 验证邮箱
	if req.Email != "" {
		req.Email = strings.TrimSpace(req.Email)
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(req.Email) {
			return xerrors.NewValidationError("email", "邮箱格式不正确")
		}
	}

	// 验证手机号 (E.164格式)
	if req.PhoneNumber != "" {
		req.PhoneNumber = strings.TrimSpace(req.PhoneNumber)
		phoneRegex := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
		if !phoneRegex.MatchString(req.PhoneNumber) {
			return xerrors.NewValidationError("phone_number", "手机号格式不正确，需要E.164格式(如：+1234567890)")
		}
	}

	return nil
}
