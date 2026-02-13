package main

import (
	"fmt"
	"os"
)

type Config struct {
	Port           string
	DatabaseURL    string
	MigrationsPath string
}

func LoadConfig() (Config, error) {
	cfg := Config{
		Port:           env("PORT", "8080"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		MigrationsPath: env("MIGRATIONS_PATH", "migrations"),
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
