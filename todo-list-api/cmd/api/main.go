package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/xenptr/go-projects/todo-list-api/internal/config"
	"github.com/xenptr/go-projects/todo-list-api/internal/db"
	"github.com/xenptr/go-projects/todo-list-api/internal/handlers"
	"github.com/xenptr/go-projects/todo-list-api/internal/repository"
	"github.com/xenptr/go-projects/todo-list-api/internal/routes"
)

func main() {
	// Load .env file if present; ignore error in production where env vars are set directly.
	_ = godotenv.Load()

	cfg := config.Load()

	db, err := db.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repo := repository.New(db)
	h := handlers.New(repo, cfg.JWTSecret)

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, h, cfg.JWTSecret)

	log.Printf("server starting on :%s", cfg.AppPort)

	server := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: mux,
	}
	log.Fatal(server.ListenAndServe())
}
