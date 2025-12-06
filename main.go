package main

import (
	"flag"
	"fmt"
	"log/slog"

	"base-service/config"
	"base-service/internal/infra"
	"base-service/internal/route"
	"base-service/util"

	_ "base-service/docs"

	"github.com/jackc/pgx/v5/pgxpool"
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
	flagConf   string
	configPath string
	envFile    string
	env        string = ""
)

func init() {
	util.LoadEnv()
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
	db := InitDatabase(&cfg.Database)
	rd := InitRedis(&cfg.Redis)
	route.InitRoute(cfg, db, rd)

	defer db.Close()
	defer rd.Close()
}

func InitRedis(conf *config.RedisConfig) *infra.RedisClient {
	client, err := infra.NewRedisClient(conf)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to connect to redis, %v", err))
		panic(err)
	}
	if client == nil {
		slog.Error("failed to connect to redis")
		panic("failed to connect to redis")
	}
	return client
}

func InitDatabase(conf *config.DatabaseConfig) *pgxpool.Pool {
	client, err := infra.NewDatabaseClient(conf)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to connect to database, %v", err))
		panic(err)
	}
	if client == nil {
		slog.Error("failed to connect to database")
		panic("failed to connect to database")
	}
	return client.GetPool()
}
