package infra

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// =============================================================================
// Infrastructure Registry
// clean-arch: Centralized registry for all infrastructure components
// =============================================================================

// Registry provides unified access to all infrastructure components.
// It manages connections lifecycle and health checking.
type Registry struct {
	mu          sync.RWMutex
	connections map[string]Connection
	database    *DatabaseClient
	cache       *RedisClient
}

// NewRegistry creates a new infrastructure registry.
func NewRegistry() *Registry {
	return &Registry{
		connections: make(map[string]Connection),
	}
}

// =============================================================================
// Registration Methods
// =============================================================================

// RegisterConnection registers a connection with the registry.
func (r *Registry) RegisterConnection(conn Connection) error {
	if conn == nil {
		return fmt.Errorf("connection is nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	name := conn.Name()
	if _, exists := r.connections[name]; exists {
		return fmt.Errorf("connection %s already registered", name)
	}

	r.connections[name] = conn
	slog.Info("Registered infrastructure connection", "name", name)
	return nil
}

// RegisterDatabase registers the database client.
func (r *Registry) RegisterDatabase(db *DatabaseClient) error {
	if db == nil {
		return fmt.Errorf("database client is nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.database = db
	r.connections[db.Name()] = db
	slog.Info("Registered database connection")
	return nil
}

// RegisterCache registers the cache client.
func (r *Registry) RegisterCache(cache *RedisClient) error {
	if cache == nil {
		return fmt.Errorf("cache client is nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.cache = cache
	r.connections[cache.Name()] = cache
	slog.Info("Registered cache connection")
	return nil
}

// =============================================================================
// Accessor Methods
// =============================================================================

// Database returns the database client.
func (r *Registry) Database() *DatabaseClient {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.database
}

// Cache returns the cache client.
func (r *Registry) Cache() *RedisClient {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.cache
}

// GetConnection returns a connection by name.
func (r *Registry) GetConnection(name string) (Connection, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	conn, ok := r.connections[name]
	return conn, ok
}

// Connections returns all registered connections.
func (r *Registry) Connections() []Connection {
	r.mu.RLock()
	defer r.mu.RUnlock()

	conns := make([]Connection, 0, len(r.connections))
	for _, conn := range r.connections {
		conns = append(conns, conn)
	}
	return conns
}

// =============================================================================
// Health Check Methods
// =============================================================================

// HealthCheckAll performs health checks on all registered connections.
func (r *Registry) HealthCheckAll(ctx context.Context) []HealthStatus {
	r.mu.RLock()
	connections := make([]Connection, 0, len(r.connections))
	for _, conn := range r.connections {
		connections = append(connections, conn)
	}
	r.mu.RUnlock()

	results := make([]HealthStatus, 0, len(connections))

	for _, conn := range connections {
		status := r.checkConnection(ctx, conn)
		results = append(results, status)
	}

	return results
}

// HealthCheck performs a health check on a specific connection.
func (r *Registry) HealthCheck(ctx context.Context, name string) (HealthStatus, error) {
	r.mu.RLock()
	conn, ok := r.connections[name]
	r.mu.RUnlock()

	if !ok {
		return HealthStatus{}, fmt.Errorf("connection %s not found", name)
	}

	return r.checkConnection(ctx, conn), nil
}

func (r *Registry) checkConnection(ctx context.Context, conn Connection) HealthStatus {
	start := time.Now()
	healthy := conn.IsHealthy(ctx)
	latency := time.Since(start)

	status := StatusHealthy
	message := "OK"

	if !healthy {
		status = StatusUnhealthy
		message = "Health check failed"
	} else if latency > 100*time.Millisecond {
		status = StatusDegraded
		message = fmt.Sprintf("High latency: %v", latency)
	}

	return HealthStatus{
		Name:      conn.Name(),
		Status:    status,
		Latency:   latency,
		Message:   message,
		Timestamp: time.Now(),
	}
}

// IsHealthy returns true if all connections are healthy.
func (r *Registry) IsHealthy(ctx context.Context) bool {
	statuses := r.HealthCheckAll(ctx)
	for _, status := range statuses {
		if status.Status == StatusUnhealthy {
			return false
		}
	}
	return true
}

// =============================================================================
// Lifecycle Methods
// =============================================================================

// Close closes all registered connections.
func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errs []error

	for name, conn := range r.connections {
		if err := conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close %s: %w", name, err))
			slog.Error("Failed to close connection", "name", name, "error", err)
		} else {
			slog.Info("Closed connection", "name", name)
		}
	}

	// Clear the registry
	r.connections = make(map[string]Connection)
	r.database = nil
	r.cache = nil

	if len(errs) > 0 {
		return fmt.Errorf("errors closing connections: %v", errs)
	}
	return nil
}

// =============================================================================
// Statistics Methods
// =============================================================================

// Stats returns statistics for all connections.
type RegistryStats struct {
	TotalConnections int                       `json:"total_connections"`
	Database         *ConnectionStats          `json:"database,omitempty"`
	Connections      map[string]map[string]any `json:"connections"`
}

// Stats returns statistics for all registered components.
func (r *Registry) Stats() RegistryStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := RegistryStats{
		TotalConnections: len(r.connections),
		Connections:      make(map[string]map[string]any),
	}

	if r.database != nil {
		dbStats := r.database.Stats()
		stats.Database = &dbStats
	}

	for name := range r.connections {
		stats.Connections[name] = map[string]any{
			"registered": true,
		}
	}

	return stats
}
