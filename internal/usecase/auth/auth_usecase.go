package auth

import (
	"context"
	"errors"

	"base-service/internal/domain/entity"
	domainerrors "base-service/internal/domain/errors"
	"base-service/internal/domain/repository"
	"base-service/internal/usecase/port"
)

// PasswordHasher defines the interface for password hashing operations.
type PasswordHasher interface {
	HashPassword(password string) (string, error)
	VerifyPassword(password, hash string) (bool, error)
}

// TokenGenerator defines the interface for JWT token operations.
type TokenGenerator interface {
	GenerateTokenPair(userID int64, username string) (*port.TokenPair, error)
	GenerateAccessToken(userID int64, username string) (*port.TokenPair, error)
	ValidateRefreshToken(token string) (userID int64, username string, err error)
	InvalidateToken(token string) error
}

type authUseCase struct {
	userRepo       repository.UserRepository
	passwordHasher PasswordHasher
	tokenGenerator TokenGenerator
}

// NewAuthUseCase creates a new authentication use case.
func NewAuthUseCase(
	userRepo repository.UserRepository,
	passwordHasher PasswordHasher,
	tokenGenerator TokenGenerator,
) port.AuthUseCase {
	return &authUseCase{
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
		tokenGenerator: tokenGenerator,
	}
}

// Register creates a new user account.
func (uc *authUseCase) Register(ctx context.Context, input *port.RegisterInput) (*port.RegisterOutput, error) {
	// Hash the password
	hashedPassword, err := uc.passwordHasher.HashPassword(input.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create domain entity
	user := &entity.User{
		Username:     input.Username,
		Email:        input.Email,
		PhoneNumber:  input.PhoneNumber,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		HashPassword: hashedPassword,
	}

	// Persist user
	createdUser, err := uc.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	// Generate tokens
	tokenPair, err := uc.tokenGenerator.GenerateTokenPair(createdUser.ID, createdUser.Username)
	if err != nil {
		return nil, err
	}

	return &port.RegisterOutput{
		User:         createdUser,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// Login authenticates a user and returns tokens.
func (uc *authUseCase) Login(ctx context.Context, input *port.LoginInput) (*port.LoginOutput, error) {
	// Find user by username or email
	user, err := uc.userRepo.FindByUsernameOrEmail(ctx, input.UsernameOrEmail)
	if err != nil {
		return nil, domainerrors.ErrInvalidCredentials
	}

	// Verify password
	valid, err := uc.passwordHasher.VerifyPassword(input.Password, user.HashPassword)
	if err != nil || !valid {
		return nil, domainerrors.ErrInvalidCredentials
	}

	// Generate tokens
	tokenPair, err := uc.tokenGenerator.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	return &port.LoginOutput{
		User:         user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// RefreshToken generates a new access token using a refresh token.
func (uc *authUseCase) RefreshToken(ctx context.Context, refreshToken string) (*port.TokenPair, error) {
	userID, username, err := uc.tokenGenerator.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	tokenPair, err := uc.tokenGenerator.GenerateAccessToken(userID, username)
	if err != nil {
		return nil, err
	}

	// Keep the same refresh token
	tokenPair.RefreshToken = refreshToken
	return tokenPair, nil
}

// Logout invalidates the user's tokens.
func (uc *authUseCase) Logout(ctx context.Context, accessToken string) error {
	return uc.tokenGenerator.InvalidateToken(accessToken)
}
