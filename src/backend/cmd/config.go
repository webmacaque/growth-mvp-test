package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port                string
	DatabaseURL         string
	MigrationsPath      string
	FrontendURL         string
	TelegramSendTimeout time.Duration
	TelegramMaxAttempts int
}

func LoadConfig() (Config, error) {
	cfg := Config{
		Port:                env("PORT", "8080"),
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		MigrationsPath:      env("MIGRATIONS_PATH", "migrations"),
		FrontendURL:         env("FRONTEND_URL", "http://localhost:5173"),
		TelegramSendTimeout: envDuration("TELEGRAM_SEND_TIMEOUT", 5*time.Second),
		TelegramMaxAttempts: envInt("TELEGRAM_MAX_ATTEMPTS", 3),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	return cfg, nil
}

func env(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func envDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)

	if v == "" {
		return fallback
	}

	d, err := time.ParseDuration(v)

	if err != nil {
		return fallback
	}

	return d
}

func envInt(key string, fallback int) int {
	v := os.Getenv(key)

	if v == "" {
		return fallback
	}

	n, err := strconv.Atoi(v)

	if err != nil || n <= 0 {
		return fallback
	}

	return n
}
