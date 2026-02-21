package middleware

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

// =============================================================================
// Middleware Contracts (Ports)
// clean-arch: These interfaces define contracts for middleware components
// =============================================================================

// TokenService defines the contract for JWT token operations.
// clean-arch: Port interface for token generation and validation
type TokenService interface {
	// GenerateAccessToken generates a new access token
	GenerateAccessToken(userID int64, username string) (*TokenPair, error)
	// GenerateTokenPair generates both access and refresh tokens
	GenerateTokenPair(userID int64, username string) (*TokenPair, error)
	// ValidateAccessToken validates an access token and returns claims
	ValidateAccessToken(token string) (*Claims, error)
	// ValidateRefreshToken validates a refresh token and returns claims
	ValidateRefreshToken(token string) (*Claims, error)
}

// PasswordHasher defines the contract for password hashing operations.
// clean-arch: Port interface for password security (separates from token logic)
type PasswordHasher interface {
	// Hash generates a secure hash of the password
	Hash(password string) (string, error)
	// Verify checks if a password matches the hash
	Verify(password, hash string) (bool, error)
}

// TokenCache defines the contract for token caching and blacklisting.
// clean-arch: Port interface for token persistence (Redis, memory, etc.)
type TokenCache interface {
	// CacheToken caches a validated token
	CacheToken(ctx context.Context, token string, userID int64, expiresAt time.Time) error
	// GetCachedToken retrieves cached token data
	GetCachedToken(ctx context.Context, token string) (userID int64, found bool)
	// InvalidateToken removes a token from cache
	InvalidateToken(ctx context.Context, token string) error
	// BlacklistToken adds a token to blacklist (for logout)
	BlacklistToken(ctx context.Context, token string, expiresAt time.Time) error
	// IsBlacklisted checks if a token is blacklisted
	IsBlacklisted(ctx context.Context, token string) bool
	// IsEnabled returns whether caching is enabled
	IsEnabled() bool
}

// RateLimiter defines the contract for rate limiting.
// clean-arch: Port interface for rate limiting strategies
type RateLimiter interface {
	// Allow checks if a request is allowed
	Allow(ctx context.Context, key string) (allowed bool, remaining int, resetAt time.Time)
	// Reset resets the rate limit for a key
	Reset(ctx context.Context, key string) error
}

// =============================================================================
// Middleware Handler Contract
// =============================================================================

// MiddlewareHandler defines the contract for middleware that can be applied to routes.
type MiddlewareHandler interface {
	// Handler returns the Fiber handler function
	Handler() fiber.Handler
	// Name returns the middleware name for logging
	Name() string
	// IsEnabled returns whether the middleware is enabled
	IsEnabled() bool
}

// =============================================================================
// Auth Context Keys (Type-Safe)
// =============================================================================

// authContextKey is a custom type for auth context keys to avoid collisions.
type authContextKey int

const (
	// userClaimsKey stores the authenticated user claims
	userClaimsKey authContextKey = iota
	// userIDKey stores the user ID
	userIDKey
	// usernameKey stores the username
	usernameKey
)

// =============================================================================
// Context Helpers for Auth
// =============================================================================

// SetUserInContext stores user information in fiber context.
func SetUserInContext(c *fiber.Ctx, claims *Claims) {
	c.Locals(userClaimsKey, claims)
	c.Locals(userIDKey, claims.UserId)
	c.Locals(usernameKey, claims.UserName)
	// Backward compatibility with string keys
	c.Locals("user", claims)
	c.Locals("user_id", claims.UserId)
	c.Locals("username", claims.UserName)
}

// GetUserFromContext retrieves user claims from fiber context.
func GetUserFromContext(c *fiber.Ctx) (*Claims, bool) {
	// Try typed key first
	if claims, ok := c.Locals(userClaimsKey).(*Claims); ok {
		return claims, true
	}
	// Fallback to string key
	if claims, ok := c.Locals("user").(*Claims); ok {
		return claims, true
	}
	return nil, false
}

// GetUserIDFromContext retrieves user ID from fiber context.
func GetUserIDFromContext(c *fiber.Ctx) (int64, bool) {
	// Try typed key first
	if userID, ok := c.Locals(userIDKey).(int64); ok {
		return userID, true
	}
	// Fallback to string key
	if userID, ok := c.Locals("user_id").(int64); ok {
		return userID, true
	}
	return 0, false
}

// GetUsernameFromContext retrieves username from fiber context.
func GetUsernameFromContext(c *fiber.Ctx) (string, bool) {
	// Try typed key first
	if username, ok := c.Locals(usernameKey).(string); ok {
		return username, true
	}
	// Fallback to string key
	if username, ok := c.Locals("username").(string); ok {
		return username, true
	}
	return "", false
}

// =============================================================================
// Username Generator Contract (for guest users)
// =============================================================================

// UsernameGenerator defines the contract for generating usernames.
type UsernameGenerator interface {
	// Generate creates a unique username from email
	Generate(ctx context.Context, email string) (string, error)
}
