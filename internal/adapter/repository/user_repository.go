package repository

import (
	"context"

	"base-service/internal/adapter/repository/mapper"
	"base-service/internal/database/user"
	"base-service/internal/domain/entity"
	domainerrors "base-service/internal/domain/errors"
	"base-service/internal/domain/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

// userRepository implements the domain.UserRepository interface.
type userRepository struct {
	pool    *pgxpool.Pool
	queries *user.Queries
}

// NewUserRepository creates a new user repository adapter.
func NewUserRepository(pool *pgxpool.Pool) repository.UserRepository {
	return &userRepository{
		pool:    pool,
		queries: user.New(pool),
	}
}

// Create creates a new user and returns the created user with ID.
func (r *userRepository) Create(ctx context.Context, u *entity.User) (*entity.User, error) {
	params := mapper.UserEntityToCreateParams(u)
	dbUser, err := r.queries.CreateUser(ctx, params)
	if err != nil {
		// TODO: Check for unique constraint violations and return appropriate domain errors
		return nil, err
	}
	return mapper.UserDBToEntity(dbUser), nil
}

// FindByID finds a user by their ID.
func (r *userRepository) FindByID(ctx context.Context, id int64) (*entity.User, error) {
	// Note: This would require adding a GetUserByID query to sqlc
	// For now, we'll return ErrUserNotFound as a placeholder
	return nil, domainerrors.ErrUserNotFound
}

// FindByUsername finds a user by their username.
func (r *userRepository) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	dbUser, err := r.queries.GetUserByUserName(ctx, username)
	if err != nil {
		return nil, domainerrors.ErrUserNotFound
	}
	return mapper.UserDBToEntity(dbUser), nil
}

// FindByEmail finds a user by their email.
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	// Note: This would require adding a GetUserByEmail query to sqlc
	// For now, we'll return ErrUserNotFound as a placeholder
	return nil, domainerrors.ErrUserNotFound
}

// FindByUsernameOrEmail finds a user by username or email.
func (r *userRepository) FindByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (*entity.User, error) {
	dbUser, err := r.queries.GetUserByUsernameOrEmail(ctx, usernameOrEmail)
	if err != nil {
		return nil, domainerrors.ErrUserNotFound
	}
	return mapper.UserDBToEntity(dbUser), nil
}

// Update updates an existing user.
func (r *userRepository) Update(ctx context.Context, u *entity.User) (*entity.User, error) {
	// Note: This would require adding an UpdateUser query to sqlc
	// For now, we'll return ErrUserNotFound as a placeholder
	return nil, domainerrors.ErrUserNotFound
}

// Delete soft-deletes a user by their ID.
func (r *userRepository) Delete(ctx context.Context, id int64) error {
	// Note: This would require adding a DeleteUser (soft delete) query to sqlc
	// For now, we'll return ErrUserNotFound as a placeholder
	return domainerrors.ErrUserNotFound
}
