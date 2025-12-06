package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	jwtValidKeyPrefix     = "jwt:valid:%s"
	jwtBlacklistKeyPrefix = "jwt:blacklist:%s"
)

// JWTCache handles caching and blacklisting of JWT tokens using Redis
type JWTCache struct {
	redis   *redis.Client
	enabled bool
}

// NewJWTCache creates a new JWT cache instance
func NewJWTCache(redisClient *redis.Client, enabled bool) *JWTCache {
	if redisClient == nil || !enabled {
		slog.Info("JWT caching is disabled")
		return &JWTCache{enabled: false}
	}

	slog.Info("JWT caching is enabled")
	return &JWTCache{
		redis:   redisClient,
		enabled: enabled,
	}
}

// hashToken creates a SHA256 hash of the token for use as cache key
// This prevents storing the actual token in Redis
func (c *JWTCache) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// IsBlacklisted checks if a token is blacklisted (logged out)
func (c *JWTCache) IsBlacklisted(ctx context.Context, token string) bool {
	if !c.enabled || c.redis == nil {
		return false
	}

	tokenHash := c.hashToken(token)
	key := fmt.Sprintf(jwtBlacklistKeyPrefix, tokenHash)

	exists, err := c.redis.Exists(ctx, key).Result()
	if err != nil {
		slog.Error("Failed to check token blacklist",
			"error", err,
			"key", key,
		)
		return false // Fail open - allow request if Redis is down
	}

	isBlacklisted := exists > 0
	if isBlacklisted {
		slog.Warn("Blocked blacklisted token",
			"token_hash", tokenHash,
		)
	}

	return isBlacklisted
}

// BlacklistToken adds a token to the blacklist (for logout)
// The token will be blacklisted until its natural expiration
func (c *JWTCache) BlacklistToken(ctx context.Context, token string, expiresAt time.Time) error {
	if !c.enabled || c.redis == nil {
		slog.Warn("JWT caching is disabled, cannot blacklist token")
		return nil
	}

	tokenHash := c.hashToken(token)
	key := fmt.Sprintf(jwtBlacklistKeyPrefix, tokenHash)

	// Calculate TTL: time until token naturally expires
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		slog.Info("Token already expired, no need to blacklist",
			"token_hash", tokenHash,
		)
		return nil
	}

	// Store in Redis with TTL matching token expiration
	err := c.redis.Set(ctx, key, "1", ttl).Err()
	if err != nil {
		slog.Error("Failed to blacklist token",
			"error", err,
			"key", key,
			"ttl", ttl,
		)
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	slog.Info("Token blacklisted successfully",
		"token_hash", tokenHash,
		"ttl", ttl,
	)

	return nil
}

// CacheValidToken caches a validated token to avoid repeated validation
// Stores minimal data: just the user ID and expiration
func (c *JWTCache) CacheValidToken(ctx context.Context, token string, userID int64, expiresAt time.Time) error {
	if !c.enabled || c.redis == nil {
		return nil
	}

	tokenHash := c.hashToken(token)
	key := fmt.Sprintf(jwtValidKeyPrefix, tokenHash)

	// Calculate TTL
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil // Don't cache expired tokens
	}

	// Store user ID with TTL matching token expiration
	err := c.redis.Set(ctx, key, userID, ttl).Err()
	if err != nil {
		slog.Error("Failed to cache valid token",
			"error", err,
			"key", key,
		)
		return err
	}

	slog.Debug("Token cached successfully",
		"token_hash", tokenHash,
		"user_id", userID,
		"ttl", ttl,
	)

	return nil
}

// GetCachedToken retrieves a cached token's user ID
// Returns 0 and false if not cached or cache miss
func (c *JWTCache) GetCachedToken(ctx context.Context, token string) (int64, bool) {
	if !c.enabled || c.redis == nil {
		return 0, false
	}

	tokenHash := c.hashToken(token)
	key := fmt.Sprintf(jwtValidKeyPrefix, tokenHash)

	userID, err := c.redis.Get(ctx, key).Int64()
	if err == redis.Nil {
		// Cache miss - not an error
		return 0, false
	}
	if err != nil {
		slog.Error("Failed to get cached token",
			"error", err,
			"key", key,
		)
		return 0, false
	}

	slog.Debug("Token cache hit",
		"token_hash", tokenHash,
		"user_id", userID,
	)

	return userID, true
}

// InvalidateToken removes a token from the valid cache
// Used when token should no longer be considered valid
func (c *JWTCache) InvalidateToken(ctx context.Context, token string) error {
	if !c.enabled || c.redis == nil {
		return nil
	}

	tokenHash := c.hashToken(token)
	key := fmt.Sprintf(jwtValidKeyPrefix, tokenHash)

	err := c.redis.Del(ctx, key).Err()
	if err != nil {
		slog.Error("Failed to invalidate token cache",
			"error", err,
			"key", key,
		)
		return err
	}

	slog.Debug("Token cache invalidated",
		"token_hash", tokenHash,
	)

	return nil
}

// GetCacheStats returns statistics about the JWT cache
func (c *JWTCache) GetCacheStats(ctx context.Context) (map[string]int64, error) {
	if !c.enabled || c.redis == nil {
		return map[string]int64{
			"enabled": 0,
		}, nil
	}

	// Count valid tokens
	validKeys, err := c.redis.Keys(ctx, "jwt:valid:*").Result()
	if err != nil {
		return nil, err
	}

	// Count blacklisted tokens
	blacklistKeys, err := c.redis.Keys(ctx, "jwt:blacklist:*").Result()
	if err != nil {
		return nil, err
	}

	return map[string]int64{
		"enabled":            1,
		"valid_tokens":       int64(len(validKeys)),
		"blacklisted_tokens": int64(len(blacklistKeys)),
	}, nil
}
