package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/xenptr/go-projects/todo-list-api/internal/auth"
	"github.com/xenptr/go-projects/todo-list-api/internal/config"
	"github.com/xenptr/go-projects/todo-list-api/internal/db"
	"github.com/xenptr/go-projects/todo-list-api/internal/handlers"
	"github.com/xenptr/go-projects/todo-list-api/internal/ratelimit"
	"github.com/xenptr/go-projects/todo-list-api/internal/redis"
	"github.com/xenptr/go-projects/todo-list-api/internal/repository"
	"github.com/xenptr/go-projects/todo-list-api/internal/routes"
)

func main() {
	// Load .env file if present; ignore error in production where env vars are set directly.
	_ = godotenv.Load()

	cfg := config.Load()

	pool, err := db.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	redisClient, err := redis.New(cfg)
	if err != nil {
		log.Printf("redis connection failed: %v", err)
		return
	}
	defer redisClient.Close()

	rateLimit := ratelimit.New(redisClient.Client)
	refreshStore := auth.NewRedisRefreshStore(redisClient.Client)

	repo := repository.New(pool)
	h := handlers.New(repo, refreshStore, cfg.JWTSecret)

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, h, cfg.JWTSecret, rateLimit)

	server := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: mux,

		// Prevent slow-loris and stalled connection attacks.
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Listen for SIGINT / SIGTERM so we can shut down cleanly.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit) // Stop sending OS signals to this channel

	// Report unexpected server errors back to main.
	serverErr := make(chan error, 1)

	go func() {
		log.Printf("server starting on :%s", cfg.AppPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Wait for either an OS signal or an unexpected server failure.
	select {
	case <-quit:
		// Graceful shutdown requested.

	case err := <-serverErr:
		log.Printf("server terminated unexpectedly: %v", err)
		return
	}

	log.Println("shutting down server...")

	// Give in-flight requests up to 10 seconds to finish.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("forced shutdown: %v", err)
		return
	}

	log.Println("server stopped")
}
