package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// SecurityHeadersConfig defines configuration for security headers
type SecurityHeadersConfig struct {
	// XSSProtection enables XSS protection header
	XSSProtection bool
	// ContentTypeNoSniff enables X-Content-Type-Options: nosniff
	ContentTypeNoSniff bool
	// FrameDeny enables X-Frame-Options: DENY
	FrameDeny bool
	// HSTSMaxAge sets Strict-Transport-Security max-age (in seconds)
	// Set to 0 to disable HSTS
	HSTSMaxAge int
	// ContentSecurityPolicy sets Content-Security-Policy header
	// Example: "default-src 'self'; script-src 'self' 'unsafe-inline'"
	ContentSecurityPolicy string
	// ReferrerPolicy sets Referrer-Policy header
	// Recommended: "strict-origin-when-cross-origin" or "no-referrer"
	ReferrerPolicy string
	// PermissionsPolicy sets Permissions-Policy header
	// Example: "geolocation=(), microphone=()"
	PermissionsPolicy string
}

// DefaultSecurityHeadersConfig provides OWASP-recommended security headers
var DefaultSecurityHeadersConfig = SecurityHeadersConfig{
	XSSProtection:      true,
	ContentTypeNoSniff: true,
	FrameDeny:          true,
	HSTSMaxAge:         31536000, // 1 year
	ContentSecurityPolicy: "default-src 'self'; " +
		"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " + // Allow inline scripts for Swagger
		"style-src 'self' 'unsafe-inline'; " + // Allow inline styles for Swagger
		"img-src 'self' data: https:; " + // Allow images from self, data URIs, and HTTPS
		"font-src 'self' data:; " + // Allow fonts from self and data URIs
		"connect-src 'self'; " + // API calls only to same origin
		"frame-ancestors 'none'", // Prevent framing (clickjacking protection)
	ReferrerPolicy:    "strict-origin-when-cross-origin",
	PermissionsPolicy: "geolocation=(), microphone=(), camera=()",
}

// SecureHeadersMiddleware adds security headers to all responses
func SecureHeadersMiddleware(config SecurityHeadersConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// X-XSS-Protection: Protect against XSS attacks
		if config.XSSProtection {
			c.Set("X-XSS-Protection", "1; mode=block")
		}

		// X-Content-Type-Options: Prevent MIME type sniffing
		if config.ContentTypeNoSniff {
			c.Set("X-Content-Type-Options", "nosniff")
		}

		// X-Frame-Options: Prevent clickjacking
		if config.FrameDeny {
			c.Set("X-Frame-Options", "DENY")
		}

		// Strict-Transport-Security: Force HTTPS
		if config.HSTSMaxAge > 0 {
			c.Set("Strict-Transport-Security", "max-age="+string(rune(config.HSTSMaxAge))+"; includeSubDomains; preload")
		}

		// Content-Security-Policy: Control resource loading
		if config.ContentSecurityPolicy != "" {
			c.Set("Content-Security-Policy", config.ContentSecurityPolicy)
		}

		// Referrer-Policy: Control referrer information
		if config.ReferrerPolicy != "" {
			c.Set("Referrer-Policy", config.ReferrerPolicy)
		}

		// Permissions-Policy: Control browser features
		if config.PermissionsPolicy != "" {
			c.Set("Permissions-Policy", config.PermissionsPolicy)
		}

		// X-Permitted-Cross-Domain-Policies: Restrict cross-domain policies
		c.Set("X-Permitted-Cross-Domain-Policies", "none")

		// Remove X-Powered-By header (security through obscurity)
		c.Set("X-Powered-By", "")

		return c.Next()
	}
}
