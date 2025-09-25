package impl

import (
	"context"
	"database/sql"

	"tsu-self/internal/entity"
	"tsu-self/internal/repository/interfaces"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

type userRepositoryImpl struct {
	db *sql.DB
}

// NewUserRepository 创建用户仓储实现
func NewUserRepository(db *sql.DB) interfaces.UserRepository {
	return &userRepositoryImpl{db: db}
}

func (r *userRepositoryImpl) GetByID(ctx context.Context, id string) (*entity.User, error) {
	return entity.FindUser(ctx, r.db, id)
}

func (r *userRepositoryImpl) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	return entity.Users(qm.Where("email = ?", email)).One(ctx, r.db)
}

func (r *userRepositoryImpl) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	return entity.Users(qm.Where("username = ?", username)).One(ctx, r.db)
}

func (r *userRepositoryImpl) Create(ctx context.Context, user *entity.User) error {
	return user.Insert(ctx, r.db, boil.Infer())
}

func (r *userRepositoryImpl) Update(ctx context.Context, user *entity.User) error {
	_, err := user.Update(ctx, r.db, boil.Infer())
	return err
}

func (r *userRepositoryImpl) Delete(ctx context.Context, id string) error {
	user := &entity.User{ID: id}
	_, err := user.Delete(ctx, r.db)
	return err
}

func (r *userRepositoryImpl) GetActiveUsers(ctx context.Context, limit, offset int) ([]*entity.User, error) {
	return entity.Users(
		qm.Where("is_banned = ?", false),
		qm.Where("deleted_at IS NULL"),
		qm.Limit(limit),
		qm.Offset(offset),
		qm.OrderBy("created_at DESC"),
	).All(ctx, r.db)
}

func (r *userRepositoryImpl) GetPremiumUsers(ctx context.Context) ([]*entity.User, error) {
	return entity.Users(
		qm.InnerJoin("user_finances uf ON users.id = uf.user_id"),
		qm.Where("uf.premium_expiry > NOW()"),
	).All(ctx, r.db)
}

func (r *userRepositoryImpl) GetBannedUsers(ctx context.Context) ([]*entity.User, error) {
	return entity.Users(
		qm.Where("is_banned = ?", true),
	).All(ctx, r.db)
}

func (r *userRepositoryImpl) CountActiveUsers(ctx context.Context) (int, error) {
	count, err := entity.Users(
		qm.Where("is_banned = ?", false),
		qm.Where("deleted_at IS NULL"),
	).Count(ctx, r.db)
	return int(count), err
}

func (r *userRepositoryImpl) CountPremiumUsers(ctx context.Context) (int, error) {
	count, err := entity.Users(
		qm.InnerJoin("user_finances uf ON users.id = uf.user_id"),
		qm.Where("uf.premium_expiry > NOW()"),
	).Count(ctx, r.db)
	return int(count), err
}