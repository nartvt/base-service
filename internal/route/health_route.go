package route

import (
	"base-service/internal/handler"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func SetupHealthRoute(app *fiber.App, db *pgxpool.Pool, redis *redis.Client) {
	healthHandler := handler.NewHealthHandler(db, redis)
	metricsHandler := handler.NewMetricsHandler(db, redis)

	// Health check endpoints
	app.Get("/healthz", healthHandler.HealthCheck)
	app.Get("/health", healthHandler.HealthCheck) // Alternative endpoint
	app.Get("/ready", healthHandler.ReadinessCheck)
	app.Get("/live", healthHandler.LivenessCheck)

	// Metrics endpoint for Prometheus
	app.Get("/metrics", metricsHandler.Metrics)
}
