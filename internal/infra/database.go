package infra

import (
	"context"
	"fmt"
	"time"

	conf "base-service/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabaseClient struct {
	db *pgxpool.Pool
}

func NewDatabaseClient(dbConfig *conf.DatabaseConfig) (*DatabaseClient, error) {
	dbURL := dbConfig.BuildConnectionStringPostgres()

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url failed: %w", err)
	}

	config.MaxConns = int32(dbConfig.MaxOpenConnections) // pgx yêu cầu int32
	config.MinConns = int32(dbConfig.MaxIdleConnections)
	config.MaxConnLifetime = dbConfig.MaxConnLifetime // ví dụ: 30 * time.Minute
	config.MaxConnIdleTime = dbConfig.MaxConnIdleTime // ví dụ: 5 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute        // tự động ping DB
	config.MaxConnLifetimeJitter = 10 * time.Second   // tránh thundering herd

	// Enable SQL query logging
	config.ConnConfig.Tracer = &SQLTracer{}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return &DatabaseClient{db: pool}, nil
}

func (d *DatabaseClient) Close() {
	if d.db != nil {
		d.db.Close()
	}
}

func (d *DatabaseClient) GetPool() *pgxpool.Pool {
	return d.db
}
