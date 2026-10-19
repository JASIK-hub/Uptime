package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"uptime/internal/api"
	"uptime/internal/infra/postgres"
	"uptime/internal/monitor"
	"uptime/internal/scheduler"
)

func main() {
	store := monitor.NewMonitorStore()
	handler := api.NewMonitorHandler(store)
	schedulerDone := make(chan struct{})

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		log.Print("startup failed: DATABASE_URL is required")
		os.Exit(1)
	}
	connectCtx, connectCancel := context.WithTimeout(ctx, 5*time.Second)
	pool, err := postgres.Connect(connectCtx, databaseURL)
	connectCancel()

	if err != nil {
		log.Printf("startup failed: %v", err)
		os.Exit(1)
	}
	defer pool.Close()
	log.Print("database connected")

	go func() {
		defer close(schedulerDone)
		scheduler.RunScheduler(ctx, store)
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /monitor", handler.CreateMonitor)
	mux.HandleFunc("GET /monitors", handler.ListMonitors)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	serverErr := make(chan error, 1)

	go func() {
		log.Print("starting HTTP server on :8080")
		serverErr <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		log.Print("shutdown requested")
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server failed: %v", err)
		}
		cancel()
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP shutdown failed: %v, forcing close...", err)
		if err := server.Close(); err != nil {
			log.Printf("HTTP close failed: %v", err)
		}
	}
	<-schedulerDone
	log.Print("server stopped")
}
