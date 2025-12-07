package repository

import (
	"context"

	"base-service/internal/domain/entity"
)

// UserRepository defines the interface for user persistence operations.
// This is a port interface that the infrastructure layer must implement.
type UserRepository interface {
	// Create creates a new user and returns the created user with ID.
	Create(ctx context.Context, user *entity.User) (*entity.User, error)

	// FindByID finds a user by their ID.
	FindByID(ctx context.Context, id int64) (*entity.User, error)

	// FindByUsername finds a user by their username.
	FindByUsername(ctx context.Context, username string) (*entity.User, error)

	// FindByEmail finds a user by their email.
	FindByEmail(ctx context.Context, email string) (*entity.User, error)

	// FindByUsernameOrEmail finds a user by username or email.
	FindByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (*entity.User, error)

	// Update updates an existing user.
	Update(ctx context.Context, user *entity.User) (*entity.User, error)

	// Delete soft-deletes a user by their ID.
	Delete(ctx context.Context, id int64) error
}
