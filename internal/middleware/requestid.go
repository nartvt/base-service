package middleware

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// =============================================================================
// Context Keys (Type-Safe)
// clean-arch: Using typed context keys prevents collisions between packages
// =============================================================================

// contextKey is a custom type for context keys to avoid collisions.
type contextKey int

const (
	// requestIDKey is the typed context key for request ID
	requestIDKey contextKey = iota
)

// Legacy constants for backward compatibility with external packages
const (
	// RequestIDHeader is the header name for request ID
	RequestIDHeader = "X-Request-ID"
	// RequestIDContextKey is the string key (deprecated, use GetRequestID instead)
	RequestIDContextKey = "request_id"
)

// RequestIDConfig defines configuration for request ID middleware
type RequestIDConfig struct {
	// Header is the header name to read/write request ID
	Header string
	// ContextKey is the context key to store request ID (deprecated)
	ContextKey string
	// Generator is a function to generate request ID
	// If nil, UUID v4 will be used
	Generator func() string
}

// DefaultRequestIDConfig provides default configuration
var DefaultRequestIDConfig = RequestIDConfig{
	Header:     RequestIDHeader,
	ContextKey: RequestIDContextKey,
	Generator:  nil, // Will use UUID v4
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware(config RequestIDConfig) fiber.Handler {
	// Set defaults
	if config.Header == "" {
		config.Header = RequestIDHeader
	}
	if config.ContextKey == "" {
		config.ContextKey = RequestIDContextKey
	}
	if config.Generator == nil {
		config.Generator = func() string {
			return uuid.New().String()
		}
	}

	return func(c *fiber.Ctx) error {
		// Check if request ID already exists in header
		requestID := c.Get(config.Header)

		// If not, generate a new one
		if requestID == "" {
			requestID = config.Generator()
		}

		// Store in context using typed key (primary)
		c.Locals(requestIDKey, requestID)
		// Also store with string key for backward compatibility
		c.Locals(config.ContextKey, requestID)

		// Set in response header
		c.Set(config.Header, requestID)

		// Add to structured logger context
		slog.Debug("Request received",
			"request_id", requestID,
			"method", c.Method(),
			"path", c.Path(),
			"ip", c.IP(),
		)

		return c.Next()
	}
}

// GetRequestID retrieves the request ID from context using typed key.
// This is the preferred method as it's type-safe.
func GetRequestID(c *fiber.Ctx) string {
	// Try typed key first
	if requestID, ok := c.Locals(requestIDKey).(string); ok {
		return requestID
	}
	// Fallback to string key for backward compatibility
	if requestID, ok := c.Locals(RequestIDContextKey).(string); ok {
		return requestID
	}
	return ""
}

// MustGetRequestID retrieves the request ID or returns a default value.
func MustGetRequestID(c *fiber.Ctx, defaultValue string) string {
	if requestID := GetRequestID(c); requestID != "" {
		return requestID
	}
	return defaultValue
}
