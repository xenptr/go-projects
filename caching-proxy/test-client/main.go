package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

func main() {
	const clients = 10

	var wg sync.WaitGroup

	start := time.Now()

	for i := 1; i <= clients; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			clientStart := time.Now()

			resp, err := http.Get("http://localhost:3000/data")
			if err != nil {
				fmt.Printf("[CLIENT %d] ERROR: %v\n", id, err)
				return
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)

			fmt.Printf(
				"[CLIENT %d] status=%d took=%v response=%s",
				id,
				resp.StatusCode,
				time.Since(clientStart),
				body,
			)
		}(i)
	}

	wg.Wait()

	fmt.Printf("\nTOTAL TIME: %v\n", time.Since(start))
}