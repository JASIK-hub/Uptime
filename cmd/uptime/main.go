package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"uptime/internal/api"
	"uptime/internal/infra/postgres"
	"uptime/internal/infra/postgres/repository"
	"uptime/internal/scheduler"

	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		slog.Info("database startup failed: ", "DATABASE_URL", "required")
		os.Exit(1)
	}

	g, gctx := errgroup.WithContext(ctx)

	connectCtx, connectCancel := context.WithTimeout(ctx, 5*time.Second)
	pool, err := postgres.Connect(connectCtx, databaseURL)
	connectCancel()

	if err != nil {
		slog.Info("database startup failed:", "err", err.Error())
		os.Exit(1)
	}
	defer pool.Close()
	slog.Info("database started")

	repository := repository.NewMonitorRepository(pool)
	handler := api.NewMonitorHandler(repository)

	scheduler := scheduler.NewSchedule(repository)

	g.Go(func() error {
		return scheduler.Run(ctx)
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /monitor", handler.CreateMonitor)
	mux.HandleFunc("GET /monitors", handler.ListMonitors)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	g.Go(func() error {
		slog.Info("starting HTTP server", "addr", ":8080")
		return server.ListenAndServe()
	})

	g.Go(func() error {
		<-gctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		return server.Shutdown(shutdownCtx)
	})

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("application stopped with error", "err", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}
