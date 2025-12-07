package middleware

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	// RequestIDHeader is the header name for request ID
	RequestIDHeader = "X-Request-ID"
	// RequestIDContextKey is the context key for request ID
	RequestIDContextKey = "request_id"
)

// RequestIDConfig defines configuration for request ID middleware
type RequestIDConfig struct {
	// Header is the header name to read/write request ID
	Header string
	// ContextKey is the context key to store request ID
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

		// Store in context for handlers to access
		c.Locals(config.ContextKey, requestID)

		// Set in response header
		c.Set(config.Header, requestID)

		// Add to structured logger context
		slog.Info("Request received",
			"request_id", requestID,
			"method", c.Method(),
			"path", c.Path(),
			"ip", c.IP(),
		)

		return c.Next()
	}
}

// GetRequestID retrieves the request ID from context
func GetRequestID(c *fiber.Ctx) string {
	if requestID, ok := c.Locals(RequestIDContextKey).(string); ok {
		return requestID
	}
	return ""
}
