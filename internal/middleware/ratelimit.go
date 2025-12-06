package middleware

import (
	"fmt"
	"log/slog"
	"time"

	"base-service/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// RateLimitFilter creates a standard rate limiter middleware for general API endpoints
// Default: 100 requests per minute per IP
func RateLimitFilter(rateLimitConfig config.RateLimitConfig) fiber.Handler {
	if !rateLimitConfig.Enabled {
		slog.Info("Rate limiting is disabled")
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	// Set defaults if not configured
	max := rateLimitConfig.Max
	if max == 0 {
		max = 100 // 100 requests per minute by default
	}

	expiration := rateLimitConfig.Expiration
	if expiration == 0 {
		expiration = 1 * time.Minute
	}

	limitReached := rateLimitConfig.LimitReached
	if limitReached == "" {
		limitReached = "Too many requests, please try again later."
	}

	slog.Info("Configuring rate limiting middleware",
		"max", max,
		"expiration", expiration,
		"skipFailedReq", rateLimitConfig.SkipFailedReq,
	)

	return limiter.New(limiter.Config{
		Max:                    max,
		Expiration:             expiration,
		SkipFailedRequests:     rateLimitConfig.SkipFailedReq,
		SkipSuccessfulRequests: rateLimitConfig.SkipSuccessReq,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Use IP address as the key
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			slog.Warn("Rate limit exceeded",
				"ip", c.IP(),
				"path", c.Path(),
				"method", c.Method(),
			)
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "rate_limit_exceeded",
				"message": limitReached,
			})
		},
	})
}

// AuthRateLimitFilter creates a stricter rate limiter for authentication endpoints
// Default: 5 requests per minute per IP (to prevent brute force attacks)
func AuthRateLimitFilter(rateLimitConfig config.RateLimitConfig) fiber.Handler {
	if !rateLimitConfig.AuthEnabled {
		slog.Info("Auth rate limiting is disabled")
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	// Set stricter defaults for auth endpoints
	max := rateLimitConfig.AuthMax
	if max == 0 {
		max = 5 // Only 5 login attempts per minute by default
	}

	expiration := rateLimitConfig.AuthExpiration
	if expiration == 0 {
		expiration = 1 * time.Minute
	}

	slog.Info("Configuring auth rate limiting middleware",
		"max", max,
		"expiration", expiration,
	)

	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: expiration,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Use IP address as the key
			// Could also use username from request body for more sophisticated limiting
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			slog.Warn("Auth rate limit exceeded",
				"ip", c.IP(),
				"path", c.Path(),
				"method", c.Method(),
			)
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "auth_rate_limit_exceeded",
				"message": fmt.Sprintf("Too many authentication attempts. Please try again in %v.", expiration),
			})
		},
	})
}
