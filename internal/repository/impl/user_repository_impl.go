package impl

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/google/uuid"

	"tsu-self/internal/entity/auth"
	"tsu-self/internal/repository/interfaces"
)

type userRepositoryImpl struct {
	db *sql.DB
}

// NewUserRepository 创建用户仓储实例
func NewUserRepository(db *sql.DB) interfaces.UserRepository {
	return &userRepositoryImpl{db: db}
}

// GetByID 根据ID获取用户
func (r *userRepositoryImpl) GetByID(ctx context.Context, userID string) (*auth.User, error) {
	user, err := auth.Users(
		qm.Where("id = ? AND deleted_at IS NULL", userID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("用户不存在: %s", userID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return user, nil
}

// GetByUsername 根据用户名获取用户
func (r *userRepositoryImpl) GetByUsername(ctx context.Context, username string) (*auth.User, error) {
	user, err := auth.Users(
		qm.Where("username = ? AND deleted_at IS NULL", username),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("用户不存在: %s", username)
	}
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *userRepositoryImpl) GetByEmail(ctx context.Context, email string) (*auth.User, error) {
	user, err := auth.Users(
		qm.Where("email = ? AND deleted_at IS NULL", email),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("用户不存在: %s", email)
	}
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return user, nil
}

// List 获取用户列表（分页）
func (r *userRepositoryImpl) List(ctx context.Context, params interfaces.UserQueryParams) ([]*auth.User, int64, error) {
	// 构建查询条件
	queries := []qm.QueryMod{
		qm.Where("deleted_at IS NULL"),
	}

	// 关键词搜索
	if params.Keyword != "" {
		keyword := "%" + params.Keyword + "%"
		queries = append(queries, qm.And("(username ILIKE ? OR email ILIKE ? OR nickname ILIKE ?)", keyword, keyword, keyword))
	}

	// 封禁状态筛选
	if params.IsBanned != nil {
		queries = append(queries, qm.And("is_banned = ?", *params.IsBanned))
	}

	// 排序
	sortBy := params.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortDir := strings.ToUpper(params.SortDir)
	if sortDir != "ASC" && sortDir != "DESC" {
		sortDir = "DESC"
	}

	// 获取总数
	total, err := auth.Users(queries...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("统计用户总数失败: %w", err)
	}

	// 分页
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	offset := (params.Page - 1) * params.PageSize

	queries = append(queries,
		qm.OrderBy(fmt.Sprintf("%s %s", sortBy, sortDir)),
		qm.Limit(params.PageSize),
		qm.Offset(offset),
	)

	// 查询用户列表
	users, err := auth.Users(queries...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询用户列表失败: %w", err)
	}

	return users, total, nil
}

// Create 创建用户
func (r *userRepositoryImpl) Create(ctx context.Context, user *auth.User) error {
	// 生成 UUID
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	// 设置时间戳
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// 插入数据库
	if err := user.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}

	return nil
}

// Update 更新用户信息
func (r *userRepositoryImpl) Update(ctx context.Context, user *auth.User) error {
	// 更新时间戳
	user.UpdatedAt = time.Now()

	// 更新数据库
	if _, err := user.Update(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("更新用户失败: %w", err)
	}

	return nil
}

// BanUser 封禁用户
func (r *userRepositoryImpl) BanUser(ctx context.Context, userID string, banUntil *string, banReason string) error {
	user, err := r.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	user.IsBanned = true
	user.BanReason.SetValid(banReason)

	if banUntil != nil && *banUntil != "" {
		// 解析时间字符串
		t, err := time.Parse(time.RFC3339, *banUntil)
		if err != nil {
			return fmt.Errorf("封禁时间格式错误: %w", err)
		}
		user.BanUntil.SetValid(t)
	} else {
		// 永久封禁
		user.BanUntil.Valid = false
	}

	return r.Update(ctx, user)
}

// UnbanUser 解禁用户
func (r *userRepositoryImpl) UnbanUser(ctx context.Context, userID string) error {
	user, err := r.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	user.IsBanned = false
	user.BanUntil.Valid = false
	user.BanReason.Valid = false

	return r.Update(ctx, user)
}

// UpdateLoginInfo 更新登录信息
func (r *userRepositoryImpl) UpdateLoginInfo(ctx context.Context, userID string, loginIP string) error {
	user, err := r.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	now := time.Now()
	user.LastLoginAt.SetValid(now)
	user.LastLoginIP.SetValid(loginIP)
	user.LoginCount++

	return r.Update(ctx, user)
}

// Exists 检查用户是否存在
func (r *userRepositoryImpl) Exists(ctx context.Context, userID string) (bool, error) {
	count, err := auth.Users(
		qm.Where("id = ? AND deleted_at IS NULL", userID),
	).Count(ctx, r.db)

	if err != nil {
		return false, fmt.Errorf("检查用户是否存在失败: %w", err)
	}

	return count > 0, nil
}
