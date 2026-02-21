package middleware

import (
	"context"
	"log/slog"
	"sync"

	"base-service/config"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// =============================================================================
// Middleware Registry
// clean-arch: Centralized registry for all middleware components
// =============================================================================

// Registry provides unified access to all middleware components.
// It manages middleware dependencies and provides easy access to services.
type Registry struct {
	mu             sync.RWMutex
	config         config.MiddlewareConfig
	tokenService   TokenService
	passwordHasher PasswordHasher
	tokenCache     TokenCache
	authMiddleware *AuthMiddleware
	handlers       map[string]MiddlewareHandler
}

// NewRegistry creates a new middleware registry with the given configuration.
func NewRegistry(cfg config.MiddlewareConfig) *Registry {
	return &Registry{
		config:   cfg,
		handlers: make(map[string]MiddlewareHandler),
	}
}

// =============================================================================
// Initialization Methods
// =============================================================================

// InitWithRedis initializes the registry with Redis-backed caching.
func (r *Registry) InitWithRedis(redisClient *redis.Client, cacheEnabled bool) *Registry {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Create token cache
	r.tokenCache = NewJWTCache(redisClient, cacheEnabled)

	// Create password hasher
	r.passwordHasher = NewArgon2PasswordHasher()

	// Create auth middleware with all dependencies
	r.authMiddleware = NewAuthMiddleware(r.config, r.tokenCache, r.passwordHasher)
	r.tokenService = r.authMiddleware

	slog.Info("Middleware registry initialized",
		"cache_enabled", cacheEnabled,
		"has_redis", redisClient != nil,
	)

	return r
}

// InitWithoutCache initializes the registry without caching.
func (r *Registry) InitWithoutCache() *Registry {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Create password hasher
	r.passwordHasher = NewArgon2PasswordHasher()

	// Create auth middleware without cache
	r.authMiddleware = NewAuthMiddleware(r.config, nil, r.passwordHasher)
	r.tokenService = r.authMiddleware

	slog.Info("Middleware registry initialized without cache")

	return r
}

// =============================================================================
// Accessor Methods
// =============================================================================

// TokenService returns the token service.
func (r *Registry) TokenService() TokenService {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.tokenService
}

// PasswordHasher returns the password hasher.
func (r *Registry) PasswordHasher() PasswordHasher {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.passwordHasher
}

// TokenCache returns the token cache.
func (r *Registry) TokenCache() TokenCache {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.tokenCache
}

// AuthMiddleware returns the auth middleware instance.
func (r *Registry) AuthMiddleware() *AuthMiddleware {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.authMiddleware
}

// Config returns the middleware configuration.
func (r *Registry) Config() config.MiddlewareConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.config
}

// =============================================================================
// Handler Methods
// =============================================================================

// AuthHandler returns the Fiber authentication middleware handler.
func (r *Registry) AuthHandler() fiber.Handler {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.authMiddleware == nil {
		// Return a pass-through handler if not initialized
		return func(c *fiber.Ctx) error {
			slog.Warn("Auth middleware not initialized, allowing request")
			return c.Next()
		}
	}

	return r.authMiddleware.AuthMiddleware()
}

// RegisterHandler registers a custom middleware handler.
func (r *Registry) RegisterHandler(name string, handler MiddlewareHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.handlers[name] = handler
	slog.Info("Registered middleware handler", "name", name)
}

// GetHandler retrieves a registered middleware handler by name.
func (r *Registry) GetHandler(name string) (MiddlewareHandler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	handler, ok := r.handlers[name]
	return handler, ok
}

// =============================================================================
// Health Check Methods
// =============================================================================

// HealthStatus represents the health status of middleware components.
type MiddlewareHealthStatus struct {
	TokenCache     string `json:"token_cache"`
	AuthMiddleware string `json:"auth_middleware"`
	Handlers       int    `json:"handlers_count"`
}

// HealthCheck returns the health status of middleware components.
func (r *Registry) HealthCheck(ctx context.Context) MiddlewareHealthStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()

	status := MiddlewareHealthStatus{
		TokenCache:     "disabled",
		AuthMiddleware: "not_initialized",
		Handlers:       len(r.handlers),
	}

	if r.tokenCache != nil && r.tokenCache.IsEnabled() {
		status.TokenCache = "enabled"
	}

	if r.authMiddleware != nil {
		status.AuthMiddleware = "initialized"
	}

	return status
}

// =============================================================================
// Statistics Methods
// =============================================================================

// Stats returns statistics about the middleware registry.
type RegistryStats struct {
	CacheEnabled   bool           `json:"cache_enabled"`
	HandlersCount  int            `json:"handlers_count"`
	TokenCacheStats map[string]int64 `json:"token_cache_stats,omitempty"`
}

// Stats returns statistics about the middleware registry.
func (r *Registry) Stats(ctx context.Context) RegistryStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := RegistryStats{
		CacheEnabled:  r.tokenCache != nil && r.tokenCache.IsEnabled(),
		HandlersCount: len(r.handlers),
	}

	// Get token cache stats if available
	if jwtCache, ok := r.tokenCache.(*JWTCache); ok {
		cacheStats, err := jwtCache.GetCacheStats(ctx)
		if err == nil {
			stats.TokenCacheStats = cacheStats
		}
	}

	return stats
}

// =============================================================================
// Convenience Methods
// =============================================================================

// HashPassword hashes a password using the registered password hasher.
func (r *Registry) HashPassword(password string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.passwordHasher == nil {
		return "", nil
	}
	return r.passwordHasher.Hash(password)
}

// VerifyPassword verifies a password using the registered password hasher.
func (r *Registry) VerifyPassword(password, hash string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.passwordHasher == nil {
		return false, nil
	}
	return r.passwordHasher.Verify(password, hash)
}

// GenerateTokenPair generates a token pair using the token service.
func (r *Registry) GenerateTokenPair(userID int64, username string) (*TokenPair, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.tokenService == nil {
		return nil, nil
	}
	return r.tokenService.GenerateTokenPair(userID, username)
}

// ValidateAccessToken validates an access token using the token service.
func (r *Registry) ValidateAccessToken(token string) (*Claims, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.tokenService == nil {
		return nil, nil
	}
	return r.tokenService.ValidateAccessToken(token)
}
