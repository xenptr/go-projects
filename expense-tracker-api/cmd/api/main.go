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
	"github.com/xenptr/go-projects/expense-tracker-api/internal/config"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/db"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/handlers"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/repository"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/routes"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	pool, err := db.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	repo := repository.New(pool)
	h := handlers.New(repo)

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, h)

	server := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: mux,

		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	serverErr := make(chan error, 1)

	go func() {
		log.Printf("server started on %v", cfg.AppPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

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

	if err = server.Shutdown(ctx); err != nil {
		log.Printf("forced shutdown: %v", err)
		return
	}

	log.Println("server stopped")
}
