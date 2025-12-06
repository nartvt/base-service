package repository

import (
	"context"

	"base-service/internal/database/user"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *user.CreateUserParams) (*user.User, error)
	GetUserByUserName(ctx context.Context, userName string) (*user.User, error)
	GetUserByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (*user.User, error)
	ValidateUserPasswordByUserName(ctx context.Context, input *user.ValidateUserPasswordByUserNameParams) (*user.User, error)
}

type userRepositoryImpl struct {
	db *pgxpool.Pool
	q  *user.Queries
}

func NewUserRepository(pool *pgxpool.Pool, db user.DBTX) UserRepository {
	return &userRepositoryImpl{
		db: pool,
		q:  user.New(db),
	}
}

func (r *userRepositoryImpl) CreateUser(ctx context.Context, user *user.CreateUserParams) (*user.User, error) {
	return r.q.CreateUser(ctx, user)
}

func (r *userRepositoryImpl) GetUserByUserName(ctx context.Context, userName string) (*user.User, error) {
	return r.q.GetUserByUserName(ctx, userName)
}

func (r *userRepositoryImpl) GetUserByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (*user.User, error) {
	return r.q.GetUserByUsernameOrEmail(ctx, usernameOrEmail)
}

func (r *userRepositoryImpl) ValidateUserPasswordByUserName(ctx context.Context, input *user.ValidateUserPasswordByUserNameParams) (*user.User, error) {
	return r.q.ValidateUserPasswordByUserName(ctx, input)
}
