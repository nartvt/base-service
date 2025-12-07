package handler

import (
	"context"
	"time"

	"base-service/internal/dto/response"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type HealthHandler interface {
	HealthCheck(c *fiber.Ctx) error
	ReadinessCheck(c *fiber.Ctx) error
	LivenessCheck(c *fiber.Ctx) error
}

type healthHandlerImpl struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewHealthHandler(db *pgxpool.Pool, redis *redis.Client) HealthHandler {
	return &healthHandlerImpl{
		db:    db,
		redis: redis,
	}
}

// @Summary Health check
// @Description Check if the service is healthy
// @Tags Health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /healthz [get]
func (h *healthHandlerImpl) HealthCheck(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	response := response.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Unix(),
		Services:  make(map[string]string),
		Version:   "2.0.0",
	}

	// Check database
	if h.db != nil {
		if err := h.db.Ping(ctx); err != nil {
			response.Services["database"] = "unhealthy: " + err.Error()
			response.Status = "degraded"
		} else {
			response.Services["database"] = "healthy"
		}
	} else {
		response.Services["database"] = "not configured"
	}

	// Check Redis
	if h.redis != nil {
		if err := h.redis.Ping(ctx).Err(); err != nil {
			response.Services["redis"] = "unhealthy: " + err.Error()
			response.Status = "degraded"
		} else {
			response.Services["redis"] = "healthy"
		}
	} else {
		response.Services["redis"] = "not configured"
	}

	// Return 503 if any service is unhealthy
	if response.Status == "degraded" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(response)
	}

	return c.JSON(response)
}

// @Summary Readiness check
// @Description Check if the service is ready to accept traffic
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /ready [get]
func (h *healthHandlerImpl) ReadinessCheck(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Check critical dependencies
	if h.db != nil {
		if err := h.db.Ping(ctx); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"ready":  false,
				"reason": "database unavailable",
			})
		}
	}

	return c.JSON(fiber.Map{
		"ready": true,
	})
}

// @Summary Liveness check
// @Description Check if the service is alive (for Kubernetes liveness probe)
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /live [get]
func (h *healthHandlerImpl) LivenessCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"alive": true,
	})
}
