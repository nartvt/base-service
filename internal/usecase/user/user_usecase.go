package user

import (
	"context"

	"base-service/internal/domain/entity"
	domainerrors "base-service/internal/domain/errors"
	"base-service/internal/domain/repository"
	"base-service/internal/usecase/port"
)

type userUseCase struct {
	userRepo repository.UserRepository
}

// NewUserUseCase creates a new user use case.
func NewUserUseCase(userRepo repository.UserRepository) port.UserUseCase {
	return &userUseCase{
		userRepo: userRepo,
	}
}

// GetProfile returns the user's profile by username.
func (uc *userUseCase) GetProfile(ctx context.Context, username string) (*entity.User, error) {
	user, err := uc.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, domainerrors.ErrUserNotFound
	}

	if user.IsDeleted() {
		return nil, domainerrors.ErrUserDeleted
	}

	return user, nil
}

// GetByID returns a user by their ID.
func (uc *userUseCase) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, domainerrors.ErrUserNotFound
	}

	if user.IsDeleted() {
		return nil, domainerrors.ErrUserDeleted
	}

	return user, nil
}

// Update updates a user's profile.
func (uc *userUseCase) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	// Verify user exists
	existing, err := uc.userRepo.FindByID(ctx, user.ID)
	if err != nil {
		return nil, domainerrors.ErrUserNotFound
	}

	if existing.IsDeleted() {
		return nil, domainerrors.ErrUserDeleted
	}

	return uc.userRepo.Update(ctx, user)
}
