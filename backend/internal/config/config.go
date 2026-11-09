package config

import (
	"os"
	"time"
)

type Config struct {
	HTTPAddr    string
	RedisAddr   string
	DatabaseURL string
	LatestTTL   time.Duration // how long the "last known position" survives in Redis
	JWTSecret   string
	TokenTTL    time.Duration
}

const DefaultJWTSecret = "dev-only-change-me-use-a-long-random-secret"

func Load() Config {
	return Config{
		HTTPAddr:    env("HTTP_ADDR", ":8080"),
		RedisAddr:   env("REDIS_ADDR", "localhost:6379"),
		DatabaseURL: env("DATABASE_URL", "postgres://bustracker:password123@localhost:5432/bustracker"),
		LatestTTL:   10 * time.Minute,
		JWTSecret:   env("JWT_SECRET", DefaultJWTSecret),
		TokenTTL:    12 * time.Hour,
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
