package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	redisClient "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"scannerAuth/internal/config"
	"scannerAuth/internal/pkg/database"
	"scannerAuth/internal/pkg/rediscache"
)

type App struct {
	Router *gin.Engine
	DB     *gorm.DB
	Redis  *redisClient.Client
	Config *config.Config
}

func NewApp() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	gin.SetMode(cfg.GinMode)

	db, _, err := database.Connect(cfg)
	if err != nil {
		return nil, err
	}

	redisClient, err := rediscache.Connect(cfg)
	if err != nil {
		return nil, err
	}

	router := gin.Default()

	app := &App{
		Router: router,
		DB:     db,
		Redis:  redisClient,
		Config: cfg,
	}

	app.setupRoutes()

	return app, nil
}

func (app *App) setupRoutes() {
	app.Router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Scanner Auth server is running",
		})
	})
}

func (app *App) Run() error {
	return app.Router.Run(fmt.Sprintf(":%s", app.Config.AppPort))
}
