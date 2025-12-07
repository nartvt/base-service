package route

import (
	"base-service/config"
	"base-service/internal/infra"
	"base-service/internal/middleware"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitRoute(cf *config.Config, pool *pgxpool.Pool, redisClient *infra.RedisClient) {
	httpClient := infra.HttpServer{
		AppName: cf.Server.Http.AppName,
		Conf:    &cf.Server.Http,
		CORS:    &cf.Middleware,
		Redis:   redisClient.Redis(),
		RedisCf: &cf.Redis,
	}
	httpClient.InitHttpServer()
	jwtCache := middleware.NewJWTCache(redisClient.Redis(), true)
	auth := middleware.NewAuthenHandler(cf.Middleware, jwtCache)

	// Health and metrics endpoints (no auth required)
	SetupHealthRoute(httpClient.App(), pool, redisClient.Redis())

	apiv1 := httpClient.App().Group("/api/v1")
	SetupUserRoute(apiv1, auth, pool, cf, redisClient.Redis())

	// Print only API routes (not middleware routes)
	httpClient.Start()
}
