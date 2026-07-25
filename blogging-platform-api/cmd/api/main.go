package main

import (
	"blogging-platform-api/internal/config"
	"blogging-platform-api/internal/database"
	"blogging-platform-api/internal/handlers"
	"blogging-platform-api/internal/store"
	"fmt"
	"log"
	"net/http"
)

func main() {
	cfg := config.Load()

	db, err := database.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print("Connected!")

	store := store.New(db)
	handler := handlers.New(store)

	mux := http.NewServeMux()

	handlers.RegisterRoutes(mux, handler)

	server := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}
