package main

import (
	"context"
	"errors"
	"fmt"
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
	if err := run(); err != nil {
		slog.Error("application failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}

	g, gctx := errgroup.WithContext(ctx)

	connectCtx, connectCancel := context.WithTimeout(ctx, 5*time.Second)
	pool, err := postgres.Connect(connectCtx, databaseURL)
	connectCancel()

	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()
	slog.Info("database started")

	repository := repository.NewMonitorRepository(pool)
	handler := api.NewMonitorHandler(repository)

	scheduler := scheduler.NewSchedule(repository)

	g.Go(func() error {
		err := scheduler.Run(gctx)
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /monitor", handler.CreateMonitor)
	mux.HandleFunc("GET /monitors", handler.ListMonitors)

	server := http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	g.Go(func() error {
		slog.Info("starting HTTP server", "addr", ":8080")
		err := server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	})

	g.Go(func() error {
		<-gctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			closeErr := server.Close()
			return errors.Join(
				fmt.Errorf("shutdown HTTP server: %w", err),
				closeErr,
			)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("run application: %w", err)
	}
	slog.Info("server stoped")
	return nil
}
