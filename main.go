package main

import (
	"flag"
	"fmt"
	"log/slog"

	"base-service/config"
	"base-service/internal/infra"
	"base-service/internal/route"

	_ "base-service/docs"
)

// @title Base Service API
// @version 1.0
// @description API documentation for Base Service
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@baseservice.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8081
// @BasePath /api

var (
	configPath string
	env        string = ""
)

func init() {
	// Define flags
	flag.StringVar(&configPath, "config", "config", "path to config file")
	flag.StringVar(&env, "env", "", "env")
	flag.Parse()
}

func LoadConfig() *config.Config {
	conf := &config.Config{}
	err := config.LoadConfig(configPath, env, conf)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to load config, %v", err))
	}
	return conf
}

func main() {
	conf := LoadConfig()
	infra.InitLogger(*conf)
	StartServer(conf)
}

func StartServer(cfg *config.Config) {
	// Initialize infrastructure registry
	registry := infra.NewRegistry()

	// Initialize and register database
	db, err := initDatabase(&cfg.Database)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		panic(err)
	}
	if err := registry.RegisterDatabase(db); err != nil {
		slog.Error("Failed to register database", "error", err)
		panic(err)
	}

	// Initialize and register cache
	cache, err := initRedis(&cfg.Redis)
	if err != nil {
		slog.Error("Failed to initialize redis", "error", err)
		panic(err)
	}
	if err := registry.RegisterCache(cache); err != nil {
		slog.Error("Failed to register cache", "error", err)
		panic(err)
	}

	// Log infrastructure stats
	logInfraStats(registry)

	// Initialize routes (backward compatible - using pool and redis client)
	route.InitRoute(cfg, db.GetPool(), cache)

	// Graceful shutdown - close all connections via registry
	defer func() {
		slog.Info("Shutting down infrastructure...")
		if err := registry.Close(); err != nil {
			slog.Error("Error closing infrastructure", "error", err)
		}
	}()
}

func initRedis(conf *config.RedisConfig) (*infra.RedisClient, error) {
	client, err := infra.NewRedisClient(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}
	if client == nil {
		return nil, fmt.Errorf("redis client is nil")
	}
	return client, nil
}

func initDatabase(conf *config.DatabaseConfig) (*infra.DatabaseClient, error) {
	client, err := infra.NewDatabaseClient(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	if client == nil {
		return nil, fmt.Errorf("database client is nil")
	}
	return client, nil
}

func logInfraStats(registry *infra.Registry) {
	stats := registry.Stats()
	slog.Info("Infrastructure initialized",
		"total_connections", stats.TotalConnections,
	)

	if stats.Database != nil {
		slog.Info("Database pool stats",
			"max_connections", stats.Database.MaxConnections,
			"current_connections", stats.Database.CurrentConnections,
			"idle_connections", stats.Database.IdleConnections,
		)
	}
}
