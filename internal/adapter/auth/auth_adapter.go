package auth

import (
	"base-service/internal/middleware"
	"base-service/internal/usecase/port"
)

// AuthAdapter wraps the existing AuthMiddleware to implement usecase interfaces.
type AuthAdapter struct {
	authen *middleware.AuthMiddleware
}

// NewAuthAdapter creates a new auth adapter.
func NewAuthAdapter(authen *middleware.AuthMiddleware) *AuthAdapter {
	return &AuthAdapter{authen: authen}
}

// HashPassword implements auth.PasswordHasher.
func (a *AuthAdapter) HashPassword(password string) (string, error) {
	return a.authen.HashPassword(password)
}

// VerifyPassword implements auth.PasswordHasher.
func (a *AuthAdapter) VerifyPassword(password, hash string) (bool, error) {
	return a.authen.VerifyPassword(password, hash)
}

// GenerateTokenPair implements auth.TokenGenerator.
func (a *AuthAdapter) GenerateTokenPair(userID int64, username string) (*port.TokenPair, error) {
	pair, err := a.authen.GenerateTokenPair(userID, username)
	if err != nil {
		return nil, err
	}
	return &port.TokenPair{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	}, nil
}

// GenerateAccessToken implements auth.TokenGenerator.
func (a *AuthAdapter) GenerateAccessToken(userID int64, username string) (*port.TokenPair, error) {
	pair, err := a.authen.GenerateAcessToken(userID, username)
	if err != nil {
		return nil, err
	}
	return &port.TokenPair{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	}, nil
}

// ValidateRefreshToken implements auth.TokenGenerator.
func (a *AuthAdapter) ValidateRefreshToken(token string) (int64, string, error) {
	claims, err := a.authen.ValidateRefreshToken(token)
	if err != nil {
		return 0, "", err
	}
	return claims.UserId, claims.UserName, nil
}

// InvalidateToken implements auth.TokenGenerator.
// Note: This is a no-op as the logout is handled directly by the middleware.
func (a *AuthAdapter) InvalidateToken(token string) error {
	// Token invalidation is handled by the middleware's Logout method
	// which requires fiber.Ctx. For use case layer, we return nil.
	return nil
}
