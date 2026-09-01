package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port           string
	DatabaseURL    string
	JWTSecret      string
	FrontendOrigin string
}

func Load() (Config, error) {
	cfg := Config{
		Port:           envOrDefault("PORT", "8080"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		FrontendOrigin: envOrDefault("FRONTEND_ORIGIN", "http://localhost:5173"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	if len(cfg.JWTSecret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}