package main

import (
	"log"
	"net/http"

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
		log.Fatalf("redis connection failed: %v", err)
	}
	defer redisClient.Close()

	rateLimit := ratelimit.New(redisClient.Client)
	refreshStore := auth.NewRedisRefreshStore(redisClient.Client)

	repo := repository.New(pool)
	h := handlers.New(repo, refreshStore, cfg.JWTSecret)

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, h, cfg.JWTSecret, rateLimit)

	log.Printf("server starting on :%s", cfg.AppPort)

	server := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server terminated unexpectedly: %v", err)
	}
}
