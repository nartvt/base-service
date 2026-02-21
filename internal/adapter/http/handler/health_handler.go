package handler

import (
	"context"
	"time"

	"base-service/internal/adapter/http/dto/response"
	"base-service/internal/common"
	"base-service/internal/infra"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// HealthHandler handles health check HTTP requests.
// clean-arch: Adapter layer - converts HTTP requests to infrastructure health checks
type HealthHandler struct {
	db       *pgxpool.Pool
	redis    *redis.Client
	registry *infra.Registry // New: unified registry for health checks
}

// NewHealthHandler creates a new health handler (backward compatible).
func NewHealthHandler(db *pgxpool.Pool, redis *redis.Client) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

// NewHealthHandlerWithRegistry creates a new health handler with registry support.
func NewHealthHandlerWithRegistry(registry *infra.Registry) *HealthHandler {
	h := &HealthHandler{
		registry: registry,
	}
	// Backward compatibility: also set individual clients
	if registry != nil {
		if db := registry.Database(); db != nil {
			h.db = db.GetPool()
		}
		if cache := registry.Cache(); cache != nil {
			h.redis = cache.Redis()
		}
	}
	return h
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

	// Use registry if available (new approach)
	if h.registry != nil {
		statuses := h.registry.HealthCheckAll(ctx)
		for _, status := range statuses {
			switch status.Status {
			case infra.StatusHealthy:
				resp.Services[status.Name] = "healthy"
			case infra.StatusDegraded:
				resp.Services[status.Name] = "degraded: " + status.Message
				if resp.Status == "healthy" {
					resp.Status = "degraded"
				}
			case infra.StatusUnhealthy:
				resp.Services[status.Name] = "unhealthy: " + status.Message
				resp.Status = "degraded"
			}
		}
	} else {
		// Backward compatible: check individual connections
		h.checkDatabaseHealth(ctx, &resp)
		h.checkRedisHealth(ctx, &resp)
	}

	// Return 503 if any service is unhealthy
	if resp.Status == "degraded" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(common.ApiResponse(resp, nil, nil))
	}

	return c.JSON(common.ApiResponse(resp, nil, nil))
}

func (h *HealthHandler) checkDatabaseHealth(ctx context.Context, resp *response.HealthResponse) {
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
}

func (h *HealthHandler) checkRedisHealth(ctx context.Context, resp *response.HealthResponse) {
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

	// Use registry if available
	if h.registry != nil {
		if !h.registry.IsHealthy(ctx) {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"ready":  false,
				"reason": "infrastructure unhealthy",
			})
		}
		return c.JSON(fiber.Map{
			"ready": true,
		})
	}

	// Backward compatible: check database
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

// @Summary Infrastructure stats
// @Description Get infrastructure statistics
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health/stats [get]
func (h *HealthHandler) InfraStats(c *fiber.Ctx) error {
	if h.registry == nil {
		return c.JSON(fiber.Map{
			"message": "registry not configured",
		})
	}

	stats := h.registry.Stats()
	return c.JSON(common.ApiResponse(stats, nil, nil))
}
