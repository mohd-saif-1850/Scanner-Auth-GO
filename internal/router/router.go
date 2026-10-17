package router

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"scannerAuth/internal/handler"
	"scannerAuth/internal/hasher"
	"scannerAuth/internal/middleware"
	"scannerAuth/internal/pkg/token"
	"scannerAuth/internal/repository/gormrepo"
	"scannerAuth/internal/service"
)

func Setup(engine *gin.Engine, db *gorm.DB, tokenMgr token.Manager, pwHasher hasher.PasswordHasher) {
	// Initialize repositories
	userRepo := gormrepo.NewUserRepository(db)
	tokenRepo := gormrepo.NewTokenRepository(db)
	txManager := gormrepo.NewTxManager(db)

	// Initialize services
	authSvc := service.NewAuthService(userRepo, tokenRepo, txManager, pwHasher, tokenMgr)
	userSvc := service.NewUserService(userRepo, tokenRepo, txManager, pwHasher)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(userSvc)
	healthHandler := handler.NewHealthHandler()

	// Middleware
	rateLimiter := middleware.NewRateLimiter()

	// Global middleware chain (in order)
	engine.Use(gin.Recovery())
	engine.Use(middleware.Logger())
	engine.Use(middleware.ErrorHandler())

	// Public routes
	engine.GET("/health", healthHandler.Check)

	api := engine.Group("/api/v1")

	// Auth routes (public, with rate limit)
	auth := api.Group("/auth")
	auth.Use(rateLimiter.RateLimit(10))
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.Refresh)

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthRequired(tokenMgr))

	auth.POST("/logout", authHandler.Logout)

	// User routes
	users := protected.Group("/users")
	users.GET("/me", userHandler.GetMe)
	users.PUT("/me", userHandler.Update)
	users.PUT("/me/password", userHandler.ChangePassword)
	users.DELETE("/me", userHandler.Delete)
}
