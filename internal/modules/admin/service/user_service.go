// File: internal/app/admin/service/user_service.go
package service

import (
	"context"
	"fmt"
	"strings"

	"tsu-self/internal/pkg/log"

	"github.com/jmoiron/sqlx"
)

// UserService 用户管理服务
type UserService struct {
	db     *sqlx.DB
	logger log.Logger
}

// NewUserService 创建用户管理服务
func NewUserService(logger log.Logger) (*UserService, error) {
	return &UserService{
		logger: logger,
	}, nil
}

func (s *UserService) UpdateUserProfile(ctx context.Context, userID string, profile map[string]interface{}) error {
	s.logger.InfoContext(ctx, "更新用户资料",
		log.String("user_id", userID),
		log.Any("profile", profile),
	)

	setParts := make([]string, 0, len(profile))
	args := make([]interface{}, 0, len(profile)+1)
	argIndex := 2

	for field, value := range profile {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
		args = append(args, value)
		argIndex++
	}

	query := fmt.Sprintf('
		UPDATE users
		SET %s
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, nickname, email, phone, created_at, updated_at')

	return s.db.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.Nickname, &user.Email, &user.Phone, &user.CreatedAt, &user.UpdatedAt)
}
