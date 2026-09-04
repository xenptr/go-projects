package main

import (
	"fmt"
	"os"

	"github.com/xenptr/go-projects/caching-proxy/internal/config"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n\n", err)
		config.PrintUsage()
		os.Exit(1)
	}

	if cfg.Clear {
		fmt.Println("Clearing cache...")
		// TODO: cache clear logic
		return
	}

	fmt.Printf("Starting %s on port %d, forwarding to %s\n", "caching-proxy", cfg.Port, cfg.Origin)
	// TODO: start proxy server
}
