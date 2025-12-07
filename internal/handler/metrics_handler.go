package handler

import (
	"fmt"
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type MetricsHandler interface {
	Metrics(c *fiber.Ctx) error
}

type metricsHandlerImpl struct {
	db        *pgxpool.Pool
	redis     *redis.Client
	startTime time.Time
}

func NewMetricsHandler(db *pgxpool.Pool, redis *redis.Client) MetricsHandler {
	return &metricsHandlerImpl{
		db:        db,
		redis:     redis,
		startTime: time.Now(),
	}
}

// @Summary Metrics endpoint
// @Description Get Prometheus-compatible metrics
// @Tags Monitoring
// @Produce plain
// @Success 200 {string} string "Prometheus metrics"
// @Router /metrics [get]
func (h *metricsHandlerImpl) Metrics(c *fiber.Ctx) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := fmt.Sprintf(`# HELP go_goroutines Number of goroutines
# TYPE go_goroutines gauge
go_goroutines %d

# HELP go_memstats_alloc_bytes Number of bytes allocated and still in use
# TYPE go_memstats_alloc_bytes gauge
go_memstats_alloc_bytes %d

# HELP go_memstats_sys_bytes Number of bytes obtained from system
# TYPE go_memstats_sys_bytes gauge
go_memstats_sys_bytes %d

# HELP go_memstats_heap_alloc_bytes Number of heap bytes allocated and still in use
# TYPE go_memstats_heap_alloc_bytes gauge
go_memstats_heap_alloc_bytes %d

# HELP go_memstats_heap_sys_bytes Number of heap bytes obtained from system
# TYPE go_memstats_heap_sys_bytes gauge
go_memstats_heap_sys_bytes %d

# HELP go_memstats_gc_completed_total Total number of completed GC cycles
# TYPE go_memstats_gc_completed_total counter
go_memstats_gc_completed_total %d

# HELP process_uptime_seconds Process uptime in seconds
# TYPE process_uptime_seconds gauge
process_uptime_seconds %d

`,
		runtime.NumGoroutine(),
		m.Alloc,
		m.Sys,
		m.HeapAlloc,
		m.HeapSys,
		m.NumGC,
		int64(time.Since(h.startTime).Seconds()),
	)

	// Add database pool metrics if available
	if h.db != nil {
		stats := h.db.Stat()
		metrics += fmt.Sprintf(`# HELP db_pool_max_connections Maximum number of database connections
# TYPE db_pool_max_connections gauge
db_pool_max_connections %d

# HELP db_pool_total_connections Total number of database connections
# TYPE db_pool_total_connections gauge
db_pool_total_connections %d

# HELP db_pool_idle_connections Number of idle database connections
# TYPE db_pool_idle_connections gauge
db_pool_idle_connections %d

# HELP db_pool_acquired_connections Number of acquired database connections
# TYPE db_pool_acquired_connections gauge
db_pool_acquired_connections %d

`,
			stats.MaxConns(),
			stats.TotalConns(),
			stats.IdleConns(),
			stats.AcquiredConns(),
		)
	}

	// Add Redis metrics if available
	if h.redis != nil {
		poolStats := h.redis.PoolStats()
		metrics += fmt.Sprintf(`# HELP redis_pool_hits Total number of times a connection was found in the pool
# TYPE redis_pool_hits counter
redis_pool_hits %d

# HELP redis_pool_misses Total number of times a connection was not found in the pool
# TYPE redis_pool_misses counter
redis_pool_misses %d

# HELP redis_pool_timeouts Total number of times a timeout occurred waiting for a connection
# TYPE redis_pool_timeouts counter
redis_pool_timeouts %d

# HELP redis_pool_total_conns Total number of connections in the pool
# TYPE redis_pool_total_conns gauge
redis_pool_total_conns %d

# HELP redis_pool_idle_conns Number of idle connections in the pool
# TYPE redis_pool_idle_conns gauge
redis_pool_idle_conns %d

`,
			poolStats.Hits,
			poolStats.Misses,
			poolStats.Timeouts,
			poolStats.TotalConns,
			poolStats.IdleConns,
		)
	}

	c.Set("Content-Type", "text/plain; version=0.0.4")
	return c.SendString(metrics)
}
