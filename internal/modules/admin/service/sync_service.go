// internal/modules/admin/service/sync_service.go
package service

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"tsu-self/internal/repository/entity"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/xerrors"
)

type SyncService struct {
	db     *sqlx.DB
	logger log.Logger
}

func NewSyncService(db *sqlx.DB, logger log.Logger) *SyncService {
	return &SyncService{
		db:     db,
		logger: logger,
	}
}

// CreateBusinessUser 创建业务用户（注册时调用）
func (s *SyncService) CreateBusinessUser(ctx context.Context, identityID, email, username string) (*entity.User, *xerrors.AppError) {
	s.logger.InfoContext(ctx, "开始创建业务用户",
		log.String("identity_id", identityID),
		log.String("email", email),
		log.String("username", username))

	// 开始数据库事务
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, xerrors.NewDatabaseError("begin_transaction", "users", err)
	}
	defer tx.Rollback()

	// 插入用户记录
	query := `
		INSERT INTO users (id, username, email, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, username, email, is_premium, diamond_count, created_at, updated_at
	`

	var userInfo entity.User
	err = tx.QueryRowxContext(ctx, query, identityID, username, email).Scan(
		&userInfo.ID,
		&userInfo.Username,
		&userInfo.Email,
		&userInfo.IsPremium,
		&userInfo.DiamondCount,
		&userInfo.CreatedAt,
		&userInfo.UpdatedAt,
	)

	if err != nil {
		s.logger.ErrorContext(ctx, "创建业务用户失败", log.Any("error", err))
		return nil, xerrors.NewDatabaseError("insert", "users", err)
	}

	// 创建用户设置记录
	settingsQuery := `
		INSERT INTO user_settings (user_id, created_at, updated_at)
		VALUES ($1, NOW(), NOW())
	`
	_, err = tx.ExecContext(ctx, settingsQuery, identityID)
	if err != nil {
		s.logger.ErrorContext(ctx, "创建用户设置失败", log.Any("error", err))
		return nil, xerrors.NewDatabaseError("insert", "user_settings", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return nil, xerrors.NewDatabaseError("commit_transaction", "users", err)
	}

	s.logger.InfoContext(ctx, "业务用户创建成功", log.String("user_id", userInfo.ID))
	return &userInfo, nil
}

// GetUserByID 根据ID获取用户信息
func (s *SyncService) GetUserByID(ctx context.Context, userID string) (*entity.User, *xerrors.AppError) {
	query := `
		SELECT id, username, email, is_premium, diamond_count, created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	var userInfo entity.User
	err := s.db.GetContext(ctx, &userInfo, query, userID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, xerrors.FromCode(xerrors.CodeUserNotFound).WithMetadata("user_id", userID)
		}
		return nil, xerrors.NewDatabaseError("select", "users", err)
	}

	return &userInfo, nil
}

// UpdateUserTraits 更新用户特征（同步 email/username）
func (s *SyncService) UpdateUserTraits(ctx context.Context, userID, email, username string) *xerrors.AppError {
	s.logger.InfoContext(ctx, "开始同步用户特征",
		log.String("user_id", userID),
		log.String("email", email),
		log.String("username", username))

	query := `
		UPDATE users
		SET username = $2, email = $3, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, userID, username, email)
	if err != nil {
		return xerrors.NewDatabaseError("update", "users", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return xerrors.NewDatabaseError("rows_affected", "users", err)
	}

	if rowsAffected == 0 {
		return xerrors.FromCode(xerrors.CodeUserNotFound).WithMetadata("user_id", userID)
	}

	s.logger.InfoContext(ctx, "用户特征同步成功", log.String("user_id", userID))
	return nil
}

// RecordLoginHistory 记录登录历史
func (s *SyncService) RecordLoginHistory(ctx context.Context, userID, clientIP, userAgent string, success bool) {
	status := "success"
	if !success {
		status = "failed"
	}

	query := `
		INSERT INTO user_login_history (user_id, login_time, ip_address, user_agent, login_method, status)
		VALUES ($1, NOW(), $2, $3, 'password', $4)
	`

	_, err := s.db.ExecContext(ctx, query, userID, clientIP, userAgent, status)
	if err != nil {
		// 登录历史记录失败不应该影响主流程，只记录日志
		s.logger.WarnContext(ctx, "记录登录历史失败",
			log.String("user_id", userID),
			log.Any("error", err))
	}
}

// UpdateLastLogin 更新最后登录时间
func (s *SyncService) UpdateLastLogin(ctx context.Context, userID, clientIP string) {
	query := `
		UPDATE users
		SET last_login_at = NOW(), last_login_ip = $2, login_count = login_count + 1, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := s.db.ExecContext(ctx, query, userID, clientIP)
	if err != nil {
		s.logger.WarnContext(ctx, "更新最后登录时间失败",
			log.String("user_id", userID),
			log.Any("error", err))
	}
}

// DeleteUser 软删除用户（Saga 回滚时使用）
func (s *SyncService) DeleteUser(ctx context.Context, userID string) *xerrors.AppError {
	query := `
		UPDATE users
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, userID)
	if err != nil {
		return xerrors.NewDatabaseError("soft_delete", "users", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return xerrors.NewDatabaseError("rows_affected", "users", err)
	}

	if rowsAffected == 0 {
		return xerrors.FromCode(xerrors.CodeUserNotFound).WithMetadata("user_id", userID)
	}

	s.logger.InfoContext(ctx, "用户软删除成功", log.String("user_id", userID))
	return nil
}

// SyncUserAfterLogin 登录后同步用户信息
func (s *SyncService) SyncUserAfterLogin(ctx context.Context, userID, clientIP string) (*entity.User, *xerrors.AppError) {
	// 获取用户信息
	userInfo, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 更新最后登录时间（异步，不影响主流程）
	go func() {
		s.UpdateLastLogin(context.Background(), userID, clientIP)
	}()

	return userInfo, nil
}