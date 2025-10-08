package service

import (
	"context"
	"database/sql"
	"fmt"

	"tsu-self/internal/entity/auth"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// UserService 用户服务
type UserService struct {
	userRepo interfaces.UserRepository
}

// NewUserService 创建用户服务
func NewUserService(db *sql.DB) *UserService {
	return &UserService{
		userRepo: impl.NewUserRepository(db),
	}
}

// GetUsers 获取用户列表
func (s *UserService) GetUsers(ctx context.Context, params interfaces.UserQueryParams) ([]*auth.User, int64, error) {
	return s.userRepo.List(ctx, params)
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(ctx context.Context, userID string) (*auth.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(ctx context.Context, userID string, updates map[string]interface{}) error {
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// 更新字段
	if username, ok := updates["username"].(string); ok && username != "" {
		// 检查用户名是否已被使用
		existing, err := s.userRepo.GetByUsername(ctx, username)
		if err == nil && existing.ID != userID {
			return fmt.Errorf("用户名已被使用: %s", username)
		}
		user.Username = username
	}

	if email, ok := updates["email"].(string); ok && email != "" {
		// 检查邮箱是否已被使用
		existing, err := s.userRepo.GetByEmail(ctx, email)
		if err == nil && existing.ID != userID {
			return fmt.Errorf("邮箱已被使用: %s", email)
		}
		user.Email = email
	}

	if nickname, ok := updates["nickname"].(string); ok {
		if nickname != "" {
			user.Nickname.SetValid(nickname)
		} else {
			user.Nickname.Valid = false
		}
	}

	if phoneNumber, ok := updates["phone_number"].(string); ok {
		if phoneNumber != "" {
			user.PhoneNumber.SetValid(phoneNumber)
		} else {
			user.PhoneNumber.Valid = false
		}
	}

	if avatarURL, ok := updates["avatar_url"].(string); ok {
		if avatarURL != "" {
			user.AvatarURL.SetValid(avatarURL)
		} else {
			user.AvatarURL.Valid = false
		}
	}

	if bio, ok := updates["bio"].(string); ok {
		if bio != "" {
			user.Bio.SetValid(bio)
		} else {
			user.Bio.Valid = false
		}
	}

	if birthDate, ok := updates["birth_date"].(string); ok {
		if birthDate != "" {
			// TODO: 解析日期字符串
			// 这里简单处理,实际应该解析为 time.Time
			// user.BirthDate.SetValid(parsedDate)
		}
	}

	if gender, ok := updates["gender"].(string); ok {
		if gender != "" {
			user.Gender.SetValid(gender)
		} else {
			user.Gender.Valid = false
		}
	}

	if timezone, ok := updates["timezone"].(string); ok {
		if timezone != "" {
			user.Timezone.SetValid(timezone)
		} else {
			user.Timezone.Valid = false
		}
	}

	if language, ok := updates["language"].(string); ok {
		if language != "" {
			user.Language.SetValid(language)
		} else {
			user.Language.Valid = false
		}
	}

	// 保存更新
	return s.userRepo.Update(ctx, user)
}

// BanUser 封禁用户
func (s *UserService) BanUser(ctx context.Context, userID string, banUntil *string, banReason string) error {
	return s.userRepo.BanUser(ctx, userID, banUntil, banReason)
}

// UnbanUser 解禁用户
func (s *UserService) UnbanUser(ctx context.Context, userID string) error {
	return s.userRepo.UnbanUser(ctx, userID)
}

// UpdateLoginInfo 更新登录信息
func (s *UserService) UpdateLoginInfo(ctx context.Context, userID string, loginIP string) error {
	return s.userRepo.UpdateLoginInfo(ctx, userID, loginIP)
}
