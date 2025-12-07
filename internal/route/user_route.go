package route

import (
	"base-service/config"
	adapterAuth "base-service/internal/adapter/auth"
	adapterHandler "base-service/internal/adapter/http/handler"
	adapterRepository "base-service/internal/adapter/repository"
	"base-service/internal/middleware"
	"base-service/internal/usecase/auth"
	"base-service/internal/usecase/user"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// SetupUserRoute sets up user and auth routes using clean architecture.
func SetupUserRoute(r fiber.Router, authHandler *middleware.AuthMiddleware, db *pgxpool.Pool, conf *config.Config, redisClient *redis.Client) {
	// === Infrastructure Layer ===
	// Create repository adapters (implements domain interfaces)
	userRepo := adapterRepository.NewUserRepository(db)

	// === Adapter Layer ===
	// Create auth adapter (wraps middleware for use case layer)
	authAdapter := adapterAuth.NewAuthAdapter(authHandler)

	// === Application Layer ===
	// Create use cases with their dependencies
	authUseCase := auth.NewAuthUseCase(userRepo, authAdapter, authAdapter)
	userUseCase := user.NewUserUseCase(userRepo)

	// === Interface Layer ===
	// Create HTTP handlers
	authHTTPHandler := adapterHandler.NewAuthHandler(authUseCase, authHandler)
	userHTTPHandler := adapterHandler.NewUserHandler(userUseCase, authHandler)

	// === Routes ===
	// Auth routes (public)
	authGroup := r.Group("/auth")
	if conf.Middleware.RateLimit.AuthEnabled {
		var redisCli *redis.Client
		if conf.Middleware.RateLimit.UseRedis && redisClient != nil {
			redisCli = redisClient
		}
		authGroup.Use(middleware.AuthRateLimitFilter(conf.Middleware.RateLimit, &conf.Redis, redisCli))
	}

	POST(authGroup, "/register", authHTTPHandler.RegisterUser)
	POST(authGroup, "/login", authHTTPHandler.LoginUser)
	POST(authGroup, "/refresh", authHTTPHandler.RefreshToken)
	POST(authGroup, "/logout", authHTTPHandler.Logout)

	// User routes (protected)
	groupUser := r.Group("/user")
	protectedRoute := groupUser.Use(authHandler.AuthMiddleware())
	GET(protectedRoute, "profile", userHTTPHandler.Profile)
}

// SetupHealthRoute sets up health and metrics routes using clean architecture.
func SetupHealthRoute(r fiber.Router, db *pgxpool.Pool, redisClient *redis.Client) {
	healthHandler := adapterHandler.NewHealthHandler(db, redisClient)
	GET(r, "/health", healthHandler.HealthCheck)
	GET(r, "/health/liveness", healthHandler.LivenessCheck)
	GET(r, "/health/readiness", healthHandler.ReadinessCheck)

	metricsHandler := adapterHandler.NewMetricsHandler(db, redisClient)
	GET(r, "/metrics", metricsHandler.Metrics)
}
