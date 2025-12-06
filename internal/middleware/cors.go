package middleware

import (
	"fmt"
	"log/slog"
	"time"

	"base-service/config"
	"base-service/util"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/csrf"
)

type Source string

const (
	HeaderName            = "X-Csrf-Token"
	SourceCookie   Source = "cookie"
	SourceHeader   Source = "header"
	SourceURLQuery Source = "query"
)

func CorsFilter(corsConfig config.CORSConfig) fiber.Handler {
	slog.Info("Configuring CORS middleware",
		"origins", corsConfig.GetOriginsString(),
		"methods", corsConfig.GetMethodsString(),
		"credentials", corsConfig.AllowCredentials,
	)
	if corsConfig.AllowCredentials && corsConfig.GetOriginsString() == "*" {
		slog.Warn("CORS: AllowCredentials is true but AllowOrigins is '*'. This is insecure and may not work in browsers.")
	}

	maxAge := corsConfig.MaxAge
	if maxAge == 0 {
		maxAge = 86400 // 24 hours default
	}

	return cors.New(cors.Config{
		AllowOrigins:     corsConfig.GetOriginsString(),
		AllowMethods:     corsConfig.GetMethodsString(),
		AllowHeaders:     corsConfig.GetHeadersString(),
		AllowCredentials: corsConfig.AllowCredentials,
		ExposeHeaders:    corsConfig.GetExposedHeadersString(),
		MaxAge:           maxAge,
	})
}

func CSRFFilter() fiber.Handler {
	return csrf.New(csrf.Config{
		KeyLookup:      fmt.Sprintf("%s:%s", SourceHeader, HeaderName),
		CookieName:     "csrf_",
		CookieSameSite: "Lax",
		Expiration:     30 * time.Minute,
		KeyGenerator:   util.UUIDFunc(),
	})
}
