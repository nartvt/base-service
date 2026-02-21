package middleware

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"base-service/config"
	"base-service/internal/common"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// =============================================================================
// Errors
// =============================================================================

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenExpired     = errors.New("token has expired")
	ErrMalformedToken   = errors.New("malformed token")
	ErrMissingToken     = errors.New("missing token")
	ErrInvalidSignature = errors.New("invalid token signature")
	ErrInvalidPassword  = errors.New("invalid password")
)

// =============================================================================
// Constants
// =============================================================================

const (
	Prefix              = "Bearer"
	RefreshTokenKeyName = "refreshToken"
	AccessTokenKeyName  = "accessToken"
	AuthorizationHeader = "Authorization"
	RefreshTokenHeader  = "RefreshToken"
)

// =============================================================================
// Token Types
// =============================================================================

// TokenPair represents access and refresh tokens.
type TokenPair struct {
	AccessToken  string    `json:"access_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	TokenType    string    `json:"token_type,omitempty"`
}

// Claims represents JWT token claims.
type Claims struct {
	UserId    int64  `json:"user_id,omitempty"`
	UserName  string `json:"username,omitempty"`
	TokenType string `json:"token_type,omitempty"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// =============================================================================
// AuthMiddleware Implementation
// clean-arch: Implements TokenService interface
// =============================================================================

// Compile-time interface compliance check
var _ TokenService = (*AuthMiddleware)(nil)

// AuthMiddleware handles JWT authentication and implements TokenService.
type AuthMiddleware struct {
	config         config.MiddlewareConfig
	tokenCache     TokenCache
	passwordHasher PasswordHasher
}

// NewAuthenHandler creates a new AuthMiddleware (backward compatible).
func NewAuthenHandler(config config.MiddlewareConfig, jwtCache *JWTCache) *AuthMiddleware {
	return &AuthMiddleware{
		config:         config,
		tokenCache:     jwtCache,
		passwordHasher: NewArgon2PasswordHasher(),
	}
}

// NewAuthMiddleware creates a new AuthMiddleware with explicit dependencies.
// clean-arch: Constructor with dependency injection
func NewAuthMiddleware(config config.MiddlewareConfig, tokenCache TokenCache, passwordHasher PasswordHasher) *AuthMiddleware {
	if passwordHasher == nil {
		passwordHasher = NewArgon2PasswordHasher()
	}
	return &AuthMiddleware{
		config:         config,
		tokenCache:     tokenCache,
		passwordHasher: passwordHasher,
	}
}

// =============================================================================
// TokenService Interface Implementation
// =============================================================================

// GenerateAccessToken generates a new access token (implements TokenService).
func (a *AuthMiddleware) GenerateAccessToken(userID int64, username string) (*TokenPair, error) {
	accessSecretConfig := a.config.Token.AccessTokenSecret
	accessExpireConfig := a.config.Token.AccessTokenExp
	accessToken, accessExp, err := a.generateToken(userID, username, Prefix, accessSecretConfig, accessExpireConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}
	return &TokenPair{
		AccessToken: accessToken,
		ExpiresAt:   accessExp,
		TokenType:   Prefix,
	}, nil
}

// GenerateTokenPair generates both access and refresh tokens (implements TokenService).
func (a *AuthMiddleware) GenerateTokenPair(userID int64, username string) (*TokenPair, error) {
	accessSecretConfig := a.config.Token.AccessTokenSecret
	accessExpireConfig := a.config.Token.AccessTokenExp
	refreshSecretConfig := a.config.Token.RefreshTokenSecret
	refreshExpireConfig := a.config.Token.RefreshTokenExp

	accessToken, accessExp, err := a.generateToken(userID, username, Prefix, accessSecretConfig, accessExpireConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, _, err := a.generateToken(userID, username, Prefix, refreshSecretConfig, refreshExpireConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExp,
		TokenType:    Prefix,
	}, nil
}

// ValidateAccessToken validates an access token (implements TokenService).
func (a *AuthMiddleware) ValidateAccessToken(token string) (*Claims, error) {
	return a.ValidateToken(token, a.config.Token.AccessTokenSecret, Prefix)
}

// ValidateRefreshToken validates a refresh token (implements TokenService).
func (a *AuthMiddleware) ValidateRefreshToken(tokenString string) (*Claims, error) {
	refreshSecretConfig := a.config.Token.RefreshTokenSecret
	return a.ValidateToken(tokenString, refreshSecretConfig, Prefix)
}

// =============================================================================
// Token Generation & Validation (Internal)
// =============================================================================

func (a *AuthMiddleware) generateToken(
	userId int64,
	username string,
	tokenType string,
	secretKey string,
	expiration time.Duration,
) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(expiration)

	claims := &Claims{
		UserId:    userId,
		UserName:  username,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   fmt.Sprintf("%d", userId),
			Issuer:    "client-side",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, expiresAt, nil
}

// ValidateToken validates a JWT token with the given secret and expected type.
func (a *AuthMiddleware) ValidateToken(tokenString, secretToken, expectedType string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSignature
		}
		return []byte(secretToken), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrMalformedToken
		}
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("invalid token type: expected %s, got %s", expectedType, claims.TokenType)
	}

	return claims, nil
}

// =============================================================================
// HTTP Middleware Handler
// =============================================================================

// AuthMiddleware returns a Fiber middleware handler for JWT authentication.
func (a *AuthMiddleware) AuthMiddleware() fiber.Handler {
	accessSecretConfig := a.config.Token.AccessTokenSecret
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		auth := c.Get(AuthorizationHeader)
		if auth == "" {
			return a.handleError(c, ErrMissingToken)
		}

		accessToken := strings.Split(auth, " ")
		if len(accessToken) != 2 || !strings.EqualFold(accessToken[0], Prefix) {
			return a.handleError(c, ErrMalformedToken)
		}

		tokenString := accessToken[1]

		// Check if token is blacklisted (logged out)
		if a.tokenCache != nil && a.tokenCache.IsEnabled() && a.tokenCache.IsBlacklisted(ctx, tokenString) {
			slog.Warn("Blocked blacklisted token attempt",
				"ip", c.IP(),
				"path", c.Path(),
			)
			return a.handleError(c, errors.New("token has been revoked"))
		}

		// Check cache first for valid token
		if a.tokenCache != nil && a.tokenCache.IsEnabled() {
			if userID, found := a.tokenCache.GetCachedToken(ctx, tokenString); found {
				// Cache hit - use cached user ID
				claims := &Claims{
					UserId: userID,
				}
				SetUserInContext(c, claims)

				slog.Debug("JWT cache hit, skipping validation",
					"user_id", userID,
					"path", c.Path(),
				)

				return c.Next()
			}
		}

		// Cache miss or caching disabled - validate token normally
		claims, err := a.ValidateToken(tokenString, accessSecretConfig, Prefix)
		if err != nil {
			return a.handleError(c, err)
		}

		// Cache the validated token
		if a.tokenCache != nil && a.tokenCache.IsEnabled() && claims.ExpiresAt != nil {
			_ = a.tokenCache.CacheToken(ctx, tokenString, claims.UserId, claims.ExpiresAt.Time)
		}

		// Use type-safe context helpers
		SetUserInContext(c, claims)

		return c.Next()
	}
}

func (a *AuthMiddleware) handleError(c *fiber.Ctx, err error) error {
	status := fiber.StatusUnauthorized
	message := "Authentication failed"

	switch {
	case errors.Is(err, ErrMissingToken):
		status = fiber.StatusBadRequest
		message = "Missing authentication token"
	case errors.Is(err, ErrMalformedToken):
		status = fiber.StatusBadRequest
		message = "Malformed authentication token"
	case errors.Is(err, ErrTokenExpired):
		message = "Token has expired"
	case errors.Is(err, ErrInvalidSignature):
		message = "Invalid token signature"
	case errors.Is(err, ErrInvalidToken):
		message = "Invalid token"
	}
	slog.Error(fmt.Sprintf("Status error: %d, message: %s", status, message))
	return common.ResponseApi(c, nil, err)
}

// =============================================================================
// Refresh Token Handler
// =============================================================================

// RefreshToken handles token refresh requests.
func (a *AuthMiddleware) RefreshToken(c *fiber.Ctx) error {
	refreshToken := c.Get(AuthorizationHeader)
	refreshSecretConfig := a.config.Token.RefreshTokenSecret
	if refreshToken == "" {
		return a.handleError(c, ErrMissingToken)
	}

	refreshToken = strings.TrimPrefix(refreshToken, Prefix)

	claims, err := a.ValidateToken(refreshToken, refreshSecretConfig, RefreshTokenKeyName)
	if err != nil {
		return a.handleError(c, err)
	}

	tokenPair, err := a.GenerateTokenPair(claims.UserId, claims.UserName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate new tokens",
		})
	}

	return c.JSON(tokenPair)
}

// =============================================================================
// Context Helpers (Backward Compatible)
// =============================================================================

// ExtractUserFromContext retrieves user claims from context (backward compatible).
func (a *AuthMiddleware) ExtractUserFromContext(c *fiber.Ctx) (*Claims, error) {
	claims, ok := GetUserFromContext(c)
	if !ok {
		return nil, errors.New("user not found in context")
	}
	return claims, nil
}

// GetUserNameFromContext retrieves username from context (backward compatible).
func (a *AuthMiddleware) GetUserNameFromContext(c *fiber.Ctx) (string, error) {
	username, ok := GetUsernameFromContext(c)
	if !ok {
		return "", errors.New("user not found in context")
	}
	return username, nil
}

// GetUserIdFromContext retrieves user ID from context (backward compatible).
func (a *AuthMiddleware) GetUserIdFromContext(c *fiber.Ctx) (int64, error) {
	userID, ok := GetUserIDFromContext(c)
	if !ok {
		return 0, errors.New("user not found in context")
	}
	return userID, nil
}

// =============================================================================
// Logout Handler
// =============================================================================

// Logout blacklists the current access token, effectively logging out the user.
func (a *AuthMiddleware) Logout(c *fiber.Ctx) error {
	// Extract token from Authorization header
	auth := c.Get(AuthorizationHeader)
	if auth == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "missing_token",
			"message": "No token provided",
		})
	}

	accessToken := strings.Split(auth, " ")
	if len(accessToken) != 2 || !strings.EqualFold(accessToken[0], Prefix) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_token_format",
			"message": "Invalid token format",
		})
	}

	tokenString := accessToken[1]

	// Validate token to get expiration time
	accessSecretConfig := a.config.Token.AccessTokenSecret
	claims, err := a.ValidateToken(tokenString, accessSecretConfig, Prefix)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "invalid_token",
			"message": "Invalid or expired token",
		})
	}

	// Blacklist the token if caching is enabled
	if a.tokenCache != nil && a.tokenCache.IsEnabled() && claims.ExpiresAt != nil {
		return a.expireTokenCache(c, tokenString, claims)
	}

	// JWT caching is disabled - can't blacklist
	slog.Warn("Logout called but JWT caching is disabled",
		"user_id", claims.UserId,
	)

	return c.JSON(fiber.Map{
		"message": "Logout acknowledged (caching disabled, token will expire naturally)",
	})
}

func (a *AuthMiddleware) expireTokenCache(c *fiber.Ctx, tokenString string, claims *Claims) error {
	ctx := c.Context()
	err := a.tokenCache.BlacklistToken(ctx, tokenString, claims.ExpiresAt.Time)
	if err != nil {
		slog.Error("Failed to blacklist token during logout",
			"error", err,
			"user_id", claims.UserId,
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "logout_failed",
			"message": "Failed to complete logout",
		})
	}

	// Also invalidate from valid token cache
	_ = a.tokenCache.InvalidateToken(ctx, tokenString)

	slog.Info("User logged out successfully",
		"user_id", claims.UserId,
		"username", claims.UserName,
	)

	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

// =============================================================================
// Password Methods (Delegates to PasswordHasher)
// =============================================================================

// HashPassword generates an argon2id hash of the password (backward compatible).
func (a *AuthMiddleware) HashPassword(password string) (string, error) {
	return a.passwordHasher.Hash(password)
}

// VerifyPassword verifies a password against an argon2id hash (backward compatible).
func (a *AuthMiddleware) VerifyPassword(password, encodedHash string) (bool, error) {
	return a.passwordHasher.Verify(password, encodedHash)
}

// PasswordHasher returns the password hasher for direct access.
func (a *AuthMiddleware) PasswordHasher() PasswordHasher {
	return a.passwordHasher
}

// =============================================================================
// Username Generator
// =============================================================================

// GenerateGuestUsername creates a unique username from email.
func (a *AuthMiddleware) GenerateGuestUsername(ctx context.Context, email string) (string, error) {
	parts := strings.Split(email, "@")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid email format")
	}

	usernamePrefix := parts[0]

	re := regexp.MustCompile(`[^a-zA-Z0-9]`)
	usernamePrefix = re.ReplaceAllString(usernamePrefix, "")

	if len(usernamePrefix) < 3 {
		usernamePrefix = usernamePrefix + "1"
	}

	guestUsername := fmt.Sprintf("guest_%s%d", usernamePrefix, time.Now().Unix())
	return guestUsername, nil
}

// =============================================================================
// Legacy Method Aliases (Backward Compatibility)
// =============================================================================

// GenerateAcessToken is an alias for GenerateAccessToken (fixes typo, backward compatible).
func (a *AuthMiddleware) GenerateAcessToken(userId int64, userName string) (*TokenPair, error) {
	return a.GenerateAccessToken(userId, userName)
}
