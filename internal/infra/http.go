package infra

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"base-service/config"
	"base-service/internal/middleware"
	"base-service/util"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	"github.com/redis/go-redis/v9"
)

type HttpServer struct {
	AppName string
	Conf    *config.ServerInfo
	CORS    *config.MiddlewareConfig
	Redis   *redis.Client
	RedisCf *config.RedisConfig
	app     *fiber.App
}

func (r *HttpServer) Start() {
	if r == nil || r.app == nil {
		slog.Error("http server is nil")
		return
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	serverShutdown := make(chan struct{})

	go func() {
		err := r.app.Listen(fmt.Sprintf("%s:%d", r.Conf.Host, r.Conf.Port))
		if err != nil {
			slog.Error("listen error", "err", err)
		}
		close(serverShutdown)
	}()

	<-shutdown
	slog.Info("Shutting down server...")

	ctx := util.ContextwithTimeout()
	if err := r.app.ShutdownWithContext(ctx); err != nil {
		slog.Info(fmt.Sprintf("Server forced to shutdown: %v\n", err))
	}

	<-serverShutdown
	slog.Info("Server gracefully stopped")
}

func (r *HttpServer) InitHttpServer() {
	app := fiber.New(r.ConfigFiber(r.Conf))

	// Security headers (first layer of defense)
	app.Use(middleware.SecureHeadersMiddleware(middleware.DefaultSecurityHeadersConfig))

	// Request ID tracking
	app.Use(middleware.RequestIDMiddleware(middleware.DefaultRequestIDConfig))

	// Apply CORS middleware with configuration
	app.Use(middleware.CorsFilter(r.CORS.CORS))

	// Apply general rate limiting to all API endpoints
	if r.CORS != nil && r.CORS.RateLimit.Enabled {
		var redisClient *redis.Client
		if r.CORS.RateLimit.UseRedis && r.Redis != nil {
			redisClient = r.Redis
		}
		app.Use(middleware.RateLimitFilter(r.CORS.RateLimit, r.RedisCf, redisClient))
	}

	r.app = app
	r.SetupSwagger()
	r.SetupPrintAPIRoutes()
}

func (r *HttpServer) SetupPrintAPIRoutes() {
	r.app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${method} ${path} ${latency}\n",
		TimeFormat: "02/01/2006 15:04:05",
		TimeZone:   "Asia/Ho_Chi_Minh",
		// Only log API routes, skip static assets and health checks
		Next: func(c *fiber.Ctx) bool {
			path := c.Path()
			// Skip logging for non-API routes
			return r.ignorePath(path)
		},
	}))
}

func (r *HttpServer) ignorePath(path string) bool {
	return path == "/" ||
		path == "/favicon.ico" ||
		path == "/healthz" ||
		path == "/metrics" ||
		strings.Contains(path, "github") ||
		len(path) >= 8 && path[:8] == "/static/" ||
		len(path) >= 7 && path[:7] == "/public/"
}

func (r *HttpServer) SetupSwagger() {
	r.app.Get("/swagger/*", swagger.New(swagger.Config{
		Title:        "Base Service API Documentation",
		DeepLinking:  true,
		DocExpansion: "none",
	}))
}

func (r *HttpServer) App() *fiber.App {
	return r.app
}

func (r *HttpServer) ConfigFiber(conf *config.ServerInfo) fiber.Config {
	return fiber.Config{
		AppName:               conf.AppName,
		EnablePrintRoutes:     true, // Disabled - we'll print only API routes manually
		DisableStartupMessage: false,
		ReadTimeout:           time.Duration(conf.ReadTimeout) * time.Second,
		WriteTimeout:          time.Duration(conf.WriteTimeout) * time.Second,
	}
}
