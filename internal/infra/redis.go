package infra

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"base-service/config"

	"github.com/redis/go-redis/v9"
)

const (
	redisName                 = "redis"
	defaultRedisHealthTimeout = 3 * time.Second
)

// Compile-time interface compliance check
var (
	_ Connection = (*RedisClient)(nil)
	_ CacheStore = (*RedisClient)(nil)
)

// RedisClient wraps redis.Client and implements Connection and CacheStore interfaces.
// clean-arch: Infrastructure adapter implementing domain ports
type RedisClient struct {
	client *redis.Client
	config *config.RedisConfig
}

// NewRedisClient creates a new RedisClient with the given configuration.
func NewRedisClient(rd *config.RedisConfig) (*RedisClient, error) {
	if rd == nil {
		return nil, fmt.Errorf("redis config is required")
	}

	client, err := initRedis(rd)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize redis: %w", err)
	}

	return &RedisClient{
		client: client,
		config: rd,
	}, nil
}

func initRedis(rd *config.RedisConfig) (*redis.Client, error) {
	opt, err := redis.ParseURL(rd.BuildRedisConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	opt.PoolSize = rd.MaxIdle
	opt.MinIdleConns = rd.MinIdle
	opt.MaxIdleConns = rd.MaxIdle
	opt.MaxActiveConns = rd.MaxIdle
	opt.DialTimeout = rd.DialTimeout
	opt.ReadTimeout = rd.ReadTimeout
	opt.WriteTimeout = rd.WriteTimeout
	opt.Password = rd.Password
	opt.DB = rd.DB

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), defaultRedisHealthTimeout)
	defer cancel()

	pong, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	slog.Info("Redis connected", "response", pong)
	return client, nil
}

// =============================================================================
// Connection Interface Implementation
// =============================================================================

// Name returns the identifier for this connection.
func (r *RedisClient) Name() string {
	return redisName
}

// Ping checks if the Redis connection is alive.
func (r *RedisClient) Ping(ctx context.Context) error {
	if r.client == nil {
		return fmt.Errorf("redis client is nil")
	}
	return r.client.Ping(ctx).Err()
}

// IsHealthy performs a health check with timeout.
func (r *RedisClient) IsHealthy(ctx context.Context) bool {
	if r.client == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(ctx, defaultRedisHealthTimeout)
	defer cancel()

	if err := r.client.Ping(ctx).Err(); err != nil {
		return false
	}
	return true
}

// Close closes the Redis connection.
func (r *RedisClient) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// =============================================================================
// CacheStore Interface Implementation
// =============================================================================

// Get retrieves a value from Redis by key.
func (r *RedisClient) Get(ctx context.Context, key string) ([]byte, error) {
	if r.client == nil {
		return nil, fmt.Errorf("redis client is nil")
	}

	val, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // Key not found, return nil without error
	}
	if err != nil {
		return nil, fmt.Errorf("redis get failed: %w", err)
	}
	return val, nil
}

// Set stores a value in Redis with the given TTL.
func (r *RedisClient) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if r.client == nil {
		return fmt.Errorf("redis client is nil")
	}

	if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}
	return nil
}

// Delete removes a key from Redis.
func (r *RedisClient) Delete(ctx context.Context, key string) error {
	if r.client == nil {
		return fmt.Errorf("redis client is nil")
	}

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis delete failed: %w", err)
	}
	return nil
}

// Exists checks if a key exists in Redis.
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	if r.client == nil {
		return false, fmt.Errorf("redis client is nil")
	}

	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists failed: %w", err)
	}
	return result > 0, nil
}

// =============================================================================
// Additional Methods (for backward compatibility and advanced operations)
// =============================================================================

// Redis returns the underlying redis.Client.
// NOTE: Use this only when you need direct access to redis features.
// Prefer using the interface methods for better abstraction.
func (r *RedisClient) Redis() *redis.Client {
	return r.client
}

// SetNX sets a value only if the key does not exist (for distributed locks).
func (r *RedisClient) SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	if r.client == nil {
		return false, fmt.Errorf("redis client is nil")
	}

	result, err := r.client.SetNX(ctx, key, value, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("redis setnx failed: %w", err)
	}
	return result, nil
}

// Incr increments the integer value of a key by one.
func (r *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	if r.client == nil {
		return 0, fmt.Errorf("redis client is nil")
	}

	result, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("redis incr failed: %w", err)
	}
	return result, nil
}

// Expire sets a timeout on a key.
func (r *RedisClient) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if r.client == nil {
		return fmt.Errorf("redis client is nil")
	}

	if err := r.client.Expire(ctx, key, ttl).Err(); err != nil {
		return fmt.Errorf("redis expire failed: %w", err)
	}
	return nil
}

// TTL returns the remaining time to live of a key.
func (r *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	if r.client == nil {
		return 0, fmt.Errorf("redis client is nil")
	}

	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("redis ttl failed: %w", err)
	}
	return ttl, nil
}
