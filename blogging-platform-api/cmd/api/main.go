package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/xenptr/go-projects/blogging-platform-api/internal/config"
	"github.com/xenptr/go-projects/blogging-platform-api/internal/database"
	"github.com/xenptr/go-projects/blogging-platform-api/internal/handlers"
	pgxstore "github.com/xenptr/go-projects/blogging-platform-api/internal/store/pgx"
)

func main() {
	cfg := config.Load()

	db, err := database.OpenPGX(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Print("Connected!")

	store := pgxstore.New(db)
	handler := handlers.New(store)

	mux := http.NewServeMux()

	handlers.RegisterRoutes(mux, handler)

	server := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}
