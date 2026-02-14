package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"growth-mvp/backend/adapters/postgres"
	"growth-mvp/backend/adapters/telegram"
	"growth-mvp/backend/api"
	"growth-mvp/backend/domain"

	"github.com/gin-contrib/cors"
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
	telegramClient := telegram.NewClient(cfg.TelegramSendTimeout)

	service := domain.NewService(integrationRepo, orderRepo, sendLogRepo, telegramClient, cfg.TelegramMaxAttempts)
	handler := api.NewHandler(service)

	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{cfg.FrontendURL},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	handler.RegisterRoutes(router)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("http shutdown failed", "error", err)
		}

	}()

	logger.Info("starting API server", "port", cfg.Port)

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
}
