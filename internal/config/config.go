package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	CORSOrigin  string
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) {
			return Config{}, fmt.Errorf("load .env: %w", err)
		}
	}

	databaseURL := os.Getenv("DB_URL")
	if databaseURL == "" {
		databaseURL = os.Getenv("DATABASE_URL")
	}
	if databaseURL == "" {
		return Config{}, fmt.Errorf("DB_URL or DATABASE_URL is required")
	}

	cfg := Config{
		DatabaseURL: databaseURL,
		JWTSecret:   os.Getenv("JWT_SECRET"),
		CORSOrigin:  os.Getenv("CORS_ORIGIN"),
	}
	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.CORSOrigin == "" {
		cfg.CORSOrigin = "*"
	}
	return cfg, nil
}
