package infra

import (
	"context"
	"fmt"
	"time"

	conf "base-service/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	databaseName           = "postgresql"
	defaultHealthTimeout   = 5 * time.Second
	defaultConnectTimeout  = 10 * time.Second
	defaultHealthCheckPeriod = 1 * time.Minute
	defaultJitter          = 10 * time.Second
)

// Compile-time interface compliance check
var (
	_ Connection      = (*DatabaseClient)(nil)
	_ TransactionalDB = (*DatabaseClient)(nil)
)

// DatabaseClient wraps pgxpool.Pool and implements Connection and TransactionalDB interfaces.
// clean-arch: Infrastructure adapter implementing domain ports
type DatabaseClient struct {
	pool   *pgxpool.Pool
	config *conf.DatabaseConfig
}

// NewDatabaseClient creates a new DatabaseClient with the given configuration.
func NewDatabaseClient(dbConfig *conf.DatabaseConfig) (*DatabaseClient, error) {
	if dbConfig == nil {
		return nil, fmt.Errorf("database config is required")
	}

	dbURL := dbConfig.BuildConnectionStringPostgres()

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url failed: %w", err)
	}

	// Connection pool settings
	config.MaxConns = int32(dbConfig.MaxOpenConnections)
	config.MinConns = int32(dbConfig.MaxIdleConnections)
	config.MaxConnLifetime = dbConfig.MaxConnLifetime
	config.MaxConnIdleTime = dbConfig.MaxConnIdleTime
	config.HealthCheckPeriod = defaultHealthCheckPeriod
	config.MaxConnLifetimeJitter = defaultJitter

	// Enable SQL query logging
	config.ConnConfig.Tracer = &SQLTracer{}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultConnectTimeout)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return &DatabaseClient{
		pool:   pool,
		config: dbConfig,
	}, nil
}

// =============================================================================
// Connection Interface Implementation
// =============================================================================

// Name returns the identifier for this connection.
func (d *DatabaseClient) Name() string {
	return databaseName
}

// Ping checks if the database connection is alive.
func (d *DatabaseClient) Ping(ctx context.Context) error {
	if d.pool == nil {
		return fmt.Errorf("database pool is nil")
	}
	return d.pool.Ping(ctx)
}

// IsHealthy performs a health check with timeout.
func (d *DatabaseClient) IsHealthy(ctx context.Context) bool {
	if d.pool == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(ctx, defaultHealthTimeout)
	defer cancel()

	if err := d.pool.Ping(ctx); err != nil {
		return false
	}

	// Additional check: verify we can acquire a connection
	conn, err := d.pool.Acquire(ctx)
	if err != nil {
		return false
	}
	conn.Release()

	return true
}

// Close closes the database connection pool.
func (d *DatabaseClient) Close() error {
	if d.pool != nil {
		d.pool.Close()
	}
	return nil
}

// =============================================================================
// TransactionalDB Interface Implementation
// =============================================================================

// BeginTx starts a new database transaction.
func (d *DatabaseClient) BeginTx(ctx context.Context) (Transaction, error) {
	if d.pool == nil {
		return nil, fmt.Errorf("database pool is nil")
	}

	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &pgxTransaction{tx: tx}, nil
}

// Stats returns the connection pool statistics.
func (d *DatabaseClient) Stats() ConnectionStats {
	if d.pool == nil {
		return ConnectionStats{}
	}

	stat := d.pool.Stat()
	return ConnectionStats{
		MaxConnections:      stat.MaxConns(),
		CurrentConnections:  stat.TotalConns(),
		IdleConnections:     stat.IdleConns(),
		AcquiredConnections: stat.AcquiredConns(),
	}
}

// =============================================================================
// Additional Methods (for backward compatibility)
// =============================================================================

// GetPool returns the underlying pgxpool.Pool.
// NOTE: Use this only when you need direct access to pgx features.
// Prefer using the interface methods for better abstraction.
func (d *DatabaseClient) GetPool() *pgxpool.Pool {
	return d.pool
}

// =============================================================================
// Transaction Implementation
// =============================================================================

// pgxTransaction wraps pgx.Tx and implements Transaction interface.
type pgxTransaction struct {
	tx pgx.Tx
}

// Commit commits the transaction.
func (t *pgxTransaction) Commit(ctx context.Context) error {
	if t.tx == nil {
		return fmt.Errorf("transaction is nil")
	}
	return t.tx.Commit(ctx)
}

// Rollback rolls back the transaction.
func (t *pgxTransaction) Rollback(ctx context.Context) error {
	if t.tx == nil {
		return fmt.Errorf("transaction is nil")
	}
	return t.tx.Rollback(ctx)
}

// Tx returns the underlying pgx.Tx for advanced operations.
func (t *pgxTransaction) Tx() pgx.Tx {
	return t.tx
}
