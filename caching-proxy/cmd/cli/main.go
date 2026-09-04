package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xenptr/go-projects/caching-proxy/internal/cache"
	"github.com/xenptr/go-projects/caching-proxy/internal/config"
	"github.com/xenptr/go-projects/caching-proxy/internal/proxy"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n\n", err)
		config.PrintUsage()
		os.Exit(1)
	}

	redisClient, err := cache.New(&cfg)
	if err != nil {
		log.Fatalf("connecting to redis: %v", err)
	}
	defer redisClient.Close()

	p, err := proxy.New(cfg.Origin, redisClient)
	if err != nil {
		log.Fatalf("creating proxy: %v", err)
	}

	if cfg.Clear {
		fmt.Println("Clearing cache...")
		if err := p.ClearAll(context.Background()); err != nil {
			log.Fatalf("clearing cache: %v", err)
		}
		fmt.Println("Cache cleared.")
		return
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listening on %s: %v", addr, err)
	}

	server := &http.Server{Handler: p}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("caching-proxy listening on %s, forwarding to %s", addr, cfg.Origin)
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case <-quit:
	case err := <-serverErr:
		log.Printf("server error: %v", err)
		return
	}

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("forced shutdown: %v", err)
		return
	}

	log.Println("server stopped")
}
