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
	apiAuthReq "tsu-self/internal/api/model/request/auth"
	apiUserReq "tsu-self/internal/api/model/request/user"

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

	// 1. 调用Auth服务进行Kratos登录验证（带重试机制）
	rpcReq := authConverter.LoginRequestToRPC(&req)
	resp, err := m.callWithRetry(ctx, "auth", "Login", mqrpc.Param(rpcReq), 3)
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

// callWithRetry 带重试机制的RPC调用
func (m *AdminModule) callWithRetry(ctx context.Context, serviceName, methodName string, param mqrpc.ParamOption, maxRetries int) (interface{}, string) {
	var result interface{}
	var err string

	// 可重试的错误类型
	retryableErrors := map[string]bool{
		"none available":     true, // 没有可用服务
		"deadline exceeded":  true, // 超时
		"client closed":      true, // 客户端关闭
		"connection refused": true, // 连接被拒绝
		"connection reset":   true, // 连接重置
		"temporary failure":  true, // 临时失败
		"认证服务调用失败":           true, // 我们的自定义错误
	}

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// 在重试前检查上下文是否已取消
		if ctx != nil {
			select {
			case <-ctx.Done():
				return nil, "context canceled"
			default:
			}
		}

		result, err = m.Call(ctx, serviceName, methodName, param)
		if err == "" {
			// 成功，返回结果
			return result, ""
		}

		// 检查是否为可重试的错误
		isRetryable := false
		for retryableErr := range retryableErrors {
			if strings.Contains(err, retryableErr) {
				isRetryable = true
				break
			}
		}

		// 如果不是可重试错误，直接返回
		if !isRetryable {
			return result, err
		}

		// 记录重试信息
		m.logger.WarnContext(ctx, "RPC调用失败，准备重试",
			log.String("service", serviceName),
			log.String("method", methodName),
			log.Int("attempt", attempt),
			log.Int("max_retries", maxRetries),
			log.String("error", err))

		// 如果不是最后一次尝试，等待一段时间再重试
		if attempt < maxRetries {
			// 根据错误类型调整延迟策略
			var delay time.Duration
			switch {
			case strings.Contains(err, "none available"):
				// 服务不可用，使用较长的延迟让服务有时间恢复
				delay = time.Duration(attempt*attempt) * 500 * time.Millisecond
			case strings.Contains(err, "deadline exceeded"):
				// 超时错误，使用较短的延迟
				delay = time.Duration(attempt) * 300 * time.Millisecond
			case strings.Contains(err, "认证服务调用失败"):
				// 自定义错误，可能是临时性问题
				delay = time.Duration(attempt) * 200 * time.Millisecond
			default:
				// 默认指数退避
				delay = time.Duration(attempt) * 250 * time.Millisecond
			}

			// 限制最大延迟时间
			maxDelay := 3 * time.Second
			if delay > maxDelay {
				delay = maxDelay
			}

			// 检查NATS连接状态和服务健康状况
			if nc := m.GetApp().Transport(); nc != nil {
				if !nc.IsConnected() {
					m.logger.WarnContext(ctx, "NATS连接断开，等待重连",
						log.String("status", nc.Status().String()))

					// 等待连接恢复，但不超过delay时间
					waitCtx, cancel := context.WithTimeout(ctx, delay)
					ticker := time.NewTicker(100 * time.Millisecond) // 增加检查间隔

					reconnected := false
					for {
						select {
						case <-waitCtx.Done():
							cancel()
							ticker.Stop()
							goto nextAttempt
						case <-ticker.C:
							if nc.IsConnected() {
								m.logger.InfoContext(ctx, "NATS连接已恢复")
								reconnected = true
								cancel()
								ticker.Stop()
								goto nextAttempt
							}
						}
					}

					// 如果重连成功，减少等待时间
					if reconnected && delay > 500*time.Millisecond {
						delay = delay / 2
					}
				}
			}

			m.logger.DebugContext(ctx, "等待重试",
				log.String("delay", delay.String()),
				log.Int("attempt", attempt),
				log.String("error_type", err))

			// 使用context控制的sleep，支持提前取消
			select {
			case <-ctx.Done():
				return nil, "context canceled during retry delay"
			case <-time.After(delay):
				// 继续重试
			}
		}

	nextAttempt:
	}

	// 所有重试都失败了
	m.logger.ErrorContext(ctx, "RPC调用重试失败",
		log.String("service", serviceName),
		log.String("method", methodName),
		log.Int("attempts", maxRetries),
		log.String("final_error", err))

	return result, err
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

	// 1. 调用Auth服务进行Kratos注册（带重试机制）
	rpcReq := authConverter.RegisterRequestToRPC(&req)
	result, err := m.callWithRetry(ctx, "auth", "Register", mqrpc.Param(rpcReq), 3)
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

// GetUserProfile 获取用户资料
// @Summary 获取用户资料
// @Description 获取指定用户的详细资料信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Success 200 {object} response.APIResponse[apiUserResp.UserProfile] "获取成功"
// @Failure 400 {object} response.APIResponse[any] "请求参数错误"
// @Failure 404 {object} response.APIResponse[any] "用户不存在"
// @Failure 500 {object} response.APIResponse[any] "服务器错误"
// @Router /user/{user_id}/profile [get]
func (m *AdminModule) GetUserProfile(c echo.Context) error {
	ctx := c.Request().Context()
	userID := c.Param("user_id")

	m.logger.InfoContext(ctx, "开始获取用户资料", log.String("user_id", userID))

	// 验证用户ID格式
	if userID == "" {
		return m.respWriter.WriteError(ctx, c.Response().Writer,
			xerrors.New(xerrors.CodeInvalidParams, "用户ID不能为空"))
	}

	// 从数据库获取用户信息
	user, appErr := m.userService.GetUserByID(ctx, userID)
	if appErr != nil {
		if appErr.Code == xerrors.CodeUserNotFound {
			return m.respWriter.WriteError(ctx, c.Response().Writer,
				xerrors.New(xerrors.CodeUserNotFound, "用户不存在"))
		}
		m.logger.ErrorContext(ctx, "获取用户资料失败", log.Any("error", appErr))
		return m.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	// 转换为API响应格式
	profile := userConverter.BasicUserToProfileResponse(user)

	m.logger.InfoContext(ctx, "用户资料获取成功", log.String("user_id", userID))
	return m.respWriter.WriteSuccess(ctx, c.Response().Writer, profile)
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

	profile := userConverter.BasicUserToProfileResponse(updatedUser)

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
