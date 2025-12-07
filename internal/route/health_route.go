package route

import (
	"base-service/internal/handler"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func SetupHealthRoute(r fiber.Router, db *pgxpool.Pool, redis *redis.Client) {
	healthHandler := handler.NewHealthHandler(db, redis)
	metricsHandler := handler.NewMetricsHandler(db, redis)

	// Health check endpoints
	GET(r, "/healthz", healthHandler.HealthCheck)
	GET(r, "/health", healthHandler.HealthCheck) // Alternative endpoint
	GET(r, "/ready", healthHandler.ReadinessCheck)
	GET(r, "/live", healthHandler.LivenessCheck)

	// Metrics endpoint for Prometheus
	GET(r, "/metrics", metricsHandler.Metrics)
}
