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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"bus-tracker/internal/config"
	"bus-tracker/internal/hub"
	"bus-tracker/internal/repository"
	"bus-tracker/internal/service"
	transport "bus-tracker/internal/transport/http"
)

func main() {
	cfg := config.Load()
	if cfg.JWTSecret == config.DefaultJWTSecret {
		slog.Warn("JWT_SECRET not set: using an insecure development secret")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Error("redis unreachable", "err", err)
		os.Exit(1)
	}
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("postgres pool failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	cache := repository.NewTelemetryCache(rdb, cfg.LatestTTL)
	store := repository.NewTelemetryStore(pool)
	svc := service.NewTelemetryService(cache, store)
	auth := service.NewAuthService(repository.NewUserStore(pool), cfg.JWTSecret, cfg.TokenTTL)
	h := hub.New()

	sub := cache.PSubscribe(ctx)
	defer sub.Close()
	go h.Run(ctx, sub.Channel())
	go svc.RunPersister(ctx)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           transport.NewRouter(transport.NewHandler(svc, h, auth), auth),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		slog.Info("listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
