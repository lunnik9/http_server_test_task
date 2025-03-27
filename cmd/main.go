package main

import (
	"context"
	"fmt"
	"http_server/config"
	"http_server/internal/multiplexer"
	"http_server/internal/multiplexer/cache"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env"
)

func main() {
	// Config setup
	var cfg config.Config
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	fmt.Println(cfg.CacheTTL)

	// Global context setup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := multiplexer.Options{
		MaxUrls:       cfg.MaxURLs,
		RequestsLimit: cfg.RequestsLimit,
		WorkerLimit:   cfg.WorkerLimit,
		FetchTimeout:  cfg.FetchTimeout,

		Retry:      cfg.Retry,
		NumRetries: cfg.RetryNum,
		Delay:      cfg.RetryDelay,
		FillRatio:  cfg.RetryFillRatio,
	}

	// Cache setup
	if cfg.Cache {
		opts.Cache = cache.NewCache(ctx, cfg.CacheTTL)
	}

	m := multiplexer.NewMultiplexer(opts)

	mux := http.NewServeMux()

	mux.HandleFunc("/fetch", m.FetchHandler)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Running server in separate goroutine
	go func() {
		log.Println("Starting server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Setup signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	shutdownInitiated := false

ShutdownLoop:
	for {
		select {
		case <-ctx.Done():
			break ShutdownLoop

		case sig := <-sigCh:
			switch sig {
			case syscall.SIGINT:
				if !shutdownInitiated {
					log.Println("Received SIGINT. Performing soft shutdown...")
					shutdownInitiated = true

					go func() {

						err := srv.Shutdown(ctx)
						if err != nil {
							log.Printf("Server shutdown failed: %v", err)
						}

						cancel()
					}()
				} else {
					log.Println("Received second SIGINT. Performing hard shutdown...")
					cancel()
				}
			case syscall.SIGTERM:
				log.Println("Received SIGTERM. Triggering shutdown...")
				cancel()
			}
		}
	}
}
