package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"growth-mvp/backend/adapters/postgres"
	"growth-mvp/backend/adapters/telegram"
	"growth-mvp/backend/api"
	"growth-mvp/backend/domain"

	"github.com/gin-gonic/gin"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := LoadConfig()

	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	defer stop()

	if err := postgres.RunMigrations(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	db, err := postgres.NewPool(ctx, cfg.DatabaseURL)

	if err != nil {
		logger.Error("failed to connect to db", "error", err)
		os.Exit(1)
	}

	defer db.Close()

	integrationRepo := postgres.NewIntegrationRepository(db)
	orderRepo := postgres.NewOrderRepository(db)
	sendLogRepo := postgres.NewSendLogRepository(db)
	telegramClient := telegram.NewClient()

	service := domain.NewService(integrationRepo, orderRepo, sendLogRepo, telegramClient)
	handler := api.NewHandler(service)

	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())
	handler.RegisterRoutes(router)

	logger.Info("starting API server", "port", cfg.Port)

	if err := router.Run(":" + cfg.Port); err != nil {
		logger.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
}
