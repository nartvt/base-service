package port

import (
	"context"

	"base-service/internal/domain/entity"
)

// RegisterInput represents input for user registration.
type RegisterInput struct {
	Username    string
	Email       string
	PhoneNumber string
	FirstName   string
	LastName    string
	Password    string
}

// LoginInput represents input for user login.
type LoginInput struct {
	UsernameOrEmail string
	Password        string
}

// RegisterOutput represents output from user registration.
type RegisterOutput struct {
	User         *entity.User
	AccessToken  string
	RefreshToken string
}

// LoginOutput represents output from user login.
type LoginOutput struct {
	User         *entity.User
	AccessToken  string
	RefreshToken string
}

// TokenPair represents a pair of access and refresh tokens.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// AuthUseCase defines the interface for authentication operations.
type AuthUseCase interface {
	// Register creates a new user account.
	Register(ctx context.Context, input *RegisterInput) (*RegisterOutput, error)

	// Login authenticates a user and returns tokens.
	Login(ctx context.Context, input *LoginInput) (*LoginOutput, error)

	// RefreshToken generates a new access token using a refresh token.
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)

	// Logout invalidates the user's tokens.
	Logout(ctx context.Context, accessToken string) error
}

// UserUseCase defines the interface for user operations.
type UserUseCase interface {
	// GetProfile returns the user's profile by username.
	GetProfile(ctx context.Context, username string) (*entity.User, error)

	// GetByID returns a user by their ID.
	GetByID(ctx context.Context, id int64) (*entity.User, error)

	// Update updates a user's profile.
	Update(ctx context.Context, user *entity.User) (*entity.User, error)
}
