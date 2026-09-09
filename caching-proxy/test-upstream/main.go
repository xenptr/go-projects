package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

var requestCount atomic.Int64

func main() {
	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		id := requestCount.Add(1)

		log.Printf("UPSTREAM HIT #%d", id)

		// Simulate a slow API/database
		time.Sleep(2 * time.Second)

		fmt.Fprintf(w, "response from upstream | request #%d\n", id)
	})

	log.Println("Upstream running on :9000")

	log.Fatal(http.ListenAndServe(":9000", nil))
}