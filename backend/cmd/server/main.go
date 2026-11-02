package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MTPeraya/stampede-shield/backend/internal/api"
	"github.com/MTPeraya/stampede-shield/backend/internal/cache"
	"github.com/MTPeraya/stampede-shield/backend/internal/chaos"
	"github.com/MTPeraya/stampede-shield/backend/internal/features"
	"github.com/MTPeraya/stampede-shield/backend/internal/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pgURL := getenv("DATABASE_URL", "postgres://stampede:stampede@postgres:5432/features?sslmode=disable")
	redisURL := getenv("REDIS_URL", "redis://redis:6379/0")
	pg, err := pgxpool.New(ctx, pgURL)
	if err != nil {
		logger.Error("create postgres pool", "error", err)
		os.Exit(1)
	}
	defer pg.Close()
	if err := pg.Ping(ctx); err != nil {
		logger.Error("ping postgres", "error", err)
		os.Exit(1)
	}

	redisOptions, err := redis.ParseURL(redisURL)
	if err != nil {
		logger.Error("parse redis URL", "error", err)
		os.Exit(1)
	}
	rdb := redis.NewClient(redisOptions)
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error("ping redis", "error", err)
		os.Exit(1)
	}

	store := features.NewStore(pg)
	counters := metrics.New()
	chaosState := chaos.New()
	redisCache := cache.New(rdb, chaosState)
	service := features.NewService(store, redisCache, counters, chaosState)
	router := api.NewRouter(service, redisCache, counters)
	server := &http.Server{Addr: ":" + getenv("PORT", "8080"), Handler: router}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	logger.Info("server listening", "addr", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
