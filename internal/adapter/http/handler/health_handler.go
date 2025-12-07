package handler

import (
	"context"
	"time"

	"base-service/internal/adapter/http/dto/response"
	"base-service/internal/common"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// HealthHandler handles health check HTTP requests.
type HealthHandler struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(db *pgxpool.Pool, redis *redis.Client) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

// @Summary Health check
// @Description Check if the service is healthy
// @Tags Health
// @Produce json
// @Success 200 {object} response.HealthResponse
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
	defer cancel()

	resp := response.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Unix(),
		Services:  make(map[string]string),
		Version:   "2.0.0",
	}

	// Check database
	if h.db != nil {
		if err := h.db.Ping(ctx); err != nil {
			resp.Services["database"] = "unhealthy: " + err.Error()
			resp.Status = "degraded"
		} else {
			resp.Services["database"] = "healthy"
		}
	} else {
		resp.Services["database"] = "not configured"
	}

	// Check Redis
	if h.redis != nil {
		if err := h.redis.Ping(ctx).Err(); err != nil {
			resp.Services["redis"] = "unhealthy: " + err.Error()
			resp.Status = "degraded"
		} else {
			resp.Services["redis"] = "healthy"
		}
	} else {
		resp.Services["redis"] = "not configured"
	}

	// Return 503 if any service is unhealthy
	if resp.Status == "degraded" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(common.ApiResponse(resp, nil, nil))
	}

	return c.JSON(common.ApiResponse(resp, nil, nil))
}

// @Summary Readiness check
// @Description Check if the service is ready to accept traffic
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health/readiness [get]
func (h *HealthHandler) ReadinessCheck(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
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
// @Router /health/liveness [get]
func (h *HealthHandler) LivenessCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"alive": true,
	})
}
