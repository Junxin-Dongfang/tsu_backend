package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"tsu-self/internal/entity/auth"
	"tsu-self/internal/modules/auth/client"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/metrics"
	"tsu-self/internal/pkg/sessioncache"
	"tsu-self/internal/pkg/xerrors"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
)

// AuthService 认证服务业务逻辑层
type AuthService struct {
	db           *sql.DB
	kratosClient *client.KratosClient
	redis        RedisClient
	sessionCache *sessioncache.Cache
}

const defaultLoginCacheTTL = 10 * time.Minute

func getLoginCacheTTL() time.Duration {
	if raw := os.Getenv("LOGIN_CACHE_TTL"); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil {
			if d > 0 {
				return d
			}
		}
	}
	return defaultLoginCacheTTL
}

// RedisClient Redis 客户端接口（用于测试时 mock）
type RedisClient interface {
	SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	GetString(ctx context.Context, key string) (string, error)
	Exists(ctx context.Context, key string) (bool, error)
	DeleteKey(ctx context.Context, keys ...string) error
}

// NewAuthService 创建认证服务实例
func NewAuthService(db *sql.DB, kratosClient *client.KratosClient, redis RedisClient) *AuthService {
	logger := log.GetLogger().With("module", "auth_service")
	cache := sessioncache.New(getLoginCacheTTL(), metrics.DefaultLoginMetrics, logger)
	return &AuthService{
		db:           db,
		kratosClient: kratosClient,
		redis:        redis,
		sessionCache: cache,
	}
}

// RegisterInput 注册输入参数
type RegisterInput struct {
	Email    string
	Username string
	Password string
}

// RegisterResult 注册结果
type RegisterResult struct {
	UserID       string
	KratosID     string
	Email        string
	Username     string
	SessionToken string // 新增：Registration Flow 返回的 session token
	NeedVerify   bool
}

// Register 用户自助注册（使用 Kratos Registration Flow）
// 这是正确的、符合 Kratos 最佳实践的注册方式
// 前端只需要提供 email + password，后端自动完成整个流程（隐藏 flow_id）
func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*RegisterResult, error) {
	// 1. 验证输入
	if err := s.validateRegisterInput(input); err != nil {
		return nil, fmt.Errorf("输入验证失败: %w", err)
	}

	// 2. 检查用户是否已存在（提前检查，避免浪费 Kratos 资源）
	exists, err := auth.Users(
		auth.UserWhere.Email.EQ(input.Email),
		auth.UserWhere.DeletedAt.IsNull(),
	).Exists(ctx, s.db)
	if err != nil {
		return nil, fmt.Errorf("检查用户是否存在失败: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("邮箱已被注册")
	}

	// 3. 创建 Registration Flow（前端不感知）
	flow, err := s.kratosClient.CreateRegistrationFlow(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建注册流程失败: %w", err)
	}

	// 4. 提交注册（完成整个 Registration Flow）
	sessionToken, kratosID, err := s.kratosClient.Register(ctx, flow.Id, input.Email, input.Username, input.Password)
	if err != nil {
		// 直接返回 Kratos 的错误信息，不再重复包装
		return nil, err
	}

	// 5. 获取 Kratos Identity 详情
	kratosIdentity, err := s.kratosClient.GetIdentity(ctx, kratosID)
	if err != nil {
		return nil, fmt.Errorf("获取 identity 详情失败: %w", err)
	}

	// 提取 traits
	email, username, err := client.GetIdentityTraits(kratosIdentity)
	if err != nil {
		return nil, fmt.Errorf("提取 identity traits 失败: %w", err)
	}

	// 6. 同步数据到业务数据库
	user := &auth.User{
		ID:         kratosID,
		Email:      email,
		Username:   username,
		Nickname:   null.String{},
		IsBanned:   false,
		LoginCount: 0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err = user.Insert(ctx, s.db, boil.Infer())
	if err != nil {
		// 数据库插入失败，记录日志（Kratos 已创建用户，无法完全回滚）
		// 这是一个一致性问题，应该通过后台同步任务修复
		return nil, fmt.Errorf("保存用户数据失败: %w", err)
	}

	return &RegisterResult{
		UserID:       user.ID,
		KratosID:     kratosID,
		Email:        user.Email,
		Username:     user.Username,
		SessionToken: sessionToken,
		NeedVerify:   false,
	}, nil
}

// RegisterViaAdminAPI 通过 Admin API 注册用户（仅用于后台管理）
// ⚠️ 注意：这不是推荐的用户注册方式，仅用于管理员批量创建用户等场景
func (s *AuthService) RegisterViaAdminAPI(ctx context.Context, input RegisterInput) (*RegisterResult, error) {
	// 1. 验证输入
	if err := s.validateRegisterInput(input); err != nil {
		return nil, fmt.Errorf("输入验证失败: %w", err)
	}

	// 2. 检查用户是否已存在
	exists, err := auth.Users(
		auth.UserWhere.Email.EQ(input.Email),
		auth.UserWhere.DeletedAt.IsNull(),
	).Exists(ctx, s.db)
	if err != nil {
		return nil, fmt.Errorf("检查用户是否存在失败: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("邮箱已被注册")
	}

	// 3. 在 Kratos 中创建 identity（使用 Admin API）
	kratosIdentity, err := s.kratosClient.CreateIdentity(ctx, input.Email, input.Username, input.Password)
	if err != nil {
		return nil, fmt.Errorf("创建 Kratos identity 失败: %w", err)
	}

	// 提取 Kratos identity 的 traits
	email, username, err := client.GetIdentityTraits(kratosIdentity)
	if err != nil {
		// 尝试回滚 Kratos identity
		_ = s.kratosClient.DeleteIdentity(ctx, kratosIdentity.Id)
		return nil, fmt.Errorf("提取 identity traits 失败: %w", err)
	}

	// 4. 同步数据到业务数据库
	user := &auth.User{
		ID:         kratosIdentity.Id,
		Email:      email,
		Username:   username,
		Nickname:   null.String{},
		IsBanned:   false,
		LoginCount: 0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err = user.Insert(ctx, s.db, boil.Infer())
	if err != nil {
		// 数据库插入失败，尝试回滚 Kratos identity
		_ = s.kratosClient.DeleteIdentity(ctx, kratosIdentity.Id)
		return nil, fmt.Errorf("保存用户数据失败: %w", err)
	}

	return &RegisterResult{
		UserID:       user.ID,
		KratosID:     kratosIdentity.Id,
		Email:        user.Email,
		Username:     user.Username,
		SessionToken: "", // Admin API 创建不返回 session token
		NeedVerify:   false,
	}, nil
}

// validateRegisterInput validates registration input
func (s *AuthService) validateRegisterInput(input RegisterInput) error {
	if input.Email == "" {
		return fmt.Errorf("email is required")
	}
	if input.Username == "" {
		return fmt.Errorf("username is required")
	}
	if len(input.Username) < 3 {
		return fmt.Errorf("username must be at least 3 characters")
	}
	if input.Password == "" {
		return fmt.Errorf("password is required")
	}
	if len(input.Password) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}
	return nil
}

// GetUserByID gets user by ID
func (s *AuthService) GetUserByID(ctx context.Context, userID string) (*auth.User, error) {
	user, err := auth.Users(
		auth.UserWhere.ID.EQ(userID),
		auth.UserWhere.DeletedAt.IsNull(),
	).One(ctx, s.db)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return user, nil
}

// GetUserByEmail gets user by email
func (s *AuthService) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	user, err := auth.Users(
		auth.UserWhere.Email.EQ(email),
		auth.UserWhere.DeletedAt.IsNull(),
	).One(ctx, s.db)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return user, nil
}

// UpdateLoginInfo updates user login information
func (s *AuthService) UpdateLoginInfo(ctx context.Context, userID string, loginIP string) error {
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	user.LastLoginAt = null.TimeFrom(time.Now())
	user.LastLoginIP = null.StringFrom(loginIP)
	user.LoginCount++
	user.UpdatedAt = time.Now()

	_, err = user.Update(ctx, s.db, boil.Infer())
	if err != nil {
		return fmt.Errorf("failed to update login info: %w", err)
	}

	// Record login history
	loginHistory := &auth.UserLoginHistory{
		UserID:       userID,
		LoginTime:    time.Now(),
		IPAddress:    loginIP,
		UserAgent:    null.String{},
		LoginMethod:  "password",
		Status:       "active",
		IsSuspicious: false,
	}

	err = loginHistory.Insert(ctx, s.db, boil.Infer())
	if err != nil {
		fmt.Printf("failed to record login history: %v\n", err)
	}

	return nil
}

// SyncUserFromKratos syncs user data from Kratos to business database
func (s *AuthService) SyncUserFromKratos(ctx context.Context, kratosID string) error {
	// 1. Get latest identity info from Kratos
	kratosIdentity, err := s.kratosClient.GetIdentity(ctx, kratosID)
	if err != nil {
		return fmt.Errorf("failed to get Kratos identity: %w", err)
	}

	// Extract traits
	email, username, err := client.GetIdentityTraits(kratosIdentity)
	if err != nil {
		return fmt.Errorf("failed to extract identity traits: %w", err)
	}

	// 2. Check if user exists
	user, err := auth.FindUser(ctx, s.db, kratosID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to query user: %w", err)
	}

	if user != nil {
		// User exists, update info
		user.Email = email
		user.Username = username
		user.UpdatedAt = time.Now()

		_, err = user.Update(ctx, s.db, boil.Infer())
		if err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
	} else {
		// User doesn't exist, create new record
		newUser := &auth.User{
			ID:         kratosID,
			Email:      email,
			Username:   username,
			IsBanned:   false,
			LoginCount: 0,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		err = newUser.Insert(ctx, s.db, boil.Infer())
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
	}

	return nil
}

// BanUser bans a user
func (s *AuthService) BanUser(ctx context.Context, userID string, reason string, banUntil *time.Time) error {
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	user.IsBanned = true
	user.BanReason = null.StringFrom(reason)
	if banUntil != nil {
		user.BanUntil = null.TimeFrom(*banUntil)
	}
	user.UpdatedAt = time.Now()

	_, err = user.Update(ctx, s.db, boil.Infer())
	if err != nil {
		return fmt.Errorf("failed to ban user: %w", err)
	}

	return nil
}

// UnbanUser unbans a user
func (s *AuthService) UnbanUser(ctx context.Context, userID string) error {
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	user.IsBanned = false
	user.BanReason = null.String{}
	user.BanUntil = null.Time{}
	user.UpdatedAt = time.Now()

	_, err = user.Update(ctx, s.db, boil.Infer())
	if err != nil {
		return fmt.Errorf("failed to unban user: %w", err)
	}

	return nil
}

// ==================== Login & Logout ====================

// LoginInput 登录输入
type LoginInput struct {
	Identifier    string // email, username, 或 phone_number
	Password      string
	SessionToken  string // 已存在的 Kratos session token,允许复用
	ClientService string // 调用方服务(admin/game),用于指标标签
}

// LoginOutput 登录输出
type LoginOutput struct {
	SessionToken string
	UserID       string
	Email        string
	Username     string
}

// Login 用户登录
// 支持使用 email, username, 或 phone_number 登录
func (s *AuthService) Login(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	serviceLabel := sessioncache.NormalizeService(input.ClientService)
	isGame := serviceLabel == "game"

	if reused, err := s.tryReuseSession(ctx, serviceLabel, input.SessionToken); err != nil {
		return nil, err
	} else if reused != nil {
		return reused, nil
	}

	sessionToken, identityID, err := s.kratosClient.LoginWithPassword(ctx, input.Identifier, input.Password)
	if err != nil {
		return nil, fmt.Errorf("登录失败: %w", err)
	}

	user, err := s.ensureUser(ctx, identityID)
	if err != nil {
		return nil, err
	}

	if user.IsBanned {
		return nil, fmt.Errorf("用户已被封禁: %s", user.BanReason.String)
	}

	if isGame {
		metrics.DefaultBusinessMetrics.IncPlayers(metrics.GetServiceName())
	}

	output := &LoginOutput{
		SessionToken: sessionToken,
		UserID:       user.ID,
		Email:        user.Email,
		Username:     user.Username,
	}

	if s.sessionCache != nil {
		s.sessionCache.Set(ctx, serviceLabel, sessioncache.Session{
			SessionToken: sessionToken,
			UserID:       user.ID,
			Username:     user.Username,
			Email:        user.Email,
		})
	}

	return output, nil
}

func (s *AuthService) ensureUser(ctx context.Context, identityID string) (*auth.User, error) {
	user, err := s.GetUserByID(ctx, identityID)
	if err == nil {
		return user, nil
	}
	if err := s.SyncUserFromKratos(ctx, identityID); err != nil {
		return nil, fmt.Errorf("同步用户数据失败: %w", err)
	}
	user, err = s.GetUserByID(ctx, identityID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}
	return user, nil
}

func (s *AuthService) tryReuseSession(ctx context.Context, serviceLabel, token string) (*LoginOutput, error) {
	if token == "" || s.sessionCache == nil {
		return nil, nil
	}

	if cached, ok := s.sessionCache.Get(ctx, serviceLabel, token); ok {
		user, err := s.ensureUser(ctx, cached.UserID)
		if err != nil {
			return nil, err
		}
		if user.IsBanned {
			return nil, fmt.Errorf("用户已被封禁: %s", user.BanReason.String)
		}
		s.sessionCache.Set(ctx, serviceLabel, sessioncache.Session{
			SessionToken: token,
			UserID:       user.ID,
			Username:     user.Username,
			Email:        user.Email,
		})
		return &LoginOutput{
			SessionToken: token,
			UserID:       user.ID,
			Email:        user.Email,
			Username:     user.Username,
		}, nil
	}

	session, err := s.kratosClient.ValidateSession(ctx, token)
	if err != nil {
		if isSessionExpiredErr(err) {
			s.sessionCache.Delete(ctx, serviceLabel, token, "expired")
			return nil, nil
		}
		// 其他错误时,记录日志但允许继续走密码登录路径
		log.GetLogger().WarnContext(ctx, "validate session failed", log.Any("error", err), log.String("service", serviceLabel))
		return nil, nil
	}

	if session == nil || session.Identity == nil {
		return nil, xerrors.NewKratosDataIntegrityError("session.identity", "validate session 未返回 identity")
	}

	user, err := s.ensureUser(ctx, session.Identity.Id)
	if err != nil {
		return nil, err
	}
	if user.IsBanned {
		return nil, fmt.Errorf("用户已被封禁: %s", user.BanReason.String)
	}

	s.sessionCache.Set(ctx, serviceLabel, sessioncache.Session{
		SessionToken: token,
		UserID:       user.ID,
		Username:     user.Username,
		Email:        user.Email,
	})

	return &LoginOutput{
		SessionToken: token,
		UserID:       user.ID,
		Email:        user.Email,
		Username:     user.Username,
	}, nil
}

func isSessionExpiredErr(err error) bool {
	var appErr *xerrors.AppError
	if errors.As(err, &appErr) {
		return appErr.Code == xerrors.CodeSessionExpired
	}
	return false
}

// LogoutInput 登出输入
type LogoutInput struct {
	SessionToken  string
	ClientService string
}

// Logout 用户登出
func (s *AuthService) Logout(ctx context.Context, input LogoutInput) error {
	serviceLabel := sessioncache.NormalizeService(input.ClientService)
	// 调用 Kratos API 撤销 Session
	if err := s.kratosClient.RevokeSession(ctx, input.SessionToken); err != nil {
		return fmt.Errorf("登出失败: %w", err)
	}

	if s.sessionCache != nil {
		s.sessionCache.Delete(ctx, serviceLabel, input.SessionToken, "logout")
	}

	// 记录玩家下线指标
	if serviceLabel == "game" {
		metrics.DefaultBusinessMetrics.DecPlayers(metrics.GetServiceName())
	}

	return nil
}

// ==================== 密码重置功能 ====================

// InitiatePasswordRecovery 用户发起密码恢复 (步骤1)
func (s *AuthService) InitiatePasswordRecovery(ctx context.Context, email string) error {
	// 1. 创建恢复流程
	flow, err := s.kratosClient.CreateRecoveryFlow(ctx)
	if err != nil {
		return fmt.Errorf("创建恢复流程失败: %w", err)
	}

	// 2. 提交邮箱,请求发送验证码
	_, err = s.kratosClient.UpdateRecoveryFlowWithCode(ctx, flow.Id, email)
	if err != nil {
		return fmt.Errorf("发送恢复验证码失败: %w", err)
	}

	// 3. 将 email → flow_id 映射存储到 Redis（10分钟过期）
	cacheKey := fmt.Sprintf("recovery:flow:%s", email)
	err = s.redis.SetWithTTL(ctx, cacheKey, flow.Id, 10*time.Minute)
	if err != nil {
		return fmt.Errorf("缓存 flow_id 失败: %w", err)
	}

	return nil
}

// VerifyRecoveryCode 用户验证恢复码 (步骤2)
// email: 用户邮箱（用于从 Redis 获取 flow_id）
// code: 验证码
// 返回：Kratos 特权 settings flow ID（用于步骤3重置密码）
func (s *AuthService) VerifyRecoveryCode(ctx context.Context, email, code string) (privilegedFlowID string, err error) {
	// 1. 从 Redis 获取 flow_id
	cacheKey := fmt.Sprintf("recovery:flow:%s", email)
	flowID, err := s.redis.GetString(ctx, cacheKey)
	if err != nil {
		return "", fmt.Errorf("恢复流程已过期，请重新发送验证码")
	}

	// 2. 验证验证码（调用 Kratos）
	// 启用 use_continue_with_transitions 后，Kratos 会：
	// - 验证验证码
	// - 在 ContinueWith 中返回 session token 和 settings flow ID
	sessionToken, privilegedFlowID, err := s.kratosClient.VerifyRecoveryCodeAndGetSessionToken(ctx, flowID, code)
	if err != nil {
		// 提供更友好的错误信息
		errMsg := err.Error()
		if strings.Contains(errMsg, "401") || strings.Contains(errMsg, "Unauthorized") {
			return "", fmt.Errorf("验证码错误或已过期，请重新获取")
		}
		if strings.Contains(errMsg, "410") {
			return "", fmt.Errorf("恢复流程已过期，请重新发起密码恢复")
		}
		if strings.Contains(errMsg, "400") {
			return "", fmt.Errorf("验证码格式错误，请输入6位数字验证码")
		}
		return "", fmt.Errorf("验证失败: %s", errMsg)
	}

	// 3. 将 privileged flow ID 与 email + session token 关联（5分钟有效）
	// 格式: email:sessionToken
	tokenKey := fmt.Sprintf("recovery:flow_token:%s", privilegedFlowID)
	tokenData := fmt.Sprintf("%s:%s", email, sessionToken)
	err = s.redis.SetWithTTL(ctx, tokenKey, tokenData, 5*time.Minute)
	if err != nil {
		return "", fmt.Errorf("缓存flow token失败: %w", err)
	}

	// 4. 清理 recovery flow_id 缓存（已完成验证，不再需要）
	s.redis.DeleteKey(ctx, cacheKey)

	return privilegedFlowID, nil
}

// ResetPassword 重置密码 (步骤3)
// flowToken: 从步骤2返回的 Kratos 特权 settings flow ID
// email: 用户邮箱（用于验证）
// newPassword: 新密码
// 🔒 安全性：完全遵循 Kratos 流程，使用特权 settings flow + session token 更新密码
func (s *AuthService) ResetPassword(ctx context.Context, flowToken, email, newPassword string) error {
	// 1. 验证 flow token 是否有效且与 email 匹配
	tokenKey := fmt.Sprintf("recovery:flow_token:%s", flowToken)
	cachedData, err := s.redis.GetString(ctx, tokenKey)
	if err != nil {
		return fmt.Errorf("验证凭证已过期或无效，请重新发起密码恢复")
	}

	// 2. 解析缓存数据（格式: email:sessionToken）
	parts := strings.SplitN(cachedData, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("验证数据异常，请重新发起密码恢复")
	}
	cachedEmail := parts[0]
	sessionToken := parts[1]

	if cachedEmail != email {
		return fmt.Errorf("验证信息与邮箱不匹配，请确认邮箱地址是否正确")
	}

	if sessionToken == "" {
		return fmt.Errorf("会话凭证无效，请重新发起密码恢复流程")
	}

	// 3. 验证密码强度（前端验证的补充）
	if len(newPassword) < 6 {
		return fmt.Errorf("密码长度至少为6位")
	}
	if len(newPassword) > 72 {
		return fmt.Errorf("密码长度不能超过72位")
	}

	// 4. 使用 Kratos 特权 Settings Flow 更新密码（正确的方式）
	// Kratos 会：
	// - 验证特权 session token（确保用户已通过 recovery 验证）
	// - 更新密码
	// - 记录审计日志
	err = s.kratosClient.UpdatePasswordWithPrivilegedFlow(ctx, flowToken, sessionToken, newPassword)
	if err != nil {
		// 直接返回 KratosClient 的错误信息（已经是用户友好的格式）
		return err
	}

	// 5. 清理缓存（密码已成功更新，flow token 已使用，防止重复使用）
	s.redis.DeleteKey(ctx, tokenKey)

	return nil
}

// ResetPasswordWithCode 验证码重置密码（验证码 + 新密码）
// email: 用户邮箱
// code: 验证码
// newPassword: 新密码
// 🔒 安全性：完全遵循 Kratos 流程，内部合并验证码验证和密码重置两步
func (s *AuthService) ResetPasswordWithCode(ctx context.Context, email, code, newPassword string) error {
	// 1. 从 Redis 获取 flow_id（验证恢复流程是否存在且未过期）
	cacheKey := fmt.Sprintf("recovery:flow:%s", email)
	flowID, err := s.redis.GetString(ctx, cacheKey)
	if err != nil {
		return fmt.Errorf("恢复流程已过期，请重新发送验证码")
	}

	// 2. 验证验证码并获取特权 session token
	sessionToken, privilegedFlowID, err := s.kratosClient.VerifyRecoveryCodeAndGetSessionToken(ctx, flowID, code)
	if err != nil {
		// 提供更友好的错误信息
		errMsg := err.Error()
		if strings.Contains(errMsg, "401") || strings.Contains(errMsg, "Unauthorized") {
			return fmt.Errorf("验证码错误或已过期，请重新获取")
		}
		if strings.Contains(errMsg, "410") {
			return fmt.Errorf("恢复流程已过期，请重新发起密码恢复")
		}
		if strings.Contains(errMsg, "400") {
			return fmt.Errorf("验证码格式错误，请输入6位数字验证码")
		}
		return fmt.Errorf("验证失败: %s", errMsg)
	}

	// 3. 验证密码强度
	if len(newPassword) < 6 {
		return fmt.Errorf("密码长度至少为6位")
	}
	if len(newPassword) > 72 {
		return fmt.Errorf("密码长度不能超过72位")
	}

	// 4. 使用特权 session token 更新密码
	err = s.kratosClient.UpdatePasswordWithPrivilegedFlow(ctx, privilegedFlowID, sessionToken, newPassword)
	if err != nil {
		return err
	}

	// 5. 清理 Redis 缓存（密码已成功更新）
	if err := s.redis.DeleteKey(ctx, cacheKey); err != nil {
		// 清理缓存失败不影响整个操作，只记录日志
		fmt.Printf("[WARN] 清理恢复流程缓存失败 (email=%s): %v\n", email, err)
	}

	return nil
}

// DeleteUser 删除用户（软删除业务 DB + 删除 Kratos identity）
// ⚠️ 注意：这是不可逆操作，会同时删除 Kratos identity 和业务数据
func (s *AuthService) DeleteUser(ctx context.Context, userID string) error {
	// 1. 检查用户是否存在
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}

	// 2. 先软删除业务数据库（设置 deleted_at）
	// 使用软删除而不是硬删除，保留审计记录
	user.DeletedAt = null.TimeFrom(time.Now())
	user.UpdatedAt = time.Now()

	_, err = user.Update(ctx, s.db, boil.Infer())
	if err != nil {
		return fmt.Errorf("软删除业务数据失败: %w", err)
	}

	// 3. 再删除 Kratos identity
	// 注意：如果 Kratos 删除失败，业务 DB 已经标记为删除
	// 这是可以接受的，因为用户在业务层面已经"删除"了
	err = s.kratosClient.DeleteIdentity(ctx, user.ID)
	if err != nil {
		// 记录错误但不回滚业务 DB 的软删除
		// 可以通过后台任务重试 Kratos 删除
		fmt.Printf("[WARN] 删除 Kratos identity 失败 (user_id=%s): %v\n", userID, err)
		// 不返回错误，因为业务层面用户已删除
	}

	return nil
}
