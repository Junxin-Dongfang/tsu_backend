package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"tsu-self/internal/entity/auth"
	"tsu-self/internal/modules/auth/client"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
)

// AuthService 认证服务业务逻辑层
type AuthService struct {
	db           *sql.DB
	kratosClient *client.KratosClient
}

// NewAuthService 创建认证服务实例
func NewAuthService(db *sql.DB, kratosClient *client.KratosClient) *AuthService {
	return &AuthService{
		db:           db,
		kratosClient: kratosClient,
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
	UserID     string
	KratosID   string
	Email      string
	Username   string
	NeedVerify bool
}

// Register 用户注册
// 流程：
// 1. 在 Kratos 中创建 identity（失败则直接返回）
// 2. 同步数据到 auth.users 表（失败则尝试删除 Kratos identity）
func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*RegisterResult, error) {
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

	// 3. 在 Kratos 中创建 identity
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
		UserID:     user.ID,
		KratosID:   kratosIdentity.Id,
		Email:      user.Email,
		Username:   user.Username,
		NeedVerify: false,
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
		UserID:      userID,
		LoginTime:   time.Now(),
		IPAddress:   loginIP,
		UserAgent:   null.String{},
		LoginMethod: "password",
		Status:      "active",
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
	Identifier string // email, username, 或 phone_number
	Password   string
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
	// 1. 调用 Kratos Public API 进行认证
	sessionToken, identityID, err := s.kratosClient.LoginWithPassword(ctx, input.Identifier, input.Password)
	if err != nil {
		return nil, fmt.Errorf("登录失败: %w", err)
	}

	// 2. 根据 Kratos ID 查询用户信息
	user, err := s.GetUserByID(ctx, identityID)
	if err != nil {
		// 如果用户不存在,可能需要从 Kratos 同步用户数据
		if err := s.SyncUserFromKratos(ctx, identityID); err != nil {
			return nil, fmt.Errorf("同步用户数据失败: %w", err)
		}
		// 重新查询
		user, err = s.GetUserByID(ctx, identityID)
		if err != nil {
			return nil, fmt.Errorf("获取用户信息失败: %w", err)
		}
	}

	// 3. 检查用户是否被封禁
	if user.IsBanned {
		return nil, fmt.Errorf("用户已被封禁: %s", user.BanReason.String)
	}

	// 4. 更新登录信息 (最后登录时间、登录次数)
	// 注意: 这里暂时不获取 IP,由调用方传入
	// 如果需要在这里更新,可以添加 loginIP 参数

	// 5. 返回登录结果
	return &LoginOutput{
		SessionToken: sessionToken,
		UserID:       user.ID,
		Email:        user.Email,
		Username:     user.Username,
	}, nil
}

// LogoutInput 登出输入
type LogoutInput struct {
	SessionToken string
}

// Logout 用户登出
func (s *AuthService) Logout(ctx context.Context, input LogoutInput) error {
	// 调用 Kratos API 撤销 Session
	if err := s.kratosClient.RevokeSession(ctx, input.SessionToken); err != nil {
		return fmt.Errorf("登出失败: %w", err)
	}

	return nil
}
