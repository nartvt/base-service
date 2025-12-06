package middleware

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
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
	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenExpired     = errors.New("token has expired")
	ErrMalformedToken   = errors.New("malformed token")
	ErrMissingToken     = errors.New("missing token")
	ErrInvalidSignature = errors.New("invalid token signature")
	ErrInvalidPassword  = errors.New("invalid password")
	Prefix              = "Bearer"
	RefreshTokenKeyName = "refreshToken"
	AccessTokenKeyName  = "accessToken"
	AuthorizationHeader = "Authorization"
	RefreshTokenHeader  = "RefreshToken"
)

// Argon2id parameters - based on OWASP recommendations for 2024
// These provide ~50-100ms hash time on modern hardware
const (
	argon2Time       = 3         // Number of iterations
	argon2Memory     = 64 * 1024 // 64 MB of memory
	argon2Threads    = 2         // Number of parallel threads
	argon2KeyLength  = 32        // Length of the derived key
	argon2SaltLength = 16        // Length of the random salt
)

type TokenPair struct {
	AccessToken  string    `json:"access_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	TokenType    string    `json:"token_type,omitempty"`
}

type Claims struct {
	UserId    int64  `json:"user_id,omitempty"`
	UserName  string `json:"username,omitempty"`
	TokenType string `json:"token_type,omitempty"` // "access" or "refresh"
	jwt.RegisteredClaims
}

type AuthenHandler struct {
	config   config.MiddlewareConfig
	jwtCache *JWTCache
}

func NewAuthenHandler(config config.MiddlewareConfig, jwtCache *JWTCache) *AuthenHandler {
	return &AuthenHandler{
		config:   config,
		jwtCache: jwtCache,
	}
}

func (a *AuthenHandler) GenerateAcessToken(userId int64, userName string) (*TokenPair, error) {
	accessScretConfig := a.config.Token.AccessTokenSecret
	accessExpireConfig := a.config.Token.AccessTokenExp
	accessToken, accessExp, err := a.generateToken(userId, userName, Prefix, accessScretConfig, accessExpireConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}
	return &TokenPair{
		AccessToken: accessToken,
		ExpiresAt:   accessExp,
		TokenType:   Prefix,
	}, nil
}

func (a *AuthenHandler) GenerateTokenPair(userId int64, userName string) (*TokenPair, error) {
	accessScretConfig := a.config.Token.AccessTokenSecret
	accessExpireConfig := a.config.Token.AccessTokenExp
	refreshScretConfig := a.config.Token.RefreshTokenSecret
	refreshExpireConfig := a.config.Token.RefreshTokenExp
	accessToken, accessExp, err := a.generateToken(userId, userName, Prefix, accessScretConfig, accessExpireConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, _, err := a.generateToken(userId, userName, Prefix, refreshScretConfig, refreshExpireConfig)
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

func (a *AuthenHandler) generateToken(
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

func (a *AuthenHandler) ValidateRefreshToken(tokenString string) (*Claims, error) {
	refreshSecretConfig := a.config.Token.RefreshTokenSecret
	return a.ValidateToken(tokenString, refreshSecretConfig, Prefix)
}

func (a *AuthenHandler) ValidateToken(tokenString, secretToken, expectedType string) (*Claims, error) {
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

func (a *AuthenHandler) AuthMiddleware() fiber.Handler {
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
		if a.jwtCache != nil && a.jwtCache.IsBlacklisted(ctx, tokenString) {
			slog.Warn("Blocked blacklisted token attempt",
				"ip", c.IP(),
				"path", c.Path(),
			)
			return a.handleError(c, errors.New("token has been revoked"))
		}

		// Check cache first for valid token
		if a.jwtCache != nil {
			if userID, found := a.jwtCache.GetCachedToken(ctx, tokenString); found {
				// Cache hit - use cached user ID
				claims := &Claims{
					UserId: userID,
					// Note: We only cache user ID for performance
					// Full claims will be from original token validation
				}
				c.Locals("user", claims)
				c.Locals("user_id", userID)

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
		if a.jwtCache != nil && claims.ExpiresAt != nil {
			_ = a.jwtCache.CacheValidToken(ctx, tokenString, claims.UserId, claims.ExpiresAt.Time)
		}

		c.Locals("user", claims)
		c.Locals("user_id", claims.UserId)
		c.Locals("username", claims.UserName)

		return c.Next()
	}
}

func (a *AuthenHandler) handleError(c *fiber.Ctx, err error) error {
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

func (a *AuthenHandler) RefreshToken(c *fiber.Ctx) error {
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

func (a *AuthenHandler) ExtractUserFromContext(c *fiber.Ctx) (*Claims, error) {
	user, ok := c.Locals("user").(*Claims)
	if !ok {
		return nil, errors.New("user not found in context")
	}
	return user, nil
}

func (r *AuthenHandler) GetUserNameFromContext(c *fiber.Ctx) (string, error) {
	userName := c.Locals("username")
	if userName == nil {
		return "", errors.New("user not found in context")
	}
	return userName.(string), nil
}

func (r *AuthenHandler) GetUserIdFromContext(c *fiber.Ctx) (int64, error) {
	userId := c.Locals("user_id")
	if userId == nil {
		return 0, errors.New("user not found in context")
	}
	return userId.(int64), nil
}

// Logout blacklists the current access token, effectively logging out the user
func (a *AuthenHandler) Logout(c *fiber.Ctx) error {
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
	if a.jwtCache != nil && claims.ExpiresAt != nil {
		return a.expireJwtCache(c, tokenString, claims)
	}

	// JWT caching is disabled - can't blacklist
	slog.Warn("Logout called but JWT caching is disabled",
		"user_id", claims.UserId,
	)

	return c.JSON(fiber.Map{
		"message": "Logout acknowledged (caching disabled, token will expire naturally)",
	})
}

func (a *AuthenHandler) expireJwtCache(c *fiber.Ctx, tokenString string, claims *Claims) error {
	ctx := c.Context()
	err := a.jwtCache.BlacklistToken(ctx, tokenString, claims.ExpiresAt.Time)
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
	_ = a.jwtCache.InvalidateToken(ctx, tokenString)

	slog.Info("User logged out successfully",
		"user_id", claims.UserId,
		"username", claims.UserName,
	)

	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

// HashPassword generates an argon2id hash of the password
// Returns: base64-encoded string in format: $argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>
func (s *AuthenHandler) HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	// Generate a cryptographically secure random salt
	salt := make([]byte, argon2SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Generate the hash using argon2id
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		argon2Time,
		argon2Memory,
		argon2Threads,
		argon2KeyLength,
	)

	// Encode salt and hash to base64
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Return in PHC string format for easy parsing and future-proofing
	// Format: $argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>
	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argon2Memory,
		argon2Time,
		argon2Threads,
		b64Salt,
		b64Hash,
	)

	return encodedHash, nil
}

// VerifyPassword verifies a password against an argon2id hash
// Returns true if the password matches, false otherwise
func (s *AuthenHandler) VerifyPassword(password, encodedHash string) (bool, error) {
	if password == "" {
		return false, errors.New("password cannot be empty")
	}
	if encodedHash == "" {
		return false, errors.New("hash cannot be empty")
	}

	// Parse the encoded hash to extract parameters, salt, and hash
	var version int
	var memory, time uint32
	var threads uint8
	var salt, hash string

	_, err := fmt.Sscanf(
		encodedHash,
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		&version,
		&memory,
		&time,
		&threads,
		&salt,
		&hash,
	)
	if err != nil {
		return false, fmt.Errorf("invalid hash format: %w", err)
	}

	// Decode salt and hash from base64
	decodedSalt, err := base64.RawStdEncoding.DecodeString(salt)
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(hash)
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	// Generate hash with the same parameters
	passwordHash := argon2.IDKey(
		[]byte(password),
		decodedSalt,
		time,
		memory,
		threads,
		uint32(len(decodedHash)),
	)

	// Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare(decodedHash, passwordHash) == 1 {
		return true, nil
	}

	return false, nil
}

func (s *AuthenHandler) GenerateGuestUsername(ctx context.Context, email string) (string, error) {
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
