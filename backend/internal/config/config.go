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
}

func Load() Config {
	return Config{
		HTTPAddr:    env("HTTP_ADDR", ":8080"),
		RedisAddr:   env("REDIS_ADDR", "localhost:6379"),
		DatabaseURL: env("DATABASE_URL", "postgres://bustracker:bustracker_dev@localhost:5432/bustracker"),
		LatestTTL:   10 * time.Minute,
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}