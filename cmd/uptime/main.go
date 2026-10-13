package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
	"uptime/internal/api"
	"uptime/internal/monitor"
	"uptime/internal/scheduler"
)

func main() {
	store := monitor.NewMonitorStore()
	handler := api.NewMonitorHandler(store)
	schedulerDone := make(chan struct{})

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

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
		fmt.Print("Server started")
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
