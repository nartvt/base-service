package route

import (
	"base-service/config"
	"base-service/internal/biz"
	"base-service/internal/handler"
	"base-service/internal/middleware"
	"base-service/internal/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupUserRoute(r fiber.Router, auth *middleware.AuthenHandler, db *pgxpool.Pool, conf *config.Config) {
	userRepository := repository.NewUserRepository(db, db)
	userBiz := biz.NewUserBiz(userRepository)
	userHandler := handler.NewUserHandler(userBiz, *auth)

	// Apply auth rate limiting to authentication endpoints
	authGroup := r.Group("/auth")
	if conf.Middleware.RateLimit.AuthEnabled {
		authGroup.Use(middleware.AuthRateLimitFilter(conf.Middleware.RateLimit))
	}
	authHandler := handler.NewAuthHandler(userBiz, *auth)
	POST(authGroup, "/register", authHandler.RegisterUser)
	POST(authGroup, "/login", authHandler.LoginUser)
	POST(authGroup, "/refresh", authHandler.RefreshToken)
	POST(authGroup, "/logout", auth.Logout) // Logout requires authentication

	groupUser := r.Group("/user")
	protectedRoute := groupUser.Use(auth.AuthMiddleware())
	GET(protectedRoute, "profile", userHandler.Profile)
}
