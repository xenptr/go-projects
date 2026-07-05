package main

import (
	"log"
	"net/http"
)

func main() {
	if err := loadTemplates(); err != nil {
		log.Fatal(err)
	}

	registerRoutes()

	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
