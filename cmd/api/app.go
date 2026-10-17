package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"scannerAuth/internal/config"
	"scannerAuth/internal/hasher"
	"scannerAuth/internal/pkg/database"
	"scannerAuth/internal/pkg/token"
	"scannerAuth/internal/router"
)

type App struct {
	Config *config.Config
}

func NewApp() (*App, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	return &App{Config: cfg}, nil
}

func (app *App) Run() error {
	// Connect to database
	db, _, err := database.Connect(app.Config)
	if err != nil {
		return fmt.Errorf("database connect: %w", err)
	}

	// Run migrations if enabled
	if app.Config.RunMigrations {
		if err := database.Migrate(db); err != nil {
			return fmt.Errorf("database migrate: %w", err)
		}
		log.Println("Migrations completed")
	}

	// Initialize password hasher
	bcryptHasher := hasher.NewBcryptHasher(app.Config.BcryptCost)

	// Initialize JWT manager
	tokenMgr, err := token.NewJWTManager(
		app.Config.JWTSecret,
		app.Config.JWTAccessTTL,
		app.Config.JWTIssuer,
		app.Config.JWTAudience,
	)
	if err != nil {
		return fmt.Errorf("init token manager: %w", err)
	}

	// Set Gin mode
	switch app.Config.GinMode {
	case "debug", "release", "test":
		// valid
	default:
		app.Config.GinMode = "release"
	}

	// Setup Gin engine
	gin.SetMode(app.Config.GinMode)
	engine := gin.Default()

	// Setup routes and middleware
	router.Setup(engine, db, tokenMgr, bcryptHasher)

	// Start server
	log.Printf("Starting auth service on :%s", app.Config.AppPort)
	return engine.Run(fmt.Sprintf(":%s", app.Config.AppPort))
}
