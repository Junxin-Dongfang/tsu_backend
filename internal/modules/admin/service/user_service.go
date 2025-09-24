// File: internal/app/admin/service/user_service.go
package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/entity"

	"github.com/jmoiron/sqlx"
)

// UserService 用户管理服务
type UserService struct {
	db     *sqlx.DB
	logger log.Logger
}

// NewUserService 创建用户管理服务
func NewUserService(db *sqlx.DB, logger log.Logger) *UserService {
	return &UserService{
		db:     db,
		logger: logger,
	}
}

// GetUserByID 根据ID获取用户信息
func (s *UserService) GetUserByID(ctx context.Context, userID string) (*entity.User, *xerrors.AppError) {
	s.logger.InfoContext(ctx, "获取用户信息", log.String("user_id", userID))

	query := `
		SELECT id, email, username, nickname, phone_number, is_premium, diamond_count, created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	var user entity.User
	err := s.db.GetContext(ctx, &user, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, xerrors.New(xerrors.CodeUserNotFound, "用户不存在")
		}
		s.logger.ErrorContext(ctx, "查询用户失败", log.Any("error", err))
		return nil, s.parseDatabaseError(err, "查询用户")
	}

	s.logger.InfoContext(ctx, "用户信息获取成功", log.String("user_id", userID))
	return &user, nil
}

func (s *UserService) UpdateUserProfile(ctx context.Context, userID string, profile map[string]interface{}) (*entity.User, *xerrors.AppError) {
	s.logger.InfoContext(ctx, "更新用户资料",
		log.String("user_id", userID),
		log.Any("profile", profile),
	)

	// 如果没有字段要更新，直接返回当前用户信息
	if len(profile) == 0 {
		return s.GetUserByID(ctx, userID)
	}

	setParts := make([]string, 0, len(profile))
	args := make([]interface{}, 0, len(profile)+1)
	argIndex := 2

	for field, value := range profile {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
		args = append(args, value)
		argIndex++
	}

	query := fmt.Sprintf(`
		UPDATE users
		SET %s
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, username, nickname, email, phone_number, is_premium, diamond_count, created_at, updated_at
		`, strings.Join(setParts, ", "))

	args = append([]interface{}{userID}, args...)

	s.logger.DebugContext(ctx, "执行 SQL 查询", log.String("query", query), log.Any("args", args))

	// 执行查询并扫描结果
	var user entity.User
	err := s.db.GetContext(ctx, &user, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, xerrors.FromCode(xerrors.CodeUserNotFound).WithMetadata("user_id", userID)
		}
		return nil, s.parseDatabaseError(err, "更新用户资料")
	}
	s.logger.InfoContext(ctx, "用户资料更新成功", log.Any("user", user))
	// 返回更新后的用户信息
	return &user, nil
}

// parseDatabaseError 解析数据库错误，返回用户友好的错误信息
func (s *UserService) parseDatabaseError(err error, operation string) *xerrors.AppError {
	errMsg := err.Error()

	s.logger.WarnContext(context.Background(), "解析数据库错误",
		log.String("operation", operation),
		log.String("error", errMsg))

	// PostgreSQL 约束错误解析
	if strings.Contains(errMsg, "duplicate key value violates unique constraint") {
		if strings.Contains(errMsg, "users_email_key") {
			return xerrors.NewValidationError("email", "该邮箱已被使用")
		}
		if strings.Contains(errMsg, "users_username_key") {
			return xerrors.NewValidationError("username", "该用户名已被使用")
		}
		if strings.Contains(errMsg, "users_phone_number_key") {
			return xerrors.NewValidationError("phone_number", "该手机号已被使用")
		}
		return xerrors.NewValidationError("data", "数据重复，请检查输入信息")
	}

	// 外键约束错误
	if strings.Contains(errMsg, "violates foreign key constraint") {
		return xerrors.NewValidationError("reference", "关联数据不存在")
	}

	// 非空约束错误
	if strings.Contains(errMsg, "violates not-null constraint") {
		return xerrors.NewValidationError("required", "必填字段不能为空")
	}

	// 默认数据库错误
	return xerrors.NewDatabaseError(operation, "database", err)
}
